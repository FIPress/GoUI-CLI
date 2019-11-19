package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

const defaultTimeout = 20 * time.Second

type Command struct {
	Cmd  string
	Args []string
	Env  []string
	Dir  string
}

func NewCommand(cmd string, args ...string) *Command {
	return &Command{Cmd: cmd, Args: args}
}

func (c *Command) Run() (err error) {
	return c.RunEx(os.Stdout, os.Stderr, 0)
}

func (c *Command) RunEx(stdout io.Writer, stderr io.Writer, timeout time.Duration) (err error) {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	//c.Args = append(c.Args,c.Env...)

	cmd := exec.Command(c.Cmd, c.Args...)

	debug("exec:", c.Cmd, c.Args)

	if stdout == nil {
		stdout = new(bytes.Buffer)
	}

	if c.Env != nil && len(c.Env) != 0 {
		cmd.Env = append(os.Environ(), c.Env...)
	}

	cmd.Dir = c.Dir
	cmd.Stdout = stdout
	if stderr == nil {
		cmd.Stderr = stdout
	} else {
		cmd.Stderr = stderr
	}
	if err = cmd.Start(); err != nil {
		return
	}

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		if cmd.Process != nil && cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
			if err = cmd.Process.Kill(); err != nil {
				err = fmt.Errorf("fail to kill process: %v", err)
				return
			}
		}

		<-done
		debug("after timeout")
		err = fmt.Errorf("execute command timeout,[duration: %v]", timeout)
		return
	case err = <-done:
		debug("after done")
	}

	debug("after wait")
	//stdout = stdoutBuf.Bytes()
	//stdoutBuf.ReadBytes()
	//stderr = stderrBuf.Bytes()

	return
}
