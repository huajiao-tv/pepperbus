package utility

import (
	"sync"
	"time"
)

type ErrorCounter struct {
	execInterval  time.Duration
	clearInterval time.Duration

	sync.RWMutex
	logCount     map[time.Time]uint64
	callbackFunc map[string]*callbackFunc

	stop chan bool
}

type callbackFunc struct {
	moreThen uint64
	fn       func(uint64)
}

func NewErrorCounter(execInterval, clearInterval time.Duration) *ErrorCounter {
	return &ErrorCounter{
		execInterval:  execInterval,
		clearInterval: clearInterval,
		logCount:      make(map[time.Time]uint64),
		callbackFunc:  make(map[string]*callbackFunc),
		stop:          make(chan bool, 1),
	}
}

func (ec *ErrorCounter) Incr(data uint64) {
	ec.Lock()
	defer ec.Unlock()
	ec.logCount[time.Now()] = data
}

func (ec *ErrorCounter) RegisterCallback(name string, moreThen uint64, fn func(uint64)) {
	ec.Lock()
	defer ec.Unlock()

	cf := &callbackFunc{
		moreThen,
		fn,
	}

	ec.callbackFunc[name] = cf
}

func (ec *ErrorCounter) UnRegisterCallback(name string) {
	ec.Lock()
	defer ec.Unlock()

	delete(ec.callbackFunc, name)
}

func (ec *ErrorCounter) clearAndCal() uint64 {
	ec.Lock()
	defer ec.Unlock()

	for t, _ := range ec.logCount {
		if time.Now().Sub(t) > ec.clearInterval {
			delete(ec.logCount, t)
		}
	}

	var datas uint64
	for _, data := range ec.logCount {
		datas += data
	}

	return datas
}

func (ec *ErrorCounter) Run() {
	execTicker, clearTicker := time.NewTicker(ec.execInterval), time.NewTicker(ec.clearInterval)
	defer execTicker.Stop()
	defer clearTicker.Stop()

	for {
		select {
		case <-ec.stop:
			return
		case <-execTicker.C:
			count := ec.clearAndCal()
			for _, cf := range ec.callbackFunc {
				if count >= cf.moreThen {
					go cf.fn(count)
				}
			}
		case <-clearTicker.C:
			ec.clearAndCal()
		}
	}
}

func (ec *ErrorCounter) Stop() {
	ec.stop <- true
}
