package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"git.huajiao.com/qmessenger/redeo"
	"git.huajiao.com/qmessenger/redeo/resp"
	"github.com/CodisLabs/codis/pkg/utils/errors"
)

type Gateway struct {
	*redeo.Server
	ip   string
	port string
	MuxServerIntf
}

func NewGateway(ip, port string, server MuxServerIntf) *Gateway {
	gw := Gateway{}
	conf := &redeo.Config{
		Timeout:      time.Second * 30,
		IdleTimeout:  time.Second * 20,
		TCPKeepAlive: time.Second * 10,
	}
	gw.Server = redeo.NewServer(conf)
	gw.MuxServerIntf = server
	gw.ip, gw.port = ip, port
	// gw.Server.Info().String()

	gw.HandleFunc("auth", gw.auth)
	gw.HandleFunc("ping", gw.ping)
	gw.HandleFunc("lpush", gw.lpush)
	gw.HandleFunc("rpop", gw.rpop)
	gw.HandleFunc("info", gw.info)
	return &gw
}

// Serve with a listener
func (gw *Gateway) Serve(lis net.Listener) error {
	flog.Trace("Gateway Listen on tcp://%s", gw.ip+":"+gw.port)
	err := gw.Server.Serve(lis)
	timeout := time.After(GraceStopTimeout)
	tick := time.Tick(time.Second)
Loop:
	for {
		select {
		case <-tick:
			flog.Error("total gateway connection left", gw.Server.Info().TotalConnections())
			if gw.Server.Info().TotalConnections() == 0 {
				flog.Error("connection == 0 ")
				break Loop
			}
		case <-timeout:
			flog.Error("time out graceful restart some connection is in active")
			break Loop
		}
	}
	flog.Error("gateway service done")
	return err
}

// Serve with ip and port
func (gw *Gateway) ListenAndServe() error {
	l, err := GlobalNet.Listen("tcp", gw.ip+":"+gw.port)
	if err != nil {
		return err
	}
	return gw.Server.Serve(l)
}

func (gw *Gateway) checkAuth(ctx context.Context, qn QueueName, tn TopicName) error {
	client := redeo.GetClient(ctx)
	if auth := client.Context().Value("auth"); auth != nil {
		val := fmt.Sprintf("%v/%v", qn, tn)
		if tn == "" {
			val = string(qn)
		}
		if v, ok := auth.(string); ok {
			if val == v || (qn == "" && v != "") {
				return nil
			}
		}
	}
	return errors.New("NOAUTH Authentication required.")
}

// auth format is: queuename</topicname>:passwd
func (gw *Gateway) auth(w resp.ResponseWriter, c *resp.Command) {
	if c.ArgN() == 0 {
		w.AppendError(redeo.WrongNumberOfArgs(c.Name))
		return
	}
	client := redeo.GetClient(c.Context())
	fields := strings.Split(c.Args[0].String(), ":")
	var qn, tn string
	if len(fields) < 2 {
		goto InvalidPwd
	}
	if qt := strings.Split(fields[0], "/"); len(qt) == 2 {
		qn, tn = qt[0], qt[1]
	} else {
		qn = qt[0]
	}
	if queue, ok := QueueConfs()[QueueName(qn)]; !ok {
		goto InvalidPwd
	} else if tn != "" {
		topic, ok := queue.Topic[TopicName(tn)]
		if !ok || topic.Password != fields[1] {
			goto InvalidPwd
		}
	} else if queue.Password != fields[1] {
		goto InvalidPwd
	}

	client.SetContext(context.WithValue(c.Context(), "auth", fields[0]))
	w.AppendOK()
	return

InvalidPwd:
	w.AppendError("ERR invalid password")
	client.SetContext(context.WithValue(c.Context(), "auth", false))
}

// test graceful restart return pid and start time
func (gw *Gateway) info(w resp.ResponseWriter, c *resp.Command) {
	time.Sleep(time.Second * 5)
	str := fmt.Sprintf("%s started at %s slept for %d second from pid %d.\n", "redeo", time.Now().String(), 5, os.Getpid())
	w.AppendInlineString(str)
}

func (gw *Gateway) ping(w resp.ResponseWriter, c *resp.Command) {
	if err := gw.checkAuth(c.Context(), "", ""); err != nil {
		w.AppendError(err.Error())
		return
	}
	w.AppendInlineString("PONG")
}

// WriteTaskDebug 调试用！将用户添加的任务记录到存储
func WriteTaskDebug(jobs *Jobs) {
	for _, job := range jobs.Slice {
		CGITaskDebug.AddTask(jobs.queueName, job.Content)
	}
}

// push jobs and return jobs id to client
func (gw *Gateway) lpush(w resp.ResponseWriter, c *resp.Command) {
	client := redeo.GetClient(c.Context())
	ip := client.RemoteAddr()

	flog.Trace(c.Name, c.Args, ip)

	var err error
	var ids []string

	var WriteGatewayResp = func(ids []string, inTraceID string, err error) {
		type resp struct {
			Jobs      []string `json:"jobs"`
			InTraceId string   `json:"inTraceId"`
			Error     string   `json:"error"`
		}
		if err != nil {
			json, _ := json.Marshal(resp{nil, inTraceID, err.Error()})
			w.AppendBulk(json)
			return
		}
		json, _ := json.Marshal(resp{ids, inTraceID, ""})
		w.AppendBulk(json)
		return
	}

	// 参数检查
	if c.ArgN() == 0 {
		WriteGatewayResp(nil, "", errors.New("Invalid Args"))
		return
	}

	jobs, ids, err := ParseJob(c.Args)
	if err != nil {
		flog.Error("Gateway lpush:", err.Error(), ip)
		WriteGatewayResp(nil, "", err)
		return
	}
	if err = gw.checkAuth(c.Context(), jobs.queueName, jobs.topicName); err != nil {
		WriteGatewayResp(nil, jobs.inTraceId, err)
		return
	}

	if err = gw.MuxServerIntf.AddJobs(jobs); err != nil {
		flog.Error("Gateway lpush jobs fail:", err.Error(), ip)
		WriteGatewayResp(nil, jobs.inTraceId, err)
	}
	WriteGatewayResp(ids, jobs.inTraceId, err)

	WriteTaskDebug(jobs)
	GatewayStat.Add(jobs.queueName, err == nil)
	return
}

// pop job according quene/topic name
func (gw *Gateway) rpop(w resp.ResponseWriter, c *resp.Command) {
	var err error
	flog.Trace(c.Name, c.Args)

	// 参数检查
	if c.ArgN() == 0 {
		w.Append("Invalid Args")
		return
	}

	fields := strings.SplitN(string(c.Args[0]), "/", 2)
	if len(fields) < 2 {
		w.Append("ParseJob fail: invalid job meta info: " + string(c.Args[0]))
		return
	}
	qName, tName := QueueName(fields[0]), TopicName(fields[1])
	if err := gw.checkAuth(c.Context(), qName, tName); err != nil {
		w.AppendError(err.Error())
		return
	}
	jobs, err := gw.MuxServerIntf.GetJobs(qName, tName, CommonQueue)
	if err != nil {
		flog.Trace("Gateway lpop: " + err.Error())
		w.Append(err.Error())
		return
	} else if jobs == nil {
		w.AppendNil()
	} else {
		w.Append(jobs.Encode())
	}
}
