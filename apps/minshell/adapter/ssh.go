package adapter

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func InteractiveWithTerminalForSSH(username, password, privateKeyPath string, host string, port int, timeout time.Duration) error {
	authMethods := make([]ssh.AuthMethod, 0, 4)
	if privateKeyPath != "" {
		pk, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return err
		}
		signer, err := ssh.ParsePrivateKey([]byte(pk))
		if err != nil {
			return err
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if password != "" {
		authMethods = append(authMethods, ssh.KeyboardInteractive(setKeyboard(password)))
		authMethods = append(authMethods, ssh.Password(password))
	}

	config := ssh.ClientConfig{
		User: username,
		Auth: authMethods,
		Config: ssh.Config{
			Ciphers: []string{
				"aes128-ctr",
				"aes192-ctr",
				"aes256-ctr",
				"aes128-gcm@openssh.com",
				"arcfour256",
				"arcfour128",
				"aes128-cbc",
			},
			KeyExchanges: []string{
				"diffie-hellman-group-exchange-sha1",
				"diffie-hellman-group1-sha1",
				"diffie-hellman-group-exchange-sha256",
				"diffie-hellman-group16-sha512",
				"diffie-hellman-group18-sha512",
				"diffie-hellman-group14-sha256",
				"diffie-hellman-group14-sha1",
			},
		},
		Timeout: timeout,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	connection, err := ssh.Dial("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)), &config)
	if err != nil {
		return err
	}
	defer connection.Close()

	session, err := connection.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer term.Restore(fd, state)

	w, h, err := term.GetSize(fd)
	if err != nil {
		return err
	}

	modes := ssh.TerminalModes{
		ssh.ECHOCTL:       1,
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}

	err = session.RequestPty(termType, h, w, modes)
	if err != nil {
		return err
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}

	if err = session.Shell(); err != nil {
		return err
	}

	go io.Copy(os.Stderr, stderr)
	go io.Copy(os.Stdout, stdout)
	go io.Copy(stdin, os.Stdin)

	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, syscall.SIGWINCH, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		for {
			s := <-signal_chan
			switch s {
			case syscall.SIGWINCH:
				fd := int(os.Stdout.Fd())
				w, h, _ = term.GetSize(fd)
				session.WindowChange(h, w)
			default:
				session.Signal(ssh.SIGTERM)
				return
			}
		}
	}()

	err = session.Wait()
	return nil
}

func setKeyboard(password string) func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
	return func(_, _ string, questions []string, _ []bool) (answers []string, err error) {
		answers = make([]string, len(questions))
		for n := range questions {
			answers[n] = password
		}
		return answers, nil
	}
}
