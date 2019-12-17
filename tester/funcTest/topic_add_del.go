package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/qmessenger/utility/msgRedis"
)

const gokeeperUpdateUrl = "http://10.143.168.10:17000/conf/manage"

type TopicAddDelTest struct {
	gateway     string
	queue       string
	topic       string
	jobExecAddr string
	sshUser     string
	sshPassword string

	gatewayClient  *msgRedis.Pool
	jobRedisClient *msgRedis.Pool
}

func NewTopicAddDelTest() *TopicAddDelTest {
	return &TopicAddDelTest{
		gateway:     gatewayAddr,
		jobExecAddr: jobExecAddr,
		sshUser:     sshUser,
		sshPassword: sshPassword,
	}
}

func (t *TopicAddDelTest) Init() error {
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

func (t *TopicAddDelTest) Run() error {
	// 初始化
	t.Init()

	// 添加topic，然后测试
	AddTopic()
	time.Sleep(2 * time.Second)
	// LPUSH
	queueTopic := t.queue + "/" + t.topic
	c := t.gatewayClient.Pop()
	args := []interface{}{queueTopic, NormalValue + "-" + t.jobExecAddr}
	_, err := c.CallN(msgRedis.RetryTimes, "LPUSH", args...)
	t.gatewayClient.Push(c)

	if err != nil {
		return errors.New("TopicAddDelTest AddTopic Test failed, gateway lpush error: " + err.Error())
	}

	time.Sleep(2 * time.Second)

	// 检查执行结果，测试用php脚本写入redis一个执行时的时间，精确到分钟，所以判断在一分钟内即可
	// 检查pepperbus_test
	c = t.jobRedisClient.Pop()
	res, _ := c.GET(NormalValue)
	t.jobRedisClient.Push(c)
	jobExecTime, _ := strconv.ParseInt(string(res), 10, 0)

	now := time.Now().Unix()
	sub := now - jobExecTime
	if jobExecTime > now {
		sub = jobExecTime - now
	}

	if sub > 30 {
		errStr := fmt.Sprintf("TopicAddDelTest AddTopic Test failed, job have not done, last exec result in pepperbus_test: %d and now is: %d", jobExecTime, now)
		return errors.New(errStr)
	}

	// 检查pepperbus_test_add_del，job执行时，新增的topic写入redis的key为job名加AddDel
	c = t.jobRedisClient.Pop()
	res, _ = c.GET(NormalValue + "AddDel")
	t.jobRedisClient.Push(c)
	jobExecTime, _ = strconv.ParseInt(string(res), 10, 0)

	now = time.Now().Unix()
	sub = now - jobExecTime
	if jobExecTime > now {
		sub = jobExecTime - now
	}

	if sub > 30 {
		errStr := fmt.Sprintf("TopicAddDelTest AddTopic Test failed, job have not done, last exec result in pepperbus_test_add_del: %d and now is: %d", jobExecTime, now)
		return errors.New(errStr)
	}

	///////////////////////////////////////
	// 删除topic，上一步新增topic写入redis的值清零，然后测试
	DelTopic()
	time.Sleep(2 * time.Second)
	c = t.jobRedisClient.Pop()
	res, _ = c.SET(NormalValue+"AddDel", "0")
	t.jobRedisClient.Push(c)

	// LPUSH
	c = t.gatewayClient.Pop()
	args = []interface{}{queueTopic, NormalValue + "-" + t.jobExecAddr}
	_, err = c.CallN(msgRedis.RetryTimes, "LPUSH", args...)
	t.gatewayClient.Push(c)

	if err != nil {
		return errors.New("TopicAddDelTest DelTopic Test failed, gateway lpush error: " + err.Error())
	}

	time.Sleep(2 * time.Second)

	// 检查执行结果，测试用php脚本写入redis一个执行时的时间，精确到分钟，所以判断在一分钟内即可
	// 检查pepperbus_test
	c = t.jobRedisClient.Pop()
	res, _ = c.GET(NormalValue)
	t.jobRedisClient.Push(c)
	jobExecTime, _ = strconv.ParseInt(string(res), 10, 0)

	now = time.Now().Unix()
	sub = now - jobExecTime
	if jobExecTime > now {
		sub = jobExecTime - now
	}

	if sub > 30 {
		errStr := fmt.Sprintf("TopicAddDelTest DelTopic Test failed, job have not done, last exec result in pepperbus_test: %d and now is: %d", jobExecTime, now)
		return errors.New(errStr)
	}

	// 检查pepperbus_test_add_del
	c = t.jobRedisClient.Pop()
	res, _ = c.GET(NormalValue + "AddDel")
	t.jobRedisClient.Push(c)
	jobExecTime, _ = strconv.ParseInt(string(res), 10, 0)

	now = time.Now().Unix()
	sub = now - jobExecTime
	if jobExecTime > now {
		sub = jobExecTime - now
	}

	if sub < 30 {
		errStr := fmt.Sprintf("TopicAddDelTest DelTopic Test failed, job have not done, last exec result in pepperbus_test_add_del: %d and now is: %d", jobExecTime, now)
		return errors.New(errStr)
	}

	return nil
}

func AddTopic() error {
	config := "{\"test500Queue\":{\"queue_name\":\"test500Queue\",\"password\":\"783ab0ce\",\"topic\":{\"topic\":{\"name\":\"topic\",\"password\":\"783ab0ce\",\"retry_times\":3,\"max_queue_length\":1000,\"run_type\":0,\"storage\":{\"type\":\"redis\",\"address\":\"127.0.0.1:6380\",\"auth\":\"\",\"max_conn_num\":30,\"max_idle_num\":10,\"max_idle_seconds\":300},\"num_of_workers\":10,\"script_entry\":\"/tmp/consume.php\",\"cgi_config_key\":\"local\"}}},\"testSuccessQueue\":{\"queue_name\":\"testSuccessQueue\",\"password\":\"783ab0ce\",\"topic\":{\"topic\":{\"name\":\"topic\",\"password\":\"783ab0ce\",\"retry_times\":3,\"max_queue_length\":1000,\"run_type\":0,\"storage\":{\"type\":\"redis\",\"address\":\"127.0.0.1:6380\",\"auth\":\"\",\"max_conn_num\":30,\"max_idle_num\":10,\"max_idle_seconds\":300},\"num_of_workers\":10,\"script_entry\":\"/tmp/consume.php\",\"cgi_config_key\":\"local\"},\"topicForAddDel\":{\"name\":\"topicForAddDel\",\"password\":\"783ab0ce\",\"retry_times\":3,\"max_queue_length\":1000,\"run_type\":0,\"storage\":{\"type\":\"redis\",\"address\":\"127.0.0.1:6380\",\"auth\":\"\",\"max_conn_num\":30,\"max_idle_num\":10,\"max_idle_seconds\":300},\"num_of_workers\":10,\"script_entry\":\"/tmp/consume.php\",\"cgi_config_key\":\"local\"}}},\"testUnsubscribedQueue\":{\"queue_name\":\"testUnsubscribedQueue\",\"password\":\"783ab0ce\",\"topic\":{\"topic\":{\"name\":\"topic\",\"password\":\"783ab0ce\",\"retry_times\":3,\"max_queue_length\":1000,\"run_type\":0,\"storage\":{\"type\":\"redis\",\"address\":\"127.0.0.1:6380\",\"auth\":\"\",\"max_conn_num\":30,\"max_idle_num\":10,\"max_idle_seconds\":300},\"num_of_workers\":10,\"script_entry\":\"/tmp/consume.php\",\"cgi_config_key\":\"local\"}}}}"

	l := strings.Split(domain, "/")
	operates := make(map[string]string)
	operates["opcode"] = "update"
	operates["file"] = l[1] + "/queue.conf"
	operates["section"] = "DEFAULT"
	operates["key"] = "queue_config"
	operates["value"] = config
	operates["type"] = ""
	operatesJson, _ := json.Marshal([]map[string]string{operates})

	values := url.Values{
		"domain":   []string{l[0]},
		"operates": []string{string(operatesJson)},
		"note":     []string{"test add topic"},
	}

	return httpPost(gokeeperUpdateUrl, values, nil)

}

func DelTopic() error {
	config := "{\"test500Queue\":{\"queue_name\":\"test500Queue\",\"password\":\"783ab0ce\",\"topic\":{\"topic\":{\"name\":\"topic\",\"password\":\"783ab0ce\",\"retry_times\":3,\"max_queue_length\":1000,\"run_type\":0,\"storage\":{\"type\":\"redis\",\"address\":\"127.0.0.1:6380\",\"auth\":\"\",\"max_conn_num\":30,\"max_idle_num\":10,\"max_idle_seconds\":300},\"num_of_workers\":10,\"script_entry\":\"/tmp/consume.php\",\"cgi_config_key\":\"local\"}}},\"testSuccessQueue\":{\"queue_name\":\"testSuccessQueue\",\"password\":\"783ab0ce\",\"topic\":{\"topic\":{\"name\":\"topic\",\"password\":\"783ab0ce\",\"retry_times\":3,\"max_queue_length\":1000,\"run_type\":0,\"storage\":{\"type\":\"redis\",\"address\":\"127.0.0.1:6380\",\"auth\":\"\",\"max_conn_num\":30,\"max_idle_num\":10,\"max_idle_seconds\":300},\"num_of_workers\":10,\"script_entry\":\"/tmp/consume.php\",\"cgi_config_key\":\"local\"}}},\"testUnsubscribedQueue\":{\"queue_name\":\"testUnsubscribedQueue\",\"password\":\"783ab0ce\",\"topic\":{\"topic\":{\"name\":\"topic\",\"password\":\"783ab0ce\",\"retry_times\":3,\"max_queue_length\":1000,\"run_type\":0,\"storage\":{\"type\":\"redis\",\"address\":\"127.0.0.1:6380\",\"auth\":\"\",\"max_conn_num\":30,\"max_idle_num\":10,\"max_idle_seconds\":300},\"num_of_workers\":10,\"script_entry\":\"/tmp/consume.php\",\"cgi_config_key\":\"local\"}}}}"

	l := strings.Split(domain, "/")
	operates := make(map[string]string)
	operates["opcode"] = "update"
	operates["file"] = l[1] + "/queue.conf"
	operates["section"] = "DEFAULT"
	operates["key"] = "queue_config"
	operates["value"] = config
	operates["type"] = "string"
	operatesJson, _ := json.Marshal([]map[string]string{operates})

	values := url.Values{
		"domain":   []string{l[0]},
		"operates": []string{string(operatesJson)},
		"note":     []string{"test add topic"},
	}

	return httpPost(gokeeperUpdateUrl, values, nil)
}
