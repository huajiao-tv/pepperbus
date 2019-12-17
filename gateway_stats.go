package main

import (
	"sync"
	"time"
)

const (
	// Gateway 统计计算间隔
	GatewayStatsCalInterval = time.Minute

	// Gateway 统计 Channel 大小
	GatewayStatsChannelBuff = 10000
)

type (
	gatewayStatItem struct {
		queue   QueueName
		success bool
	}

	GatewayStatsData struct {
		// 统计间隔
		TimePeriod time.Duration
		// 成功次数
		SuccessCount int64
		// 失败次数
		FailCount int64
	}

	GatewayStats struct {
		// 上一分钟数据
		data map[QueueName]*GatewayStatsData
		// 当前数据，需要加锁访问
		cur map[QueueName]*GatewayStatsData

		itemChan chan *gatewayStatItem
		stop     chan bool
		sync.RWMutex
	}
)

func NewGatewayStats() *GatewayStats {
	return &GatewayStats{
		data:     make(map[QueueName]*GatewayStatsData),
		cur:      make(map[QueueName]*GatewayStatsData),
		itemChan: make(chan *gatewayStatItem, GatewayStatsChannelBuff),
		stop:     make(chan bool),
	}
}

func (s *GatewayStats) Run() {
	go func() {
		ticker := time.NewTicker(GatewayStatsCalInterval)
		defer ticker.Stop()
		for {
			select {
			case <-s.stop:
				return
			case <-ticker.C:
				s.Lock()
				s.data = s.cur
				s.cur = make(map[QueueName]*GatewayStatsData)
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
				s.cur[item.queue] = &GatewayStatsData{
					TimePeriod: GatewayStatsCalInterval,
				}
			}
			if item.success {
				s.cur[item.queue].SuccessCount++
			} else {
				s.cur[item.queue].FailCount++
			}
			s.Unlock()

			// add external stats
			addQueueStats(string(item.queue), item.success)
		}
	}
}

func (s *GatewayStats) Add(queue QueueName, addJobsIsSuccess bool) {
	if queue == "" {
		return
	}
	item := &gatewayStatItem{
		queue:   queue,
		success: addJobsIsSuccess,
	}

	select {
	case s.itemChan <- item:
	default:
		flog.Error("GatewayStats channel is full")
	}
}

func (s *GatewayStats) Output() map[QueueName]*GatewayStatsData {
	s.RLock()
	defer s.RUnlock()
	return s.data
}

func (s *GatewayStats) Stop() {
	close(s.stop)
}
