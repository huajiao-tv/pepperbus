package main

import (
	"errors"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"git.huajiao.com/qmessenger/pepperbus/codec/fastcgi"
	"github.com/johntech-o/idgen"
)

// 新建 CGI 连接失败后，等待时间，单位秒
const (
	NewConnRetryDuration    = 1
	NewConnRetryMaxDuration = 60
)

// CgiConn 封装了 FCGIClient 包，并为连接池做统计
type CgiConn struct {
	id         uint64
	client     *fastcgi.FCGIClient
	usedCounts int64
	created    time.Time

	needClose bool

	cgiConnPool *CgiConnPool

	sync.Mutex
}

// Send 会调用 FCGIClient.PostForm 方法，同时统计使用次数与统计时间
func (c *CgiConn) Send(env map[string]string, params url.Values) (*http.Response, error) {
	c.Lock()
	defer c.Unlock()
	c.usedCounts++
	resp, err := c.client.PostForm(env, params)
	if err != nil {
		c.needClose = true
	}
	return resp, err
}

// Close 将连接 put 回连接池
func (c *CgiConn) Close() {
	c.Lock()
	defer c.Unlock()

	// 判断该连接是否需要直接关闭
	if c.needClose {
		c.cgiConnPool.closeConn(c)
		return
	}

	// 判断 cgi 连接是否可再次使用和使用次数和使用时间是否超限
	if c.usedCounts > c.cgiConnPool.reuseCounts ||
		time.Now().Sub(c.created) > c.cgiConnPool.reuseLife {
		c.cgiConnPool.closeConn(c)
		return
	}

	// 重新放回连接池
	c.cgiConnPool.put(c)
}

// CgiConnPool 连接池结构，每个 Cgi 地址会有一个 ConnPool
type CgiConnPool struct {
	store chan *CgiConn

	// 所有的连接数
	connNum int64

	address        string
	connectTimeout time.Duration
	writeTimeout   time.Duration
	readTimeout    time.Duration

	initConnNum int64
	maxConnNum  int64

	reuseCounts int64
	reuseLife   time.Duration
}

// NewCgiConnPool 创建一个新的连接池，每个 Cgi 地址需要创建一个
// 需要带上该地址的初始连接数与最大连接数
// 初始化时，会建立初始连接
// 当超过最大连接数，连接池将不再新建连接
func NewCgiConnPool(address string, connectTimeout, writeTimeout, readTimeout time.Duration,
	initConnNum, maxConnNum int64,
	reuseCounts int64, reuseLife time.Duration) *CgiConnPool {

	if initConnNum > maxConnNum {
		initConnNum = maxConnNum
	}

	pool := &CgiConnPool{
		store: make(chan *CgiConn, maxConnNum),

		initConnNum: initConnNum,
		maxConnNum:  maxConnNum,

		address:        address,
		connectTimeout: connectTimeout,
		writeTimeout:   writeTimeout,
		readTimeout:    readTimeout,

		reuseCounts: reuseCounts,
		reuseLife:   reuseLife,
	}

	for i := int64(0); i < initConnNum; i++ {
		err := pool.newConn()
		if err != nil {
			flog.Error("CgiConnPool init newConn error", address, err.Error())
			continue
		}
	}

	return pool
}

func (cp *CgiConnPool) newConnReal() (*CgiConn, error) {
	var (
		err  error
		fcgi *fastcgi.FCGIClient

		cgiConn = &CgiConn{
			created:     time.Now(),
			cgiConnPool: cp,
		}

		newCgiClient = func() (*fastcgi.FCGIClient, error) {
			keepAlive := true
			if cp.reuseCounts <= 1 {
				keepAlive = false
			}
			client, err := fastcgi.DialTimeout("tcp", cp.address, cp.connectTimeout, keepAlive)
			if err != nil {
				flog.Error("CgiConnPool newConn error", cp.address, keepAlive, err.Error())
				return nil, err
			}
			// TODO: 设置一个最大值，配置化
			client.SetWriteDeadline(90 * time.Minute)
			client.SetReadDeadline(90 * time.Minute)
			return client, nil
		}
	)

	connNum := atomic.AddInt64(&cp.connNum, 1)
	if connNum > cp.maxConnNum {
		err = errors.New("too many connection")
		goto Err
	}

	fcgi, err = newCgiClient()
	if err != nil {
		goto Err
	}
	cgiConn.id = idgen.GenId()
	cgiConn.client = fcgi

	return cgiConn, nil

Err:
	atomic.AddInt64(&cp.connNum, -1)
	return nil, err
}

func (cp *CgiConnPool) newConn() error {
	cgiConn, err := cp.newConnReal()
	if err != nil {
		return err
	}
	cp.store <- cgiConn
	return nil
}

func (cp *CgiConnPool) put(cgiConn *CgiConn) error {
	// 尝试往连接池里放，如果满了，直接关闭连接
	select {
	case cp.store <- cgiConn:
	default:
		cp.closeConn(cgiConn)
	}

	return nil
}

func (cp *CgiConnPool) closeConn(cgiConn *CgiConn) {
	cgiConn.client.Close()
	atomic.AddInt64(&cp.connNum, -1)
}

// GetConn 从连接池中取出一个连接，返回给使用者
func (cp *CgiConnPool) GetConn() *CgiConn {
	retryDuration := NewConnRetryDuration
	for {
		select {
		case cgiConn := <-cp.store:
			// PHP-FPM 会关闭 5 秒没有发任何包的连接，这里判断一下这个连接是否新建后闲置了 5 秒
			if cgiConn.usedCounts == 0 && time.Now().Sub(cgiConn.created) >= (5*time.Second-500*time.Millisecond) {
				cp.closeConn(cgiConn)
				continue
			}
			if cgiConn.usedCounts > cp.reuseCounts ||
				time.Now().Sub(cgiConn.created) > cp.reuseLife {
				// 已经超限了，关闭连接
				cp.closeConn(cgiConn)
				continue
			}
			return cgiConn
		default:
			flog.Warn("CgiConnPool GetConn failed, try to newConn", cp.address, retryDuration)
			time.Sleep(time.Duration(retryDuration) * time.Second)
			retryDuration *= 2
			if retryDuration >= NewConnRetryMaxDuration {
				retryDuration = NewConnRetryMaxDuration
			}
			cp.newConn()
		}
	}
}

// CgiConnPoolManager 用于管理多个连接池
type CgiConnPoolManager struct {
	pool map[string]*CgiConnPool
	sync.Mutex
}

// NewCgiConnPoolManager 创建一个新的连接池管理器
func NewCgiConnPoolManager() *CgiConnPoolManager {
	return &CgiConnPoolManager{
		pool: make(map[string]*CgiConnPool),
	}
}

// Init 通过配置初始化所有的连接池
func (pm *CgiConnPoolManager) Init(confs map[string]*CgiConfig) {
	pm.Lock()
	defer pm.Unlock()

	for key, conf := range confs {
		pm.pool[key] = NewCgiConnPool(conf.Address, conf.ConnectTimeout, conf.WriteTimeout, conf.ReadTimeout,
			conf.InitConnNum, conf.MaxConnNum, conf.ReuseCounts, conf.ReuseLife)
	}
}

// Get 通过地址获取连接池
func (pm *CgiConnPoolManager) Get(key string) (*CgiConnPool, error) {
	pm.Lock()
	defer pm.Unlock()

	if _, ok := pm.pool[key]; !ok {
		return nil, errors.New("pool not found")
	}

	return pm.pool[key], nil
}
