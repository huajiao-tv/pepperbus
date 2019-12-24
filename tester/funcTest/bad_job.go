package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/huajiao-tv/pepperbus/utility/msgRedis"
)

type BadJobTest struct {
	gateway     string
	logPath     string
	sshUser     string
	sshPassword string
	queue       string
	topic       string
	local       bool

	gatewayClient *msgRedis.Pool
}

func NewBadJobTest() *BadJobTest {
	return &BadJobTest{
		gateway:     gatewayAddr,
		logPath:     logPath,
		sshUser:     sshUser,
		sshPassword: sshPassword,
		local:       local,
	}
}

func (t *BadJobTest) Init() error {
	// queue和topic需在配置中设置，但是job执行的脚本没有订阅
	queue := "testSuccessQueue"
	topic := "topic"
	auth := queue + ":783ab0ce"

	t.queue = queue
	t.topic = topic
	t.gatewayClient = msgRedis.NewPool(t.gateway, auth, 20, 20, 20)

	return nil
}

func (t *BadJobTest) Run() error {
	// 初始化
	t.Init()

	// LPUSH，job在解析时只有queue和topic不是组织为queue/topic的结构时才会出错
	queueTopic := t.queue
	c := t.gatewayClient.Pop()
	args := []interface{}{queueTopic, NormalValue}
	_, err := c.CallN(msgRedis.RetryTimes, "LPUSH", args...)
	t.gatewayClient.Push(c)
	if err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	// 检查执行结果，发送任务给cgi时，如果topic未定阅，会记录日志
	cmdArr := [][]string{
		{"grep", "ParseJob fail", t.logPath},
		{"grep", queueTopic},
		{"wc", "-l"},
	}
	cmdStr := fmt.Sprintf("grep 'ParseJob fail' %s | grep %s | wc -l", t.logPath, queueTopic)
	ip := strings.Split(t.gateway, ":")[0]
	path := os.Getenv("GOBIN")
	res, err := execCmd(t.local, path+"/remote_exec.sh", ip, t.sshUser, t.sshPassword, cmdStr, cmdArr)

	if err != nil {
		return err
	}
	// 远程调用和本地调用输出不同
	if res != "1" {
		str := fmt.Sprintf("BadJobTest failed because no 'ParseJob fail' found in %s for testSuccessQueue", t.logPath)
		return errors.New(str)
	}

	return nil
}
