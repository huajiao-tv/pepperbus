package msgRedis

import (
	"fmt"
	"testing"
	"time"
)

func TestSigle(t *testing.T) {
	p := NewPool("10.16.15.121:9731", "", 10, 10, 10)
	c := p.Pop()
	// fmt.Println(c.CallN(2, "HMGET", "a", "a"))

	fmt.Println(c.CallN(2, "HSET", "zkey1106", "tagList", "sss"))

	kv1 := make(map[string]interface{})
	kv1["key2"] = "keys2"
	kv1["key3"] = 1234

	key := "asfset"

	fmt.Println(c.HSET(key, "field", "value"))
	// fmt.Println(c.HMGET(key+"adf", []string{"sfjk", "noexists"}))
	// fmt.Println(c.HGETALLMAP(key))
}

func TestPipeSmallBuffer(t *testing.T) {
	c, e := Dial("10.16.15.121:9731", "", ConnectTimeout, ReadTimeout, WriteTimeout, false, nil)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()
	// test pipeline
	for i := 0; i < 1000; i++ {
		c.PipeSend("INCR", "Zincr")
	}
	fmt.Println(c.PipeExec())
}

func TestPoolNoAuth(t *testing.T) {
	p := NewPool("10.16.15.121:9731", "", 10, 10, 10)
	connSlice := make([]*Conn, 0, 25)
	for i := 0; i < 5; i++ {
		fmt.Println("Active=", p.Actives())
		fmt.Println("Idles=", p.Idles())
		c := p.Pop()
		connSlice = append(connSlice, c)
	}

	fmt.Println("Conn slice len:", len(connSlice))
	for _, c := range connSlice {
		fmt.Println("Active=", p.Actives())
		fmt.Println("Idles=", p.Idles())
		p.Push(c)
	}
}

func TestPoolAuth(t *testing.T) {
	p := NewPool("10.16.15.121:9991", "1234567890", 10, 10, 10)
	fmt.Println(p.Actives())
	fmt.Println(p.Idles())
	c := p.Pop()
	if c == nil {
		fmt.Println("Pop nil")
		return
	}
	defer p.Push(c)
	// fmt.Println(c.Info())
	fmt.Println(c.Call("SET", "zyh0924", "abcdefghijklmnopqrstuvwxyz"))
	fmt.Println(c.Call("GET", "zyh0924"))
}

