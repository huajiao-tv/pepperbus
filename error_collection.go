package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/huajiao-tv/pepperbus/utility/msgRedis"
)

const (
	// ErrorCollectionKeyFMT 定义错误收集存储的 Key 名字
	ErrorCollectionKeyFMT = "EC:%s:%s"
)

// ErrorCollection 收集 CGI 错误信息入存储
type ErrorCollection struct {
	c        *StorageConfig
	redisCli *msgRedis.MultiPool
	address  string
}

// ErrorCollectionData 收集的结果结构
type ErrorCollectionData struct {
	Time     string
	Jobs     string
	SysCode  string
	SysError string
	UserCode string
	UserMsg  string
}

// NewErrorCollection 新建一个收集实例
func NewErrorCollection(c *StorageConfig) *ErrorCollection {
	errorCollection := &ErrorCollection{
		c: c,
	}

	// 初始化 Redis
	switch c.Type {
	case "redis":
		redisAddressWithAuth := fmt.Sprintf("%s:%s", c.Address, c.Auth)
		errorCollection.redisCli = msgRedis.NewMultiPool(
			[]string{redisAddressWithAuth},
			c.MaxConnNum,
			c.MaxIdleNum,
			c.MaxIdleSeconds,
		)
		errorCollection.address = redisAddressWithAuth
	}

	return errorCollection
}

// AddError 记录一条错误，当 Enable 为 False 时，直接返回
//
func (ec *ErrorCollection) AddError(queue QueueName, topic TopicName,
	jobs *Jobs, respCode string, respErrorCode string, respBody string, errMsg error) {

	if !netQueueConf().ErrorCollectionEnable {
		return
	}

	key := fmt.Sprintf(ErrorCollectionKeyFMT, queue, topic)

	tmp := respBody
	hostname, _ := os.Hostname()

	respBody = fmt.Sprintf("%s ------ [Host] %s", tmp, hostname)

	var errorInfo string
	if errMsg != nil {
		errorInfo = fmt.Sprintf("%v,,,,%s,,,,%s,,,,%s,,,,%s,,,,%s", time.Now().Unix(), jobs, respCode, respErrorCode, respBody, errMsg)
	} else {
		errorInfo = fmt.Sprintf("%v,,,,%s,,,,%s,,,,%s,,,,%s", time.Now().Unix(), jobs, respCode, respErrorCode, respBody)
	}
	_, err := ec.redisCli.Call(ec.address).LPUSH(key, []string{errorInfo})
	if err != nil {
		flog.Error("ErrorCollection AddError Failed", err)
	}
}

// GetError 返回存储的错误信息，有条数限制，默认为最新的10条
func (ec *ErrorCollection) GetError(queue QueueName, topic TopicName, length int) []*ErrorCollectionData {
	if !netQueueConf().ErrorCollectionEnable {
		return nil
	}

	if length <= 0 {
		length = 10
	}
	key := fmt.Sprintf(ErrorCollectionKeyFMT, queue, topic)
	tmp, err := ec.redisCli.Call(ec.address).LRANGE(key, 0, length-1)
	if err != nil {
		flog.Error("redisCli GetError Failed ", ec.address, err)
		return nil
	}

	var ret []*ErrorCollectionData
	for _, td := range tmp {
		tb, ok := td.([]byte)
		if !ok {
			continue
		}
		data := string(tb)

		splitData := strings.Split(data, ",,,,")
		splitDataLen := len(splitData)
		if splitDataLen != 6 && splitDataLen != 5 {
			continue
		}

		errorData := &ErrorCollectionData{
			Time:     splitData[0],
			Jobs:     splitData[1],
			SysCode:  splitData[2],
			UserCode: splitData[3],
			UserMsg:  splitData[4],
		}

		if splitDataLen == 6 {
			errorData.SysError = splitData[5]
		}

		ret = append(ret, errorData)
	}

	return ret
}
