package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type (
	QueueName string
	TopicName string

	QueueConfig struct {
		QueueName QueueName                  `json:"queue_name"`
		Password  string                     `json:"password"`
		Topic     map[TopicName]*TopicConfig `json:"topic"`
	}

	ServerConfig struct {
		GatewayIP   string
		GatewayPort string
		AdminIP     string
		AdminPort   string
	}

	StorageConfig struct {
		Type           string `json:"type"`
		Address        string `json:"address"`
		Auth           string `json:"auth"`
		MaxConnNum     int    `json:"max_conn_num"`
		MaxIdleNum     int    `json:"max_idle_num"`
		MaxIdleSeconds int64  `json:"max_idle_seconds"`
	}

	// topic: name:offset
	TopicConfig struct {
		Name           string         `json:"name"`
		Password       string         `json:"password"`
		RetryTimes     int            `json:"retry_times"`
		MaxQueueLength int            `json:"max_queue_length"`
		RunType        int            `json:"run_type"`
		Storage        *StorageConfig `json:"storage"`
		StartFrom      int            `json:"-"`
		Status         int            `json:"-"`
		*CgiConfig
		*WorkerConfig
	}

	CgiConfig struct {
		Address        string        `json:"address"`
		ConnectTimeout time.Duration `json:"connect_timeout"`
		ReadTimeout    time.Duration `json:"read_timeout"`
		WriteTimeout   time.Duration `json:"write_timeout"`
		ScriptEntry    string        `json:"script_entry"`
	}

	WorkerConfig struct {
		Script      string `json:"script"`
		Entry       string `json:"entry"`
		MaxChildren int    `json:"max_children"`
		MaxTask     int    `json:"max_task"`
	}
)

func main() {
	storageConfig := &StorageConfig{
		Type:           "redis",
		Address:        "10.208.255.222:5814",
		MaxConnNum:     30,
		MaxIdleNum:     10,
		MaxIdleSeconds: 300,
	}

	topicConfigs := make(map[TopicName]*TopicConfig)
	topicConfig := TopicConfig{
		Name:           "BenTestTopic0",
		Password:       "783ab0ce",
		RetryTimes:     3,
		MaxQueueLength: 1000,
		Storage:        storageConfig,
		RunType:        0,
		CgiConfig: &CgiConfig{
			Address:        "127.0.0.1:9000",
			ConnectTimeout: 3 * time.Second,
			ReadTimeout:    3 * time.Second,
			WriteTimeout:   3 * time.Second,
			ScriptEntry:    "/data/example/php/consume.php",
		},
	}
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("BenTestTopic%d", i)
		topicConfig.Name = name
		t := topicConfig
		topicConfigs[TopicName(name)] = &t
	}

	queueConfigs := make(map[string]*QueueConfig)
	queueConfig := QueueConfig{
		QueueName: QueueName("queue1"),
		Password:  "783ab0ce",
		Topic:     topicConfigs,
	}
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("BenTestQueue%d", i)
		queueConfig.QueueName = QueueName(name)
		q := queueConfig
		queueConfigs[name] = &q
	}

	data, err := json.Marshal(queueConfigs)
	if err != nil {
		panic(err)
	}

	fmt.Println("config string = ", string(data))
	fmt.Println("tags_mapping string = {}\n[1.1.1.1]\ntags []string =\nshard_id int = 1")
}
