package main

/*
gokeeper config ---- queue.conf

config string = """{"test500Queue":{"queue_name":"test500Queue","password":"783ab0ce","topic":{"topic":{"name":"topic","password":"783ab0ce","retry_times":3,"max_queue_length":1000,"run_type":0,"storage":{"type":"redis","address":"127.0.0.1:6380","auth":"","max_conn_num":30,"max_idle_num":10,"max_idle_seconds":300},"num_of_workers":10,"script_entry":"/tmp/consume.php","cgi_config_key":"local"}}},"testSuccessQueue":{"queue_name":"testSuccessQueue","password":"783ab0ce","topic":{"topic":{"name":"topic","password":"783ab0ce","retry_times":3,"max_queue_length":1000,"run_type":0,"storage":{"type":"redis","address":"127.0.0.1:6380","auth":"","max_conn_num":30,"max_idle_num":10,"max_idle_seconds":300},"num_of_workers":10,"script_entry":"/tmp/consume.php","cgi_config_key":"local"}}},"testUnsubscribedQueue":{"queue_name":"testUnsubscribedQueue","password":"783ab0ce","topic":{"topic":{"name":"topic","password":"783ab0ce","retry_times":3,"max_queue_length":1000,"run_type":0,"storage":{"type":"redis","address":"127.0.0.1:6380","auth":"","max_conn_num":30,"max_idle_num":10,"max_idle_seconds":300},"num_of_workers":10,"script_entry":"/tmp/consume.php","cgi_config_key":"local"}}}}"""

tags_mapping string = """{"live": {"testSuccessQueue": [],"testUnsubscribedQueue": [],"test500Queue":[]}}"""

# cgi 全局配置
cgi_config string = """{"local": {"address":"127.0.0.1:9000","connect_timeout":3000000000,"read_timeout":3000000000,"write_timeout":3000000000,"init_conn_num":10,"max_conn_num":20,"reuse_counts":100,"reuse_life":30000000000}}"""

[127.0.0.1:12017]
tags []string = live
shard_id int = 1

*/

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

var (
	domain       string
	testcase     string
	gatewayAddr  string
	gatewayAdmin string
	redisAddr    string
	jobExecAddr  string
	sshUser      string
	sshPassword  string
	logPath      string
	jobCount     int
	local        bool

	httpClient *http.Client
)

const NormalValue = "Normal"
const RestartValue = "Restart"
const GatewayRestartPath = "/manage/service?op=restart"
const GatewayStopPath = "/manage/service?op=stop"

type gatewayResp struct {
	Jobs      []string `json:"jobs"`
	InTraceId string   `json:"inTraceId"`
	Error     string   `json:"error"`
}

type Test interface {
	Run() error
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&domain, "domain", "help", "gokeeper domain")
	flag.StringVar(&testcase, "tc", "help", "test case name")
	flag.StringVar(&gatewayAddr, "ghost", "127.0.0.1:12017", "gateway( in redis protcol ) addr")
	flag.StringVar(&gatewayAdmin, "gadmin", "127.0.0.1:19840", "gateway admin addr")
	flag.StringVar(&jobExecAddr, "jhost", "127.0.0.1:6380", "job exec redis addr")
	flag.StringVar(&sshPassword, "pass", "", "password of ssh for remote result checking")
	flag.StringVar(&redisAddr, "rhost", "127.0.0.1:6379", "redis addr")
	flag.StringVar(&sshUser, "user", "", "user of ssh for remote result checking")
	flag.StringVar(&logPath, "lpath", "/data/pepperbus/log/pepperbus_trace.log", "log path")
	flag.IntVar(&jobCount, "jobCount", 100, "jobCount for ServiceRestart test or Cover Test")
	flag.BoolVar(&local, "local", false, "local flag")

	flag.Parse()

	httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(network, addr, time.Duration(1000)*time.Millisecond)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost: 100,
		},
		Timeout: time.Duration(60000) * time.Millisecond,
	}
}

func main() {
	var t Test

	switch testcase {
	case "normal":
		t = NewNormalTest(1)
	case "nNormal":
		t = NewNormalTest(jobCount)
	case "topicUnsubscribe":
		t = NewTopicUnsubscribedTest()
	case "jobException":
		t = NewJobExceptionTest()
	case "serviceRestart":
		t = NewServiceRestartTest(jobCount)
	case "badJob":
		t = NewBadJobTest()
	case "topicAddDel":
		t = NewTopicAddDelTest()
	case "phpEof":
		t = NewPHPEofTest(jobCount * 10)
	default:
		fmt.Println(usage())
		return
	}

	switch testcase {
	case "serviceRestart", "topicAddDel", "phpEof":
		if gatewayAddr != "127.0.0.1:12017" || gatewayAdmin != "127.0.0.1:19840" {
			fmt.Println("------------------------------------------------------------------------------\n")
			fmt.Println("Test case topicAddDel serviceRestart and phpEof can only run in local !!!!!\n")
			fmt.Println("------------------------------------------------------------------------------")
			return
		}
	default:
	}

	err := t.Run()
	fmt.Println("---------------------------------------------")
	if err != nil {
		fmt.Println(testcase, " case run FAILED! Error: ", err)
	} else {
		fmt.Println(testcase, " case run SUCCESS!")
	}
	fmt.Println("---------------------------------------------")

}

func usage() string {
	return `
funcTest is a test tool for functions of pepperbus

Usage:
	funcTest -tc testcase -ghost "127.0.0.1:12017" -rhost "127.0.0.1:6379" -lpath "/data/pepperbus/log/pepperbus_trace.log"

Parameters Description:
	-tc 	necessary 	must in values of help normal topicNotSubscribe jobException serviceRestart and jobEncodeError 
	-user	necessary	user of ssh for checking remote result in file
	-pass	necessary	passpord of ssh for checking remote result in file
	-ghost	unnecessary	gateway address default "127.0.0.1:12017"
	-gadmin unnecessary	gateway admin address default "127.0.0.1:19840"
	-rhost	unnecessary	redis address default "127.0.0.1:6379"
	-jhost	unnecessary	job exec address default "127.0.0.1"
	-lpath	unnecessary	redis address default "/data/pepperbus/log/pepperbus_trace.log"

The test cases are:
	normal			Normal function test of pepperbus
	nNormal			N times normal function test of pepperbus
	topicUnsubscribe	Push a job to a not sbuscribed topic, then pepperbus will only log it
	jobException		Push a normal job to pepperbus, but cgi return exceptions, then pepperbus will put it to retry queue
	serviceRestart		Restart pepperbus, the service will finish all doing jobs then stop
	badJob			Push a bad constructed job to pepperbus, then pepperbus will only log it
	topicAddDel		After topic add and del options, the service must be work correctly
	phpEof			When php eof occurs, jobs will still run correctly and EOF log occurs

	`
}

type Jar struct {
	cookies []*http.Cookie
}

func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies = cookies
}
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies
}

func httpPost(uri string, values url.Values, response interface{}) error {
	jar := &Jar{
		cookies: []*http.Cookie{&http.Cookie{Name: "manage", Value: "true"}},
	}

	httpClient.Jar = jar

	resp, err := httpClient.PostForm(uri, values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("%+v", string(body))
	if err != nil {
		return err
	}
	if response == nil {
		return nil
	}
	if err = json.Unmarshal(body, response); err != nil {
		return err
	}

	return nil
}
