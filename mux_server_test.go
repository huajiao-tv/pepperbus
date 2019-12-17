package main

import (
	"context"
	"errors"
	"testing"
)

// TestMux for test inject
type TestMux struct{}

func NewTestMux() *TestMux {
	return &TestMux{}
}

func (tm *TestMux) Close() {
}

func (tm *TestMux) Serve() {
}

func (tm *TestMux) Storage() storager {
	return nil
}

func (tm *TestMux) SendPing() error {
	return nil
}

func (tm *TestMux) SendJobs(jobs *Jobs) *WorkerResponse {
	return nil
}

func (tm *TestMux) AddJobs(jobs *Jobs, typ QueueTyp) error {
	if jobs.queueName == "TestQueue" {
		return nil
	} else {
		return errors.New("TestError")
	}
}

func (tm *TestMux) GetJobsFromQueue(q QueueName, t TopicName, typ QueueTyp) (*Jobs, error) {
	if q == "TestQueue" {
		return &Jobs{}, nil
	} else {
		return &Jobs{}, errors.New("TestError")
	}
}

func (tm *TestMux) Stats() *MuxStats {
	return nil
}

// test proc
var TestMuxManager = NewMuxServer(context.Background())

func init() {
	go TestMuxManager.Serve()
}

func getConf() *QueueConfig {
	topic := make(map[TopicName]*TopicConfig)
	topic["TestTopic"] = &TopicConfig{
		Name:           "TestTopic",
		RetryTimes:     3,
		MaxQueueLength: 100,
		StartFrom:      0,
		Status:         0,
	}

	return &QueueConfig{
		QueueName: "TestQueue0",
		//Topic:     topic, empty for no mux serve run
	}
}

func getConf1() *QueueConfig {
	return &QueueConfig{
		QueueName: "TestQueue1",
	}
}

