package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	NameSpace = "pepperbus"
)

var (
	globalRegistry = prometheus.NewRegistry()

	topicLen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: NameSpace,
			Name:      "topic_length",
			Help:      "topic length",
		},
		[]string{"queue", "topic"},
	)
	topicRetryLen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: NameSpace,
			Name:      "topic_retry_length",
			Help:      "topic retry length",
		},
		[]string{"queue", "topic"},
	)
	topicTimeoutLen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: NameSpace,
			Name:      "topic_timeout_length",
			Help:      "topic timeout length",
		},
		[]string{"queue", "topic"},
	)

	cpuUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: NameSpace,
			Name:      "cpu_usage",
			Help:      "cpu usage",
		},
	)
	memUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: NameSpace,
			Name:      "mem_usage",
			Help:      "memory usage",
		},
	)
	loadAverage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: NameSpace,
			Name:      "load_average",
			Help:      "machine load average",
		},
	)
	queueAddCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: NameSpace,
			Name:      "queue_add_count",
			Help:      "queue add operation count",
		},
		[]string{"queue"},
	)
	jobSuccessCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: NameSpace,
			Name:      "job_success_count",
			Help:      "succeed job count",
		},
		[]string{"queue", "topic"},
	)
	jobFailCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: NameSpace,
			Name:      "job_fail_count",
			Help:      "failed jobs count",
		},
		[]string{"queue", "topic"},
	)
	jobConsumes = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  NameSpace,
			Name:       "job_consume_milliseconds",
			Help:       "Cgi job consume milliseconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"queue", "topic"},
	)
)

func init() {
	globalRegistry.MustRegister(topicLen)
	globalRegistry.MustRegister(topicRetryLen)

	prometheus.MustRegister(cpuUsage)
	prometheus.MustRegister(memUsage)
	prometheus.MustRegister(loadAverage)
	prometheus.MustRegister(queueAddCount)
	prometheus.MustRegister(jobSuccessCount)
	prometheus.MustRegister(jobFailCount)
	prometheus.MustRegister(jobConsumes)
}

func updateTopicLength() {
	for qn, queue := range QueueConfs() {
		for tn, _ := range queue.Topic {
			mux, err := MuxManager.GetMuxer(qn, tn)
			if err != nil {
				continue
			}
			normal, retry, timeout, err := mux.Storage().GetLength(qn, tn)
			if err != nil {
				continue
			}
			topicLen.WithLabelValues(string(qn), string(tn)).Set(float64(normal))
			topicRetryLen.WithLabelValues(string(qn), string(tn)).Set(float64(retry))
			topicTimeoutLen.WithLabelValues(string(qn), string(tn)).Set(float64(timeout))
		}
	}
}

func updateCurrentStats() {
	MachineStat.RLock()
	cpuUsage.Set(MachineStat.CPU)
	memUsage.Set(MachineStat.Memory)
	loadAverage.Set(MachineStat.Load)
	MachineStat.RUnlock()
}

func addQueueStats(queue string, success bool) {
	if success {
		queueAddCount.WithLabelValues(queue).Inc()
	}
}

func addJobStats(queue, topic string, duration time.Duration, success bool) {
	if success {
		jobConsumes.WithLabelValues(queue, topic).Observe(float64(duration.Nanoseconds() / 1e6))
		jobSuccessCount.WithLabelValues(queue, topic).Inc()
	} else {
		jobFailCount.WithLabelValues(queue, topic).Inc()
	}
}
