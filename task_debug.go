package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/qmessenger/utility/msgRedis"
)

const (
	// TaskDebugKeyFMT 定义任务调试存储的 Key 名字
	TaskDebugKeyFMT = "TD:%s"
)

// TaskDebug 收集 Job 信息入存储，只在测试环境信用
type TaskDebug struct {
	c        *StorageConfig
	redisCli *msgRedis.MultiPool
	address  string
}

// TaskDebugData 收集的数据结构
type TaskDebugData struct {
	Time string
	Jobs string
}

// NewTaskDebug 新建一个收集实例
func NewTaskDebug(c *StorageConfig) *TaskDebug {
	td := &TaskDebug{
		c: c,
	}

	// 初始化 Redis
	switch c.Type {
	case "redis":
		redisAddressWithAuth := fmt.Sprintf("%s:%s", c.Address, c.Auth)
		td.redisCli = msgRedis.NewMultiPool(
			[]string{redisAddressWithAuth},
			c.MaxConnNum,
			c.MaxIdleNum,
			c.MaxIdleSeconds,
		)
		td.address = redisAddressWithAuth
	}

	return td
}

// AddTask 记录一条任务到存储
func (td *TaskDebug) AddTask(queue QueueName, task string) {
	if !netQueueConf().TaskDebugEnable {
		return
	}

	key := fmt.Sprintf(TaskDebugKeyFMT, queue)

	info := fmt.Sprintf("%v,,,,%s", time.Now().Unix(), task)
	_, err := td.redisCli.Call(td.address).LPUSH(key, []string{info})
	if err != nil {
		flog.Error("TaskDebug AddTask Failed", err)
	}

	count, _ := td.redisCli.Call(td.address).LLEN(key)
	// 删除无用的历史
	if count >= 100 {
		td.redisCli.Call(td.address).LTRIM(key, 0, 50)
	}
}

// GetTask 返回记录的历史任务
func (td *TaskDebug) GetTask(queue QueueName, length int) []*TaskDebugData {
	if !netQueueConf().TaskDebugEnable {
		return nil
	}

	if length <= 0 {
		length = 10
	}
	key := fmt.Sprintf(TaskDebugKeyFMT, queue)
	tmp, err := td.redisCli.Call(td.address).LRANGE(key, 0, length-1)
	if err != nil {
		return nil
	}

	var ret []*TaskDebugData
	for _, td := range tmp {
		tb, ok := td.([]byte)
		if !ok {
			continue
		}
		data := string(tb)

		splitData := strings.Split(data, ",,,,")
		splitDataLen := len(splitData)
		if splitDataLen != 2 {
			continue
		}

		retData := &TaskDebugData{
			Time: splitData[0],
			Jobs: splitData[1],
		}

		ret = append(ret, retData)
	}

	return ret
}
