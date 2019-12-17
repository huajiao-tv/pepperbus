package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

var testContext, _ = func() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}()

// TestStorage and TestSuccessRemote for test inject
type TestSuccessRemote struct {
}

func (tc *TestSuccessRemote) SendPing() (*WorkerResponse, error) {
	return &WorkerResponse{Code: ResponseSuccess, Msg: "TestQueue/TestTopic"}, nil
}

func (tc *TestSuccessRemote) SendJobs(jobs *Jobs) (*WorkerResponse, error) {
	if jobs.queueName == "TestQueue" {
		return &WorkerResponse{Code: ResponseSuccess}, nil
	} else if jobs.queueName == "TestError" {
		return &WorkerResponse{}, errors.New("TestError")
	} else {
		return &WorkerResponse{Code: RemoteResponseFail, Msg: "TestFailed"}, nil
	}
}

type TestErrorRemote struct {
}

func (tc *TestErrorRemote) SendPing() (*WorkerResponse, error) {
	return &WorkerResponse{}, errors.New("TestError")
}

func (tc *TestErrorRemote) SendJobs(jobs *Jobs) (*WorkerResponse, error) {
	return &WorkerResponse{Code: ResponseSuccess}, nil
}

type TestFailedRemote struct {
}

func (tc *TestFailedRemote) SendPing() (*WorkerResponse, error) {
	return &WorkerResponse{Code: RemoteResponseFail, Msg: "TestFailed"}, nil
}

func (tc *TestFailedRemote) SendJobs(jobs *Jobs) (*WorkerResponse, error) {
	return &WorkerResponse{Code: ResponseSuccess}, nil
}

type TestTopicNotRegistedRemote struct {
}

func (tc *TestTopicNotRegistedRemote) SendPing() (*WorkerResponse, error) {
	return &WorkerResponse{Code: ResponseSuccess, Msg: "Queue/TopicNotRegisted"}, nil
}

func (tc *TestTopicNotRegistedRemote) SendJobs(jobs *Jobs) (*WorkerResponse, error) {
	return &WorkerResponse{Code: ResponseSuccess}, nil
}

type TestStorage struct {
}

func (ts *TestStorage) GetJobsFromQueue(qn QueueName, tn TopicName, typ QueueTyp) (*Jobs, error) {
	if qn == "TestQueue" {
		return &Jobs{}, nil
	} else {
		return &Jobs{}, errors.New("TestError")
	}
}

func (ts *TestStorage) AddQueue(jobs *Jobs, typ QueueTyp) error {
	if jobs.queueName == "TestQueue" {
		return nil
	} else {
		return errors.New("TestError")
	}
}

func (ts *TestStorage) RetryQueue(qn QueueName, tn TopicName, count int) (int64, error) {
	return 1, nil
}

func (ts *TestStorage) CleanQueue(qn QueueName, tn TopicName, typ QueueTyp) (int64, error) {
	return 1, nil
}

func (ts *TestStorage) GetLength(qn QueueName, tn TopicName) (int64, int64, error) {
	return 1, 1, nil
}

// test proc
var testMux = NewMux(testContext, "TestQueue", "TestTopic", &TestSuccessRemote{}, &TestStorage{}, 10)

func TestMuxServerRun(t *testing.T) {
	// todo, have not find an available method for test running go proc, so just test the code is running as normal
	testMux.Serve()
	t.Log("no bad news is good news")
}

func TestMuxServerClose(t *testing.T) {
	// todo, have not find an available method for test running go proc, so just test the code is running as normal
	testMux.Close()
	t.Log("no bad news is good news")
}

