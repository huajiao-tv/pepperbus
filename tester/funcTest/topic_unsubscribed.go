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

type TopicUnsubscribedTest struct {
	gateway     string
	logPath     string
	sshUser     string
	sshPassword string
	queue       string
	topic       string
	local       bool

	gatewayClient *msgRedis.Pool
}

func NewTopicUnsubscribedTest() *TopicUnsubscribedTest {
	return &TopicUnsubscribedTest{
		gateway:     gatewayAddr,
		logPath:     logPath,
		sshUser:     sshUser,
		sshPassword: sshPassword,
		local:       local,
	}
}

func (t *TopicUnsubscribedTest) Init() error {
	// queue和topic需在配置中设置，但是job执行的脚本没有订阅
	queue := "testUnsubscribedQueue"
	topic := "topic"
	auth := queue + ":783ab0ce"

	t.queue = queue
	t.topic = topic
	t.gatewayClient = msgRedis.NewPool(t.gateway, auth, 20, 20, 20)

	return nil
}

func (t *TopicUnsubscribedTest) Run() error {
	// 初始化
	t.Init()

	// LPUSH
	queueTopic := t.queue + "/" + t.topic
	c := t.gatewayClient.Pop()
	args := []interface{}{queueTopic, NormalValue}
	respJson, err := c.CallN(msgRedis.RetryTimes, "LPUSH", args...)
	t.gatewayClient.Push(c)

	if err != nil {
		return errors.New("TopicUnsubscribedTest failed, gateway lpush failed, error: " + err.Error())
	}
	var resp gatewayResp
	json.Unmarshal(respJson.([]byte), &resp)
	traceId := resp.Jobs[0]

	time.Sleep(3 * time.Second)

	// 检查执行结果，发送任务给cgi时，如果topic未定阅，会记录日志
	cmdArr := [][]string{
		{"grep", "topic no register", t.logPath},
		{"grep", traceId},
		{"wc", "-l"},
	}
	cmdStr := fmt.Sprintf("grep 'topic no register' %s | grep %s | wc -l", t.logPath, traceId)
	ip := strings.Split(t.gateway, ":")[0]
	path := os.Getenv("GOBIN")
	res, err := execCmd(t.local, path+"/remote_exec.sh", ip, t.sshUser, t.sshPassword, cmdStr, cmdArr)

	if err != nil {
		return err
	}
	if res != "1" {
		str := fmt.Sprintf("TopicUnsubscribedTest failed because no 'topic no register' found in %s for traceid: %s", t.logPath, traceId)
		return errors.New(str)
	}

	return nil
}
