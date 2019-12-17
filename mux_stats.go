package main

import (
	"sync"
	"time"
)

const (
	// Mux 统计计算间隔
	MuxStatsCalInterval = 60 * time.Second

	// Mux 统计 Channel 大小
	MuxStatsChannelBuff = 10000
)

type (
	muxStatItem struct {
		queue QueueName
		topic TopicName

		// 消费是否成功
		consumerIsSuccess bool
		// 消费耗时
		consumerTC time.Duration
	}
	MuxStatsData struct {
		// 统计间隔
		TimePeriod time.Duration
		// 执行成功次数
		SuccessCount int64
		// 执行失败次数
		FailCount int64
		// 最长耗时
		MaxTime time.Duration
		// 最短耗时
		MinTime time.Duration
		// 总耗时
		TotalTime time.Duration
	}
	MuxStats struct {
		data map[QueueName]map[TopicName]*MuxStatsData
		cur  map[QueueName]map[TopicName]*MuxStatsData

		itemChan chan *muxStatItem
		stop     chan bool
		sync.RWMutex
	}
)

func NewMuxStats() *MuxStats {
	return &MuxStats{
		data:     make(map[QueueName]map[TopicName]*MuxStatsData),
		cur:      make(map[QueueName]map[TopicName]*MuxStatsData),
		itemChan: make(chan *muxStatItem, MuxStatsChannelBuff),
		stop:     make(chan bool, 1),
	}
}

func (s *MuxStats) Run() {
	go func() {
		ticker := time.NewTicker(MuxStatsCalInterval)
		defer ticker.Stop()
		for {
			select {
			case <-s.stop:
				return
			case <-ticker.C:
				s.Lock()
				s.data = s.cur
				s.cur = make(map[QueueName]map[TopicName]*MuxStatsData)
				s.Unlock()
			}
		}
	}()

	for {
		select {
		case <-s.stop:
			return
		case item := <-s.itemChan:
			s.Lock()
			if _, ok := s.cur[item.queue]; !ok {
				s.cur[item.queue] = make(map[TopicName]*MuxStatsData)
			}
			if _, ok := s.cur[item.queue][item.topic]; !ok {
				s.cur[item.queue][item.topic] = &MuxStatsData{}
			}
			if item.consumerIsSuccess {
				s.cur[item.queue][item.topic].SuccessCount++
			} else {
				s.cur[item.queue][item.topic].FailCount++
			}
			s.cur[item.queue][item.topic].TotalTime += item.consumerTC
			if item.consumerTC > s.cur[item.queue][item.topic].MaxTime {
				s.cur[item.queue][item.topic].MaxTime = item.consumerTC
			} else if s.cur[item.queue][item.topic].MinTime == 0 || item.consumerTC < s.cur[item.queue][item.topic].MinTime {
				s.cur[item.queue][item.topic].MinTime = item.consumerTC
			}
			s.Unlock()

			// add external stats
			addJobStats(string(item.queue), string(item.topic), item.consumerTC, item.consumerIsSuccess)
		}
	}
}

func (s *MuxStats) Add(queue QueueName, topic TopicName, consumerIsSuccess bool, consumerTC time.Duration) {
	if queue == "" || topic == "" {
		return
	}

	item := &muxStatItem{
		queue:             queue,
		topic:             topic,
		consumerIsSuccess: consumerIsSuccess,
		consumerTC:        consumerTC,
	}

	select {
	case s.itemChan <- item:
	default:
		flog.Error("MuxStats channel is full")
	}
}

func (s *MuxStats) Output() map[QueueName]map[TopicName]*MuxStatsData {
	s.RLock()
	defer s.RUnlock()
	return s.data
}

func (s *MuxStats) Stop() {
	close(s.stop)
}
