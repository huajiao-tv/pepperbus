package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/qmessenger/utility/msgRedis"
)

type (
	storage struct {
		redisCli *msgRedis.MultiPool
		address  string
	}
)

const MuxConsumeDuration = time.Second

type storager interface {
	GetJobsFromQueue(qn QueueName, tn TopicName, typ QueueTyp) (*Jobs, error)
	AddQueue(jobs *Jobs, typ QueueTyp) error
	RetryQueue(qn QueueName, tn TopicName, count int) (int64, error)
	CleanQueue(qn QueueName, tn TopicName, typ QueueTyp) (int64, error)
	GetLength(qn QueueName, tn TopicName) (int64, int64, int64, error)
}

// @todo redis lib support callbyDefaultAddress which do not need to pass a address param
func NewStorage(config *StorageConfig) *storage {
	s := new(storage)
	switch config.Type {
	case "redis":
		redisAddressWithAuth := fmt.Sprintf("%s:%s", config.Address, config.Auth)
		s.redisCli = msgRedis.NewMultiPool(
			[]string{redisAddressWithAuth},
			config.MaxConnNum,
			config.MaxIdleNum,
			config.MaxIdleSeconds,
		)
		s.address = redisAddressWithAuth
	}
	return s
}

var GetKey = func(qn QueueName, tn TopicName, typ QueueTyp) string {
	switch {
	case typ.IsRetryQueue() == true:
		return fmt.Sprintf("%s::%s::%s", qn, tn, "list_retry")
	case typ.IsTimeoutQueue() == true:
		return fmt.Sprintf("%s::%s::%s", qn, tn, "list_timeout")
	default:
		return fmt.Sprintf("%s::%s::%s", qn, tn, "list")
	}
}

//
func (s *storage) AddQueue(jobs *Jobs, typ QueueTyp) error {
	var (
		e  error
		vs []string // value slice of job
	)
	key := GetKey(jobs.queueName, jobs.topicName, typ)
	if typ.IsRetryQueue() {
		jobs.retrying = true
	}
	// @todo redis maybe nil catch this condition
	for _, job := range jobs.Slice {
		if jobs.retrying {
			job.Retry += 1
		}
		// @todo incr fail or add fail then retrying add the data in a few seconds
		v, err := job.Encode()
		if err != nil {
			flog.Error("AddQueue fail: " + err.Error())
			continue
		}
		vs = append(vs, string(v))

	}
	if len(vs) == 0 {
		return errors.New("empty jobs")

	}
	_, e = s.redisCli.Call(s.address).LPUSH(key, vs)
	if e != nil {
		flog.Error("AddQueue failed: ", vs, e)
	}
	return nil
}

func (s *storage) RetryQueue(qn QueueName, tn TopicName, count int) (n int64, err error) {
	srcKey := GetKey(qn, tn, RetryQueue)
	destKey := GetKey(qn, tn, CommonQueue)
	for {
		_, err = s.redisCli.Call(s.address).RPOPLPUSH(srcKey, destKey)
		if err == msgRedis.ErrKeyNotExist {
			return n, nil
		} else if err != nil {
			return n, err
		}

		n++
		if count > 0 && n >= int64(count) {
			return n, nil
		}
	}
}

func (s *storage) CleanQueue(qn QueueName, tn TopicName, typ QueueTyp) (len int64, err error) {
	key := GetKey(qn, tn, typ)
	len, err = s.redisCli.Call(s.address).LLEN(key)
	if err != nil {
		return
	}
	_, err = s.redisCli.Call(s.address).DEL(key)
	return
}

func (s *storage) GetLength(qn QueueName, tn TopicName) (normal int64, retry int64, timeout int64, err error) {
	normalKey := GetKey(qn, tn, CommonQueue)
	normal, err = s.redisCli.Call(s.address).LLEN(normalKey)
	if err != nil {
		return
	}
	retryKey := GetKey(qn, tn, RetryQueue)
	retry, err = s.redisCli.Call(s.address).LLEN(retryKey)

	timeoutKey := GetKey(qn, tn, TimeoutQueue)
	timeout, err = s.redisCli.Call(s.address).LLEN(timeoutKey)
	return
}

func (s *storage) GetJobsFromQueue(q QueueName, c TopicName, typ QueueTyp) (*Jobs, error) {
	jobs := &Jobs{"", GenTraceId(), q, c, false, []*Job{}}
	k := GetKey(jobs.queueName, jobs.topicName, typ)
	content, err := s.redisCli.Call(s.address).RPOP(k)
	if err == msgRedis.ErrKeyNotExist {
		return nil, nil
	} else if err != nil {
		flog.Error("GetJobsFromQueue key", k, "err", err.Error())
		return nil, err
	}
	job := new(Job)
	job.Decode(content)
	jobs.Slice = append(jobs.Slice, job)
	return jobs, nil
}
