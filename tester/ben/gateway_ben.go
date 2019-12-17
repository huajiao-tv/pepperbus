package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/qmessenger/utility/msgRedis"
)

type testResult struct {
	MinTime        time.Duration
	MaxTime        time.Duration
	TotalTime      time.Duration
	TotalRequests  int64
	FailedRequests int64
}

func (this *testResult) success(start time.Time) {
	cost := time.Now().Sub(start)
	this.TotalTime += cost
	this.TotalRequests++
	if cost > this.MaxTime {
		this.MaxTime = cost
	} else if cost < this.MinTime {
		this.MinTime = cost
	}
}

func (this *testResult) fail() {
	this.FailedRequests++
	this.TotalRequests++
}

func (this testResult) String() string {
	if this.TotalRequests != 0 {
		return fmt.Sprintf("Request result: min:%fs, max:%fs, average:%fs, requests:%d(failed:%d)",
			this.MinTime.Seconds(), this.MaxTime.Seconds(), this.TotalTime.Seconds()/float64(this.TotalRequests), this.TotalRequests, this.FailedRequests)
	} else {
		return "Error: not run"
	}
}

// queue和topic为压测专用，压测配置中配置为：BenTestQueue1~BenTestQueue100，BenTestTopic1~BenTestTopic3，密码均为：BenTest
func bench_test_gateway(gn, n int, tc string) {
	for i := 0; i < gn; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			result := &testResult{MinTime: 1e9}
			for {
				qIdx := rand.Intn(benTestQueueCount)
				tIdx := rand.Intn(benTestTopicCount)
				c := redisClients[qIdx].Pop()
				queueTopic := fmt.Sprintf("%s%d/%s%d", queueNamePrefix, qIdx, topicNamePrefix, tIdx)

				start := time.Now()
				err := error(nil)
				switch tc {
				case "lpush":
					args := []interface{}{queueTopic, "valuelpush1", "valuelpush2"}
					_, err = c.CallN(msgRedis.RetryTimes, "LPUSH", args...)
				case "rpop":
					_, err = c.RPOP(queueTopic)
				case "lrem":
					_, err = c.LREM(queueTopic, 1, "valuelpush1")
				case "ping":
					_, err = c.Call("PING")
				default:
					err = nil
				}

				redisClients[qIdx].Push(c)

				if err != nil {
					fmt.Println(tc, "err is ", err)
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
