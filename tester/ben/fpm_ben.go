package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/huajiao-tv/pepperbus/codec/fastcgi"
)

const (
	CgiResponseSuccess  = "200"
	CgiResponseFail     = "400"
	CgiResponseNotFound = "404"
	CgiInnerError       = "500"
	CgiTimeout          = "502"
)

// queue和topic为压测专用，压测配置中配置为：BenTestQueue1~BenTestQueue100，BenTestTopic1~BenTestTopic3，密码均为：BenTest
func bench_test_cgi(gn, n int) {
	for i := 0; i < gn; i++ {
		vs := url.Values{
			"type":       []string{"data"},
			"outTraceId": []string{"traceid"},
		}
		vs.Add("IS_PEPPER_BUS_REQUEST", "true")

		env := make(map[string]string)
		env["SCRIPT_FILENAME"] = "/data/example/php/consume.php"
		//env["SCRIPT_FILENAME"] = "/data/log/test.php"
		timeout := time.Duration(3000000000)

		go func(goid, runnum int) {
			defer wg.Done()

			result := &testResult{MinTime: 1e9}
			for {
				start := time.Now()
				err := sendToCgi(cgiNetwork, cgiAddr, timeout, timeout, timeout, env, vs)

				if err != nil {
					fmt.Println("sendToCgi err is ", err)
					result.fail()
				} else {
					result.success(start)
				}

				if runnum--; runnum == 0 {
					break
				}
			}

			fmt.Println(result.String())
		}(i, n)
	}
}

func sendToCgi(network, address string, connectTimeout, writeTimeout, readTimeout time.Duration, env map[string]string, params url.Values) error {
	var err error
	cgi, err := fastcgi.DialTimeout(network, address, connectTimeout, true)
	if err != nil {
		return err
	}
	// @todo optimize defer
	defer cgi.Close()
	cgi.SetWriteDeadline(writeTimeout)
	cgi.SetReadDeadline(readTimeout)
	resp, err := cgi.PostForm(env, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	code := resp.Header.Get("Code")
	//errCode := resp.Header.Get("Error-Code")
	if code == "" {
		code = CgiInnerError
	}
	//body, err := ioutil.ReadAll(resp.Body)
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}
