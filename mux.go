package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/huajiao-tv/pepperbus/define"
	"github.com/huajiao-tv/pepperbus/utility"
)

type QueueTyp int

func (qt QueueTyp) IsRetryQueue() bool {
	return qt == RetryQueue
}

func (qt QueueTyp) IsTimeoutQueue() bool {
	return qt == TimeoutQueue
}

const (
	CommonQueue  = 0
	RetryQueue   = 1
	TimeoutQueue = 2

	PingSleepDuration    = 3
	PingSleepDurationMax = 60

	CmdQueueAdd = "queue_add"
	CmdQueueDel = "queue_del"

	MuxLoopInterval = time.Millisecond * 1000
)

type (
	QueueName string
	TopicName string

	Cmd struct {
		conf   *QueueConfig
		op     string
		result chan error
	}

	Mux struct {
		queue QueueName
		topic TopicName

		storager
		RemoteSender
		context.Context
		cancel context.CancelFunc

		mu sync.Mutex

		numOfWorkers uint64
		numOfCgi     int64
	}
)

type RemoteSender interface {
	SendPing() (*WorkerResponse, error)
	SendJobs(jobs *Jobs) (*WorkerResponse, error)
}

type Muxer interface {
	AddJobs(jobs *Jobs, typ QueueTyp) error
	GetJobsFromQueue(q QueueName, t TopicName, typ QueueTyp) (*Jobs, error)
	Storage() storager
	Serve()
	SendPing() error
	SendJobs(jobs *Jobs) *WorkerResponse
	Close()
}

func NewMux(parentCtx context.Context, qn QueueName, cn TopicName, cgi RemoteSender, storage storager, numOfWorkers uint64) *Mux {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Mux{
		queue:        qn,
		topic:        cn,
		storager:     storage,
		RemoteSender: cgi,
		Context:      ctx,
		cancel:       cancel,
		mu:           sync.Mutex{},
		numOfWorkers: numOfWorkers,
	}
}

func (m *Mux) Close() {
	m.cancel()
}

