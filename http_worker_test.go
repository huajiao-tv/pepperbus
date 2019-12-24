package main

import (
	"context"
	"strings"
	"testing"
	"time"
)

var httpContext, _ = func() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}()

var httpMux = NewHttpWorker(httpContext, "TestQueue", "TestTopic", 20, &TestStorage{}, &TestSuccessRemote{})

func TestHttpMuxRun(t *testing.T) {
	// todo, have not find an available method for test running go proc, so just test the code is running as normal
	httpMux.Serve()
	t.Log("no news")
}

func TestHttpMuxClose(t *testing.T) {
	// todo, have not find an available method for test running go proc, so just test the code is running as normal
	httpMux.Close()
	t.Log("no news")
}

func TestHttpMuxAddJobs(t *testing.T) {
	err := httpMux.AddJobs(&Jobs{queueName: "TestQueue"}, 1)
	if err != nil {
		t.Fatalf("AddJobs Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestHttpMuxAddJobsErrorCase(t *testing.T) {
	err := httpMux.AddJobs(&Jobs{queueName: "TestError"}, 1)
	if err == nil {
		t.Fatalf("AddJobs Case want err = TestError returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestHttpMuxGetJobsFromQueue(t *testing.T) {
	_, err := httpMux.GetJobsFromQueue("TestQueue", "TestTopic", 1)
	if err != nil {
		t.Fatalf("GetJobsFromQueue Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestHttpMuxGetJobsFromQueueErrorCase(t *testing.T) {
	_, err := httpMux.GetJobsFromQueue("TestError", "TestTopic", 1)
	if err == nil {
		t.Fatalf("GetJobsFromQueue Case want err = TestError returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestHttpMuxSendJobs(t *testing.T) {
	resp := httpMux.SendJobs(&Jobs{queueName: "TestQueue"})
	if resp.Code != "200" {
		t.Fatalf("SendJobs Case want resp.Code = nil returns, but receive resp = %v", resp)
	} else {
		t.Log(resp)
	}
}

func TestHttpMuxSendJobsErrorCase(t *testing.T) {
	resp := httpMux.SendJobs(&Jobs{queueName: "TestError"})
	if resp.Msg != "TestError" {
		t.Fatalf("SendJobs Case want resp.Msg = TestError returns, but receive resp = %v", resp)
	} else {
		t.Log(resp)
	}
}

func TestHttpMuxSendJobsFailedCase(t *testing.T) {
	resp := httpMux.SendJobs(&Jobs{queueName: "TestFailed"})
	if resp.Msg != "TestFailed" {
		t.Fatalf("SendJobs Case want resp.Msg = TestFailed returns, but receive resp = %v", resp)
	} else {
		t.Log(resp)
	}
}

func TestHttpMuxSendToRetryQueue(t *testing.T) {
	err := httpMux.sendToRetryQueue(&Jobs{queueName: "TestQueue", outTraceId: "1"})
	if err != nil {
		t.Fatalf("SendToRetryQueue Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestHttpMuxSendToRetryQueueErrorCase(t *testing.T) {
	err := httpMux.sendToRetryQueue(&Jobs{queueName: "TestError", outTraceId: "1"})
	if err != nil && err.Error() == "TestError" {
		t.Log(err)
	} else {
		t.Fatalf("SendToRetryQueue Error Case want err = TestError returns, but receive err = %v", err)
	}
}

func TestHttpLoopPing(t *testing.T) {
	var sigChan = make(chan string, 1)
	go func() {
		httpMux.loopPing()
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

func TestHttpSendPing(t *testing.T) {
	err := httpMux.SendPing()
	if err != nil {
		t.Fatalf("SendPing Case want err = nil returns, but receive err = %v", err)
	} else {
		t.Log(err)
	}
}

func TestHttpSendPingError(t *testing.T) {
	httpMux.RemoteSender = &TestErrorRemote{}
	err := httpMux.SendPing()
	if err != nil && err.Error() == "TestError" {
		t.Log(err)
	} else {
		t.Fatalf("SendPing Case want err = nil returns, but receive err = %v", err)
	}
}

func TestHttpSendPingFailed(t *testing.T) {
	httpMux.RemoteSender = &TestFailedRemote{}
	err := httpMux.SendPing()
	if err != nil && strings.Contains(err.Error(), "TestFailed") {
		t.Log(err)
	} else {
		t.Fatalf("SendPing Case want err = nil returns, but receive err = %v", err)
	}
}

func TestHttpPingTopicNotRegisted(t *testing.T) {
	httpMux.RemoteSender = &TestTopicNotRegistedRemote{}
	err := httpMux.SendPing()
	if err != nil && err.Error() == "topic no register: TestQueue/TestTopic response is [Queue/TopicNotRegisted]" {
		t.Log(err)
	} else {
		t.Fatalf("SendPing Case want err = 'topic no register: TestQueue/TestTopic' returns, but receive err = %v", err)
	}
}
