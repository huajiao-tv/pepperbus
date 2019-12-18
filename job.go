package main

import (
	"strings"

	"encoding/json"
	"fmt"

	"github.com/CodisLabs/codis/pkg/utils/errors"
	"github.com/huajiao-tv/redeo/resp"
)

type Job struct {
	Id               string `json:"id"`
	InTraceId        string `json:"inTraceId"`
	Content          string `json:"content"`
	Retry            int    `json:"retrying"`
	RetryTimes       int    `json："retry_times"`        //重试次数
	RetryTime        int64  `json："retry_time"`         //重试时间（执行时间）
	IsRetry          bool   `json："is_retry"`           //是否需要重试
	BeginRetry       bool   `json："begin_retry"`        //标记是否已经进入重试机制
	RetryResidueTime int64  `json："retry_residue_time"` //剩余重试时间
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

func (jobs *Jobs) String() string {
	// jobs send from storage to cgi
	if jobs.inTraceId == "" {
		return fmt.Sprintf("%s/%s: outTraceId: %s retrying: %t jobs: %v", jobs.queueName, jobs.topicName, jobs.outTraceId, jobs.retrying, jobs.Slice)
	}
	// jobs from client to gateway
	return fmt.Sprintf("%s/%s: inTraceId: %s retrying: %t jobs: %v", jobs.queueName, jobs.topicName, jobs.inTraceId, jobs.retrying, jobs.Slice)

}

// Parse Job from request
func ParseJob(contents []resp.CommandArgument) (*Jobs, []string, error) {
	var ids []string
	slice := make([]*Job, 0, 10)
	fields := strings.SplitN(string(contents[0]), "/", 2)
	if len(fields) < 2 {
		return nil, nil, errors.New("ParseJob fail: invalid job meta info: " + string(contents[0]))
	}
	inTraceId := GenTraceId()
	qName, _ := QueueName(fields[0]), TopicName(fields[1])
	for _, content := range contents[1:] {
		job := &Job{
			InTraceId: inTraceId,
			Id:        GenId(),
			Content:   string(content),
		}
		slice = append(slice, job)
		ids = append(ids, job.Id)
	}
	return &Jobs{inTraceId, "", qName, "", false, slice}, ids, nil
}