func TestWrite(t *testing.T) {
	c, e := Dial("10.16.15.121:9731", "", ConnectTimeout, ReadTimeout, WriteTimeout, false, nil)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()

	// test commands
	key := "zyh1008"
	args := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}
	fmt.Println(c.SET(key, "zzzaaa"))
	fmt.Println(c.GET(key))
	return

	fmt.Println(c.SADD(key, args))
	fmt.Println(c.SMEMBERS(key))
	fmt.Println(c.DEL(key))

	// test pipeline
	c.PipeSend("SET", "a", "zyh")
	c.PipeSend("SET", "b", "zyh")
	c.PipeSend("SET", "c", "zyh")
	fmt.Println(c.PipeExec())

	// test transaction
	c.MULTI()
	c.TransSend("SET", "a", "zyh2")
	c.TransSend("SET", "b", "zyh3")
	fmt.Println(c.TransExec())
}
func TestCommands(t *testing.T) {
	c, e := Dial("10.16.15.121:9731", "", ConnectTimeout, ReadTimeout, WriteTimeout, false, nil)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()

	fmt.Println("SETS.**************************SETS*********************.SETS")
	keyz := "setsZYH"
	key1z := "setsZYH1"
	fmt.Println(c.SADD(keyz, []string{"a", "b", "c"}))
	fmt.Println(c.SADD(key1z, []string{"c", "d", "e"}))
	fmt.Println(c.SREM(keyz, []string{"a", "b", "c", "z"}))
	return

	fmt.Println("STRINGS.************************STRINGS**********************.STRINGS")
	fmt.Println(c.SET("key", "value"))
	fmt.Println(c.OBJECT("encoding", "key"))
	fmt.Println(c.DEL("keysa"))
	fmt.Println(c.EXISTS("key"))
	fmt.Println(c.EXPIRE("key", 1000))
	fmt.Println(c.EXPIREAT("key", time.Now().Unix()+100))
	fmt.Println(c.KEYS("*"))
	fmt.Println(c.MIGRATE("10.16.15.121", "9731", "key", "1", 5000, true, true))
	fmt.Println(c.SELECT(1))
	fmt.Println(c.MOVE("key", "3"))
	fmt.Println(c.PERSIST("key"))
	fmt.Println(c.PEXPIRE("key", 10e6))
	fmt.Println(c.PEXPIREAT("key", time.Now().Unix()*1000+100000))
	fmt.Println(c.PTTL("key"))
	fmt.Println(c.RANDOMKEY())
	fmt.Println(c.SET("key", "value"))
	fmt.Println(c.RENAME("key", "newkey"))
	fmt.Println(c.RENAMENX("newkey", "key"))
	fmt.Println(c.DUMP("key"))
	b, _ := c.DUMP("key")
	fmt.Println(c.RESTORE("key111", 999, string(b)))
	fmt.Println(c.TTL("key"))
	fmt.Println(c.TYPE("key"))
	// return

	key := "zyh1009"
	fmt.Println("STRINGS.************************STRINGS**********************.STRINGS")
	fmt.Println(c.Call("FLUSHDB"))
	fmt.Println(c.GET(key))
	fmt.Println(c.SET(key, "zyh1009"))
	fmt.Println(c.GET(key))
	fmt.Println(c.SETBIT(key, 10, 0))
	fmt.Println(c.GETBIT(key, 10))
	fmt.Println(c.GETRANGE(key, 0, 3))
	fmt.Println(c.MGET([]string{key, "key5"}))
	kv := make(map[string]string)
	kv["key2"] = "key2a"
	kv["key3"] = "key3a"
	fmt.Println(c.MSET(kv))
	fmt.Println(c.MSETNX(kv))
	fmt.Println(c.PSETEX(key, 100000, "abc"))
	fmt.Println(c.SETEX(key, 110, "abc1"))
	fmt.Println(c.SETRANGE(key, 1, "aaaa"))
	fmt.Println(c.STRLEN(key))

	// hashes
	key = "hashesZYH"
	fields := []string{"f1", "f2"}
	fmt.Println("HASHES.************************HASHES**********************.HASHES")
	fmt.Println(c.HDEL(key, fields))
	fmt.Println(c.HEXISTS(key, fields[0]))
	fmt.Println(c.HEXISTS(key, "noexists"))
	fmt.Println(c.HGET(key, fields[0]))
	fmt.Println(c.HGET(key, "noexists"))
	fmt.Println(c.HGETALL(key))
	fmt.Println(c.HGETALL("noexists"))
	fmt.Println(c.HINCRBY(key, "field", 1))
	fmt.Println(c.HINCRBYFLOAT(key, "field1", 1.1))
	fmt.Println(c.HKEYS(key))
	fmt.Println(c.HLEN(key))
	fmt.Println(c.HMGET(key, []string{key, "noexists"}))

	kv1 := make(map[string]interface{})
	kv1["key2"] = "keys2"
	kv1["key3"] = "keyss3"

	fmt.Println(c.HMSET(key, kv1))
	fmt.Println(c.HSET(key, "fieldFloat", 1.5))
	fmt.Println(c.HSETNX(key, "fieldFloat11", 1.6))
	fmt.Println(c.HVALS(key))

	fmt.Println("LISTS.***********************LISTS***********************.LISTS")
	key = "listsZYH"
	key1 := "lists1"
	fmt.Println(c.BLPOP([]string{"key11", "key21"}, 1))
	fmt.Println(c.BRPOP([]string{"key11", "key12"}, 1))
	fmt.Println(c.LPUSH(key, []string{"a", "b", "c"}))
	fmt.Println(c.RPUSH(key, []string{"x", "y", "z", "a"}))
	fmt.Println(c.BRPOPLPUSH(key, key1, 1))
	fmt.Println(c.LINDEX(key, 4))
	fmt.Println(c.LINSERT(key, "after", "a", "d"))
	fmt.Println(c.LLEN(key))
	fmt.Println(c.LPUSHX("noexists", "value"))
	fmt.Println(c.LRANGE(key, 0, 10))
	fmt.Println(c.LREM(key, -1, "a"))
	fmt.Println(c.LSET(key, 0, "0"))
	fmt.Println(c.LTRIM(key, 1, -1))
	fmt.Println(c.RPOP(key))
	fmt.Println(c.RPOPLPUSH(key, key1))
	fmt.Println(c.RPUSHX(key, "value"))

	fmt.Println("SETS.**************************SETS*********************.SETS")
	key = "setsZYH"
	key1 = "setsZYH1"
	keys := []string{key, key1}
	fmt.Println(c.SADD(key, []string{"a", "b", "c"}))
	fmt.Println(c.SADD(key1, []string{"c", "d", "e"}))
	fmt.Println(c.SREM(key, []string{"a"}))
	fmt.Println(c.SISMEMBER(key, "b"))
	fmt.Println(c.SMEMBERS(key))
	fmt.Println(c.SCARD(key))
	fmt.Println(c.SINTER([]string{key, key1}))
	fmt.Println(c.SINTERSTORE("interSets", []string{key, key1}))
	fmt.Println(c.SDIFF([]string{key, key1}))
	fmt.Println(c.SDIFFSTORE("diffSets", []string{key, key1}))
	fmt.Println(c.SMOVE("srcKey", "desKey", "c"))
	fmt.Println(c.SPOP(key))

	fmt.Println("SRANDMEMBER........")
	fmt.Println(c.SRANDMEMBER(key, 5))
	fmt.Println(c.SUNION(keys))
	fmt.Println(c.SUNIONSTORE("unionSets", keys))

	fmt.Println("SORTED SETS.********************SORTED SETS******************.SORTED SETS")
	key = "zsetsZYH"
	key1 = "zsetsZYH1"
	keyScores := make(map[string]interface{}, 4)
	keyScores["k1"] = 1.12
	keyScores["k2"] = 2
	keyScores["k3"] = 19
	fmt.Println(c.ZADD(key, keyScores))
	fmt.Println(c.ZCARD(key))
	fmt.Println(c.ZCOUNT(key, 1, 3))
	fmt.Println(c.ZINCRBY(key, -0.9, "k1"))
	fmt.Println(c.ZADD(key1, keyScores))
	fmt.Println(c.ZINTERSTORE("out1", 2, []string{key, key1}, true, []int{2, 3}, true, "min"))
	// fmt.Println(c.ZLEXCOUNT(key, "-", "+"))
	fmt.Println(c.ZRANGE(key, 0, 2, true))
	// fmt.Println(c.ZRANGEBYLEX(key, "[a", "[z", true, 0, 1))
	// fmt.Println(c.ZREVRANGEBYLEX(key, "(z", "[a", true, 0, 4))
	fmt.Println(c.ZRANGEBYSCORE(key, 0, 3.1, true, true, 0, 5))
	fmt.Println(c.ZRANK(key, "k1"))
	fmt.Println(c.ZREM(key1, []string{"k2"}))
	// fmt.Println(c.ZREMRANGEBYLEX(key, "-", "+"))
	fmt.Println(c.ZREMRANGEBYSCORE(key, 0, 1))
	fmt.Println(c.ZREVRANGE(key, 0, -1, true))
	fmt.Println(c.ZREVRANGEBYSCORE(key, 4, -1.1, true, true, 0, 7))
	fmt.Println(c.ZREVRANK(key, "k2"))
	fmt.Println(c.ZSCORE(key, "k3"))

	fmt.Println(c.ZADD("key", keyScores))
	fmt.Println(c.ZADD("key1", keyScores))
	fmt.Println(c.ZUNIONSTORE("out2", 2, []string{"key", "key1"}, false, nil, false, ""))
}

