package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	gokeeper "git.huajiao.com/qmessenger/gokeeper/client"
	"git.huajiao.com/qmessenger/gokeeper/client/discovery"
	"git.huajiao.com/qmessenger/pepperbus/logic/data"
	"github.com/davecgh/go-spew/spew"
	"github.com/johntech-o/idgen"
	"github.com/qmessenger/utility/logger"
)

func init() {
	gokeeper.Debug = false
}

// 定义 Worker 的运行模式
const (
	RunTypeCGI  = 0
	RunTypeHttp = 1
	RunTypeCmd  = 2
)

type (
	// StaticConfigType 静态配置
	StaticConfigType struct {
		LogDir       string
		BackupLogDir string
	}

	// ServerConfig 运行服务的相关配置
	ServerConfig struct {
		GatewayIP   string
		GatewayPort string
		AdminIP     string
		AdminPort   string
	}

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
		HttpConfig     string         `json:"http_config"`
		*WorkerConfig
		IsRetry   bool  `json:"is_retry"`   //topip配置是否重试
		RetryTime int64 `json:"retry_time"` // 重试有效时间
	}

	// CgiConfig CGI 的所有相关全局配置，Topic 只做指向
	CgiConfig struct {
		Address        string        `json:"address"`
		ConnectTimeout time.Duration `json:"connect_timeout"`
		ReadTimeout    time.Duration `json:"read_timeout"`
		WriteTimeout   time.Duration `json:"write_timeout"`
		InitConnNum    int64         `json:"init_conn_num"`
		MaxConnNum     int64         `json:"max_conn_num"`
		ReuseCounts    int64         `json:"reuse_counts"`
		ReuseLife      time.Duration `json:"reuse_life"`
	}

	// WorkerConfig 当运行类型为 Worker 时，会用这个配置
	WorkerConfig struct {
		Script      string `json:"script"`
		Entry       string `json:"entry"`
		MaxChildren int    `json:"max_children"`
		MaxTask     int    `json:"max_task"`
	}

	// HttpConfig 当运行类型为 httpWorker时，会使用这个配置
	HttpConfig struct {
		RequestTimeout  time.Duration `json:"request_timeout"`
		IdleConnTimeout time.Duration `json:"idle_conn_timeout"`
	}

	// TagsMapping 对应 tags 开启的 queue 与 topic 映射关系
	// 当 topic 空时，表示开启全部
	// map[tag][queueName][]string{"topic1", "topic2"}
	TagsMapping map[string]map[string][]string
)

// RootPath 可执行程序所运行的路径
var RootPath = func() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath
}()

// StaticConfig 静态配置
var StaticConfig = &StaticConfigType{
	LogDir:       filepath.Join(RootPath, "log/"),
	BackupLogDir: filepath.Join(RootPath, "log/backup/"),
}

var flog = DefaultLogger()

// QueueConfsStorage 原子替换存储
var QueueConfsStorage atomic.Value

// QueueConfs 全局 Queue 配置
var QueueConfs = func() map[QueueName]*QueueConfig {
	data, ok := QueueConfsStorage.Load().(map[QueueName]*QueueConfig)
	if !ok {
		return make(map[QueueName]*QueueConfig)
	}
	return data
}

// ServerConfStorage 原子替换存储
var ServerConfStorage atomic.Value

// ServerConf 全局 Server 配置
var ServerConf = func() *ServerConfig {
	for {
		conf, ok := ServerConfStorage.Load().(*ServerConfig)
		if !ok {
			flog.Warn("Invalid ServerConfig Waiting for gokeeper response", conf)
			<-time.After(time.Millisecond * 50)
			continue
		}
		return conf
	}
}

// gokeeper
var (
	KeeperAddr    string
	Domain        string
	NodeID        string
	DiscoveryAddr string

	Component     = "main"
	ComponentTags = make(map[string]string)
)

func netGlobalConf() *data.Global {
	return data.CurrentGlobal()
}

func netQueueConf() *data.Queue {
	return data.CurrentQueue()
}

func initDiscovery(addr, domain, nodeID string) {
	instance := discovery.NewInstance(nodeID, domain, map[string]string{"admin": nodeID})
	client := discovery.New(
		addr,
		discovery.WithRegistry(instance),
		discovery.WithRegistryTTL(60*time.Second),
	)
	client.SignalDeregister(true, syscall.SIGINT, syscall.SIGTERM)
	client.Work()
}

func initSetting() error {
	flag.StringVar(&KeeperAddr, "k", "", "keeper address ip:port")
	flag.StringVar(&DiscoveryAddr, "kd", "", "keeper discovery address ip:port")
	flag.StringVar(&Domain, "d", "", "domain name")
	flag.StringVar(&NodeID, "n", "", "current node id")
	flag.Parse()

	// 日志包
	os.MkdirAll(StaticConfig.BackupLogDir, os.ModePerm)
	filename := filepath.Join(StaticConfig.LogDir, fmt.Sprintf("%s-%s", "pepperbus", NodeID))
	l, err := logger.NewLogger(filename, "pepperbus", StaticConfig.BackupLogDir)
	if err != nil {
		panic(err)
	}
	flog = l

	// 分割 Domain 和 子目录配置
	var sections []string
	domainSplit := strings.SplitN(Domain, "/", 2)
	if len(domainSplit) != 2 {
		sections = []string{
			"/global.conf",
			"/queue.conf",
			"/queue.conf/" + NodeID,
		}
	} else {
		sections = []string{
			domainSplit[1] + "/global.conf",
			domainSplit[1] + "/queue.conf",
			domainSplit[1] + "/queue.conf/" + NodeID,
		}
	}

	domain := domainSplit[0]

	keeperCli := gokeeper.New(KeeperAddr, domain, NodeID, Component, sections, ComponentTags)
	keeperCli.LoadData(data.ObjectsContainer).RegisterCallback(UpdateSetting)
	if err := keeperCli.Work(); err != nil {
		return err
	}

	// CGI 连接池初始化
	cgiConfs, err := parseCgiTransportConfig()
	if err != nil {
		return err
	}
	CgiPoolManager.Init(cgiConfs)

	// CGI 错误收集
	errorCollectionConfig, err := parseErrorCollection()
	if err != nil {
		return err
	}
	CGIErrorCollection = NewErrorCollection(errorCollectionConfig)

	// CGI 任务调试
	taskDebugConfig, err := parseTaskDebug()
	if err != nil {
		return err
	}
	CGITaskDebug = NewTaskDebug(taskDebugConfig)

	updateVersionAndShardID()

	HttpConfigManager, err = parseHttpTransportConfig()
	if err != nil {
		return err
	}

	// 服务发现
	initDiscovery(DiscoveryAddr, Domain, NodeID)
	return nil
}

