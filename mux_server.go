package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

type MuxServerIntf interface {
	AddJobs(jobs *Jobs) error
	GetJobs(qn QueueName, tn TopicName, typ QueueTyp) (*Jobs, error)
	AddQueue(conf *QueueConfig) error
	DelQueue(conf *QueueConfig) error
	UpdateQueue(conf *QueueConfig) error
	GetMuxer(qn QueueName, tn TopicName) (Muxer, error)
	UpdateConfigs(map[QueueName]*QueueConfig) error
	Serve()
}

// a group of mux
type MuxServer struct {
	cmdC chan *Cmd

	context.Context
	sync.RWMutex // protect following
	mux          map[QueueName]map[TopicName]Muxer
	queueConfig  map[QueueName]*QueueConfig
}

func NewMuxServer(ctx context.Context) *MuxServer {
	return &MuxServer{
		mux:         make(map[QueueName]map[TopicName]Muxer),
		queueConfig: make(map[QueueName]*QueueConfig),
		cmdC:        make(chan *Cmd, 1000),
		Context:     ctx,
	}
}

// one loop per MuxServer
func (ms *MuxServer) Serve() {
	for {
		select {
		case <-ms.Done():
			flog.Warn("mux server done: ", ms.Context.Err())
			return
		case cmd := <-ms.cmdC:
			qconf, op, queueName := cmd.conf, cmd.op, cmd.conf.QueueName
			switch op {
			case CmdQueueAdd:
				// create a new mux server
				ms.Lock()
				if _, ok := ms.queueConfig[queueName]; !ok {
					ms.mux[queueName] = make(map[TopicName]Muxer)
					ms.queueConfig[queueName] = qconf
					for topicName, topicConf := range qconf.Topic {
						storage := NewStorage(topicConf.Storage)
						switch topicConf.RunType {
						case RunTypeCGI:
							mux := NewMux(ms.Context, queueName, topicName, NewCgiManager(topicConf.CgiConfigKey, topicConf.ScriptEntry), storage, topicConf.NumOfWorkers)
							ms.mux[queueName][topicName] = mux
							consumerEnable := !netQueueConf().ConsumerDisable && (qconf.ConsumerEnable || topicConf.ConsumerEnable)
							if consumerEnable {
								go mux.Serve()
							}
							flog.Warn("add queue ["+string(queueName)+"], topic ["+string(topicName)+"],", "cgiConf,", topicConf.CgiConfigKey, "consumerEnable", consumerEnable)

						case RunTypeHttp:
							sender := &HttpClientsManager{
								path: topicConf.ScriptEntry,
								Topic : topicConf,
								client: &http.Client{
									Transport: &http.Transport{
										MaxIdleConns:       int(topicConf.NumOfWorkers),
										IdleConnTimeout:    HttpConfigManager.IdleConnTimeout,
										DisableCompression: true,
									},
									Timeout: HttpConfigManager.RequestTimeout,
								},
							}
							mux := NewHttpWorker(ms.Context, queueName, topicName, int64(topicConf.NumOfWorkers), storage, sender)
							ms.mux[queueName][topicName] = mux
							consumerEnable := !netQueueConf().ConsumerDisable && (qconf.ConsumerEnable || topicConf.ConsumerEnable)
							if consumerEnable {
								go mux.Serve()
							}
							flog.Warn("add queue ["+string(queueName)+"], topic ["+string(topicName)+"],", "runType", topicConf.RunType, "consumerEnable", consumerEnable)
						case RunTypeCmd:
							mux := NewWorker(ms.Context, queueName, storage, topicConf)
							ms.mux[queueName][topicName] = mux
							consumerEnable := qconf.ConsumerEnable || topicConf.ConsumerEnable
							if consumerEnable {
								go mux.Serve()
							}
							flog.Warn("add queue ["+string(queueName)+"], topic ["+string(topicName)+"],", "workerConf", topicConf.WorkerConfig, "runType", topicConf.RunType, "consumerEnable", consumerEnable)
						}
					}
				}
				ms.Unlock()
			case CmdQueueDel:
				ms.Lock()
				delete(ms.queueConfig, queueName)
				for _, muxer := range ms.mux[queueName] {
					muxer.Close()
				}
				delete(ms.mux, queueName)
				ms.Unlock()
			}
		}
	}
}

