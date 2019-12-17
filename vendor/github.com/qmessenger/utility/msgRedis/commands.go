package msgRedis

// all return integer is int64
import (
	"errors"
	"strconv"
	"strings"
)

/******************* connections *******************/
func (c *Conn) AUTH(password string) (bool, error) {
	v, e := c.Call("AUTH", password)
	if e != nil {
		Debug("AUTH failed:"+e.Error(), c.Address)
		return false, e
	}

	r, ok := v.([]byte)
	if !ok {
		return false, errors.New("invaild response type")
	}

	if len(r) == 2 && r[0] == 'O' && r[1] == 'K' {
		return true, nil
	}
	return false, errors.New("invaild response string:" + string(r))
}

func (c *Conn) QUIT() {

}

func (c *Conn) IsAlive() bool {
	// v, e := c.CallN(RetryTimes, "PING")
	v, e := c.Call("PING")
	if e != nil {
		return false
	}
	r, ok := v.([]byte)
	if !ok {
		return false
	}
	if len(r) == 4 && r[0] == 'P' && r[1] == 'O' && r[2] == 'N' && r[3] == 'G' {
		return true
	}
	return false
}

func (c *Conn) Info() ([]byte, error) {
	v, e := c.Call("INFO")
	if e != nil {
		return nil, e
	}
	r, ok := v.([]byte)
	if !ok {
		return nil, errors.New("invalid type")
	}
	return r, nil
}

/******************* keys commands *******************/
func (c *Conn) DEL(key string) (int64, error) {
	n, e := c.CallN(RetryTimes, "DEL", key)
	if e != nil {
		return -1, e
	}
	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}
	return n.(int64), nil
}

func (c *Conn) DELMulti(keys []string) (int64, error) {
	args := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		args[i] = keys[i]
	}
	n, e := c.CallN(RetryTimes, "DEL", args...)
	if e != nil {
		return -1, e
	}
	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}
	return n.(int64), nil
}

func (c *Conn) DUMP(key string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "DUMP", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}
	return v.([]byte), nil
}

