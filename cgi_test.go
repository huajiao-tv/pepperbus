package main

import (
	//"crypto/md5"
	"fmt"
	//"io"
	"net"
	"net/http"
	"net/http/fcgi"
	"net/url"
	"time"
)

// cgi server for test
type FastCGIServer struct{}

var ipPort = "127.0.0.1:59000"

func (s FastCGIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(100000000)

	queue := req.FormValue("queue")

	if queue == "TestQueue" {
		resp.Header().Set("Code", "200")
		fmt.Fprintln(resp, "TestQueue")
	} else {
		resp.Header().Set("Code", "400")
		fmt.Fprintln(resp, "TestErrorQueue")
	}
}

// TestCgiClients for test inject
type TestCgiClients struct {
}

func (tcc *TestCgiClients) sendToCgi(env map[string]string, params url.Values) (*WorkerResponse, error) {
	return &WorkerResponse{}, nil
}

// test proc
func init() {
	go func() {
		listener, err := net.Listen("tcp", ipPort)
		if err != nil {
			// handle error
			fmt.Println("listener creatation failed: ", err)
		}

		srv := new(FastCGIServer)
		fcgi.Serve(listener, srv)
	}()

	time.Sleep(time.Microsecond * 50)
	fmt.Println("===========================================")
	fmt.Println("===========================================")
	fmt.Println("===========================================")
	fmt.Println()
	fmt.Println("------------ FAKE CGI UNIT TEST -----------")
	fmt.Println()
	fmt.Println("===========================================")
	fmt.Println("===========================================")
	fmt.Println("===========================================")
}

// func TestSendToCGIBuffer(t *testing.T) {
// 	// no need to test
// 	// TestCgiManager := NewCgiManager(&CgiConfig{})
// 	// TestCgiManager.clients = &TestCgiClients{}
// 	// TestCgiManager.SendToCGIBuffer(&Jobs{})
// 	//t.Log("no bad news is good news")
// 	t.Log("no need to test")
// }
//
// func TestSendToCgi(t *testing.T) {
// 	env := make(map[string]string)
// 	env["REQUEST_METHOD"] = "GET"
// 	env["SERVER_PROTOCOL"] = "HTTP/1.1"
//
// 	params := url.Values{
// 		"queue": []string{string("TestQueue")},
// 		"topic": []string{string("TestTopic")},
// 	}
//
// 	clients := NewCgiClients(ipPort, 3*time.Second, 3*time.Second, 3*time.Second)
//
// 	res, err := clients.sendToCgi(env, params)
//
// 	if err != nil {
// 		t.Fatalf("AddJobs Case want err = nil returns, but receive err = %v", err)
// 	} else if strings.TrimRight(res.Msg, "\n") == "TestQueue" {
// 		t.Log(res, err)
// 	} else {
// 		t.Fatalf("AddJobs Case want r.Msg = TestQueue returns, but receive r = %v", res)
// 	}
// }
//
// func TestSendToCgiTimeoutCase(t *testing.T) {
// 	env := make(map[string]string)
// 	env["REQUEST_METHOD"] = "GET"
// 	env["SERVER_PROTOCOL"] = "HTTP/1.1"
//
// 	params := url.Values{
// 		"queue": []string{string("TestTimeoutQueue")},
// 		"topic": []string{string("TestTimeoutTopic")},
// 	}
//
// 	clients := NewCgiClients(ipPort, time.Nanosecond, time.Nanosecond, time.Nanosecond)
//
// 	res, err := clients.sendToCgi(env, params)
// 	if err.Error() == "dial tcp 127.0.0.1:59000: i/o timeout" {
// 		t.Log(res, err)
// 	} else {
// 		t.Fatalf("AddJobs Case want err = 'dial tcp 127.0.0.1:59000: i/o timeout' returns, but receive r = %v", res)
// 	}
// }
//
// func TestSendToCgiErrorCase(t *testing.T) {
// 	env := make(map[string]string)
// 	env["REQUEST_METHOD"] = "GET"
// 	env["SERVER_PROTOCOL"] = "HTTP/1.1"
//
// 	params := url.Values{
// 		"queue": []string{string("TestErrorQueue")},
// 		"topic": []string{string("TestErrorTopic")},
// 	}
//
// 	clients := NewCgiClients(ipPort, 3*time.Second, 3*time.Second, 3*time.Second)
//
// 	res, err := clients.sendToCgi(env, params)
//
// 	if strings.TrimRight(res.Msg, "\n") == "TestErrorQueue" {
// 		t.Log(res, err)
// 	} else {
// 		t.Fatalf("AddJobs Case want r.Msg = TestErrorQueue returns, but receive r = %v", res)
// 	}
// }
//
// // todo, have not finish the test case: cgi.PostForm error and ioutil.ReadAll error in cgi.go
