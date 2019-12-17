package main

import (
	"net"
	"time"

	"git.huajiao.com/qmessenger/redeo"
	"git.huajiao.com/qmessenger/redeo/resp"

	//"github.com/qmessenger/utility/msgRedis"
	"strings"
	"testing"
)

// TestRedis for mocking redis server
const COMMON_QUEUE_LEN = 1
const RETRY_QUEUE_LEN = 2

type TestRedisServer struct {
	*redeo.Server
	ip   string
	port string
}

func NewTestRedis(ip, port string) *TestRedisServer {
	trs := TestRedisServer{}
	conf := &redeo.Config{
		Timeout:      time.Second * 3,
		IdleTimeout:  time.Second * 20,
		TCPKeepAlive: time.Second * 10,
	}
	trs.Server = redeo.NewServer(conf)
	trs.ip, trs.port = ip, port
	trs.HandleFunc("lpush", trs.lpush)
	trs.HandleFunc("rpop", trs.rpop)
	trs.HandleFunc("rpoplpush", trs.rpoplpush)
	trs.HandleFunc("llen", trs.llen)
	trs.HandleFunc("del", trs.del)
	return &trs
}

func (trs *TestRedisServer) Serve() error {
	lis, err := net.Listen("tcp", trs.ip+":"+trs.port)
	if err != nil {
		panic(err)
	}
	defer lis.Close()
	return trs.Server.Serve(lis)
}

func (trs *TestRedisServer) lpush(w resp.ResponseWriter, c *resp.Command) {
	fields := strings.SplitN(string(c.Args[0]), "/", 2)
	if fields[0] == "ErrorQueue::ErrorTopic::list_retry" {
		w.AppendError("ErrorQueue")
		return
	}
	w.AppendInt(1)
	return
}

func (trs *TestRedisServer) rpop(w resp.ResponseWriter, c *resp.Command) {
	fields := strings.SplitN(string(c.Args[0]), "/", 2)
	if fields[0] == "EmptyQueue::EmptyTopic::list_retry" {
		w.AppendNil()
		return
	}

	if fields[0] == "ErrorQueue::ErrorTopic::list_retry" {
		w.AppendError("ErrorQueue")
		return
	}

	jobs := &Jobs{
		queueName: "TestQueue",
		Slice:     jobsSlice,
	}

	w.Append(jobs.Encode())
	return
}

func (trs *TestRedisServer) rpoplpush(w resp.ResponseWriter, c *resp.Command) {
	fields := strings.SplitN(string(c.Args[0]), "/", 2)
	if fields[0] == "KeyNotExistQueue::TestTopic::list_retry" {
		w.AppendNil()
		return
	}
	if fields[0] == "ErrorQueue::ErrorTopic::list_retry" {
		w.AppendError("ErrorQueue")
		return
	}
	w.AppendBulkString("1")
	return
}

func (trs *TestRedisServer) llen(w resp.ResponseWriter, c *resp.Command) {
	fields := strings.SplitN(string(c.Args[0]), "/", 2)
	if fields[0] == "TestQueue::TestTopic::list" {
		w.AppendInt(COMMON_QUEUE_LEN)
		return
	}

	if fields[0] == "TestQueue::TestTopic::list_retry" {
		w.AppendInt(RETRY_QUEUE_LEN)
		return
	}

	if fields[0] == "CommonErrorQueue::TestTopic::list" {
		w.AppendError("CommonErrorQueue")
		return
	}

	if fields[0] == "RetryErrorQueue::TestTopic::list" {
		w.AppendInt(COMMON_QUEUE_LEN)
		return
	}

	if fields[0] == "RetryErrorQueue::TestTopic::list_retry" {
		w.AppendError("ErrorQueue")
		return
	}

	w.AppendError("Unknown")
	return
}

func (trs *TestRedisServer) del(w resp.ResponseWriter, c *resp.Command) {
	fields := strings.SplitN(string(c.Args[0]), "/", 2)
	// 为了测试CleanQueue，del的判断需要参考llen的判断

	// llen返回正确，del返回正确
	if fields[0] == "TestQueue::TestTopic::list" {
		w.AppendInt(1)
		return
	}

	// llen返回正确，del返回错误
	if fields[0] == "RetryErrorQueue::TestTopic::list" {
		w.AppendError("RetryErrorQueue")
		return
	}

	w.AppendError("Unknown")
	return
}

// test proc
var testStorage *storage
var ip = "127.0.0.1"
var port = "56379"
var jobsSlice = make([]*Job, 0, 4)

func init() {
	// start mock redisServer
	testRedisServer := NewTestRedis(ip, port)
	go testRedisServer.Serve()
	time.Sleep(time.Microsecond * 50)

	// new testStorage
	storageConfig := &StorageConfig{
		MaxConnNum:     30,
		MaxIdleNum:     10,
		MaxIdleSeconds: 300,
		Type:           "redis",
		Address:        ip + ":" + port,
	}

	testStorage = NewStorage(storageConfig)

	// mock jobs
	job := &Job{
		Id:        "1",
		InTraceId: "1",
		Content:   "test",
		Retry:     1,
	}
	jobsSlice = append(jobsSlice, job)

}

func TestGetKey(t *testing.T) {
	key := GetKey("TestQueue", "TestTopic", 0)
	if key != "TestQueue::TestTopic::list" {
		t.Fatalf("AddJobs Case want key = 'TestQueue::TestTopic::list' returns, but receive key = %s", key)
	} else {
		t.Log(key)
	}
}

