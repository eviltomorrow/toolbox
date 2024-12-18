package ssh

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/eviltomorrow/toolbox/apps/ssh-server/conf"
	"github.com/eviltomorrow/toolbox/apps/ssh-server/domain/user"
	"github.com/eviltomorrow/toolbox/lib/zlog"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	config      *ssh.ServerConfig
	done        chan struct{}
	listen      net.Listener
	wg          sync.WaitGroup
	inFlightSem chan struct{}

	Port int
}

func NewServer(server *conf.Server) (*Server, error) {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			remoteAddr := c.RemoteAddr().String()
			for _, blockIP := range server.BlackList {
				if blockIP == "" {
					continue
				}
				if strings.Contains(remoteAddr, blockIP) {
					return nil, fmt.Errorf("address[%s] has blocked", remoteAddr)
				}
			}
			username := c.User()
			passowrd := string(pass)
			zlog.Debug("auth info", zap.String("username", username), zap.String("password", passowrd))
			if ok := user.Auth(username, passowrd); ok {
				return nil, nil
			}
			return nil, fmt.Errorf("login[user=%s] failure", username)
		},
	}

	privateKey, err := os.ReadFile(server.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("read private-key failure, nest error: %v", err)
	}

	key, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("parse private-key failure, nest error: %v", err)
	}

	config.AddHostKey(key)

	inFlightSem := make(chan struct{}, server.MaximumLoginLimit)
	for i := 0; i < server.MaximumLoginLimit; i++ {
		inFlightSem <- struct{}{}
	}
	s := &Server{config: config, done: make(chan struct{}, 1), inFlightSem: inFlightSem, Port: server.Port}
	return s, nil
}

func (s *Server) Serve() error {
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.Port))
	if err != nil {
		return err
	}
	s.listen = listen

	go func() {
	loop:
		for {
			conn, err := listen.Accept()
			if err != nil {
				select {
				case <-s.done:
					break loop
				default:
					zlog.Error("accept conn failure", zap.Error(err))
					continue
				}
			}

			s.wg.Add(1)
			go func() {
				defer func() {
					conn.Close()
				}()

				s.handleConn(conn, s.config)
				s.wg.Done()
			}()
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	s.done <- struct{}{}
	s.listen.Close()
	s.wg.Wait()
	return nil
}

func (s *Server) accept() bool {
	select {
	case <-s.inFlightSem:
		return true
	default:
		return false
	}
}

func (s *Server) release() {
	select {
	case s.inFlightSem <- struct{}{}:
	default:
	}
}

func (s *Server) handleConn(conn net.Conn, config *ssh.ServerConfig) {
	servconn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		zlog.Error("New server conn failure", zap.Error(err))
		return
	}

	zlog.Info("New a connection", zap.String("client_version", string(servconn.ClientVersion())), zap.String("remote_addr", servconn.RemoteAddr().String()))

	go ssh.DiscardRequests(reqs)
	handleChannels(chans)
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go func() {
			if err := handleChannel(newChannel); err != nil {
				zlog.Error("Handle channel failure", zap.Error(err))
			}
		}()
	}
}

func handleChannel(newChannel ssh.NewChannel) error {
	// Since we're handling a shell, we expect a
	// channel type of "session". The also describes
	// "x11", "direct-tcpip" and "forwarded-tcpip"
	// channel types.
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return fmt.Errorf("unknown channel type: %v", t)
	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	connection, requests, err := newChannel.Accept()
	if err != nil {
		return fmt.Errorf("accept new channel failure, nest error: %v", err)
	}

	dir, err := os.UserHomeDir()
	bash := &exec.Cmd{
		Path: "/bin/bash",
		Dir:  dir,
	}
	shouldClose := func() {
		connection.Close()
		if _, err := bash.Process.Wait(); err != nil {
			zlog.Error("bash process wait failure", zap.Error(err))
		}
	}

	handler, err := pty.Start(bash)
	if err != nil {
		shouldClose()
		return fmt.Errorf("start pty failure, nest error: %v", err)
	}

	// pipe session to bash and visa-versa
	var once sync.Once
	go func() {
		io.Copy(connection, handler)
		once.Do(shouldClose)
	}()
	go func() {
		io.Copy(handler, connection)
		once.Do(shouldClose)
	}()

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				// We only accept the default shell
				// (i.e. no command in the Payload)
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}

			case "pty-req":
				termLen := req.Payload[3]
				w, h := parseDims(req.Payload[termLen+4:])
				SetWinsize(handler.Fd(), w, h)
				// Responding true (OK) here will let the client
				// know we have a pty ready for input
				req.Reply(true, nil)

			case "window-change":
				w, h := parseDims(req.Payload)
				SetWinsize(handler.Fd(), w, h)

			default:
			}
		}
	}()
	return nil
}

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

// ======================

// Winsize stores the Height and Width of a terminal.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

// SetWinsize sets the size of the given pty.
func SetWinsize(fd uintptr, w, h uint32) {
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}

// Borrowed from https://github.com/creack/termios/blob/master/win/win.go