func parseCgiTransportConfig() (map[string]*CgiConfig, error) {
	configs := make(map[string]*CgiConfig)
	flog.Warn("debug parseCgiTransportConfig: ", netQueueConf().CgiConfig)
	if err := json.Unmarshal([]byte(netQueueConf().CgiConfig), &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

func parseHttpTransportConfig() (*HttpConfig, error) {
	config := &HttpConfig{}
	flog.Warn("debug parseHttpTransportConfig: ", netQueueConf().HttpConfig)
	if netQueueConf().HttpConfig == "" {
		return config, nil
	}
	err := json.Unmarshal([]byte(netQueueConf().HttpConfig), &config)
	if err != nil {
		flog.Error("error http config init failed", err.Error())
		return nil, err
	}
	return config, nil

}

func parseQueueConfig() (map[QueueName]*QueueConfig, error) {
	configs := make(map[QueueName]*QueueConfig)
	flog.Warn("debug parseQueueConfig: ", netQueueConf().QueueConfig)
	err := json.Unmarshal([]byte(netQueueConf().QueueConfig), &configs)
	if err != nil {
		return nil, err
	}

	tagsMapping := make(TagsMapping)
	flog.Warn("debug parseTagsMapping: ", netQueueConf().TagsMapping)
	err = json.Unmarshal([]byte(netQueueConf().TagsMapping), &tagsMapping)
	if err != nil {
		return nil, err
	}

	// 取该机器需要运行的 tags
	tags := netQueueConf().Tags
	flog.Warn("debug tags: ", tags)

	for _, tag := range tags {
		if tagMapping, ok := tagsMapping[tag]; !ok {
			continue
		} else {
			// 循环标记需要运行的 queue 和 topic
			for queue, topicList := range tagMapping {
				queueName := QueueName(queue)
				// 当 queue 下 topic 为空时，认为该 queue 下的所有 topic 都为开启状态
				if len(topicList) == 0 {
					if _, ok := configs[queueName]; ok {
						configs[queueName].ConsumerEnable = true
					}
					continue
				}

				// 为了防止 tags 中指定的 queueName 不存在，这里加个判断
				if _, ok := configs[queueName]; !ok {
					continue
				}

				// 循环标记需要运行的 topic
				for _, topic := range topicList {
					topicName := TopicName(topic)

					if _, ok := configs[queueName].Topic[topicName]; ok {
						configs[queueName].Topic[topicName].ConsumerEnable = true
					}
				}
			}
		}
	}

	flog.Warn("debug finally configs: ", spew.Sdump(configs))
	return configs, err
}

func parseErrorCollection() (*StorageConfig, error) {
	config := &StorageConfig{}
	flog.Warn("debug parseErrorCollection: ", netQueueConf().ErrorCollectionStorage)
	if netQueueConf().ErrorCollectionStorage == "" {
		return config, nil
	}

	err := json.Unmarshal([]byte(netQueueConf().ErrorCollectionStorage), &config)
	if err != nil {
		flog.Error("error collection config init failed", err.Error())
		return nil, err
	}

	return config, nil
}

func parseTaskDebug() (*StorageConfig, error) {
	config := &StorageConfig{}
	flog.Warn("debug parseTaskDebug: ", netQueueConf().TaskDebugStorage)
	if netQueueConf().TaskDebugStorage == "" {
		return config, nil
	}

	err := json.Unmarshal([]byte(netQueueConf().TaskDebugStorage), &config)
	if err != nil {
		flog.Error("task debug config init failed", err.Error())
		return nil, err
	}

	return config, nil
}

func updateServer() {
	c := &ServerConfig{
		GatewayIP:   netGlobalConf().GatewayIp,
		GatewayPort: netGlobalConf().GatewayPort,
		AdminIP:     netGlobalConf().AdminIp,
		AdminPort:   netGlobalConf().AdminPort,
	}
	ServerConfStorage.Store(c)
}

func updateQueue() {
	c, err := parseQueueConfig()
	if err != nil {
		panic("load queue config failed! error:" + err.Error())
	}
	QueueConfsStorage.Store(c)
}

func updateVersionAndShardID() {
	if (idgen.SetVersion(1) != nil) || (idgen.SetShardId(1) != nil) {
		panic("idgen init failed!")
	}
}

// UpdateSetting Gokeeper 更新调用的回调方法
func UpdateSetting() {
	updateServer()
	updateQueue()
	updateServiceId()
	MuxManager.UpdateConfigs(QueueConfs())
}

func DefaultLogger() *logger.Logger {
	os.MkdirAll("./log/backup", os.ModePerm)
	l, err := logger.NewLogger("./log/pepperbus", "pepperbus", "./log/backup/")
	if err != nil {
		fmt.Println("Use Default Logger Error: ", err)
	}
	return l
}
