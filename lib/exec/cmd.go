package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var (
	ErrTimeout = errors.New("execute timeout")
	EnvShell   = "/bin/bash"
)

func RunCmd(cmd string, _ []string, timeout time.Duration) ([]byte, []byte, error) {
	return runC(cmd, timeout)
}

func runC(c string, timeout time.Duration) ([]byte, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var (
		eg  = make(chan error)
		cmd = exec.Command(EnvShell, "-c", c)
	)
	defer func() {
		close(eg)
	}()

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start execute cmd failure, nest error: %v", err)
	}

	go func() {
		if err := cmd.Wait(); err != nil {
		}
		eg <- nil
	}()

	select {
	case <-ctx.Done():
		cmd.Process.Signal(syscall.SIGINT)
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-eg
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("%w, nest timeout: %v", ErrTimeout, timeout)

	case err := <-eg:
		return stdout.Bytes(), stderr.Bytes(), err
	}
}