func TestGetKeyRetry(t *testing.T) {
	key := GetKey("TestQueue", "TestTopic", 1)
	if key != "TestQueue::TestTopic::list_retry" {
		t.Fatalf("AddJobs Case want key = 'TestQueue::TestTopic::list_retry' returns, but receive key = %s", key)
	} else {
		t.Log(key)
	}
}

func TestAddQueue(t *testing.T) {
	jobs := &Jobs{
		queueName: "TestQueue",
		Slice:     jobsSlice,
	}

	err := testStorage.AddQueue(jobs, 1)
	if err != nil {
		t.Fatalf("AddQueue Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestAddQueueEmptyCase(t *testing.T) {
	jobs := &Jobs{
		queueName: "EmptyQueue",
	}

	err := testStorage.AddQueue(jobs, 1)
	if err.Error() != "empty jobs" {
		t.Fatalf("AddQueue Empty Case want err = 'empty jobs' returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestGetJobsFromQueue(t *testing.T) {
	_, err := testStorage.GetJobsFromQueue("TestQueue", "TestTopic", 1)

	if err != nil {
		t.Fatalf("GetJobsFromQueue Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestGetJobsFromQueueEmptyCase(t *testing.T) {
	jobs, err := testStorage.GetJobsFromQueue("EmptyQueue", "EmptyTopic", 1)

	if err != nil || jobs != nil {
		t.Fatalf("GetJobsFromQueue Empty Case want err = nil and jobs = nil returns, but receive err = %v and jobs = %v", err, jobs)
	} else {
		t.Log(err)
	}
}

func TestGetJobsFromQueueErrorCase(t *testing.T) {
	_, err := testStorage.GetJobsFromQueue("ErrorQueue", "ErrorTopic", 1)

	if err == nil {
		t.Fatalf("GetJobsFromQueue Error Case want err = 'ErrorQueue' returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestRetryQueue(t *testing.T) {
	count := 3
	res, err := testStorage.RetryQueue("TestQueue", "TestTopic", count)
	if err == nil && res == int64(count) {
		t.Log(res, err)
	} else {
		t.Fatalf("TestRetryQueue Case want err = nil and res = %d returns, but receive err = %v, res = %d", count, err, res)
	}
}

func TestRetryQueueKeyNotExistCase(t *testing.T) {
	count := 3
	res, err := testStorage.RetryQueue("KeyNotExistQueue", "TestTopic", count)
	if err == nil && res < int64(count) {
		t.Log(res, err)
	} else {
		t.Fatalf("TestRetryQueue Case want err = nil and res = %d returns, but receive err = %v, res = %d", count, err, res)
	}
}

func TestRetryQueueErrorCase(t *testing.T) {
	count := 3
	res, err := testStorage.RetryQueue("ErrorQueue", "ErrorTopic", count)
	if err != nil && err.Error() == "CommonError:ErrorQueue" {
		t.Log(res, err)
	} else {
		t.Fatalf("TestRetryQueue Case want err = CommonError:ErrorQueue returns, but receive err = %v, res = %d", err, res)
	}
}

func TestGetLength(t *testing.T) {
	res, res1, err := testStorage.GetLength("TestQueue", "TestTopic")
	if err == nil && res == COMMON_QUEUE_LEN && res1 == RETRY_QUEUE_LEN {
		t.Log(res, res1, err)
	} else {
		t.Fatalf("GetLengthCommon Case want err = nil and res = %d returns, but receive err = %v, res = %v, res1 = %v", COMMON_QUEUE_LEN, err, res, res1)
	}
}

func TestGetLengthCommonError(t *testing.T) {
	res, res1, err := testStorage.GetLength("CommonErrorQueue", "TestTopic")
	if err != nil {
		t.Log(res, res1, err)
	} else {
		t.Fatalf("GetLengthCommon Error Case want err != nil returns, but receive err = %v, res = %v, res1 = %v", err, res, res1)
	}
}

func TestGetLengthRetryError(t *testing.T) {
	res, res1, err := testStorage.GetLength("RetryErrorQueue", "TestTopic")
	if err != nil && res == COMMON_QUEUE_LEN {
		t.Log(res, res1, err)
	} else {
		t.Fatalf("GetLengthRetry Error Case want err != nil and res = %d returns, but receive err = %v, res = %v, res1 = %v", COMMON_QUEUE_LEN, err, res, res1)
	}
}

func TestCleanQueue(t *testing.T) {
	res, err := testStorage.CleanQueue("TestQueue", "TestTopic", 0)
	if err != nil {
		t.Fatalf("CleanQueue Case want err = nil returns, but receive err = %v, res = %d", err, res)
	} else {
		t.Log(res, err)
	}
}

func TestCleanQueueDelError(t *testing.T) {
	res, err := testStorage.CleanQueue("RetryErrorQueue", "TestTopic", 0)
	if err != nil && err.Error() == "CommonError:RetryErrorQueue" {
		t.Log(res, err)
	} else {
		t.Fatalf("CleanQueueDelError Case want err = RetryErrorQueue returns, but receive err = %v, res = %d", err, res)
	}
}

func TestCleanQueueLenError(t *testing.T) {
	res, err := testStorage.CleanQueue("CommonErrorQueue", "TestTopic", 0)
	if err != nil && err.Error() == "CommonError:CommonErrorQueue" {
		t.Log(res, err)
	} else {
		t.Fatalf("CleanQueueLenError Case want err = CommonErrorQueue returns, but receive err = %v, res = %d", err, res)
	}
}