func TestQPS(t *testing.T) {
	p := NewPool("10.16.15.121:9731", "", 10, 10, 10)
	for i := 0; i < 50; i++ {
		c := p.Pop()
		if c != nil {
			go call(c)
		}
	}

	go func() {
		for {
			fmt.Println("QPS:", p.QPS())
		}
	}()

	go func() {
		for {
			fmt.Println("QPSAvg:", p.QPSAvg())
		}
	}()

	// select {}
	time.Sleep(15e9)
}

func call(c *Conn) {
	for {
		key := "zyh1009"
		// c.Call("FLUSHDB")
		c.GET(key)
		c.SET(key, "zyh1009")
		c.GET(key)
		c.SETBIT(key, 10, 0)
		c.GETBIT(key, 10)
		c.GETRANGE(key, 0, 3)
		c.MGET([]string{key, "key5"})
		kv := make(map[string]string)
		kv["key2"] = "key2a"
		kv["key3"] = "key3a"
		c.MSET(kv)
		c.MSETNX(kv)
		c.PSETEX(key, 100000, "abc")
		c.SETEX(key, 110, "abc1")
		c.SETRANGE(key, 1, "aaaa")
		c.STRLEN(key)

		// hashes
		key = "hashesZYH"
		fields := []string{"f1", "f2"}
		c.HDEL(key, fields)
		c.HEXISTS(key, "noexists")
		c.HGET(key, fields[0])
		c.HGET(key, "noexists")
		c.HGETALL(key)
		c.HGETALL("noexists")
		c.HINCRBY(key, "field", 1)
		c.HINCRBYFLOAT(key, "field1", 1.1)
		c.HKEYS(key)
		c.HLEN(key)
		c.HMGET(key, []string{key, "noexists"})

		kv1 := make(map[string]interface{})
		kv1["key2"] = "keys2"
		kv1["key3"] = "keyss3"

		c.HMSET(key, kv1)
		c.HSET(key, "fieldFloat", 1.5)
		c.HSETNX(key, "fieldFloat11", 1.6)
		c.HVALS(key)

	}
}