func TestMuxAddJobs(t *testing.T) {
	err := testMux.AddJobs(&Jobs{queueName: "TestQueue"}, 1)
	if err != nil {
		t.Fatalf("AddJobs Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxAddJobsErrorCase(t *testing.T) {
	err := testMux.AddJobs(&Jobs{queueName: "TestError"}, 1)
	if err == nil {
		t.Fatalf("AddJobs Case want err = TestError returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxGetJobsFromQueue(t *testing.T) {
	_, err := testMux.GetJobsFromQueue("TestQueue", "TestTopic", 1)
	if err != nil {
		t.Fatalf("GetJobsFromQueue Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxGetJobsFromQueueErrorCase(t *testing.T) {
	_, err := testMux.GetJobsFromQueue("TestError", "TestTopic", 1)
	if err == nil {
		t.Fatalf("GetJobsFromQueue Case want err = TestError returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxSendJobs(t *testing.T) {
	resp := testMux.SendJobs(&Jobs{queueName: "TestQueue"})
	if resp.Code != "200" {
		t.Fatalf("SendJobs Case want resp.Code = nil returns, but receive resp = %v", resp)
	} else {
		t.Log(resp)
	}
}

func TestMuxSendJobsErrorCase(t *testing.T) {
	resp := testMux.SendJobs(&Jobs{queueName: "TestError"})
	if resp.Msg != "TestError" {
		t.Fatalf("SendJobs Case want resp.Msg = TestError returns, but receive resp = %v", resp)
	} else {
		t.Log(resp)
	}
}

func TestMuxSendJobsFailedCase(t *testing.T) {
	resp := testMux.SendJobs(&Jobs{queueName: "TestFailed"})
	if resp.Msg != "TestFailed" {
		t.Fatalf("SendJobs Case want resp.Msg = TestFailed returns, but receive resp = %v", resp)
	} else {
		t.Log(resp)
	}
}

func TestMuxSendToRetryQueue(t *testing.T) {
	err := testMux.sendToRetryQueue(&Jobs{queueName: "TestQueue", outTraceId: "1"})
	if err != nil {
		t.Fatalf("SendToRetryQueue Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestMuxSendToRetryQueueErrorCase(t *testing.T) {
	err := testMux.sendToRetryQueue(&Jobs{queueName: "TestError", outTraceId: "1"})
	if err != nil && err.Error() == "TestError" {
		t.Log(err)
	} else {
		t.Fatalf("SendToRetryQueue Error Case want err = TestError returns, but receive err = %v", err)
	}
}

func TestLoopPingCGI(t *testing.T) {
	var sigChan = make(chan string, 1)
	go func() {
		testMux.loopPingCGI()
		close(sigChan)
	}()

	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case _, ok := <-sigChan:
			// channel关闭，RobotGiftPromoteBehaviour停止
			if !ok {
				t.Log(nil)
				return
			}
		case <-timer.C:
			t.Fatalf("LoopPingCGI case should be finished in 1 second, but it haven't finish now!!!")
			return
		}
	}
}

func TestSendPing(t *testing.T) {
	err := testMux.SendPing()
	if err != nil {
		t.Fatalf("SendPing Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestSendPingError(t *testing.T) {
	testMux.RemoteSender = &TestErrorRemote{}
	err := testMux.SendPing()
	if err != nil && err.Error() == "TestError" {
		t.Log(err)
	} else {
		t.Fatalf("SendPing Case want err = nil returns, but receive err = %v", err)
	}
}

func TestSendPingFailed(t *testing.T) {
	testMux.RemoteSender = &TestFailedRemote{}
	err := testMux.SendPing()
	if err != nil && strings.Contains(err.Error(), "TestFailed") {
		t.Log(err)
	} else {
		t.Fatalf("SendPing Case want err = nil returns, but receive err = %v", err)
	}
}

func TestSendPingTopicNotRegisted(t *testing.T) {
	testMux.RemoteSender = &TestTopicNotRegistedRemote{}
	err := testMux.SendPing()
	if err != nil && err.Error() == "topic no register: TestQueue/TestTopic" {
		t.Log(err)
	} else {
		t.Fatalf("SendPing Case want err = 'topic no register: TestQueue/TestTopic' returns, but receive err = %v", err)
	}
}
