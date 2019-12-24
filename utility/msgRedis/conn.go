package msgRedis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ConnectTimeout    = 10e9
	ReadTimeout       = 60e9
	WriteTimeout      = 60e9
	DefaultBufferSize = 64

	RetryWaitSeconds = time.Second
	RetryTimes       = 2

	TypeError        = '-'
	TypeSimpleString = '+'
	TypeBulkString   = '$'
	TypeIntegers     = ':'
	TypeArrays       = '*'
)

var (
	ErrNil           = errors.New("nil data return")
	ErrBadType       = errors.New("invalid return type")
	ErrBadTcpConn    = errors.New("invalid tcp conn")
	ErrBadTerminator = errors.New("invalid terminator")
	ErrResponse      = errors.New("bad call")
	ErrNilPool       = errors.New("conn not belongs to any pool")
	ErrKeyNotExist   = errors.New(CommonErrPrefix + "key not exist")
	ErrBadArgs       = errors.New(CommonErrPrefix + "request args invalid")
	ErrEmptyDB       = errors.New(CommonErrPrefix + "empty db")
	ErrResponseType  = errors.New(CommonErrPrefix + "response type error")

	CommonErrPrefix = "CommonError:"
)

//
type Conn struct {
	sync.RWMutex
	Address        string
	keepAlive      bool
	isIdle         bool
	pipeCount      int
	lastActiveTime int64
	buffer         []byte
	conn           *net.TCPConn
	rb             *bufio.Reader
	wb             *bufio.Writer
	readTimeout    time.Duration
	writeTimeout   time.Duration
	pool           *Pool
	err            error // 表示该条链接是否已经出错
	isOnce         bool  // 用于判断每次调用后，是否自动放回连接池 ，true自动放回无需开发者显示操作，默认为false
}

func NewConn(conn *net.TCPConn, connectTimeout, readTimeout, writeTimeout time.Duration, keepAlive bool, pool *Pool, Address string) *Conn {
	return &Conn{
		conn:           conn,
		lastActiveTime: time.Now().Unix(),
		keepAlive:      keepAlive,
		isIdle:         true,
		buffer:         make([]byte, DefaultBufferSize),
		rb:             bufio.NewReader(conn),
		wb:             bufio.NewWriter(conn),
		readTimeout:    readTimeout,
		writeTimeout:   writeTimeout,
		pool:           pool,
		Address:        Address,
		isOnce:         false,
	}
}

func Connect(addr string, connectTimeout, readTimeout, writeTimeout time.Duration) (*Conn, error) {
	addrPass := strings.Split(addr, ":")
	address := ""
	password := ""
	if len(addrPass) == 3 {
		address = addrPass[0] + ":" + addrPass[1]
		password = addrPass[2]
	} else if len(addrPass) == 2 {
		address = addr
	} else {
		return nil, errors.New("invalid address pattarn")
	}
	return Dial(address, password, connectTimeout, readTimeout, writeTimeout, false, nil)
}

// connect with timeout
func Dial(address, password string, connectTimeout, readTimeout, writeTimeout time.Duration, keepAlive bool, pool *Pool) (*Conn, error) {
	c, e := net.DialTimeout("tcp", address, connectTimeout)
	if e != nil {
		return nil, e
	}
	if _, ok := c.(*net.TCPConn); !ok {
		return nil, ErrBadTcpConn
	}

	if password != "" {
		address = address + ":" + password
	}
	conn := NewConn(c.(*net.TCPConn), connectTimeout, readTimeout, writeTimeout, keepAlive, pool, address)
	if password != "" {
		if _, e := conn.AUTH(password); e != nil {
			return nil, e
		}
	}
	return conn, nil
}

func (c *Conn) Copy(conn *Conn) {
	c.Address = conn.Address
	c.keepAlive = conn.keepAlive
	c.pipeCount = conn.pipeCount
	c.lastActiveTime = conn.lastActiveTime
	c.buffer = conn.buffer
	c.conn = conn.conn
	c.rb = conn.rb
	c.wb = conn.wb
	c.readTimeout = conn.readTimeout
	c.writeTimeout = conn.writeTimeout
	c.pool = conn.pool
	c.isOnce = conn.isOnce
	c.isIdle = conn.isIdle
	c.err = nil
}

func (c *Conn) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Conn) CallN(retry int, command string, args ...interface{}) (interface{}, error) {
	if c.err != nil {
		return nil, c.err
	}
	// 记录下当前连接是否需要调用完自动放回
	isOnce := c.isOnce
	var ret interface{}
	var e error
	for i := 0; i < retry; i++ {
		ret, e = c.Call(command, args...)
		if c.err != nil && c.pool != nil && i+1 < retry {
			// 如果isOnce参数为true，会在Call函数中放回
			// isOnce为true时，说明在call函数中已经被放进了连接池内
			if !isOnce {
				c.pool.Push(c)
			}
			conn := c.pool.Pop()
			if conn == nil {
				return nil, e
			}
			c.Copy(conn)
			// 如果CallN是通过Once调用，第一次Call失败，isOnce会被置为false
			// 此时从pool中再	取一条conn（pool中的isOnce是false）
			// 再调用的时候就不会被自动放回，造成泄漏
			c.isOnce = isOnce
			continue
		}
		break
	}
	return ret, e
}