func TestScript(t *testing.T) {
	p := NewPool("10.16.15.121:9731", "", 10, 10, 10)
	c := p.Pop()
	if c == nil {
		return
	}

	scriptA := `
		local n = redis.call("exists",KEYS[1])
		if n == 1 then
			redis.call("expire",KEYS[1],ARGV[1])
		else 
			redis.call("set", KEYS[1], ARGV[2])
		end

		local ttl = redis.call("ttl",KEYS[1])
		local key = redis.call("get",KEYS[1])

		local rTable = {}
		rTable[1] = ttl
		rTable[2] = key
		return rTable
	`
	sha1, e := c.SCRIPTLOAD(scriptA)
	if e != nil {
		fmt.Println("script load error = ", e.Error())
		return
	}

	sha := string(sha1)

	p.AddScriptSha1("true", sha)
	fmt.Println("script sha1 = ", sha)

	v, e := c.EVAL(scriptA, 1, []string{"luaKey"}, []string{"123", "luaValue"})
	if e != nil {
		fmt.Println("eval = ", e.Error())
		return
	}
	fmt.Println("EVAL = ", v)
	return
	v, e = c.EVALSHA(p.GetScriptSha1("true"), 1, []string{"luaKey"}, []string{"123", "luaValue"})
	if e != nil {
		fmt.Println("evalsha = ", e.Error())
		return
	}
	fmt.Println("EVALSHA = ", v)
	fmt.Println(c.SCRIPTEXISTS([]string{sha, "avd"}))
	fmt.Println(c.SCRIPTKILL())
	fmt.Println(c.SCRIPTFLUSH())
}

