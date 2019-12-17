package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
	"encoding/json"
	"fmt"
	"strings"

	"git.huajiao.com/qmessenger/pepperbus/define"
	"github.com/CodisLabs/codis/pkg/utils/errors"
	"git.huajiao.com/qmessenger/pepperbus/utility"
)

// 包含调用连接池、发送http请求、解析返回结果
type HttpClientsManager struct {
	client *http.Client
	Topic  *TopicConfig
	path   string
}

func (hcm *HttpClientsManager) SendPing() (*WorkerResponse, error) {
	params := url.Values{
		"type": []string{RequestTypePing},
	}
	return hcm.request(hcm.path, params)
}

func (hcm *HttpClientsManager) SendJobs(jobs *Jobs) (*WorkerResponse, error) {
	jbs, _ := json.Marshal(jobs.Slice)
	params := url.Values{
		"type":       []string{RequestTypeData},
		"outTraceId": []string{jobs.outTraceId},
		"jobs":       []string{string(jbs)},
		"queue":      []string{string(jobs.queueName)},
		"topic":      []string{string(jobs.topicName)},
	}
	return hcm.request(hcm.path, params)
}

// @ todo log nothing,caller log the error
func (hcm *HttpClientsManager) request(path string, values url.Values) (*WorkerResponse, error) {

	values.Add("IS_PEPPER_BUS_REQUEST", "true")
	request, err := http.NewRequest("POST", path, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	var hc map[string]string
	if hcm.Topic.HttpConfig != "" {
		err = json.Unmarshal([]byte(hcm.Topic.HttpConfig), &hc)
		if err != nil {
			return nil, err
		}
		request.Host = hc["host"]
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("IS-PEPPER-BUS-REQUEST", "true")

	resp, err := hcm.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	code := resp.Header.Get("Code")
	errCode := resp.Header.Get("Error-Code")
	consumeTopics := resp.Header.Get("Consume-Topics")
	if code == "" {
		code = InnerError
	}
	cr := &WorkerResponse{code, errCode, consumeTopics, ""}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		flog.Error("read all error", err.Error())
		cr.Msg = err.Error()
		return cr, err
	}
	cr.Msg = string(body)
	return cr, nil
}

func (hw *HttpWorker) Stats() *MuxStats {
	return nil
}

type HttpWorker struct {
	queue QueueName
	topic TopicName
	storager
	context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	numOfChildren int64
	client        http.Client
	RemoteSender
}

func NewHttpWorker(parentCtx context.Context, qn QueueName, tn TopicName, numOfChildren int64, storage storager, sender RemoteSender) *HttpWorker {
	ctx, cancel := context.WithCancel(parentCtx)
	return &HttpWorker{
		queue:         qn,
		topic:         tn,
		storager:      storage,
		Context:       ctx,
		cancel:        cancel,
		mu:            sync.Mutex{},
		numOfChildren: numOfChildren,
		RemoteSender:  sender,
	}
}

func (hw *HttpWorker) AddJobs(jobs *Jobs, typ QueueTyp) error {
	return hw.storager.AddQueue(jobs, typ)
}

func (hw *HttpWorker) GetJobsFromQueue(q QueueName, t TopicName, typ QueueTyp) (*Jobs, error) {
	return hw.storager.GetJobsFromQueue(q, t, typ)
}

func (hw *HttpWorker) Storage() storager {
	return hw.storager
}

func (hw *HttpWorker) Serve() {
	conf := QueueConfs()
	for i := 0; i < int(hw.numOfChildren); i++ {
		go func() {
		Begin:
			hw.loopPing()
			for {

				select {
				case <-hw.Done():
					flog.Warn("mux done: ", hw.queue, hw.topic, hw.Context.Err())
					return
				default:
					jobs, err := hw.GetJobsFromQueue(hw.queue, hw.topic, CommonQueue)
					if err != nil {
						time.Sleep(MuxLoopInterval * 2)
						continue
					} else if jobs == nil {
						time.Sleep(MuxLoopInterval)
						continue
					}
					flog.Trace("mux GetJobsFromQueue success: ", jobs)

					atomic.AddInt64(&hw.numOfChildren, 1)

					start := time.Now()
					for _, val := range jobs.Slice {
						val.RetryTime = start.Unix()
						val.RetryTimes = val.RetryTimes + 1
						val.IsRetry = conf[hw.queue].Topic[hw.topic].IsRetry
						if !val.BeginRetry {
							val.RetryResidueTime = conf[hw.queue].Topic[hw.topic].RetryTime
						}
					}
					// 发送任务给 http worker
					resp := hw.SendJobs(jobs)

					// 统计执行时间
					done := time.Now().Sub(start)
					// 添加统计
					MuxStat.Add(hw.queue, hw.topic, resp.Code == ResponseSuccess, done)

					atomic.AddInt64(&hw.numOfChildren, -1)

					switch resp.Code {
					case ResponseSuccess:
						flog.Trace("send to cgi success jobs: ", jobs)
					case ResponseNotFound:
						flog.Error("mux serve sendToCGI 404", "topic no register", jobs)
						hw.sendToRetryQueue(jobs)
						// 没订阅的话，回到 loopPingCGI()
						goto Begin
					case BadGateway:
						flog.Error("mux serve sendToCGI bad gateway: ", resp.Msg, " jobs: ", jobs)
						hw.sendToRetryQueue(jobs)

						// 网络错误重新ping
						if strings.Contains(resp.Msg, "connection refused") {
							goto Begin
						}
					default:
						flog.Error("mux serve sendToCGI fail: ", resp.Msg, " jobs: ", jobs)
						hw.sendToRetryQueue(jobs)
					}

					// 记录错误日志
					if resp.Code != ResponseSuccess {
						CGIErrorCollection.AddError(hw.queue, hw.topic, jobs, resp.Code, resp.ErrorCode, resp.Msg, err)
					}
				}
			}
		}()
	}

	// 不需要重试时，不需要进入循环判断
	if !conf[hw.queue].Topic[hw.topic].IsRetry {
		select {
		case <-hw.Done():
			flog.Warn("hw retry done: ", hw.queue, hw.topic, hw.Context.Err())
			return
		}
	}

	// 读取重试队列，并判断是否能重试
	// 如果可以，则再次加入队列
	for {
		select {
		case <-hw.Done():
			flog.Warn("hw retry done: ", hw.queue, hw.topic, hw.Context.Err())
			return
		default:
			jobs, err := hw.GetJobsFromQueue(hw.queue, hw.topic, RetryQueue)
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
						hw.AddJobs(jobs, CommonQueue)
					}
				} else {
					hw.sendToRetryQueue(jobs)
					time.Sleep(10 * time.Millisecond)
				}
			} else {
				hw.sendToTimeoutQueue(jobs)
				time.Sleep(10 * time.Millisecond)
			}

		}
	}
}

