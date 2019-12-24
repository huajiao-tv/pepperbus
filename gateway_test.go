package main

import (
	"encoding/json"
	"testing"
	"time"

	"errors"

	"github.com/huajiao-tv/pepperbus/utility/msgRedis"
)

type gwResp struct {
	Jobs      []string `json:"jobs"`
	InTraceId string   `json:"inTraceId"`
	Error     string   `json:"error"`
}

// TestMuxServer for test inject
type TestMuxServer struct{}

func NewTestMuxServer() *TestMuxServer {
	return &TestMuxServer{}
}

func (tms *TestMuxServer) Serve() {

}

func (tms *TestMuxServer) AddJobs(jobs *Jobs) error {
	if jobs.queueName == "TestQueue" {
		return nil
	} else {
		return errors.New("TestError")
	}
}

func (tms *TestMuxServer) GetJobs(qn QueueName, tn TopicName, typ QueueTyp) (*Jobs, error) {
	if qn == "TestQueue" {
		return &Jobs{}, nil
	} else {
		return &Jobs{}, errors.New("TestError")
	}
}

func (tms *TestMuxServer) AddQueue(conf *QueueConfig) error {
	return nil
}

func (tms *TestMuxServer) DelQueue(conf *QueueConfig) error {
	return nil
}

func (tms *TestMuxServer) UpdateQueue(conf *QueueConfig) error {
	return nil
}

func (tms *TestMuxServer) AuthQueue(qn QueueName, tn TopicName, pwd string) error {
	return nil
}

func (tms *TestMuxServer) GetMuxer(qn QueueName, tn TopicName) (Muxer, error) {
	return &Mux{}, nil
}

func (tms *TestMuxServer) UpdateConfigs(map[QueueName]*QueueConfig) error {
	return nil
}

// @todo gateway have an implicit reliance on QueueConfs needs to refactor
func init() {
	conf := QueueConfs()
	conf["TestQueue"] = &QueueConfig{
		QueueName:      "TestQueue",
		Password:       "password",
		ConsumerEnable: true,
		Topic:          map[TopicName]*TopicConfig{"default": {Password: "password"}},
	}
	conf["TestError"] = &QueueConfig{
		QueueName:      "TestError",
		Password:       "password",
		ConsumerEnable: true,
		Topic:          map[TopicName]*TopicConfig{"error": {Password: "password"}},
	}
	QueueConfsStorage.Store(conf)
	tMuxServer := NewTestMuxServer()
	// @todo 测试的端口与本地运行的端口分开  12017容易被本地运行端口冲突
	gw := NewGateway("127.0.0.1", "12017", tMuxServer)
	go func() {
		err := gw.ListenAndServe()
		panic(err)
	}()
	time.Sleep(time.Second)
}

func TestGatewayPING(t *testing.T) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestQueue:password", 2, 2, 2)
	c := redisClient.Pop()
	defer redisClient.Push(c)
	res, err := c.Call("PING")
	if err != nil {
		t.Fatalf("Ping Case want err = nil returns which decided by gateway_test, but receive res = %v and err = %v", res, err)
	} else {
		t.Log(string(res.([]byte)), err)
	}

}

func TestGatewayLPUSH(t *testing.T) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestQueue:password", 2, 2, 2)
	c := redisClient.Pop()
	defer redisClient.Push(c)
	args := []interface{}{"TestQueue/default", "valuelpush1", "valuelpush2"}
	res, err := c.CallN(msgRedis.RetryTimes, "LPUSH", args...)
	resp := &gwResp{}
	json.Unmarshal(res.([]uint8), resp)
	if resp.Error == "" {
		t.Log(string(res.([]uint8)), err)
	} else {
		t.Fatalf("LPUSH Case want res.Error = '' but receive res = %+v", resp)
	}
}