func TestSCAN(t *testing.T) {
	c, e := Dial("10.16.15.121:9731", "", ConnectTimeout, ReadTimeout, WriteTimeout, false, nil)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()

	// test scan
	fmt.Println("........................TESTING SCAN.......................")
	cursor, elements, e := c.SCAN(0, false, "", false, 0)
	if e != nil {
		fmt.Println(e)
		return
	}

	eles := []string{}
	for _, iface := range elements {
		eles = append(eles, string(iface.([]byte)))
	}
	fmt.Printf("CURSOR=%d, elements=%v\n", cursor, eles)
	for cursor != 0 {
		cursor, elements, e = c.SCAN(cursor, false, "", false, 0)
		if e != nil {
			fmt.Println(e)
			return
		}
		eles := []string{}
		for _, iface := range elements {
			eles = append(eles, string(iface.([]byte)))
		}
		fmt.Printf("CURSOR=%d, elements=%v\n", cursor, eles)
	}

	// test sscan
	fmt.Println("........................TESTING sets SCAN.......................")
	key := "setsZYH"
	cursor, elements, e = c.SSCAN(key, 0, true, "*", true, 15)
	if e != nil {
		fmt.Println(e)
		return
	}

	eles = []string{}
	for _, iface := range elements {
		eles = append(eles, string(iface.([]byte)))
	}

	fmt.Printf("CURSOR=%d, elements=%v\n", cursor, eles)
	for cursor != 0 {
		cursor, elements, e = c.SSCAN(key, cursor, true, "*", true, 15)
		if e != nil {
			fmt.Println(e)
			return
		}
		eles := []string{}
		for _, iface := range elements {
			eles = append(eles, string(iface.([]byte)))
		}
		fmt.Printf("CURSOR=%d, elements=%v\n", cursor, eles)
	}

	// test hscan
	fmt.Println("........................TESTING Hashes SCAN.......................")
	key = "hashesZYH"
	cursor, elements, e = c.HSCAN(key, 0, true, "*1*", false, 15)
	if e != nil {
		fmt.Println(e)
		return
	}

	eles = []string{}
	for _, iface := range elements {
		eles = append(eles, string(iface.([]byte)))
	}

	fmt.Printf("CURSOR=%d, elements=%v\n", cursor, eles)
	for cursor != 0 {
		cursor, elements, e = c.HSCAN(key, cursor, true, "*", false, 15)
		if e != nil {
			fmt.Println(e)
			return
		}
		eles := []string{}
		for _, iface := range elements {
			eles = append(eles, string(iface.([]byte)))
		}
		fmt.Printf("CURSOR=%d, elements=%v\n", cursor, eles)
	}

	// test hscan
	fmt.Println("........................TESTING Sorted Sets SCAN.......................")
	key = "zsetsZYH"
	cursor, elements, e = c.ZSCAN(key, 0, true, "*1*", false, 15)
	if e != nil {
		fmt.Println(e)
		return
	}

	eles = []string{}
	for _, iface := range elements {
		eles = append(eles, string(iface.([]byte)))
	}

	fmt.Printf("CURSOR=%d, elements=%v\n", cursor, eles)
	for cursor != 0 {
		cursor, elements, e = c.ZSCAN(key, cursor, true, "*1*", false, 15)
		if e != nil {
			fmt.Println(e)
			return
		}
		eles := []string{}
		for _, iface := range elements {
			eles = append(eles, string(iface.([]byte)))
		}
		fmt.Printf("CURSOR=%d, elements=%v\n", cursor, eles)
	}
}

func BenchmarkStrings(b *testing.B) {
	c, e := Dial("10.16.15.121:9731", "", ConnectTimeout, ReadTimeout, WriteTimeout, false, nil)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SET("key", "valuevaluevaluevaluevalue")
	}
}

func BenchmarkSets(b *testing.B) {
	c, e := Dial("10.16.15.121:9731", "", ConnectTimeout, ReadTimeout, WriteTimeout, false, nil)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SADD("keySets", []string{"valuevaluevaluevaluevalue"})
	}
}

func BenchmarkHashes(b *testing.B) {
	c, e := Dial("10.16.15.121:9731", "", ConnectTimeout, ReadTimeout, WriteTimeout, false, nil)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()

	kv := make(map[string]interface{})
	kv["f1"] = "f1"
	kv["f2"] = "f2"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.HMSET("keyHashes", kv)
	}
}

//
//
//
//
//
//

//
//
