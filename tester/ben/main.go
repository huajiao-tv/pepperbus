package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/huajiao-tv/pepperbus/codec/fastcgi"
	"github.com/huajiao-tv/pepperbus/utility/msgRedis"
)

var (
	testcase     string
	gatewayAddr  string
	cgiAddr      string
	cgiNetwork   string
	gonum        int
	num          int
	cgiKeepAlive bool
	wg           *sync.WaitGroup
	help         string

	queueNamePrefix   string
	topicNamePrefix   string
	benTestPasspord   string
	benTestQueueCount int
	benTestTopicCount int

	redisClients map[int]*msgRedis.Pool

	// conns map[int]*fastcgi.FCGIClient
	conns chan *fastcgi.FCGIClient
)

const CONN_POOL_SIZE = 128 // cgi

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&help, "help", "", "help test case")
	flag.StringVar(&testcase, "tc", "help", "test case name")
	flag.StringVar(&gatewayAddr, "ghost", "127.0.0.1:12017", "gateway( in redis protcol ) addr")
	flag.StringVar(&cgiNetwork, "cnetwork", "tcp", "cgi network(tcp | unix)")
	flag.StringVar(&cgiAddr, "chost", "127.0.0.1:9001", "cgi addr")
	flag.IntVar(&num, "num", 1, "test num every goroutinue")
	flag.IntVar(&gonum, "gonum", runtime.NumCPU(), "goroutinue number")
	flag.BoolVar(&cgiKeepAlive, "cgikeep", false, "cgi keepalive flag, default false")

	flag.Parse()

	// queue和topic为压测专用，压测配置中配置为：BenTestQueue0~BenTestQueue99，BenTestTopic0~BenTestTopic2，密码均为：BenTest
	queueNamePrefix = "BenTestQueue"
	topicNamePrefix = "BenTestTopic"
	benTestPasspord = "783ab0ce"
	benTestQueueCount = 40
	benTestTopicCount = 1

	if testcase == "lpush" || testcase == "rpop" || testcase == "lrem" || testcase == "ping" {
		redisClients = make(map[int]*msgRedis.Pool)
		for i := 0; i < benTestQueueCount; i++ {
			auth := fmt.Sprintf("%s%d:%s", queueNamePrefix, i, benTestPasspord)
			redisClients[i] = msgRedis.NewPool(gatewayAddr, auth, 300, 20, 20)
		}
	}

}

func main() {
	rand.Seed(time.Now().UnixNano())

	if help != "" {
		fmt.Println(testcase_help(help))
		return
	}

	wg = new(sync.WaitGroup)
	wg.Add(gonum)

	begin := time.Now()

	switch testcase {
	// for gateway_ben
	case "lpush":
		bench_test_gateway(gonum, num, "lpush")
	case "rpop":
		bench_test_gateway(gonum, num, "rpop")
	case "lrem":
		bench_test_gateway(gonum, num, "lrem")
	case "ping":
		bench_test_gateway(gonum, num, "ping")
	case "cgi":
		bench_test_cgi(gonum, num)

	// default for help
	case "help":
		fallthrough
	default:
		fmt.Println(usage())
		return
	}

	wg.Wait()
	end := time.Now()

	duration := end.Sub(begin)
	sec := duration.Seconds()
	fmt.Println(float64(gonum*num)/sec, sec)

	// 清理
	if testcase == "cgi" && cgiKeepAlive {
		// for _, conn := range conns {
		// 	if conn != nil {
		// 		conn.Close()
		// 	}
		// }
		for i := 0; i < CONN_POOL_SIZE; i++ {
			select {
			case conn, ok := <-conns:
				if !ok {
					return
				}
				conn.Close()
			}
		}
	}
}

func usage() string {
	return `
ben is a benchmark tool for pepperbus 

Usage:
	ben -tc testcase [arguments] [-ghost "127.0.0.1:12017"] [-chost "127.0.0.1:9000"] [-gonum 24] [-num 1]

The test cases are:
	lpush        benchmarks lpush operation capability of specified gateway
	rpop         benchmarks rpop operation capability of specified gateway
	lrem         benchmarks lrem operation capability of specified gateway
	ping         benchmarks ping operation capability of specified gateway
	cgi          benchmarks cgi operation capability of specified php-fpm

Use "ben -help [test case]" for more information about a test case.
	`
}

func testcase_help(tc string) string {
	switch tc {
	case "lpush", "rpop", "lrem", "ping":
		return `
This test case test ` + tc + ` operation capability of specified gateway

Usage:
	ben -tc ` + tc + ` [-ghost "127.0.0.1:12017"] [-gonum 24] [-num 1]

More information of arguments:
	host        specified gateway address in tcp protcol, must include correct port
	              default value is "127.0.0.1:12017"
	gonum       the go routine number, default is the number of cpu
	num         test num each go routinue, default is 1
		`
	default:
		return "\nwrong test case \n" + usage()
	}
}
