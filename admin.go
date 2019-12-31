package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/schema"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//
var (
	HttpReadTimeout  = time.Duration(10) * time.Second
	HttpWriteTimeout = time.Duration(10) * time.Second
)

func AdminServer(lis net.Listener) error {
	http.HandleFunc("/metrics", PrometheusHandler)
	http.HandleFunc("/global/metrics", GlobalPrometheusHandler)
	http.HandleFunc("/", ServiceHandler)
	server := &http.Server{
		ReadTimeout:  HttpReadTimeout,
		WriteTimeout: HttpWriteTimeout,
	}
	flog.Trace("AdminServe Listen on tcp: ", server.Addr)
	if err := server.Serve(lis); err != nil {
		return err
	}
	return nil
}

type Req struct {
	Queue QueueName `schema:"queue"`
	Topic TopicName `schema:"topic"`
}

// Resp define response data
type Resp struct {
	ErrorCode int         `json:"error_code"`
	Error     string      `json:"error"`
	Data      interface{} `json:"data"`
}

// HelloWorldAction say i'm alive
func (s *ServiceController) HelloWorldAction() {
	s.response.Write([]byte("hello world"))
}

func (s *ServiceController) QueueConfigAction() {
	conf := QueueConfs()
	s.renderJSON(Resp{Data: conf})
}

type GetCGIErrorReq struct {
	Req
	Count int `schema:"count"`
}

func (s *ServiceController) GetCGIErrorAction() {
	request := GetCGIErrorReq{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&request, s.request.Form); err != nil {

		s.renderJSON(Resp{http.StatusBadRequest, err.Error(), ""})
		return
	}

	data := CGIErrorCollection.GetError(request.Queue, request.Topic, request.Count)
	s.renderJSON(Resp{Data: data})
}

type GetCGITaskDebugReq struct {
	Req
	Count int `schema:"count"`
}

func (s *ServiceController) GetCGITaskDebugAction() {

	request := GetCGITaskDebugReq{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&request, s.request.Form); err != nil {
		s.renderJSON(Resp{http.StatusBadRequest, err.Error(), ""})
		return
	}

	data := CGITaskDebug.GetTask(request.Queue, request.Count)
	s.renderJSON(Resp{Data: data})
}

type RecoverReq struct {
	Req
	Count int `schema:"count"`
}

func (s *ServiceController) RecoverAction() {
	request := RecoverReq{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&request, s.request.Form); err != nil {
		s.renderJSON(Resp{http.StatusBadRequest, err.Error(), ""})
		return
	}
	mux, err := MuxManager.GetMuxer(request.Queue, request.Topic)
	if err != nil {
		s.renderJSON(Resp{http.StatusBadRequest, err.Error(), ""})
		return
	}
	n, err := mux.Storage().RetryQueue(request.Queue, request.Topic, request.Count)
	if err != nil {
		s.renderJSON(Resp{http.StatusInternalServerError, err.Error(), ""})
		return
	}
	s.renderJSON(Resp{Data: n})
}

//清除统计数据

func (s *ServiceController) CleanStatisticsAction() {
	request := Req{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&request, s.request.Form); err != nil {
		s.renderJSON(Resp{http.StatusBadRequest, err.Error(), ""})
		return
	}
	suc := 0
	queue, topic := string(request.Queue), string(request.Topic)
	jobConsumes.WithLabelValues(queue, topic).Observe(float64(time.Second.Nanoseconds() * 0))
	jobSuccessCount.WithLabelValues(queue, topic).Add(0)
	jobFailCount.WithLabelValues(queue, topic).Add(0)
	if jobFailCount.DeleteLabelValues(queue, topic) {
		suc++
	}
	if jobSuccessCount.DeleteLabelValues(queue, topic) {
		suc++
	}
	if jobConsumes.DeleteLabelValues(queue, topic) {
		suc++
	}
	if suc == 0 {
		s.renderJSON(Resp{http.StatusBadRequest, "清除失败", ""})
	} else if suc > 0 && suc < 3 {
		s.renderJSON(Resp{http.StatusBadRequest, "清除不完整", ""})
	} else {
		s.renderJSON(Resp{Data: suc})
	}

}