// 添加队列
func (ms *MuxServer) AddQueue(conf *QueueConfig) error {
	ms.cmdC <- &Cmd{conf: conf, op: CmdQueueAdd}
	return nil
}

// 删除队列
func (ms *MuxServer) DelQueue(conf *QueueConfig) error {
	ms.cmdC <- &Cmd{conf: conf, op: CmdQueueDel}
	return nil
}

// 更新队列，先删除后添加
func (ms *MuxServer) UpdateQueue(conf *QueueConfig) error {
	ms.cmdC <- &Cmd{conf: conf, op: CmdQueueDel}
	ms.cmdC <- &Cmd{conf: conf, op: CmdQueueAdd}
	return nil
}

// return the status of queue info
func (ms *MuxServer) QueueStatus() {
	ms.RLock()
	defer ms.RUnlock()
	for q, cmap := range ms.mux {
		fmt.Println(q, len(ms.mux))
		for c, _ := range cmap {
			fmt.Println(c, len(cmap))
		}
	}
}

// add job to mux Storage
// one topic of queue refer to one mux server
//todo  添加job的时候初始化重试数据
func (ms *MuxServer) AddJobs(jobs *Jobs) error {
	var err error
	ms.RLock()
	defer ms.RUnlock()
	if _, ok := ms.mux[jobs.queueName]; !ok || len(ms.mux[jobs.queueName]) == 0 {
		return errors.New("queue or Topic not exist " + string(jobs.queueName))
	}

	for tName, mux := range ms.mux[jobs.queueName] {

		jobs.topicName = tName
		for _,job := range jobs.Slice {
			if(ms.queueConfig[jobs.queueName].Topic[jobs.topicName].IsRetry == true) {
				job.RetryResidueTime = ms.queueConfig[jobs.queueName].Topic[jobs.topicName].RetryTime
				job.RetryTimes = 0
				//job.Retry = ms.queueConfig[jobs.queueName].Topic[jobs.topicName].RetryTimes
				job.IsRetry = true
				job.BeginRetry = false
			}else{
				job.RetryResidueTime = 0
				job.RetryTimes = 0
				//job.Retry = 0
				job.IsRetry = false
				job.BeginRetry = false
			}
		}
		e := mux.AddJobs(jobs, CommonQueue)
		if e != nil {
			err = e
			flog.Error("mux add jobs fail: queue:"+string(jobs.queueName)+" Topic "+string(jobs.topicName)+" error "+err.Error(), "jobs", jobs)
		}
		flog.Trace("AddJobs success: ", jobs)
	}
	return err
}

// get Jobs from mux storage
func (ms *MuxServer) GetJobs(qn QueueName, tn TopicName, typ QueueTyp) (*Jobs, error) {
	ms.RLock()
	defer ms.RUnlock()
	if _, ok := ms.mux[qn]; !ok || len(ms.mux[qn]) == 0 {
		return nil, errors.New("queue not exist " + string(qn))
	}
	mux, ok := ms.mux[qn][tn]
	if !ok {
		return nil, errors.New("topic not exist " + string(tn))
	}
	return mux.GetJobsFromQueue(qn, tn, typ)
}

func (ms *MuxServer) GetMuxer(qn QueueName, tn TopicName) (Muxer, error) {
	ms.RLock()
	defer ms.RUnlock()
	if _, ok := ms.mux[qn]; !ok || len(ms.mux[qn]) == 0 {
		return nil, errors.New("queue not exist " + string(qn))
	}
	mux, ok := ms.mux[qn][tn]
	if !ok {
		return nil, errors.New("topic not exist " + string(tn))
	}
	return mux, nil
}

func (ms *MuxServer) UpdateConfigs(queueConfs map[QueueName]*QueueConfig) error {
	ms.RLock()
	defer ms.RUnlock()

	for name, conf := range queueConfs {
		if old, ok := ms.queueConfig[name]; !ok {
			ms.AddQueue(conf)
		} else if !reflect.DeepEqual(conf, old) {
			ms.UpdateQueue(conf)
		}
	}
	for name, conf := range ms.queueConfig {
		if _, ok := queueConfs[name]; !ok {
			ms.DelQueue(conf)
		}
	}
	return nil
}
