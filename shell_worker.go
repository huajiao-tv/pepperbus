package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Worker struct {
	queue QueueName
	conf  *TopicConfig

	storager
	context.Context
	cancel context.CancelFunc

	mu            sync.Mutex
	numOfChildren int
}

func NewWorker(parentCtx context.Context, qn QueueName, storage storager, config *TopicConfig) *Worker {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Worker{
		queue:         qn,
		conf:          config,
		storager:      storage,
		Context:       ctx,
		cancel:        cancel,
		mu:            sync.Mutex{},
		numOfChildren: 0,
	}
}

func (w *Worker) AddJobs(jobs *Jobs, typ QueueTyp) error {
	return w.storager.AddQueue(jobs, typ)
}

func (w *Worker) GetJobsFromQueue(q QueueName, t TopicName, typ QueueTyp) (*Jobs, error) {
	return w.storager.GetJobsFromQueue(q, t, typ)
}

func (w *Worker) Storage() storager {
	return w.storager
}

func (w *Worker) Serve() {

	time.Sleep(time.Second * 5)

	args := []string{
		w.conf.Script,
		"-q", fmt.Sprintf("%s/%s", w.queue, w.conf.Name),
		"-h", ServerConf().GatewayIP,
		"-p", ServerConf().GatewayPort,
		"-a", w.conf.Password,
		"-t", strconv.Itoa(w.conf.MaxTask),
		"-H", w.conf.Entry,
	}

	for i := 0; i < w.conf.MaxChildren; i++ {
		go func() {
			// todo get last pid and stop id
			lastPid := -1
			StopProcess(lastPid)

			for {
				select {
				case <-w.Context.Done():
					return
				default:
				}

				var buf bytes.Buffer
				proc := NewProcess(w.Context, &buf, "/usr/local/bin/php", args...)

				if err := proc.Start(); err != nil {
					flog.Error("start worker err: " + err.Error())
					time.Sleep(2 * MuxLoopInterval)
					continue
				}

				if err := proc.Wait(); err != nil {
					flog.Error("exec worker err", err.Error(), buf.String())
				} else {
					flog.Error("exec worker done", buf.String())
				}
			}
		}()
	}
}

func (w *Worker) SendPing() error {
	return nil
}

func (w *Worker) SendJobs(jobs *Jobs) *WorkerResponse {
	return nil
}

func (w *Worker) Stats() *MuxStats {
	return nil
}

func (w *Worker) Close() {
	w.cancel()
}
