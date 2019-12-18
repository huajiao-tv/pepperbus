package main

import (
	"encoding/json"
	"time"
)

type (
	QueueName string
	TopicName string

	// QueueConfig 所有的队列配置
	QueueConfig struct {
		QueueName      QueueName                  `json:"queue_name"`
		Password       string                     `json:"password"`
		Topic          map[TopicName]*TopicConfig `json:"topic"`
		ConsumerEnable bool                       `json:"-"`
	}

	// StorageConfig 存储相关配置
	StorageConfig struct {
		Type           string `json:"type"`
		Address        string `json:"address"`
		Auth           string `json:"auth"`
		MaxConnNum     int    `json:"max_conn_num"`
		MaxIdleNum     int    `json:"max_idle_num"`
		MaxIdleSeconds int64  `json:"max_idle_seconds"`
	}

	// TopicConfig 具体消费实例相关配置
	TopicConfig struct {
		Name           string         `json:"name"`
		Password       string         `json:"password"`
		RetryTimes     int            `json:"retry_times"`
		MaxQueueLength int            `json:"max_queue_length"`
		RunType        int            `json:"run_type"`
		Storage        *StorageConfig `json:"storage"`
		StartFrom      int            `json:"-"`
		Status         int            `json:"-"`
		ConsumerEnable bool           `json:"-"`
		NumOfWorkers   uint64         `json:"num_of_workers"`
		ScriptEntry    string         `json:"script_entry"`
		CgiConfigKey   string         `json:"cgi_config_key"`
		*WorkerConfig
	}

	// CgiConfig CGI 的所有相关全局配置，Topic 只做指向
	CgiConfig struct {
		Address        string        `json:"address"`
		ConnectTimeout time.Duration `json:"connect_timeout"`
		ReadTimeout    time.Duration `json:"read_timeout"`
		WriteTimeout   time.Duration `json:"write_timeout"`
		InitConnNum    uint64        `json:"init_conn_num"`
		MaxConnNum     uint64        `json:"max_conn_num"`
		ReuseCounts    uint64        `json:"reuse_counts"`
		ReuseLife      time.Duration `json:"reuse_life"`
	}

	// WorkerConfig 当运行类型为 Worker 时，会用这个配置
	WorkerConfig struct {
		Script      string `json:"script"`
		Entry       string `json:"entry"`
		MaxChildren int    `json:"max_children"`
		MaxTask     int    `json:"max_task"`
	}
)

func main() {
	// redisAddress := "127.0.0.1:6379"
	redisAddress := "pepperbus_redis:6379"

	storageConfig := &StorageConfig{
		Type:           "redis",
		Address:        redisAddress,
		MaxConnNum:     30,
		MaxIdleNum:     10,
		MaxIdleSeconds: 300,
	}

	topic1Config := &TopicConfig{
		Name:           "topic1",
		Password:       "783ab0ce",
		RetryTimes:     3,
		MaxQueueLength: 1000,
		Storage:        storageConfig,
		RunType:        0,
		NumOfWorkers:   5,
		CgiConfigKey:   "local",
		ScriptEntry:    "/data/example/php/consume.php",
		// ScriptEntry: "/Users/specode/go/src/github.com/huajiao-tv/pepperbus/example/php/consume.php",
	}

	topic2Config := &TopicConfig{
		Name:           "topic2",
		Password:       "783ab0ce",
		RetryTimes:     3,
		MaxQueueLength: 1000,
		Storage:        storageConfig,
		RunType:        1,
		NumOfWorkers:   10,
		CgiConfigKey:   "local",
		ScriptEntry:    "http://127.0.0.1:18080/addJobs",
	}

	topicConfig := map[TopicName]*TopicConfig{
		"topic1": topic1Config,
		"topic2": topic2Config,
	}

	queueConfig := map[string]*QueueConfig{
		"queue1": {
			QueueName: QueueName("queue1"),
			Password:  "783ab0ce",
			Topic:     topicConfig,
		},
	}

	data, err := json.Marshal(queueConfig)
	if err != nil {
		panic(err)
	}

	println(string(data))
}
