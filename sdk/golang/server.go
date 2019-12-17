package busworker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/johntech-o/grace/gracehttp"
)

// resp define response data
type Resp struct {
	ErrorCode int         `json:"error_code"`
	Code      int         `json:"code"`
	Data      interface{} `json:"data"`
}

const (
	ExecSuccess  = 200
	ExecFail     = 500
	ExecNotFound = 404
)

const (
	DefaultTimeout = time.Second * 3600
)

type WorkerHandler func(r *http.Request) *Resp

var Mux = http.NewServeMux()

var Router = make(map[string]*WorkerRouter)

var HttpRouter = make(map[string]http.HandlerFunc)

type WorkerServer struct {
	router map[string]*WorkerRouter
}

func Serve(address string) {
	for pattern, router := range Router {
		func(pattern string, router *WorkerRouter) {
			Mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
				r.ParseForm()
				resp := &Resp{Code: ExecSuccess, ErrorCode: 0, Data: ""}
				var result []byte
				if r.Header.Get("IS-PEPPER-BUS-REQUEST") != "true" {
					resp = &Resp{ErrorCode: ExecFail, Code: ExecFail, Data: "not invalid header"}
					goto final
				}
				switch r.FormValue("type") {
				case "ping":
					fmt.Println("into ping")
					var sl []string
					for k := range router.queueTopicSet {
						sl = append(sl, k)
					}
					w.Header().Set("Consume-Topics", strings.Join(sl, ","))
					s := ""
					for key, _ := range router.queueTopicSet {
						s+=key+","
					}
					result = []byte(s)
				case "data":
					r.ParseForm()
					queue, topic := r.Form.Get("queue"), r.Form.Get("topic")
					if handlerFunc, ok := router.queueTopicSet[queue+"/"+topic]; ok {
						resp = handlerFunc(r)
					} else {
						resp = &Resp{ErrorCode: ExecNotFound, Code: ExecNotFound, Data: ""}
					}
					resu, err := json.Marshal(resp)
					if err != nil {
						resp = &Resp{
							ErrorCode: ExecFail,
							Code:      ExecFail,
						}
						resu, _ = json.Marshal(result)
					}
					result = resu
				}

			final:
				w.Header().Set("Code", strconv.Itoa(resp.Code))
				w.Header().Set("Error-code", strconv.Itoa(resp.ErrorCode))
				w.Write(result)
			})
		}(pattern, router)
	}
	if _, ok := HttpRouter["/"]; !ok {
		RegisterHttpHandler("/", NotFound)
	}
	for pattern, handler := range HttpRouter {
		func(pattern string, handler http.HandlerFunc) {
			Mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
				request.ParseForm()
				if request.FormValue("IS_PEPPER_BUS_REQUEST") == "true" {
					writer.Header().Set("Code", strconv.Itoa(ExecNotFound))
					writer.Header().Set("Error-code", strconv.Itoa(ExecNotFound))
				}
				handler(writer, request)
			})
		}(pattern, handler)
	}
	server := &http.Server{
		ReadTimeout:  DefaultTimeout,
		WriteTimeout: DefaultTimeout,
		IdleTimeout:  DefaultTimeout,
		Addr:         address,
	}
	server.Handler = Mux
	gracehttp.Serve(server)
}

type WorkerRouter struct {
	pattern       string
	queueTopicSet map[string]WorkerHandler
	router        map[string]string
}

func NewWorkerRouter(pattern string) *WorkerRouter {
	w := &WorkerRouter{
		pattern:       pattern,
		queueTopicSet: make(map[string]WorkerHandler),
	}
	Router[pattern] = w
	return w
}

func (wr *WorkerRouter) RegisterJobHandler(queueName, topicName string, handlerFunc WorkerHandler) {
	qt := queueName + "/" + topicName
	if _, ok := wr.queueTopicSet[qt]; ok {
		log.Fatalf("register queue topic Fail: repeat register the same queue %s topic %s", queueName, topicName)
	}
	wr.queueTopicSet[qt] = handlerFunc
	return
}

func RegisterHttpHandler(pattern string, handlerFunc http.HandlerFunc) {
	HttpRouter[pattern] = handlerFunc
}

func GetRouter() map[string]*WorkerRouter {
	return Router
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	result, _ := json.Marshal(Resp{ErrorCode: ExecNotFound, Code: ExecNotFound, Data: ""})
	w.Write(result)
}

func testFunc(r *http.Request) *Resp{
	println("hello world!")
	return &Resp{
		ErrorCode:0,
		Code:200,
		Data:"hello",
	}
}