func (c *Conn) EXISTS(key string) (bool, error) {
	n, e := c.CallN(RetryTimes, "EXISTS", key)
	if e != nil {
		return false, e
	}

	if _, ok := n.(int64); !ok {
		return false, ErrResponseType
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) EXPIRE(key string, seconds int64) (bool, error) {
	n, e := c.CallN(RetryTimes, "EXPIRE", key, seconds)
	if e != nil {
		return false, e
	}

	if _, ok := n.(int64); !ok {
		return false, ErrResponseType
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) EXPIREAT(key string, timestamp int64) (bool, error) {
	n, e := c.CallN(RetryTimes, "EXPIREAT", key, timestamp)
	if e != nil {
		return false, e
	}

	if _, ok := n.(int64); !ok {
		return false, ErrResponseType
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) KEYS(pattern string) ([][]byte, error) {
	v, e := c.CallN(RetryTimes, "KEYS", pattern)
	if e != nil {
		return nil, e
	}
	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

// since 2.6.0  COPY and REPLACE will be available in 3.0
func (c *Conn) MIGRATE(host, port, key, destDB string, timeout int, COPY, REPLACE bool) (bool, error) {
	v, e := c.CallN(RetryTimes, "MIGRATE", host, port, key, destDB, timeout)
	if e != nil {
		return false, e
	}

	if _, ok := v.([]byte); !ok {
		return false, ErrResponseType
	}

	r, ok := v.([]byte)
	if !ok {
		return false, errors.New("invaild response type")
	}

	if len(r) == 2 && r[0] == 'O' && r[1] == 'K' {
		return true, nil
	}
	return false, errors.New("migrate false")
}

func (c *Conn) SELECT(index int) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "SELECT", index)
	if e != nil {
		return nil, e
	}
	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) MOVE(key, db string) (bool, error) {
	n, e := c.CallN(RetryTimes, "MOVE", key, db)
	if e != nil {
		return false, e
	}
	if _, ok := n.(int64); !ok {
		return false, ErrResponseType
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) OBJECT(subcommand, key string) (interface{}, error) {
	v, e := c.CallN(RetryTimes, "OBJECT", subcommand, key)
	if e != nil {
		return nil, e
	}

	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.(int64); ok {
		return v.(int64), nil
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) PERSIST(key string) (bool, error) {
	n, e := c.CallN(RetryTimes, "PERSIST", key)
	if e != nil {
		return false, e
	}

	if _, ok := n.(int64); !ok {
		return false, ErrResponseType
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) PEXPIRE(key string, milliseconds int64) (bool, error) {
	n, e := c.CallN(RetryTimes, "EXPIRE", key, milliseconds)
	if e != nil {
		return false, e
	}

	if _, ok := n.(int64); !ok {
		return false, ErrResponseType
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) PEXPIREAT(key string, milliTimestamp int64) (bool, error) {
	n, e := c.CallN(RetryTimes, "EXPIREAT", key, milliTimestamp)
	if e != nil {
		return false, e
	}

	if _, ok := n.(int64); !ok {
		return false, ErrResponseType
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) PTTL(key string) (int64, error) {
	n, e := c.CallN(RetryTimes, "PTTL", key)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) RANDOMKEY() ([]byte, error) {
	v, e := c.CallN(RetryTimes, "RANDOMKEY")
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrEmptyDB
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) RENAME(key, newkey string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "RENAME", key, newkey)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) RENAMENX(key, newkey string) (bool, error) {
	n, e := c.CallN(RetryTimes, "RENAMENX", key, newkey)
	if e != nil {
		return false, e
	}

	if _, ok := n.(int64); !ok {
		return false, ErrResponseType
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

// with dump
func (c *Conn) RESTORE(key string, ttl int, serializedValue string) (bool, error) {
	v, e := c.CallN(RetryTimes, "RESTORE", key, ttl, serializedValue)
	if e != nil {
		return false, e
	}

	if _, ok := v.([]byte); !ok {
		return false, ErrResponseType
	}

	r := v.([]byte)
	if len(r) == 2 && r[0] == 'O' && r[1] == 'K' {
		return true, nil
	}
	return false, nil
}

func (c *Conn) SORT() {}

func (c *Conn) TTL(key string) (int64, error) {
	n, e := c.CallN(RetryTimes, "TTL", key)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) TYPE(key string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "TYPE", key)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) SCAN(cursor int, match bool, pattern string, isCount bool, count int) (int, []interface{}, error) {
	args := make([]interface{}, 0, 5)
	args = append(args, cursor)
	if match {
		args = append(args, "MATCH", pattern)
	}
	if isCount {
		args = append(args, "COUNT", count)
	}
	v, e := c.CallN(RetryTimes, "SCAN", args...)
	if e != nil {
		return 0, nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return 0, nil, ErrResponseType
	}

	r := v.([]interface{})
	// return cursor
	rCursor, _ := strconv.Atoi(string(r[0].([]byte)))
	return rCursor, r[1].([]interface{}), nil
}

/******************* sets commands *******************/
func (c *Conn) SADD(key string, values []string) (int64, error) {
	args := make([]interface{}, len(values)+1)
	args[0] = key
	for i := 0; i < len(values); i++ {
		args[i+1] = values[i]
	}
	n, e := c.CallN(RetryTimes, "SADD", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) SREM(key string, values []string) (int64, error) {
	args := make([]interface{}, len(values)+1)
	args[0] = key
	for i := 0; i < len(values); i++ {
		args[i+1] = values[i]
	}
	n, e := c.CallN(RetryTimes, "SREM", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) SISMEMBER(key, value string) (int64, error) {
	v, e := c.CallN(RetryTimes, "SISMEMBER", key, value)
	if e != nil {
		return 0, e
	}

	if _, ok := v.(int64); !ok {
		return -1, ErrResponseType
	}

	return v.(int64), nil
}

func (c *Conn) SMEMBERS(key string) ([][]byte, error) {
	v, e := c.CallN(RetryTimes, "SMEMBERS", key)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

// 0说明key不存在
func (c *Conn) SCARD(key string) (int64, error) {
	v, e := c.CallN(RetryTimes, "SCARD", key)
	if e != nil {
		return 0, e
	}

	if _, ok := v.(int64); !ok {
		return -1, ErrResponseType
	}

	return v.(int64), nil
}

func (c *Conn) SINTER(keys []string) ([][]byte, error) {
	args := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		args[i] = keys[i]
	}
	v, e := c.CallN(RetryTimes, "SINTER", args...)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

func (c *Conn) SINTERSTORE(key string, keys []string) (int64, error) {
	args := make([]interface{}, len(keys)+1)
	args[0] = key
	for i := 0; i < len(keys); i++ {
		args[i+1] = keys[i]
	}
	n, e := c.CallN(RetryTimes, "SINTERSTORE", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) SDIFF(keys []string) ([][]byte, error) {
	args := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		args[i] = keys[i]
	}
	v, e := c.CallN(RetryTimes, "SDIFF", args...)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

func (c *Conn) SDIFFSTORE(key string, keys []string) (int64, error) {
	args := make([]interface{}, len(keys)+1)
	args[0] = key
	for i := 0; i < len(keys); i++ {
		args[i+1] = keys[i]
	}
	n, e := c.CallN(RetryTimes, "SDIFFSTORE", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

// TODO:return bool
func (c *Conn) SMOVE(srcKey, desKey, member string) (int64, error) {
	n, e := c.CallN(RetryTimes, "SMOVE", srcKey, desKey, member)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) SPOP(key string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "SPOP", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) SRANDMEMBER(key string, count int) ([][]byte, error) {
	if count == 0 {
		v, e := c.CallN(RetryTimes, "SRANDMEMBER", key)
		if e != nil {
			return nil, e
		}
		members := make([][]byte, 1)
		members[0] = v.([]byte)
		return members, nil
	}
	v, e := c.CallN(RetryTimes, "SRANDMEMBER", key, count)
	if e != nil {
		return nil, e
	}

	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

func (c *Conn) SUNION(keys []string) ([][]byte, error) {
	args := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		args[i] = keys[i]
	}
	v, e := c.CallN(RetryTimes, "SUNION", args...)
	if e != nil {
		return nil, e
	}
	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

func (c *Conn) SUNIONSTORE(key string, keys []string) (int64, error) {
	args := make([]interface{}, len(keys)+1)
	args[0] = key
	for i := 0; i < len(keys); i++ {
		args[i+1] = keys[i]
	}
	n, e := c.CallN(RetryTimes, "SUNIONSTORE", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

/******************* strings commands *******************/
func (c *Conn) APPEND(key, value string) (int64, error) {
	n, e := c.CallN(RetryTimes, "APPEND", key, value)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), e
}

func (c *Conn) BITCOUNT(key string) (int64, error) {
	n, e := c.CallN(RetryTimes, "BITCOUNT", key)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), e
}

// 2.6.0
func (c *Conn) BITOP(op, dest string, keys []string) (int64, error) {
	args := make([]interface{}, len(keys)+2)
	args[0] = op
	args[1] = dest
	for i := 0; i < len(keys); i++ {
		args[i+2] = keys[i]
	}
	n, e := c.CallN(RetryTimes, "BITOP", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), e
}

// 2.8.7 TODO
func (c *Conn) BITPOS() {}

func (c *Conn) DECR(key string) (int64, error) {
	n, e := c.CallN(RetryTimes, "DECR", key)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), e
}

func (c *Conn) DECRBY(key string, num int) (int64, error) {
	n, e := c.CallN(RetryTimes, "DECRBY", key, num)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), e
}

func (c *Conn) INCR(key string) (int64, error) {
	n, e := c.CallN(RetryTimes, "INCR", key)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), e
}

func (c *Conn) INCRBY(key string, num int) (int64, error) {
	n, e := c.CallN(RetryTimes, "INCRBY", key, num)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), e
}

func (c *Conn) INCRBYFLOAT(key string, f float64) ([]byte, error) {
	n, e := c.CallN(RetryTimes, "INCRBYFLOAT", key, f)
	if e != nil {
		return nil, e
	}

	if _, ok := n.([]byte); !ok {
		return nil, ErrResponseType
	}

	return n.([]byte), e
}

func (c *Conn) SET(key, value string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "SET", key, value)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

// 应该返回interface还是[]byte?
func (c *Conn) GET(key string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "GET", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) GETBIT(key string, pos int) (int64, error) {
	n, e := c.CallN(RetryTimes, "GETBIT", key, pos)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), e
}

func (c *Conn) GETRANGE(key string, start, end int) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "GETRANGE", key, start, end)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) GETSET(key, value string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "GETSET", key, value)
	if e != nil {
		return nil, e
	}

	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) MGET(keys []string) ([]interface{}, error) {
	args := make([]interface{}, len(keys))
	for k, v := range keys {
		args[k] = v
	}
	v, e := c.CallN(RetryTimes, "MGET", args...)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	return v.([]interface{}), nil
}

func (c *Conn) MSET(kv map[string]string) ([]byte, error) {
	args := make([]interface{}, 2*len(kv))
	i := 0
	for k, v := range kv {
		args[i] = k
		args[i+1] = v
		i = i + 2
	}
	v, e := c.CallN(RetryTimes, "MSET", args...)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) MSETNX(kv map[string]string) (int64, error) {
	args := make([]interface{}, 2*len(kv))
	i := 0
	for k, v := range kv {
		args[i] = k
		args[i+1] = v
		i = i + 2
	}
	v, e := c.CallN(RetryTimes, "MSETNX", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := v.(int64); !ok {
		return -1, ErrResponseType
	}

	return v.(int64), e
}

func (c *Conn) PSETEX(key string, millonseconds int64, value string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "PSETEX", key, millonseconds, value)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) SETBIT(key string, pos, value int) (int64, error) {
	n, e := c.CallN(RetryTimes, "SETBIT", key, pos, value)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), e
}

func (c *Conn) SETEX(key string, seconds int64, value string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "SETEX", key, seconds, value)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) SETNX(key, value string) (int64, error) {
	v, e := c.CallN(RetryTimes, "SETNX", key, value)
	if e != nil {
		return -1, e
	}

	if _, ok := v.(int64); !ok {
		return -1, ErrResponseType
	}

	return v.(int64), e
}

func (c *Conn) SETRANGE(key string, offset int, value string) (int64, error) {
	v, e := c.CallN(RetryTimes, "SETRANGE", key, offset, value)
	if e != nil {
		return -1, e
	}

	if _, ok := v.(int64); !ok {
		return -1, ErrResponseType
	}

	return v.(int64), e
}

func (c *Conn) STRLEN(key string) (int64, error) {
	v, e := c.CallN(RetryTimes, "STRLEN", key)
	if e != nil {
		return -1, e
	}

	if _, ok := v.(int64); !ok {
		return -1, ErrResponseType
	}

	return v.(int64), e
}

func (c *Conn) SSCAN(key string, cursor int, match bool, pattern string, isCount bool, count int) (int, []interface{}, error) {
	args := make([]interface{}, 0, 6)
	args = append(args, key, cursor)
	if match {
		args = append(args, "MATCH", pattern)
	}
	if isCount {
		args = append(args, "COUNT", count)
	}
	v, e := c.CallN(RetryTimes, "SSCAN", args...)
	if e != nil {
		return 0, nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return 0, nil, ErrResponseType
	}

	r := v.([]interface{})
	// return cursor
	rCursor, _ := strconv.Atoi(string(r[0].([]byte)))
	return rCursor, r[1].([]interface{}), nil
}

/******************* hashes commands *******************/
func (c *Conn) HDEL(key string, fields []string) (int64, error) {
	args := make([]interface{}, len(fields)+1)
	args[0] = key
	for i := 0; i < len(fields); i++ {
		args[i+1] = fields[i]
	}
	n, e := c.CallN(RetryTimes, "HDEL", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) HEXISTS(key string, field string) (bool, error) {
	n, e := c.CallN(RetryTimes, "HEXISTS", key, field)
	if e != nil {
		return false, e
	}

	if _, ok := n.(int64); !ok {
		return false, ErrResponseType
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) HGET(key string, field string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "HGET", key, field)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) HGETALL(key string) ([]interface{}, error) {
	v, e := c.CallN(RetryTimes, "HGETALL", key)
	if e != nil {
		return nil, e
	}

	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	return v.([]interface{}), nil
}

// 返回结果用map组织
func (c *Conn) HGETALLMAP(key string) (map[string]string, error) {
	v, e := c.HGETALL(key)
	if e != nil {
		return nil, e
	}
	result := make(map[string]string)
	if len(v) > 0 && len(v)%2 == 0 {
		for i := 0; i < len(v); {
			if _, ok := v[i].([]byte); ok {
				result[string(v[i].([]byte))] = string(v[i+1].([]byte))
				i = i + 2
			}
		}
	} else {
		return result, ErrKeyNotExist
	}
	return result, nil
}

func (c *Conn) HINCRBY(key string, field string, increment int) (int64, error) {
	n, e := c.CallN(RetryTimes, "HINCRBY", key, field, increment)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) HINCRBYFLOAT(key string, field string, increment float64) ([]byte, error) {
	n, e := c.CallN(RetryTimes, "HINCRBYFLOAT", key, field, increment)
	if e != nil {
		return nil, e
	}

	if _, ok := n.([]byte); !ok {
		return nil, ErrResponseType
	}

	return n.([]byte), nil
}

func (c *Conn) HKEYS(key string) ([][]byte, error) {
	v, e := c.CallN(RetryTimes, "HKEYS", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

func (c *Conn) HLEN(key string) (int64, error) {
	n, e := c.CallN(RetryTimes, "HLEN", key)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) HMGET(key string, fields []string) ([]interface{}, error) {
	args := make([]interface{}, len(fields)+1)
	args[0] = key
	for i := 0; i < len(fields); i++ {
		args[i+1] = fields[i]
	}
	v, e := c.CallN(RetryTimes, "HMGET", args...)
	if e != nil {
		return nil, e
	}

	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	return v.([]interface{}), nil
}

func (c *Conn) HMSET(key string, kv map[string]interface{}) ([]byte, error) {
	args := make([]interface{}, 2*len(kv)+1)
	args[0] = key
	i := 1
	for k, v := range kv {
		args[i] = k
		args[i+1] = v
		i = i + 2
	}
	v, e := c.CallN(RetryTimes, "HMSET", args...)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) HSET(key, field string, value interface{}) (int64, error) {
	n, e := c.CallN(RetryTimes, "HSET", key, field, value)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) HSETNX(key, field string, value interface{}) (int64, error) {
	n, e := c.CallN(RetryTimes, "HSETNX", key, field, value)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) HVALS(key string) ([]interface{}, error) {
	v, e := c.CallN(RetryTimes, "HVALS", key)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	return v.([]interface{}), nil
}

func (c *Conn) HSCAN(key string, cursor int, match bool, pattern string, isCount bool, count int) (int, []interface{}, error) {
	args := make([]interface{}, 0, 6)
	args = append(args, key, cursor)
	if match {
		args = append(args, "MATCH", pattern)
	}
	if isCount {
		args = append(args, "COUNT", count)
	}
	v, e := c.CallN(RetryTimes, "HSCAN", args...)
	if e != nil {
		return 0, nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return 0, nil, ErrResponseType
	}

	r := v.([]interface{})
	// return cursor
	rCursor, _ := strconv.Atoi(string(r[0].([]byte)))
	return rCursor, r[1].([]interface{}), nil
}

/******************* lists commands *******************/
func (c *Conn) BLPOP(keys []string, timeout int) ([]interface{}, error) {
	args := make([]interface{}, len(keys)+1)
	for k, v := range keys {
		args[k] = v
	}
	args[len(keys)] = timeout

	v, e := c.CallN(RetryTimes, "BLPOP", args...)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	return v.([]interface{}), nil
}

func (c *Conn) BRPOP(keys []string, timeout int) ([]interface{}, error) {
	args := make([]interface{}, len(keys)+1)
	for k, v := range keys {
		args[k] = v
	}
	args[len(keys)] = timeout

	v, e := c.CallN(RetryTimes, "BRPOP", args...)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	return v.([]interface{}), nil
}

func (c *Conn) BRPOPLPUSH(source, dest string, timeout int) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "BRPOPLPUSH", source, dest, timeout)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) LINDEX(key string, index int) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "LINDEX", key, index)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) LINSERT(key, dir, pivot, value string) (int64, error) {
	if strings.ToLower(dir) != "before" && strings.ToLower(dir) != "after" {
		return -1, errors.New(CommonErrPrefix + "dir only can be (before or after)")
	}
	n, e := c.CallN(RetryTimes, "LINSERT", key, dir, pivot, value)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) LLEN(key string) (int64, error) {
	n, e := c.CallN(RetryTimes, "LLEN", key)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) LPOP(key string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "LPOP", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) LPUSH(key string, values []string) (int64, error) {
	args := make([]interface{}, len(values)+1)
	args[0] = key
	for i, v := range values {
		args[i+1] = v
	}
	n, e := c.CallN(RetryTimes, "LPUSH", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) LPUSHX(key, value string) (int64, error) {
	n, e := c.CallN(RetryTimes, "LPUSHX", key, value)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) LRANGE(key string, start, end int) ([]interface{}, error) {
	v, e := c.CallN(RetryTimes, "LRANGE", key, start, end)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	return v.([]interface{}), nil
}

func (c *Conn) LREM(key string, count int, value string) (int64, error) {
	n, e := c.CallN(RetryTimes, "LREM", key, count, value)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) LSET(key string, index int, value string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "LSET", key, index, value)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) LTRIM(key string, start, end int) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "LTRIM", key, start, end)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) RPOP(key string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "RPOP", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) RPOPLPUSH(source, dest string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "RPOPLPUSH", source, dest)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) RPUSH(key string, values []string) (int64, error) {
	args := make([]interface{}, len(values)+1)
	args[0] = key
	for i, v := range values {
		args[i+1] = v
	}
	n, e := c.CallN(RetryTimes, "RPUSH", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) RPUSHX(key, value string) (int64, error) {
	n, e := c.CallN(RetryTimes, "RPUSHX", key, value)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

/******************* sorted sets commands *******************/
func (c *Conn) ZADDSpec(key string, score, value string) (int64, error) {
	n, e := c.CallN(RetryTimes, "ZADD", key, score, value)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) ZADD(key string, keyScore map[string]interface{}) (int64, error) {
	args := make([]interface{}, 1+2*len(keyScore))
	args[0] = key
	i := 1
	for k, s := range keyScore {
		args[i] = s
		args[i+1] = k
		i = i + 2
	}
	n, e := c.CallN(RetryTimes, "ZADD", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) ZCARD(key string) (int64, error) {
	n, e := c.CallN(RetryTimes, "ZCARD", key)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) ZCOUNT(key string, min, max float64) (int64, error) {
	n, e := c.CallN(RetryTimes, "ZCOUNT", key, min, max)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

// increment could be int, float ,string
func (c *Conn) ZINCRBY(key string, increment interface{}, member string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "ZINCRBY", key, increment, member)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) ZINTERSTORE(destination string, numkeys int, keys []string, weights bool, ws []int, aggregate bool, ag string) (int64, error) {
	args := make([]interface{}, 2+numkeys)
	args[0] = destination
	args[1] = numkeys
	if len(keys) < numkeys {
		return -1, ErrBadArgs
	}
	for i := 0; i < numkeys; i++ {
		args[i+2] = keys[i]
	}
	if weights == true {
		if len(ws) < numkeys {
			return -1, ErrBadArgs
		}
		args = append(args, "WEIGHTS")
		for i := 0; i < numkeys; i++ {
			args = append(args, ws[i])
		}
	}

	if aggregate == true {
		args = append(args, "AGGREGATE", ag)
	}
	n, e := c.CallN(RetryTimes, "ZINTERSTORE", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

// since 2.8.9
// func (c *Conn) ZLEXCOUNT(key, min, max string) (int64, error) {
// 	n, e := c.CallN(RetryTimes,"ZLEXCOUNT", key, min, max)
// 	if e != nil {
// 		return -1, e
// 	}
// 	return n.(int64), nil
// }

func (c *Conn) ZRANGE(key string, start, stop int, withscores bool) ([]interface{}, error) {
	if withscores == true {
		v, e := c.CallN(RetryTimes, "ZRANGE", key, start, stop, "WITHSCORES")
		if e != nil {
			return nil, e
		}
		return v.([]interface{}), nil
	}
	v, e := c.CallN(RetryTimes, "ZRANGE", key, start, stop)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	return v.([]interface{}), nil
}

// since 2.8.9
// func (c *Conn) ZRANGEBYLEX(key, min, max string, limit bool, offset, count int) ([]interface{}, error) {
// 	args := make([]interface{}, 3)
// 	args[0] = key
// 	args[1] = min
// 	args[2] = max
// 	if limit {
// 		args = append(args, "LIMIT", offset, count)
// 	}
// 	v, e := c.CallN(RetryTimes,"ZRANGEBYLEX", args...)
// 	if e != nil {
// 		return nil, e
// 	}
// 	return v.([]interface{}), nil
// }

// 2.8.9
// func (c *Conn) ZREVRANGEBYLEX(key, max, min string, limit bool, offset, count int) ([]interface{}, error) {
// 	args := make([]interface{}, 3)
// 	args[0] = key
// 	args[1] = max
// 	args[2] = min
// 	if limit {
// 		args = append(args, "LIMIT", offset, count)
// 	}
// 	v, e := c.CallN(RetryTimes,"ZREVRANGEBYLEX", args...)
// 	if e != nil {
// 		return nil, e
// 	}
// 	return v.([]interface{}), nil
// }

func (c *Conn) ZRANGEBYSCORE(key string, min, max interface{}, withScores, limit bool, offset, count interface{}) ([]interface{}, error) {
	args := make([]interface{}, 3)
	args[0] = key
	args[1] = min
	args[2] = max
	if withScores {
		args = append(args, "WITHSCORES")
	}
	if limit {
		args = append(args, "LIMIT", offset, count)
	}
	v, e := c.CallN(RetryTimes, "ZRANGEBYSCORE", args...)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	return v.([]interface{}), nil
}

// if key,or member not exists return bulk string nil, else return integer
func (c *Conn) ZRANK(key, member string) (int64, error) {
	n, e := c.CallN(RetryTimes, "ZRANK", key, member)
	if e != nil {
		return -1, e
	}
	if _, ok := n.(int64); ok {
		return n.(int64), nil
	}

	return -1, ErrKeyNotExist
}

func (c *Conn) ZREM(key string, members []string) (int64, error) {
	args := make([]interface{}, 1+len(members))
	args[0] = key
	i := 1
	for _, m := range members {
		args[i] = m
		i++
	}
	n, e := c.CallN(RetryTimes, "ZREM", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

// since 2.8.9
// func (c *Conn) ZREMRANGEBYLEX(key, min, max string) (int64, error) {
// 	n, e := c.CallN(RetryTimes,"ZREMRANGEBYLEX", key, min, max)
// 	if e != nil {
// 		return -1, e
// 	}
// 	return n.(int64), nil
// }

func (c *Conn) ZREMRANGEBYRANK(key string, min, max interface{}) (int64, error) {
	n, e := c.CallN(RetryTimes, "ZREMRANGEBYRANK", key, min, max)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) ZREMRANGEBYSCORE(key string, min, max interface{}) (int64, error) {
	n, e := c.CallN(RetryTimes, "ZREMRANGEBYSCORE", key, min, max)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) ZREVRANGE(key string, start, stop int, withscores bool) ([]interface{}, error) {
	if withscores == true {
		v, e := c.CallN(RetryTimes, "ZREVRANGE", key, start, stop, "WITHSCORES")
		if e != nil {
			return nil, e
		}
		return v.([]interface{}), nil
	}
	v, e := c.CallN(RetryTimes, "ZREVRANGE", key, start, stop)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}
	return v.([]interface{}), nil
}

func (c *Conn) ZREVRANGEBYSCORE(key string, max, min interface{}, withScores, limit bool, offset, count interface{}) ([]interface{}, error) {
	args := make([]interface{}, 3)
	args[0] = key
	args[1] = max
	args[2] = min
	if withScores {
		args = append(args, "WITHSCORES")
	}
	if limit {
		args = append(args, "LIMIT", offset, count)
	}
	v, e := c.CallN(RetryTimes, "ZREVRANGEBYSCORE", args...)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}
	return v.([]interface{}), nil
}

func (c *Conn) ZREVRANK(key, member string) (int64, error) {
	n, e := c.CallN(RetryTimes, "ZREVRANK", key, member)
	if e != nil {
		return -1, e
	}
	if _, ok := n.(int64); ok {
		return n.(int64), nil
	}

	return -1, ErrKeyNotExist
}

func (c *Conn) ZSCORE(key, member string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "ZSCORE", key, member)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) ZUNIONSTORE(destination string, numkeys int, keys []string, weights bool, ws []int, aggregate bool, ag string) (int64, error) {
	args := make([]interface{}, 2+numkeys)
	args[0] = destination
	args[1] = numkeys
	if len(keys) < numkeys {
		return -1, ErrBadArgs
	}
	for i := 0; i < numkeys; i++ {
		args[i+2] = keys[i]
	}
	if weights == true {
		if len(ws) < numkeys {
			return -1, ErrBadArgs
		}
		args = append(args, "WEIGHTS")
		for i := 0; i < numkeys; i++ {
			args = append(args, ws[i])
		}
	}

	if aggregate == true {
		args = append(args, "AGGREGATE", ag)
	}
	n, e := c.CallN(RetryTimes, "ZUNIONSTORE", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) ZSCAN(key string, cursor int, match bool, pattern string, isCount bool, count int) (int, []interface{}, error) {
	args := make([]interface{}, 0, 6)
	args = append(args, key, cursor)
	if match {
		args = append(args, "MATCH", pattern)
	}
	if isCount {
		args = append(args, "COUNT", count)
	}
	v, e := c.CallN(RetryTimes, "ZSCAN", args...)
	if e != nil {
		return 0, nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return 0, nil, ErrResponseType
	}

	r := v.([]interface{})
	// return cursor
	rCursor, _ := strconv.Atoi(string(r[0].([]byte)))
	return rCursor, r[1].([]interface{}), nil
}

/******************* hyperLogLog *******************/
// since 2.8.9
func (c *Conn) PFADD(key string, elements []string) (int64, error) {
	args := make([]interface{}, 1+len(elements))
	args[0] = key
	i := 1
	for _, element := range elements {
		args[i] = element
		i++
	}
	n, e := c.CallN(RetryTimes, "PFADD", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) PFCOUNT(keys []string) (int64, error) {
	args := make([]interface{}, len(keys))
	i := 0
	for _, key := range keys {
		args[i] = key
		i++
	}
	n, e := c.CallN(RetryTimes, "PFCOUNT", args...)
	if e != nil {
		return -1, e
	}

	if _, ok := n.(int64); !ok {
		return -1, ErrResponseType
	}

	return n.(int64), nil
}

func (c *Conn) PFMERGE(destKey string, sourceKeys []string) ([]byte, error) {
	args := make([]interface{}, 1+len(sourceKeys))
	args[0] = destKey
	i := 1
	for _, sourceKey := range sourceKeys {
		args[i] = sourceKey
		i++
	}
	n, e := c.CallN(RetryTimes, "PFMERGE", args...)
	if e != nil {
		return nil, e
	}

	if _, ok := n.([]byte); !ok {
		return nil, ErrResponseType
	}

	return n.([]byte), nil
}

/******************* scripting *******************/
func (c *Conn) EVAL(script string, numkeys int, keys []string, scriptArgs []string) (interface{}, error) {
	args := make([]interface{}, 2+numkeys+len(scriptArgs))
	args[0] = script
	args[1] = numkeys
	i := 2
	for _, key := range keys {
		args[i] = key
		i++
	}
	for _, scriptArg := range scriptArgs {
		args[i] = scriptArg
		i++
	}

	v, e := c.CallN(RetryTimes, "EVAL", args...)
	if e != nil {
		return nil, e
	}
	return v, nil
}

func (c *Conn) EVALSHA(sha1 string, numkeys int, keys []string, scriptArgs []string) (interface{}, error) {
	args := make([]interface{}, 2+numkeys+len(scriptArgs))
	args[0] = sha1
	args[1] = numkeys
	i := 2
	for _, key := range keys {
		args[i] = key
		i++
	}
	for _, scriptArg := range scriptArgs {
		args[i] = scriptArg
		i++
	}

	v, e := c.CallN(RetryTimes, "EVALSHA", args...)
	if e != nil {
		return nil, e
	}
	return v, nil
}

func (c *Conn) SCRIPTEXISTS(scripts []string) ([]interface{}, error) {
	args := make([]interface{}, 1+len(scripts))
	args[0] = "EXISTS"
	i := 1
	for _, script := range scripts {
		args[i] = script
		i++
	}
	v, e := c.CallN(RetryTimes, "SCRIPT", args...)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, ErrResponseType
	}

	return v.([]interface{}), nil
}

func (c *Conn) SCRIPTFLUSH() ([]byte, error) {
	v, e := c.CallN(RetryTimes, "SCRIPT", "FLUSH")
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) SCRIPTKILL() ([]byte, error) {
	v, e := c.CallN(RetryTimes, "SCRIPT", "KILL")
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}

func (c *Conn) SCRIPTLOAD(script string) ([]byte, error) {
	v, e := c.CallN(RetryTimes, "SCRIPT", "LOAD", script)
	if e != nil {
		return nil, e
	}

	if _, ok := v.([]byte); !ok {
		return nil, ErrResponseType
	}

	return v.([]byte), nil
}
