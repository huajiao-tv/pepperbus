package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
	"time"
)

const BusBufferSize = 1000

const (
	RequestTypePing = "ping"
	RequestTypeData = "data"
	ResponseSuccess = "200"

	RemoteResponseFail = "400"
	ResponseNotFound   = "404"
	InnerError         = "500"
	BadGateway         = "502"
)

// CgiManager 服务于每个 topic 的 CGI 相关部分处理
// 包含调用连接池、发送 CGI 请求、解析返回结果
type CgiManager struct {
	cgiConfigKey string
	scriptEntry  string
}

func NewCgiManager(cgiConfigKey, scriptEntry string) *CgiManager {
	cm := &CgiManager{
		cgiConfigKey: cgiConfigKey,
		scriptEntry:  scriptEntry,
	}
	return cm
}

// GetEnv 用于获取 CGI 请求的一些基本参数
func (cm *CgiManager) GetEnv() map[string]string {
	return map[string]string{
		"SCRIPT_FILENAME": cm.scriptEntry,
	}
}

func (cm *CgiManager) GetParams(typ string) url.Values {
	return url.Values{
		"IS_PEPPER_BUS_REQUEST": []string{"true"},
		"type":                  []string{typ},
	}
}

type CGIStatus struct {
	Pool               string
	ProcessManager     string
	StartTime          string
	StartSince         string
	AcceptedConn       string
	ListenQueue        string
	MaxListenQueue     string
	ListenQueueLen     string
	IdleProcesses      string
	ActiveProcesses    string
	TotalProcesses     string
	MaxActiveProcesses string
	MaxChildrenReached string
	SlowRequests       string
}

func (cm *CgiManager) SendStatus() (*CGIStatus, error) {
	cgi, err := cm.getConn()
	if err != nil {
		flog.Error("get cgi client form pool failed", cm.cgiConfigKey, err.Error())
		return nil, err
	}
	defer cgi.Close()

	resp, err := cgi.Send(map[string]string{
		"SCRIPT_FILENAME": "/status",
		"SCRIPT_NAME":     "/status",
	}, url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dataByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	dataLines := strings.Split(string(dataByte), "\n")

	ret := &CGIStatus{}
	for _, line := range dataLines {
		if line == "" {
			continue
		}
		tmp := strings.Split(line, ":")
		key := strings.TrimSpace(tmp[0])
		value := strings.TrimSpace(tmp[1])
		switch key {
		case "pool":
			ret.Pool = value
		case "process manager":
			ret.ProcessManager = value
		case "start time":
			ret.StartTime = value
		case "start since":
			ret.StartSince = value
		case "accepted conn":
			ret.AcceptedConn = value
		case "listen queue":
			ret.ListenQueue = value
		case "max listen queue":
			ret.MaxListenQueue = value
		case "listen queue len":
			ret.ListenQueueLen = value
		case "idle processes":
			ret.IdleProcesses = value
		case "active processes":
			ret.ActiveProcesses = value
		case "total processes":
			ret.TotalProcesses = value
		case "max active processes":
			ret.MaxActiveProcesses = value
		case "max children reached":
			ret.MaxChildrenReached = value
		case "slow requests":
			ret.SlowRequests = value
		}
	}

	return ret, nil
}

// SendPing 发送 SDK 级 Ping 指令
func (cm *CgiManager) SendPing() (*WorkerResponse, error) {
	return cm.Request(cm.GetEnv(), cm.GetParams(RequestTypePing), false)
}

// SendJobs 发送需要执行的实际任务
func (cm *CgiManager) SendJobs(jobs *Jobs) (*WorkerResponse, error) {
	jbs, _ := json.Marshal(jobs.Slice)
	params := cm.GetParams(RequestTypeData)
	params.Add("outTraceId", jobs.outTraceId)
	params.Add("queue", string(jobs.queueName))
	params.Add("topic", string(jobs.topicName))
	params.Add("jobs", string(jbs))
	return cm.Request(cm.GetEnv(), params, false)
}

func (cm *CgiManager) getConn() (*CgiConn, error) {
	pool, err := CgiPoolManager.Get(cm.cgiConfigKey)
	if err != nil {
		return nil, err
	}
	return pool.GetConn(), nil
}

func (cm *CgiManager) newConnReal() (*CgiConn, error) {
	pool, err := CgiPoolManager.Get(cm.cgiConfigKey)
	if err != nil {
		return nil, err
	}
	return pool.newConnReal()
}

func (cm *CgiManager) request(env map[string]string, params url.Values, newConn bool) (*WorkerResponse, error) {
	var (
		cgi *CgiConn
		err error
	)
	if newConn {
		cgi, err = cm.newConnReal()
		if err != nil {
			flog.Error("get cgi client form new conn real failed", cm.cgiConfigKey, err.Error())
			return nil, err
		}
	} else {
		cgi, err = cm.getConn()
		if err != nil {
			flog.Error("get cgi client form pool failed", cm.cgiConfigKey, err.Error())
			return nil, err
		}
	}
	defer cgi.Close()

	resp, err := cgi.Send(env, params)
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

// Request 发送 CGI 请求，并处理返回结果
// 包含判断 EOF 错误逻辑
func (cm *CgiManager) Request(env map[string]string, params url.Values, newConn bool) (*WorkerResponse, error) {
	var (
		err        error
		retryTimes int
	)
	params.Add("IS_PEPPER_BUS_REQUEST", "true")

Request:
	resp, err := cm.request(env, params, newConn)
	if err != nil {
		newConn = true
		if err == io.EOF {
			// 当错误为 EOF 时，重试
			if retryTimes < 5 {
				retryTimes++
				flog.Error("request to cgi fail eof retry", cm.cgiConfigKey, retryTimes, env)
				time.Sleep(200 * time.Millisecond)
				goto Request
			}
		}
		if strings.Contains(err.Error(), "reset by peer") {
			// 当错误为 reset by peer 时，重试
			if retryTimes < 5 {
				retryTimes++
				flog.Error("request to cgi fail reset by peer retry", cm.cgiConfigKey, retryTimes, env)
				time.Sleep(200 * time.Millisecond)
				goto Request
			}
		}
		if strings.Contains(err.Error(), "broken pipe") {
			// 当错误为 broken pipe 时，重试
			if retryTimes < 5 {
				retryTimes++
				flog.Error("request to cgi fail broken pipe retry", cm.cgiConfigKey, retryTimes, env)
				time.Sleep(200 * time.Millisecond)
				goto Request
			}
		}
		flog.Error("request to cgi fail", cm.cgiConfigKey, err.Error(), env)
		return nil, err
	}

	return resp, nil
}

// WorkerResponse 与 SDK 约定的返回结果
type WorkerResponse struct {
	Code          string
	ErrorCode     string
	ConsumeTopics string
	Msg           string
}

// IsSuccess 返回该结果是否是成功的返回
// SDK 级
func (cr *WorkerResponse) IsSuccess() bool {
	if cr.Code == ResponseSuccess {
		return true
	}
	return false
}

func (cr WorkerResponse) String() string {
	return fmt.Sprintf("worker execute Code: %s, ErrorCode: %s, Msg: %s", cr.Code, cr.ErrorCode, cr.Msg)
}