// call redis command with request => response model
func (c *Conn) Call(command string, args ...interface{}) (interface{}, error) {
	// 如果连接已经被标记出错，直接返回
	// 应用场景，OnceConn没有获取到连接，会新建一个err非nil的Conn结构
	if c.err != nil {
		return nil, c.err
	}

	c.lastActiveTime = time.Now().Unix()
	// start := time.Now()
	if c.pool != nil {
		c.pool.callMu.Lock()
		c.pool.CallNum++
		c.pool.callMu.Unlock()
	}
	var e error
	// 如果链接网络出错，标记该条链接已出错，并立刻关闭该条链接
	defer func() {
		if e != nil && !strings.Contains(e.Error(), CommonErrPrefix) {
			if c.pool != nil {
				if command != "PING" {
					c.pool.mu.Lock()
					c.pool.CallNetErrNum++
					c.pool.mu.Unlock()
				} else {
					c.pool.mu.Lock()
					c.pool.PingErrNum++
					c.pool.mu.Unlock()
				}
			}
			c.err = e
		}

		// 需要自动放回
		if c.isOnce {
			c.isOnce = false
			if c.pool != nil {
				c.pool.Push(c)
			}
		}

	}()

	if c.writeTimeout > 0 {
		if e = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); e != nil {
			return nil, e
		}
	}
	if e = c.writeRequest(command, args); e != nil {
		return nil, e
	}

	if e = c.wb.Flush(); e != nil {
		return nil, e
	}

	if c.readTimeout > 0 {
		if e = c.conn.SetReadDeadline(time.Now().Add(c.writeTimeout)); e != nil {
			return nil, e
		}
	}
	response, e := c.readResponse()
	if e != nil {
		return nil, e
	}
	return response, e
}

// write response
func (c *Conn) writeRequest(command string, args []interface{}) error {
	var e error
	if e = c.writeLen('*', 1+len(args)); e != nil {
		return e
	}

	if e = c.writeString(command); e != nil {
		return e
	}

	for _, arg := range args {
		if e != nil {
			return e
		}
		switch data := arg.(type) {
		case int:
			e = c.writeInt64(int64(data))
		case int64:
			e = c.writeInt64(data)
		case float64:
			e = c.writeFloat64(data)
		case string:
			e = c.writeString(data)
		case []byte:
			e = c.writeBytes(data)
		case bool:
			if data {
				e = c.writeString("1")
			} else {
				e = c.writeString("0")
			}
		case nil:
			e = c.writeString("")
		default:
			e = c.writeString(fmt.Sprintf("%v", data))
		}
	}
	return e
}

// reuse one buffer
func (c *Conn) writeLen(prefix byte, n int) error {
	pos := len(c.buffer) - 1
	c.buffer[pos] = '\n'
	pos--
	c.buffer[pos] = '\r'
	pos--
	if n == 0 {
		c.buffer[pos] = '0'
		c.buffer[pos-1] = prefix
		if _, e := c.wb.Write(c.buffer[pos-1:]); e != nil {
			return e
		}
		return nil
	}

	for i := n; i != 0 && pos >= 0; i = i / 10 {
		c.buffer[pos] = byte(i%10 + '0')
		pos--
	}
	c.buffer[pos] = prefix
	_, e := c.wb.Write(c.buffer[pos:])
	if e != nil {
		return e
	}
	return nil
}

// write
func (c *Conn) writeBytes(b []byte) error {
	var e error
	if e = c.writeLen('$', len(b)); e != nil {
		return e
	}
	if _, e = c.wb.Write(b); e != nil {
		return e
	}
	if _, e = c.wb.WriteString("\r\n"); e != nil {
		return e
	}
	return nil
}

func (c *Conn) writeString(s string) error {
	var e error
	if e = c.writeLen('$', len(s)); e != nil {
		return e
	}
	if _, e = c.wb.WriteString(s); e != nil {
		return e
	}
	if _, e = c.wb.WriteString("\r\n"); e != nil {
		return e
	}
	return nil

}

func (c *Conn) writeFloat64(f float64) error {
	// Negative precision means "only as much as needed to be exact."
	return c.writeBytes(strconv.AppendFloat([]byte{}, f, 'g', -1, 64))
}

func (c *Conn) writeInt64(n int64) error {
	return c.writeBytes(strconv.AppendInt([]byte{}, n, 10))
}

