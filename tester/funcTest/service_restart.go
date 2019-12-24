package main

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/huajiao-tv/pepperbus/utility/msgRedis"
)

type ServiceRestartTest struct {
	gateway      string
	gatewayAdmin string
	queue        string
	topic        string
	jobExecAddr  string
	sshUser      string
	sshPassword  string

	jobCount int
	finish   chan error

	gatewayClient  *msgRedis.Pool
	jobRedisClient *msgRedis.Pool
}

func NewServiceRestartTest(jobCount int) *ServiceRestartTest {
	return &ServiceRestartTest{
		gateway:      gatewayAddr,
		gatewayAdmin: gatewayAdmin,
		jobExecAddr:  jobExecAddr,
		sshUser:      sshUser,
		sshPassword:  sshPassword,
		jobCount:     jobCount,
		finish:       make(chan error, jobCount),
	}
}

func (t *ServiceRestartTest) Init() error {
	// queue和topic需在配置中设置
	queue := "testSuccessQueue"
	topic := "topic"
	auth := queue + ":783ab0ce"

	t.queue = queue
	t.topic = topic
	t.gatewayClient = msgRedis.NewPool(t.gateway, auth, 20, 20, 20)

	auth = ""
	jRedis := strings.Split(t.jobExecAddr, ":")
	if len(jRedis) == 3 {
		auth = jRedis[2]
	}
	t.jobRedisClient = msgRedis.NewPool(jRedis[0]+":"+jRedis[1], auth, 20, 20, 20)

	return nil
}

func (t *ServiceRestartTest) Run() error {
	// 初始化
	t.Init()

	ServiceRestartUrl := "http://" + t.gatewayAdmin + GatewayRestartPath

	for i := 0; i < t.jobCount; i++ {
		if i == 5 {
			go httpPost(ServiceRestartUrl, url.Values{}, nil)
		}

		go t.doRun(i)
	}

	ticker := time.NewTicker(time.Second * 60)

	errs := make([]error, 0, t.jobCount)
	count := t.jobCount
	for {
		select {
		case err := <-t.finish:
			count = count - 1
			if err != nil {
				errs = append(errs, err)
			}
		case <-ticker.C:
			return errors.New("ServiceRestart Test failed, timeout!")
		}

		if count == 0 {
			break
		}
	}

	errCount := len(errs)
	if errCount != 0 {
		str := ""
		for i := 0; i < errCount; i++ {
			str += "|"
			str += errs[i].Error()
		}

		return errors.New(fmt.Sprintf("ServiceRestart Test Failed, %d in %d tests failed. All error message are: %s", errCount, t.jobCount, str))
	}

	return nil
}

func (t *ServiceRestartTest) doRun(idx int) {
	// LPUSH
	queueTopic := t.queue + "/" + t.topic
	c := t.gatewayClient.Pop()
	val := fmt.Sprintf("%s%d", RestartValue, idx)
	args := []interface{}{queueTopic, val + "-" + t.jobExecAddr}
	_, err := c.CallN(msgRedis.RetryTimes, "LPUSH", args...)
	t.gatewayClient.Push(c)

	if err != nil {
		t.finish <- errors.New("ServiceRestart Test failed, gateway lpush error: " + err.Error())
		return
	}

	time.Sleep(3 * time.Second)

	// 检查执行结果，测试用php脚本写入redis一个执行时的时间，精确到分钟，所以判断在一分钟内即可
	key := val
	c = t.jobRedisClient.Pop()
	res, _ := c.GET(key)
	t.jobRedisClient.Push(c)
	jobExecTime, _ := strconv.ParseInt(string(res), 10, 0)

	now := time.Now().Unix()

	sub := now - jobExecTime
	if jobExecTime > now {
		sub = jobExecTime - now
	}

	if sub > 30 {
		errStr := fmt.Sprintf("ServiceRestart Test failed, job have not done, idx: %d, last exec result: %d and now is: %d", idx, jobExecTime, now)
		t.finish <- errors.New(errStr)
		return
	}

	t.finish <- nil
}
