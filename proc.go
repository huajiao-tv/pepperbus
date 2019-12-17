package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

type process struct {
	*exec.Cmd
	ctx context.Context
	buf *bytes.Buffer
}

func NewProcess(ctx context.Context, buf *bytes.Buffer, name string, args ...string) *process {
	p := &process{
		Cmd: exec.Command(name, args...),
		ctx: ctx,
		buf: buf,
	}
	p.Stdout = p.buf
	p.Stderr = p.buf
	return p
}

func (p *process) Pid() int {
	if p.Process != nil {
		return p.Process.Pid
	} else {
		return -1
	}
}

func (p *process) Wait() error {
	// monitor process
	go p.monitor()

	// wait for exit
	if err := p.Cmd.Wait(); err != nil {
		fmt.Println("exec worker err", p.Pid(), err.Error(), p.buf.String())
		return err
	}

	//
	fmt.Println("worker done:", p.Pid(), p.buf.String())
	return nil
}

func (p *process) monitor() error {
	for {
		select {
		case <-p.ctx.Done():
			fmt.Println("worker cancel from monitor", p.Pid())
			return StopProcess(p.Pid())
		default:
			if p.ProcessState != nil && p.ProcessState.Exited() {
				fmt.Println("worker done from monitor", p.Pid())
				return nil
			} else {
				time.Sleep(time.Second)
			}
		}
	}
}

func StopProcess(pid int) error {
	if pid < 0 {
		return nil
	}
	// send signal to process
	if syscall.Kill(pid, syscall.SIGINT) == syscall.ESRCH {
		return nil
	}
	// wait 10 seconds for exiting
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		if syscall.Kill(pid, syscall.Signal(0)) == syscall.ESRCH {
			return nil
		}
	}
	// force kill
	if syscall.Kill(pid, syscall.SIGKILL) == syscall.ESRCH {
		return nil
	}
	time.Sleep(time.Second)
	// try again
	if err := syscall.Kill(pid, syscall.Signal(0)); err == syscall.ESRCH {
		return nil
	} else {
		return err
	}
}