// read
func (c *Conn) readResponse() (interface{}, error) {
	var e error
	p, e := c.readLine()
	if e != nil {
		return nil, e
	}
	resType := p[0]
	p = p[1:]
	switch resType {
	case TypeError:
		// 错误操作，非网络错误，不应该重建连接
		return nil, errors.New(CommonErrPrefix + string(p))
	case TypeIntegers:
		return c.parseInt(p)
	case TypeSimpleString:
		return p, nil
	case TypeBulkString:
		return c.parseBulkString(p)
	case TypeArrays:
		return c.parseArray(p)
	default:
	}
	return nil, errors.New(CommonErrPrefix + "Err type")
}

func (c *Conn) readLine() (b []byte, e error) {
	defer func() {
		if r := recover(); r != nil {
			println("Recovered in readLine", r)
			e = errors.New("readLine painc")
		}
	}()
	// var e error
	p, e := c.rb.ReadBytes('\n')
	if e != nil {
		return nil, e
	}

	i := len(p) - 2
	if i <= 0 {
		return nil, ErrBadTerminator
	}
	return p[:i], nil
}

func (c *Conn) parseInt(p []byte) (int64, error) {
	n, e := strconv.ParseInt(string(p), 10, 64)
	if e != nil {
		return 0, errors.New(CommonErrPrefix + e.Error())
	}
	return n, nil
}

func (c *Conn) parseBulkString(p []byte) (interface{}, error) {
	n, e := strconv.ParseInt(string(p), 10, 64)
	if e != nil {
		return []byte{}, errors.New(CommonErrPrefix + e.Error())
	}
	if n == -1 {
		return nil, nil
	}

	result := make([]byte, n+2)
	_, e = io.ReadFull(c.rb, result)
	return result[:n], e
}

func (c *Conn) parseArray(p []byte) ([]interface{}, error) {
	n, e := strconv.ParseInt(string(p), 10, 64)
	if e != nil {
		return nil, errors.New(CommonErrPrefix + e.Error())
	}

	if n == -1 {
		return nil, nil
	}

	result := make([]interface{}, n)
	var i int64
	for ; i < n; i++ {
		result[i], e = c.readResponse()
		if e != nil {
			return nil, e
		}
	}
	return result, nil
}

// pipeline与transactions没有用callN，失败没有重试
// pipeline
func (c *Conn) PipeSend(command string, args ...interface{}) error {
	c.pipeCount++
	return c.writeRequest(command, args)
}

func (c *Conn) PipeExec() ([]interface{}, error) {
	var e error
	if e = c.wb.Flush(); e != nil {
		return nil, e
	}
	n := c.pipeCount
	ret := make([]interface{}, c.pipeCount)
	c.pipeCount = 0
	for i := 0; i < n; i++ {
		ret[i], e = c.readResponse()
	}
	return ret, e
}

// Transactions
func (c *Conn) MULTI() error {
	ret, e := c.Call("MULTI")
	if e != nil {
		return e
	}
	if _, ok := ret.([]byte); !ok {
		return ErrBadType
	}
	r := ret.([]byte)
	if len(r) == 2 && r[0] == 'O' && r[1] == 'K' {
		return nil
	}
	return errors.New("invalid return:" + string(r))
}

func (c *Conn) TransSend(command string, args ...interface{}) error {
	ret, e := c.Call(command, args...)
	if e != nil {
		return e
	}
	if _, ok := ret.([]byte); !ok {
		return ErrBadType
	}
	r := ret.([]byte)
	if len(r) == 6 &&
		r[0] == 'Q' && r[1] == 'U' && r[2] == 'E' && r[3] == 'U' && r[4] == 'E' && r[5] == 'D' {
		return nil
	}
	return errors.New("invalid return:" + string(r))
}

func (c *Conn) TransExec() ([]interface{}, error) {
	ret, e := c.Call("EXEC")
	if e = c.wb.Flush(); e != nil {
		return nil, e
	}
	if ret == nil {
		// nil indicate transaction failed
		return nil, ErrNil
	}
	return ret.([]interface{}), e
}

func (c *Conn) Discard() error {
	ret, e := c.Call("DISCARD")
	if e != nil {
		return e
	}
	if _, ok := ret.([]byte); !ok {
		return ErrBadType
	}
	r := ret.([]byte)
	if len(r) == 2 && r[0] == 'O' && r[1] == 'K' {
		return nil
	}
	return errors.New("invalid return:" + string(r))
}

func (c *Conn) Watch(keys []string) error {
	args := make([]interface{}, len(keys))
	for i, key := range args {
		args[i] = key
	}
	ret, e := c.Call("WATCH", args...)
	if e != nil {
		return e
	}
	if _, ok := ret.([]byte); !ok {
		return ErrBadType
	}
	r := ret.([]byte)
	if len(r) == 2 && r[0] == 'O' && r[1] == 'K' {
		return nil
	}
	return errors.New("invalid return:" + string(r))
}
