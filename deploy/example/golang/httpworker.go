package main

import (
	"flag"
	"net/http"
	"time"

	"fmt"

	"encoding/json"
	"io/ioutil"
	"net/url"


	"strings"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"git.huajiao.com/qmessenger/pepperbus/sdk/golang"
)

var (
	address = flag.String("l", ":18000", "Zero address to bind to.")
	now     = time.Now()
)

func main() {
	ra := busworker.NewWorkerRouter("/your/path/a")
	ra.RegisterJobHandler("queue1", "topic1", TestSuccess)
	ra.RegisterJobHandler("queue1", "topic2", TestSuccess)
	ra.RegisterJobHandler("queue1", "topic3", TestFail)

	rb := busworker.NewWorkerRouter("/your/path/b")
	rb.RegisterJobHandler("queue1", "topic1", TestSuccess)

	fmt.Println(busworker.GetRouter())
	fmt.Printf("listen on: %s \n", *address)
	go busworker.Serve(*address)

	// success case
	time.Sleep(time.Millisecond)
	sendTestRequest("/your/path/a", paramsQueue1Topic1)
	sendTestRequest("/your/path/a", paramsQueue1Topic2)
	sendTestRequest("/your/path/b", paramsQueue1Topic1)

	// execute fail case
	sendTestRequest("/your/path/a", paramsQueue1Topic3)

	// not found case
	sendTestRequest("/your/path/b", paramsQueue1Topic2)
	sendTestRequest("/not/job/request", paramsQueue1Topic1)

	// ping case
	sendTestRequest("/your/path/a", paramsPing)
	sendTestRequest("/your/path/b", paramsPing)

}

// Code is ExecSuccess to tell pepperbus exec success
func TestSuccess(r *http.Request) *busworker.Resp {
	r.ParseForm()
	fmt.Println("TestSuccess receive form: ", r.Form)
	resp := &busworker.Resp{
		Code:      busworker.ExecSuccess,
		ErrorCode: 0,
		Data: struct {
			FieldsA string
			FieldsB string
		}{"TestSuccess", "success"},
	}
	return resp
}

// Code is not ExecSuccess pepperbus will move this job to retry queue
// error code for pepperbus to trace detail type of error
func TestFail(r *http.Request) *busworker.Resp {
	r.ParseForm()
	fmt.Println("TestFail receive form: ", r.Form)
	resp := busworker.Resp{
		Code:      busworker.ExecFail,
		ErrorCode: 123,
		Data: struct {
			FieldsA string
			FieldsB string
		}{"TestFail", "fail"},
	}
	return &resp
}

type Job struct {
	Id        string `json:"id"`
	InTraceId string `json:"inTraceId"`
	Content   string `json:"content"`
	Retry     int    `json:"retrying"`
}

var jobs = func() string {
	jbs, _ := json.Marshal([]*Job{{"123", "xxxxx", "{json:json}", 1}})
	return string(jbs)
}()

var paramsQueue1Topic1 = url.Values{
	"type":                  []string{"data"},
	"outTraceId":            []string{"xxxxx"},
	"jobs":                  []string{jobs},
	"queue":                 []string{"queue1"},
	"topic":                 []string{"topic1"},
	"IS_PEPPER_BUS_REQUEST": []string{"true"},
}
var paramsQueue1Topic2 = url.Values{
	"type":                  []string{"data"},
	"outTraceId":            []string{"xxxxx"},
	"jobs":                  []string{jobs},
	"queue":                 []string{"queue1"},
	"topic":                 []string{"topic2"},
	"IS_PEPPER_BUS_REQUEST": []string{"true"},
}

var paramsQueue1Topic3 = url.Values{
	"type":                  []string{"data"},
	"outTraceId":            []string{"xxxxx"},
	"jobs":                  []string{jobs},
	"queue":                 []string{"queue1"},
	"topic":                 []string{"topic3"},
	"IS_PEPPER_BUS_REQUEST": []string{"true"},
}

var paramsPing = url.Values{
	"type":                  []string{"ping"},
	"IS_PEPPER_BUS_REQUEST": []string{"true"},
}

func sendTestRequest(path string, params url.Values) {
	fmt.Printf("Test: %s type: %s \n", path, params.Get("type"))

	request, err := http.NewRequest("Post", "http://127.0.0.1:18000"+path, strings.NewReader(params.Encode()))
	if err != nil {
		log.Error(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("IS-PEPPER-BUS-REQUEST", "true")
	r, err := http.DefaultClient.Do(request)

	if err != nil {
		log.Error(err)
		return
	}
	defer r.Body.Close()
	code := r.Header.Get("Code")
	errCode := r.Header.Get("Error-Code")
	fmt.Printf("code %s errCode %s Header %#v \n", code, errCode, r.Header)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read all error", err.Error())
	}
	fmt.Println(string(body) + "\n")
}
