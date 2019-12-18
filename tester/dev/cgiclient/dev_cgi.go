package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/url"
	"strconv"

	"time"

	"strings"

	"fmt"

	"git.huajiao.com/qmessenger/pepperbus/codec/fastcgi"
	"github.com/CodisLabs/codis/pkg/utils/log"
)

const (
	RequestTypePing = "ping"
	RequestTypeData = "data"

	CgiResponseSuccess  = "200"
	CgiResponseFail     = "400"
	CgiResponseNotFound = "404"
	CgiInnerError       = "500"
	CgiTimeout          = "502"
)

var (
	TimeOutSecond string
	address       string
	scriptFile    string
	queueTopic    string
)

type CgiResponse struct {
	Code      string
	ErrorCode string
	Msg       string
}

type Job struct {
	Id        string `json:"id"`
	InTraceId string `json:"inTraceId"`
	Content   string `json:"content"`
	Retry     int    `json:"retrying"`
}

func (j Job) String() string {
	s, e := json.Marshal(j)
	if e != nil {
		return e.Error()
	}
	return string(s)
}

func (j *Job) Encode() ([]byte, error) {
	b, e := json.Marshal(*j)
	return b, e
}

func (j *Job) Decode(data []byte) error {
	e := json.Unmarshal(data, j)
	return e
}

type QueueName string

type TopicName string

type Jobs struct {
	inTraceId  string
	outTraceId string
	queueName  QueueName
	topicName  TopicName
	retrying   bool
	Slice      []*Job `json:"jobs"`
}

func (jobs *Jobs) Encode() []byte {
	b, e := json.Marshal(*jobs)
	if e != nil {
		// @todo return a empty Jobs struct
		return []byte("{}")
	}
	return b
}

func main() {
	flag.StringVar(&TimeOutSecond, "t", "5s", "timeout duration")
	flag.StringVar(&address, "h", "127.0.0.1:9001", "cgi address")
	flag.StringVar(&scriptFile, "f", "/Users/johntech/code/gocode/src/git.huajiao.com/qmessenger/pepperbus/example/php/consume.php", "cgi script file")
	flag.StringVar(&queueTopic, "qt", "queue1/topic1", "queue name and topic name")
	flag.Parse()

	timeout, err := time.ParseDuration(TimeOutSecond)
	if err != nil {
		panic("invalid input timeout -t")
	}
	fields := strings.Split(queueTopic, "/")
	if len(fields) != 2 {
		panic("invalid qt")
	}
	test := func() {
		env := make(map[string]string)
		env["SCRIPT_FILENAME"] = scriptFile
		env["SERVER_SOFTWARE"] = "go/fastcgi"
		env["REMOTE_ADDR"] = address
		i := 0
		env["key"] = strconv.Itoa(i) //  parse to $SERVER params of php-fpm
		jobs := &Jobs{
			inTraceId:  "111111",
			outTraceId: "222222",
			queueName:  QueueName(fields[0]),
			topicName:  TopicName(fields[1]),
			Slice: []*Job{
				{Id: "123", InTraceId: "111111", Content: "hello world", Retry: 0}},
		}

		jbs, _ := json.Marshal(jobs.Slice)
		vs := url.Values{
			"type":       []string{RequestTypeData},
			"outTraceId": []string{jobs.outTraceId},
			"queue":      []string{string(jobs.queueName)},
			"topic":      []string{string(jobs.topicName)},
			"jobs":       []string{string(jbs)},
		}
		client := NewCgiClients(address, timeout, timeout, timeout)
		res, err := client.sendToCgi(env, vs)
		fmt.Println(res.Msg, err, 1)
		res, err = client.sendToCgi(env, vs)
		fmt.Println(res.Msg, err, 1)
		res, err = client.sendToCgi(env, vs)
		fmt.Println(res.Msg, err, 1)
	}
	go test()
	test()
	time.Sleep(time.Second * 100)

}

// conservative set-up in case mess up read/write deadline
const (
	ConnReuseLife   = time.Second * 5
	ConnReuseCounts = 40
)

type CgiClients struct {
	cgi          *fastcgi.FCGIClient // in case be used in race condition
	exeCounter   int
	lastUsedTime time.Time
	cgiError     error

	address        string
	connectTimeout time.Duration
	writeTimeout   time.Duration
	readTimeout    time.Duration
}

func NewCgiClients(address string, connectTimeout time.Duration, writeTimeout time.Duration, readTimeout time.Duration) *CgiClients {
	cli := CgiClients{
		address:        address,
		connectTimeout: connectTimeout,
		writeTimeout:   writeTimeout,
		readTimeout:    readTimeout,
	}
	return &cli
}

// reuse cgi connection
func (c *CgiClients) getConn() (*fastcgi.FCGIClient, error) {
	var newConn = func() (*fastcgi.FCGIClient, error) {
		var err error
		c.cgi, err = fastcgi.DialTimeout("tcp", c.address, c.connectTimeout, true)
		if err != nil {
			return nil, err
		}
		fmt.Println("connect success")
		c.cgi.SetWriteDeadline(c.writeTimeout + ConnReuseLife)
		c.cgi.SetReadDeadline(c.readTimeout + ConnReuseLife)
		c.lastUsedTime = time.Now()
		c.exeCounter = 1
		return c.cgi, nil
	}
	// first time create a connection
	if c.cgi == nil {
		return newConn()
	}
	// receate a new conn when out of threshold or error happened
	if time.Now().Sub(c.lastUsedTime) >= ConnReuseLife ||
		c.exeCounter > ConnReuseCounts ||
		c.cgiError != nil {
		c.cgi.Close()
		return newConn()
	}
	fmt.Println("reuse connection")
	// reuse connection
	c.lastUsedTime = time.Now()
	c.exeCounter += 1
	return c.cgi, nil
}

// @todo the situation which response body contains php notice return  the notice to dashboard
func (c *CgiClients) sendToCgi(env map[string]string, params url.Values) (*CgiResponse, error) {
	var (
		err error
		cgi *fastcgi.FCGIClient
		cr  *CgiResponse
	)
	params.Add("IS_PEPPER_BUS_REQUEST", "true")
	cgi, err = c.getConn()
	defer func() {
		if err != nil {
			// network error may happen, mark and recreate connection
			c.cgiError = err
		}
	}()
	if err != nil {
		log.Error("cgi connect to fail ", c.address, err.Error())
		return cr, err
	}
	resp, err := cgi.PostForm(env, params)
	if err != nil {
		log.Errorf("request to cgi fail %s %s %+v ", c.address, err.Error(), env)
		return cr, err
	}
	defer resp.Body.Close()
	code := resp.Header.Get("Code")
	errCode := resp.Header.Get("Error-Code")
	if code == "" {
		code = CgiInnerError
	}
	cr = &CgiResponse{code, errCode, ""}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("read all error", err.Error())
		cr.Msg = err.Error()
		return cr, err
	}
	cr.Msg = string(body)
	return cr, nil
}