func (m *Mux) Serve() {
	conf := QueueConfs()
	//todo 添加重试机制
	for i := 0; i < int(m.numOfWorkers); i++ {
		go func() {
		Begin:
			m.loopPingCGI()
			for {
				select {
				case <-m.Done():
					flog.Warn("mux done: ", m.queue, m.topic, m.Context.Err())
					return
				default:
					jobs, err := m.GetJobsFromQueue(m.queue, m.topic, CommonQueue)
					if err != nil {
						time.Sleep(MuxLoopInterval * 2)
						continue
					} else if jobs == nil {
						time.Sleep(MuxLoopInterval)
						continue
					}
					flog.Trace("mux GetJobsFromQueue success: ", jobs)

					atomic.AddInt64(&m.numOfCgi, 1)

					start := time.Now()
					for _, val := range jobs.Slice {
						val.RetryTime = time.Now().Unix()
						val.RetryTimes = val.RetryTimes + 1
						val.IsRetry = conf[m.queue].Topic[m.topic].IsRetry

						if !val.BeginRetry {
							val.RetryResidueTime = conf[m.queue].Topic[m.topic].RetryTime
						}
					}
					// 发送任务给 CGI
					resp := m.SendJobs(jobs)
					//resp := new (WorkerResponse)
					//resp.Code = ResponseNotFound
					// 统计执行时间
					done := time.Now().Sub(start)
					// 添加统计
					MuxStat.Add(m.queue, m.topic, resp.Code == ResponseSuccess, done)

					atomic.AddInt64(&m.numOfCgi, -1)

					switch resp.Code {
					case ResponseSuccess:
						flog.Trace("send to cgi success jobs: ", jobs)
					case ResponseNotFound:
						flog.Error("mux serve sendToCGI 404", "topic no register", jobs)
						m.sendToRetryQueue(jobs)
						// 没订阅的话，回到 loopPingCGI()
						goto Begin
					case BadGateway:
						flog.Error("mux serve sendToCGI bad gateway: ", resp.Msg, " jobs: ", jobs)
						m.sendToRetryQueue(jobs)

						// 网络错误重新ping
						if strings.Contains(resp.Msg, "connection refused") {
							goto Begin
						}
					default:
						flog.Error("mux serve sendToCGI fail: ", resp.Msg, " jobs: ", jobs)
						m.sendToRetryQueue(jobs)
					}

					// 记录错误日志
					if resp.Code != ResponseSuccess {
						CGIErrorCollection.AddError(m.queue, m.topic, jobs, resp.Code, resp.ErrorCode, resp.Msg, err)
					}
				}
			}
		}()
	}

	if len(conf) == 0 || conf[m.queue].Topic[m.topic] == nil {
		return
	}
	// 不需要重试时，不需要进入循环判断
	if !conf[m.queue].Topic[m.topic].IsRetry {
		select {
		case <-m.Done():
			flog.Warn("mux done: ", m.queue, m.topic, m.Context.Err())
			return
		}
	}

	// 读取重试队列，并判断是否能重试
	// 如果可以，则再次加入队列
	for {
		select {
		case <-m.Done():
			flog.Warn("mux retry done: ", m.queue, m.topic, m.Context.Err())
			return
		default:
			jobs, err := m.GetJobsFromQueue(m.queue, m.topic, RetryQueue)
			if err != nil {
				time.Sleep(MuxLoopInterval * 2)
				continue
			} else if jobs == nil {
				time.Sleep(MuxLoopInterval)
				continue
			}

			jobsSlice := jobs.Slice
			if len(jobsSlice) == 0 {
				continue
			}

			flog.Trace("mux GetJobsFromRetryQueue success: ", jobs)

			for _, job := range jobsSlice {
				job.BeginRetry = true
			}

			nowTime := time.Now().Unix()

			// 检查重试
			if jobsSlice[0].IsRetry && jobsSlice[0].RetryResidueTime > 0 {
				used := jobsSlice[0].RetryTimes
				durTime := int64((used - 1) * define.RetryDur)
				if jobsSlice[0].RetryTime == 0 {
					jobsSlice[0].RetryTime = nowTime
				}
				nextTime := jobsSlice[0].RetryTime + durTime
				if nowTime >= nextTime {
					for _, val := range jobsSlice {
						val.RetryResidueTime = val.RetryResidueTime - durTime
						m.AddJobs(jobs, CommonQueue)
					}
				} else {
					m.sendToRetryQueue(jobs)
					time.Sleep(10 * time.Millisecond)
				}
			} else {
				m.sendToTimeoutQueue(jobs)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func (m *Mux) loopPingCGI() {
	pingSleepDuration := PingSleepDuration
	for {
		select {
		case <-m.Done():
			flog.Warn("ping done: ", m.queue, m.topic, m.Context.Err())
			return
		default:
			err := m.SendPing()
			if err == nil {
				return
			}
			flog.Error("mux ping failed: ", m.queue, m.topic, err.Error())
			time.Sleep(time.Duration(pingSleepDuration) * time.Second)
			pingSleepDuration += pingSleepDuration
			if pingSleepDuration >= PingSleepDurationMax {
				pingSleepDuration = PingSleepDurationMax
			}
		}
	}
}

func (m *Mux) sendToRetryQueue(jobs *Jobs) error {
	err := m.storager.AddQueue(jobs, RetryQueue)
	if err != nil {
		flog.Error("mux add jobs to retrying queue fail errors: " + err.Error() + " traceId: " + jobs.outTraceId)
		return err
	}
	flog.Error("mux save jobs to retry queue success outTraceId: ", jobs.outTraceId)
	return nil
}

func (m *Mux) sendToTimeoutQueue(jobs *Jobs) error {
	err := m.storager.AddQueue(jobs, TimeoutQueue)
	if err != nil {
		flog.Error("mux add jobs to timeout queue fail errors: " + err.Error() + " traceId: " + jobs.outTraceId)
		return err
	}
	flog.Error("mux save jobs to timeout queue success outTraceId: ", jobs.outTraceId)
	return nil
}

// add job to mux Storage
func (m *Mux) AddJobs(jobs *Jobs, typ QueueTyp) error {
	return m.storager.AddQueue(jobs, typ)
}

func (m *Mux) Storage() storager {
	return m.storager
}

// get jobs from mux storage
func (m *Mux) GetJobsFromQueue(q QueueName, t TopicName, typ QueueTyp) (*Jobs, error) {
	return m.storager.GetJobsFromQueue(q, t, typ)
}

func (m *Mux) SendPing() error {
	resp, err := m.RemoteSender.SendPing()
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return errors.New(resp.String())
	}

	// 这里同时检查了 Body 和 Header 防止 PHP 出现 Notice 后污染 Body
	queueTopic := strings.Split(resp.ConsumeTopics, ",")
	queueTopic = append(queueTopic, strings.Split(resp.Msg, ",")...)
	if utility.InArray(fmt.Sprintf("%v/%v", m.queue, m.topic), queueTopic) {
		return nil
	}

	errInfo := fmt.Sprintf("topic no register: %s/%s", m.queue, m.topic)
	return errors.New(errInfo)
}

// @todo why use a code of cgitimeout?
func (m *Mux) SendJobs(jobs *Jobs) *WorkerResponse {
	resp, err := m.RemoteSender.SendJobs(jobs)
	if err != nil {
		resp = &WorkerResponse{}
		resp.Code = BadGateway
		resp.Msg = err.Error()
	}
	return resp
}