func TestMuxServerAddQueue(t *testing.T) {
	conf := getConf()

	err := TestMuxManager.AddQueue(conf)
	if err != nil {
		t.Fatalf("AddQueue Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxServerQueueStatus(t *testing.T) {
	//err := TestMuxManager.QueueStatus(conf)
	//if err != nil {
	//	t.Fatalf("QueueStatus Case want err = nil returns, but receive err = %v", err)
	//} else {
	//	t.Log(err)
	//}
}

func TestMuxServerAddJobs(t *testing.T) {
	TestMuxManager.Lock()
	TestMuxManager.mux["TestQueue"] = make(map[TopicName]Muxer)
	TestMuxManager.mux["TestQueue"]["TestTopic"] = NewTestMux()
	TestMuxManager.Unlock()

	err := TestMuxManager.AddJobs(&Jobs{queueName: "TestQueue"})

	if err != nil {
		t.Fatalf("AddJobs Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxServerAddJobsErrorCaseNoQueue(t *testing.T) {
	err := TestMuxManager.AddJobs(&Jobs{queueName: "NoQueue"})

	if err == nil {
		t.Fatalf("AddJobs Error Case NoQueue want err = 'queue or Topic not exist NoQueue' returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxServerAddJobsErrorCaseNoTopic(t *testing.T) {
	TestMuxManager.Lock()
	TestMuxManager.mux["NoTopicQueue"] = make(map[TopicName]Muxer)
	TestMuxManager.Unlock()
	err := TestMuxManager.AddJobs(&Jobs{queueName: "NoTopicQueue"})

	if err == nil {
		t.Fatalf("AddJobs Error Case NoQueue want err = 'queue or Topic not exist NoTopicQueue' returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxServerAddJobsErrorCaseFailed(t *testing.T) {
	TestMuxManager.Lock()
	TestMuxManager.mux["ErrorQueue"] = make(map[TopicName]Muxer)
	TestMuxManager.mux["ErrorQueue"]["TestTopic"] = NewTestMux()
	TestMuxManager.Unlock()

	err := TestMuxManager.AddJobs(&Jobs{queueName: "ErrorQueue"})

	if err == nil {
		t.Fatalf("AddJobs Case want err = TestError returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxServerGetJobs(t *testing.T) {
	TestMuxManager.Lock()
	TestMuxManager.mux["TestQueue"] = make(map[TopicName]Muxer)
	TestMuxManager.mux["TestQueue"]["TestTopic"] = NewTestMux()
	TestMuxManager.Unlock()
	jobs, err := TestMuxManager.GetJobs("TestQueue", "TestTopic", 0)

	if err != nil {
		t.Fatalf("GetJobs Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(jobs, err)
	}
}

func TestMuxServerGetJobsNoQueueCase(t *testing.T) {
	jobs, err := TestMuxManager.GetJobs("NoQueue", "TestTopic", 0)

	if err != nil && err.Error() == "queue not exist NoQueue" {
		t.Log(jobs, err)
	} else {
		t.Fatalf("GetJobs No Queue Case want err = 'queue not exist NoQueue' returns, but receive err = %v", err)
	}
}

func TestMuxServerGetJobsNoTopicCase(t *testing.T) {
	TestMuxManager.Lock()
	TestMuxManager.mux["TestQueue"] = make(map[TopicName]Muxer)
	TestMuxManager.mux["TestQueue"]["TestTopic"] = NewTestMux()
	TestMuxManager.Unlock()
	jobs, err := TestMuxManager.GetJobs("TestQueue", "NoTopic", 0)

	if err != nil && err.Error() == "topic not exist NoTopic" {
		t.Log(jobs, err)
	} else {
		t.Fatalf("GetJobs NoTopic Case want err = 'queue not exist NoTopic' returns, but receive err = %v", err)
	}
}

func TestMuxServerGetMuxer(t *testing.T) {
	TestMuxManager.Lock()
	TestMuxManager.mux["TestQueue"] = make(map[TopicName]Muxer)
	TestMuxManager.mux["TestQueue"]["TestTopic"] = NewTestMux()
	TestMuxManager.Unlock()
	muxer, err := TestMuxManager.GetMuxer("TestQueue", "TestTopic")

	if err != nil {
		t.Fatalf("GetMuxer Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(muxer, err)
	}
}

func TestMuxServerGetMuxerNoQueueCase(t *testing.T) {
	muxer, err := TestMuxManager.GetMuxer("NoQueue", "TestTopic")

	if err != nil && err.Error() == "queue not exist NoQueue" {
		t.Log(muxer, err)
	} else {
		t.Fatalf("GetMuxer No Queue Case want err = 'queue not exist NoQueue' returns, but receive err = %v", err)
	}
}

func TestMuxServerGetMuxerNoTopicCase(t *testing.T) {
	TestMuxManager.Lock()
	TestMuxManager.mux["TestQueue"] = make(map[TopicName]Muxer)
	TestMuxManager.mux["TestQueue"]["TestTopic"] = NewTestMux()
	TestMuxManager.Unlock()
	muxer, err := TestMuxManager.GetMuxer("TestQueue", "NoTopic")

	if err != nil && err.Error() == "topic not exist NoTopic" {
		t.Log(muxer, err)
	} else {
		t.Fatalf("GetMuxer NoTopic Case want err = 'queue not exist NoTopic' returns, but receive err = %v", err)
	}
}

func TestMuxServerUpdateQueue(t *testing.T) {
	conf := getConf1()

	err := TestMuxManager.UpdateQueue(conf)
	if err != nil {
		t.Fatalf("UpdateQueue Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxServerDelQueue(t *testing.T) {
	conf := getConf1()

	err := TestMuxManager.DelQueue(conf)
	if err != nil {
		t.Fatalf("DelQueue Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxServerUpdateConfigs(t *testing.T) {
	conf := make(map[QueueName]*QueueConfig)
	conf["TestQueue1"] = getConf1()

	err := TestMuxManager.UpdateConfigs(conf)
	if err != nil {
		t.Fatalf("UpdateConfigs Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}