//清除错误记录
func (s *ServiceController) CleanFailedJobsAction() {
	request := Req{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&request, s.request.Form); err != nil {
		s.renderJSON(Resp{http.StatusBadRequest, err.Error(), ""})
		return
	}
	mux, err := MuxManager.GetMuxer(request.Queue, request.Topic)
	if err != nil {
		s.renderJSON(Resp{http.StatusBadRequest, err.Error(), ""})
		return
	}
	n, err := mux.Storage().CleanQueue(request.Queue, request.Topic, RetryQueue)
	if err != nil {
		s.renderJSON(Resp{http.StatusInternalServerError, err.Error(), ""})
		return
	}
	s.renderJSON(Resp{Data: n})
}

// Parse id meta info from id string
func (s *ServiceController) GetIdInfo() {
	id := s.request.FormValue("id")
	busId, err := ParseTraceId(id)
	if err != nil {
		s.renderJSON(Resp{400, err.Error(), ""})
		return
	}
	s.renderJSON(Resp{Data: busId})
}

//
func (s *ServiceController) GetQueueLengthAction() {
	request := Req{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&request, s.request.Form); err != nil {
		s.renderJSON(Resp{http.StatusBadRequest, err.Error(), ""})
		return
	}
	mux, err := MuxManager.GetMuxer(request.Queue, request.Topic)
	if err != nil {
		s.renderJSON(Resp{http.StatusBadRequest, err.Error(), ""})
		return
	}
	normal, retry, timeout, err := mux.Storage().GetLength(request.Queue, request.Topic)
	if err != nil {
		s.renderJSON(Resp{http.StatusInternalServerError, err.Error(), ""})
		return
	}
	s.renderJSON(Resp{
		Data: map[string]int64{
			"normal":  normal,
			"retry":   retry,
			"timeout": timeout,
		}})
}

type TopicDataReq struct {
	Req
	Type  string `schema:"type"`
	Value string `schema:"value"`
}

func (s *ServiceController) SendToCgiAction() {
	request := TopicDataReq{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&request, s.request.Form); err != nil {
		s.renderJSON(Resp{http.StatusBadRequest, err.Error(), ""})
		return
	}
	mux, err := MuxManager.GetMuxer(request.Queue, request.Topic)
	if err != nil {
		s.renderJSON(Resp{http.StatusBadRequest, err.Error(), ""})
		return
	}

	switch request.Type {
	case RequestTypePing:
		err := mux.SendPing()
		if err != nil {
			s.renderJSON(Resp{Data: err.Error()})
		} else {
			s.renderJSON(Resp{Data: "ok"})
		}
	default:
		inTraceID := GenTraceId()
		resp := mux.SendJobs(&Jobs{
			inTraceId: inTraceID,
			queueName: QueueName(request.Queue),
			topicName: TopicName(request.Topic),
			Slice: []*Job{
				{
					Id:        GenId(),
					InTraceId: inTraceID,
					Content:   request.Value,
				},
			},
		})
		s.renderJSON(Resp{Data: resp})
	}
}

func (s *ServiceController) GetMuxStatsAction() {
	s.renderJSON(Resp{
		Data: MuxStat.Output(),
	})
}

func (s *ServiceController) GetGatewayStatsAction() {
	s.renderJSON(Resp{
		Data: GatewayStat.Output(),
	})
}

func (s *ServiceController) GetMachineStatsAction() {
	MachineStat.RLock()
	defer MachineStat.RUnlock()
	s.renderJSON(Resp{
		Data: struct {
			CPU    float64
			Load   float64
			Memory float64
		}{
			CPU:    MachineStat.CPU,
			Load:   MachineStat.Load,
			Memory: MachineStat.Memory,
		},
	})
}