// @todo move to storage
func (hw *HttpWorker) sendToRetryQueue(jobs *Jobs) error {
	err := hw.storager.AddQueue(jobs, RetryQueue)
	if err != nil {
		flog.Error("mux add jobs to retrying queue fail errors: " + err.Error() + " traceId: " + jobs.outTraceId)
		return err
	}
	flog.Error("mux save jobs to retry queue success outTraceId: ", jobs.outTraceId)
	return nil
}

func (hw *HttpWorker) sendToTimeoutQueue(jobs *Jobs) error {
	err := hw.storager.AddQueue(jobs, TimeoutQueue)
	if err != nil {
		flog.Error("mux add jobs to timeout queue fail errors: " + err.Error() + " traceId: " + jobs.outTraceId)
		return err
	}
	flog.Error("mux save jobs to timeout queue success outTraceId: ", jobs.outTraceId)
	return nil
}

func (hw *HttpWorker) loopPing() {
	pingSleepDuration := PingSleepDuration
	for {
		select {
		case <-hw.Done():
			flog.Warn("ping done: ", hw.queue, hw.topic, hw.Context.Err())
			return
		default:
			err := hw.SendPing()
			if err == nil {
				return
			}
			flog.Error("mux ping failed: ", hw.queue, hw.topic, err.Error())
			time.Sleep(time.Duration(pingSleepDuration) * time.Second)
			pingSleepDuration += pingSleepDuration
			if pingSleepDuration >= PingSleepDurationMax {
				pingSleepDuration = PingSleepDurationMax
			}
		}
	}
}

func (hw *HttpWorker) SendPing() error {
	resp, err := hw.RemoteSender.SendPing()
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return errors.New(resp.String())
	}
	queueTopic := strings.Split(resp.Msg, ",")
	if utility.InArray(fmt.Sprintf("%v/%v", hw.queue, hw.topic), queueTopic) {
		return nil
	}

	errInfo := fmt.Sprintf("topic no register: %s/%s response is %s", hw.queue, hw.topic, queueTopic)
	return errors.New(errInfo)
}

func (hw *HttpWorker) SendJobs(jobs *Jobs) *WorkerResponse {
	resp, err := hw.RemoteSender.SendJobs(jobs)
	if err != nil {
		resp = &WorkerResponse{}
		resp.Code = BadGateway
		resp.Msg = err.Error()
	}
	return resp
}

func (hw *HttpWorker) Close() {
	hw.cancel()
}