func TestGatewayLPUSHErrorCase(t *testing.T) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestError:password", 2, 2, 2)
	c := redisClient.Pop()
	defer redisClient.Push(c)

	args := []interface{}{"TestError/error", "valuelpush1", "valuelpush2"}
	res, err := c.CallN(msgRedis.RetryTimes, "LPUSH", args...)

	resp := &gwResp{}
	json.Unmarshal(res.([]uint8), resp)

	if resp.Error == "TestError" {
		t.Log(resp, err)
	} else {
		t.Fatalf("LPUSH Error Case want err = TestError returns which decided by gateway_test, but receive res = %v and err = %v", string(res.([]uint8)), err)
	}
}

func TestGatewayLPUSHAuthErrorCase(t *testing.T) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestQueue:password", 2, 2, 2)
	c := redisClient.Pop()
	defer redisClient.Push(c)
	args := []interface{}{"TestError/error", "valuelpush1", "valuelpush2"}
	res, err := c.CallN(msgRedis.RetryTimes, "LPUSH", args...)

	resp := &gwResp{}
	json.Unmarshal(res.([]uint8), resp)

	if resp.Error == "NOAUTH Authentication required." {
		t.Log(res, err)
	} else {
		t.Fatalf("LPUSH Auth Error Case want err = TestError returns which decided by gateway_test, but receive res = %v and err = %v", string(res.([]uint8)), err)
	}
}

func TestGatewayRPOP(t *testing.T) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestQueue/default:password", 2, 2, 2)
	c := redisClient.Pop()
	defer redisClient.Push(c)
	res, err := c.RPOP("TestQueue/default")
	t.Log(string(res), err)
}

func TestGatewayRPOPErrorCase(t *testing.T) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestError/error:password", 2, 2, 2)
	c := redisClient.Pop()
	defer redisClient.Push(c)
	res, err := c.RPOP("TestError/error")
	if (string(res)) == "TestError" {
		t.Log(string(res), err)
	} else {
		t.Fatalf("LPOP Error Case want res = TestError returns which decided by gateway_test, but receive res = %s and err = %v", (string(res)), err)
	}
}

func TestGatewayRPOPAuthErrorCase(t *testing.T) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestError:password", 2, 2, 2)
	c := redisClient.Pop()
	defer redisClient.Push(c)
	args := []interface{}{"TestError/error", "valuelpush1", "valuelpush2"}
	res, err := c.CallN(msgRedis.RetryTimes, "RPOP", args...)

	if err != nil && err.Error() == "CommonError:NOAUTH Authentication required." {
		t.Log(res, err)
	} else {
		t.Fatalf("LPUSH Auth Error Case want err = TestError returns which decided by gateway_test, but receive res = %v and err = %v", res, err)
	}
}

func TestGatewayLREM(t *testing.T) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestQueue/default:password", 2, 2, 2)
	c := redisClient.Pop()
	defer redisClient.Push(c)
	res, err := c.LREM("TestQueue/default", 0, "valuelpush1")

	t.Log(res, err)
}

// follow is just for test of benchmark
func BenchmarkGatewayPING(b *testing.B) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestQueue/default:password", 2, 2, 2)
	c := redisClient.Pop()
	redisClient.Push(c)
	for i := 0; i < b.N; i++ {
		c.Call("PING")
	}
}

func BenchmarkPINGParallel(b *testing.B) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestQueue/default:password", 2, 2, 2)
	c := redisClient.Pop()
	redisClient.Push(c)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Call("PING")
		}
	})
}

func BenchmarkGatewayLPUSH(b *testing.B) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestQueue/default:password", 2, 2, 2)
	c := redisClient.Pop()
	redisClient.Push(c)
	for i := 0; i < b.N; i++ {
		c.Call("TestQueue/default", []string{"valuelpush1", "valuelpush2"})
	}
}

func BenchmarkLPUSHParallel(b *testing.B) {
	redisClient := msgRedis.NewPool("127.0.0.1:12017", "TestQueue/default:password", 2, 2, 2)
	c := redisClient.Pop()
	redisClient.Push(c)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Call("TestQueue/default", []string{"valuelpush1", "valuelpush2"})
		}
	})
}
