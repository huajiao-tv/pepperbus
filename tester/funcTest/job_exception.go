package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/huajiao-tv/pepperbus/utility/msgRedis"
)

type JobExceptionTest struct {
	gateway     string
	redis       string
	logPath     string
	sshUser     string
	sshPassword string
	queue       string
	topic       string
	local       bool

	gatewayClient *msgRedis.Pool
	redisClient   *msgRedis.Pool
}

func NewJobExceptionTest() *JobExceptionTest {
	return &JobExceptionTest{
		gateway:     gatewayAddr,
		redis:       redisAddr,
		logPath:     logPath,
		sshUser:     sshUser,
		sshPassword: sshPassword,
		local:       local,
	}
}

func (t *JobExceptionTest) Init() error {
	// queue和topic需在配置中设置
	queue := "test500Queue"
	topic := "topic"
	auth := queue + ":783ab0ce"

	t.queue = queue
	t.topic = topic
	t.gatewayClient = msgRedis.NewPool(t.gateway, auth, 20, 20, 20)

	l := strings.Split(t.redis, ":")
	if len(l) == 3 {
		auth = l[2]
	} else {
		auth = ""
	}
	t.redisClient = msgRedis.NewPool(l[0]+":"+l[1], auth, 20, 20, 20)

	return nil
}

func (t *JobExceptionTest) Run() error {
	// 初始化
	t.Init()

	// LPUSH
	queueTopic := t.queue + "/" + t.topic
	c := t.gatewayClient.Pop()
	args := []interface{}{queueTopic, NormalValue}
	respJson, err := c.CallN(msgRedis.RetryTimes, "LPUSH", args...)
	t.gatewayClient.Push(c)
	if err != nil {
		return err
	}

	var resp gatewayResp
	json.Unmarshal(respJson.([]byte), &resp)
	traceId := resp.Jobs[0]

	time.Sleep(3 * time.Second)

	// 检查执行结果，发送任务给cgi时，如果任务执行异常，会记录日志，并将任务放入retry对列
	// 日志检查
	cmdArr := [][]string{
		{"grep", "sendToCGI fail", t.logPath},
		{"grep", traceId},
		{"wc", "-l"},
	}
	cmdStr := fmt.Sprintf("grep 'sendToCGI fail' %s | grep %s | wc -l", t.logPath, traceId)
	ip := strings.Split(t.gateway, ":")[0]
	path := os.Getenv("GOBIN")
	res, err := execCmd(t.local, path+"/remote_exec.sh", ip, t.sshUser, t.sshPassword, cmdStr, cmdArr)

	if err != nil {
		return err
	}
	if res != "1" {
		str := fmt.Sprintf("JobExceptionTest failed because no 'sendToCGI fail' found in %s for traceid: %s", t.logPath, traceId)
		return errors.New(str)
	}

	// retry对列检查，检查
	key := fmt.Sprintf("%s::%s::%s", t.queue, t.topic, "list_retry")
	c = t.redisClient.Pop()
	_, err = c.RPOP(key)
	t.redisClient.Push(c)

	return nil
}
