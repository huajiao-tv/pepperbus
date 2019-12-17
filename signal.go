package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// todo get value from gokeeper
	GraceStopTimeout = time.Second * 5
)

func signalHandler(llist ...net.Listener) {
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGUSR2)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			signal.Stop(ch)
			flog.Error("receive signal TERM/INT stop the process")
			for _, lis := range llist {
				lis.Close()
			}
			GlobalCancelFunc()
			<-time.After(GraceStopTimeout)
			flog.Error("signal handler done now end the main process")
			return
		case syscall.SIGUSR2:
			flog.Error("receive signal USR2 restart process")
			count, err := GlobalNet.StartProcess()
			if err != nil {
				flog.Error("restart process fail", err)
				return
			}
			for _, lis := range llist {
				lis.Close()
			}
			GlobalCancelFunc()

			flog.Error("signal handler done now end the main process", count, err)
		}
	}
}