func (s *ServiceController) CGIStatusAction() {
	configKey := s.request.FormValue("configKey")
	mgr := NewCgiManager(configKey, "")
	status, err := mgr.SendStatus()
	if err != nil {
		s.renderJSON(Resp{400, err.Error(), ""})
		return
	}
	s.renderJSON(Resp{
		Data: status,
	})

}

// test graceful restart , get pid in duration time to ensure if op succeed
func (s *ServiceController) GetPidAction() {
	duration, err := time.ParseDuration(s.request.FormValue("duration"))
	if err != nil {
		s.renderJSON(Resp{400, err.Error(), ""})
		return
	}
	time.Sleep(duration)
	s.renderJSON(Resp{
		Data: struct {
			Pid int
			Now time.Time
		}{
			Pid: os.Getpid(),
			Now: time.Now(),
		},
	})
}

func (s *ServiceController) ManageServiceAction() {
	c, err := s.request.Cookie("manage")
	if err != nil {
		s.renderJSON(Resp{400, err.Error(), ""})
		return
	}
	if c.Value != "true" {
		s.renderJSON(Resp{400, "cookie not valid", c.Value})
		return
	}
	s.renderJSON(Resp{
		Data: struct {
			Pid    int
			Result string
		}{
			Pid:    os.Getpid(),
			Result: "call success operation in action",
		},
	})
	// wait the response can send back to clients safely
	go func() {
		<-time.After(time.Millisecond * 50)
		if s.request.FormValue("op") == "restart" {
			if err := syscall.Kill(os.Getpid(), syscall.SIGUSR2); err != nil {
				flog.Error("failed to restart service", err)
			}
			return
		}
		if s.request.FormValue("op") == "stop" {
			if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
				flog.Error("failed to term service", err)
			}
			return
		}
	}()
}

// ServiceController ...
type ServiceController struct {
	response http.ResponseWriter
	request  *http.Request
}

func isAllowedIP(ip string) bool {
	for _, aip := range netGlobalConf().AdminIpLimit {
		if ip == aip {
			return true
		}
	}

	return false
}

// ServiceHandler provide routing parser
func ServiceHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	pathInfo := strings.Trim(r.URL.Path, "/")
	if pathInfo == "favicon.ico" {
		return
	}
	parts := strings.Split(pathInfo, "/")
	for k, v := range parts {
		parts[k] = strings.Title(v)
	}
	action := strings.Join(parts, "")
	service := &ServiceController{response: w, request: r}
	controller := reflect.ValueOf(service)

	method := controller.MethodByName(action + "Action")
	if !method.IsValid() {
		method = controller.MethodByName("HelloWorldAction")
	}
	method.Call([]reflect.Value{})
}

func PrometheusHandler(w http.ResponseWriter, r *http.Request) {
	updateCurrentStats()
	promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}

func GlobalPrometheusHandler(w http.ResponseWriter, r *http.Request) {
	updateTopicLength()
	promhttp.HandlerFor(globalRegistry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}

func (s *ServiceController) renderJSON(result Resp) {
	r, err := json.Marshal(result)
	if err != nil {
		result = Resp{
			ErrorCode: -1,
			Error:     err.Error(),
		}
		r, _ = json.Marshal(result)
	}
	s.response.Write(r)
	return
}

func (s *ServiceController) required(args ...string) bool {
	for _, v := range args {
		if s.query(v) == "" {
			s.renderJSON(Resp{ErrorCode: 1, Error: fmt.Sprintf("%s is required", v)})
			return false
		}
	}
	return true
}

func (s *ServiceController) query(q string) string {
	var v string
	if v = s.request.URL.Query().Get(q); v == "" {
		v = s.request.FormValue(q)
	}
	return strings.Trim(v, " ")
}
