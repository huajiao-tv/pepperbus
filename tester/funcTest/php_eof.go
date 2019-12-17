package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/qmessenger/utility/msgRedis"
)

type PHPEofTest struct {
	gateway     string
	queue       string
	topic       string
	jobExecAddr string
	logPath     string
	sshUser     string
	sshPassword string
	local       bool

	jobCount int
	finish   chan error

	gatewayClient  *msgRedis.Pool
	jobRedisClient *msgRedis.Pool
}

func NewPHPEofTest(jobCount int) *PHPEofTest {
	return &PHPEofTest{
		gateway:     gatewayAddr,
		jobExecAddr: jobExecAddr,
		logPath:     logPath,
		sshUser:     sshUser,
		sshPassword: sshPassword,
		local:       local,

		jobCount: jobCount,
		finish:   make(chan error, jobCount),
	}
}

func (t *PHPEofTest) Init() error {
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

	//
	c := t.jobRedisClient.Pop()
	c.SET("nNormalCount", "0")
	t.jobRedisClient.Push(c)

	return nil
}

func (t *PHPEofTest) Run() error {
	// 初始化
	t.Init()

	for i := 0; i < t.jobCount; i++ {
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
			return errors.New("PHPEofTest failed, timeout!")
		}

		if count == 0 {
			break
		}
	}

	c := t.jobRedisClient.Pop()
	res, _ := c.GET("nNormalCount")
	t.jobRedisClient.Push(c)
	execCount, _ := strconv.ParseInt(string(res), 10, 0)

	if int(execCount) > t.jobCount {
		return errors.New(fmt.Sprintf("PHPEofTest Failed, job exec count overflow, %d jobs exec %d times! Check EOF retry!", t.jobCount, execCount))
	}

	errCount := len(errs)
	if errCount != 0 {
		str := ""
		for i := 0; i < errCount; i++ {
			str += "|"
			str += errs[i].Error()
		}

		return errors.New(fmt.Sprintf("PHPEofTest Failed, %d in %d tests failed. All error message are: %s", errCount, t.jobCount, str))
	}

	// 检查执行结果，发送任务给cgi出现EOF时，会有日志记录，检查日志不适用于远程测试
	cmdArr := [][]string{
		{"grep", "request to cgi fail eof retry", t.logPath},
		{"wc", "-l"},
	}
	cmdStr := fmt.Sprintf("grep 'request to cgi fail eof retry' %s | wc -l", t.logPath)
	ip := strings.Split(t.gateway, ":")[0]
	path := os.Getenv("GOBIN")
	logRes, err := execCmd(t.local, path+"/remote_exec.sh", ip, t.sshUser, t.sshPassword, cmdStr, cmdArr)

	if err != nil {
		return err
	}
	logCount, err := strconv.ParseInt(string(logRes), 10, 0)
	if err != nil {
		return err
	}

	if logCount < 1 {
		str := fmt.Sprintf("PHPEofTest failed because no 'cgi fail eof' found in %s. PHPEofTest need gokeeper and php-fpm config in special way!!!!!!", t.logPath)
		return errors.New(str)
	}

	return nil
}

func (t *PHPEofTest) doRun(idx int) {
	//
	c := t.jobRedisClient.Pop()
	key := fmt.Sprintf("%s%d", NormalValue, idx)
	c.SET(key, "0")
	t.jobRedisClient.Push(c)

	// LPUSH
	queueTopic := t.queue + "/" + t.topic
	c = t.gatewayClient.Pop()
	val := fmt.Sprintf("%s%d", NormalValue, idx)
	args := []interface{}{queueTopic, val + "-" + t.jobExecAddr}
	_, err := c.CallN(msgRedis.RetryTimes, "LPUSH", args...)
	t.gatewayClient.Push(c)

	if err != nil {
		t.finish <- errors.New("PHPEofTest failed, gateway lpush error: " + err.Error())
		return
	}

	time.Sleep(3 * time.Second)

	// 检查执行结果，测试用php脚本redis一个执行时的时间，精确到分钟，所以判断在一分钟内即可
	c = t.jobRedisClient.Pop()
	res, _ := c.GET(val)
	t.jobRedisClient.Push(c)
	jobExecTime, _ := strconv.ParseInt(string(res), 10, 0)

	now := time.Now().Unix()

	sub := now - jobExecTime
	if jobExecTime > now {
		sub = jobExecTime - now
	}

	if sub > 180 {
		errStr := fmt.Sprintf("PHPEofTest failed, job have not done, idx:%d, last exec result: %d and now is: %d", idx, jobExecTime, now)
		t.finish <- errors.New(errStr)
		return
	}

	t.finish <- nil
}
