package msgRedis

import (
	"fmt"
	"runtime"
	// "strconv"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func TestMultiOnce(t *testing.T) {
	addresses := []string{"10.16.15.121:9991:1234567890"}
	addr := "10.16.15.121:9991:1234567890"
	mp := NewMultiPool(addresses, 50, 20, 9)

	go func() {
		for {
			fmt.Println("-----------" + mp.Info())
		}
	}()

	time.Sleep(time.Second)

	for i := 0; i < 100; i++ {
		mp.Call(addr).SET("a", "a")
	}

	time.Sleep(time.Second)

	for i := 0; i < 100; i++ {
		c := mp.PopByAddr(addr)
		if c != nil {
			c.SET("a", "a")
		} else {
			fmt.Println("c==nil")
		}
		mp.Push(c)
	}

	time.Sleep(10e9)
}

func TestMultiPush(t *testing.T) {
	addresses := []string{"10.16.15.121:9931:1234567890", "10.16.15.121:9991:1234567890"}
	addr := "10.16.15.121:9991:1234567890"
	// addr := "10.16.15.121:9731"
	mp := NewMultiPool(addresses, 50, 20, 9)

	c := mp.PopByAddr(addr)
	time.Sleep(1 * time.Second)
	mp.Push(c)

	// fmt.Println(mp.AddPool("10.16.15.121:9901", 10, 60))
	// fmt.Println(mp.servers)
	// fmt.Println(mp.pools)
	// fmt.Println(mp.Info())
	// go func() {
	// 	for i := 0; i < 29; i++ {
	// 		fmt.Println(mp.Info())
	// 		time.Sleep(time.Second)
	// 	}
	// }()

	// for i := 0; i < 50; i++ {
	// 	go func() {
	// 		c := mp.PopByAddr(addr)
	// 		time.Sleep(1 * time.Second)
	// 		mp.Push(c)
	// 	}()
	// }
	time.Sleep(30e9)
}

func TestMultiPop(t *testing.T) {
	addresses := []string{"10.16.15.121:9731", "10.16.15.121:9991:1234567890"}
	// addr := "10.16.15.121:9991@1234567890"
	addr := "10.16.15.121:9731"
	mp := NewMultiPool(addresses, 50, 20, 9)
	fmt.Println(mp.AddPool("10.16.15.121:9901", 10, 20, 60))
	fmt.Println(mp.servers)
	go func() {
		for i := 0; i < 29; i++ {
			fmt.Println(mp.Info())
			// time.Sleep(time.Second)
			mp.AddPool("10.16.15.121:9901", 10, 10, 60)
			// mp.DelPool("10.16.15.121:9901")
			mp.ReplacePool("10.16.15.121:9901", "10.16.15.121:9801", 20, 10, 8)
		}
	}()

	// }
	for i := 0; i < 50; i++ {
		go func() {
			c := mp.PopByAddr(addr)
			if c == nil {
				t.Error("c==nil....................")
				return
			}

			fmt.Println(c.SET("key", "value"))
			mp.Push(c)
		}()
		time.Sleep(time.Duration(5e9))
	}
}

func TestMultiPool(t *testing.T) {
	addresses := []string{"10.16.15.121:9731", "10.16.15.121:9991:1234567890"}
	addr := "10.16.15.121:9991:1234567890"
	mp := NewMultiPool(addresses, 20, 10, 20)
	fmt.Println(mp.AddPool("10.16.15.121:9901", 10, 10, 60))
	go func() {
		for i := 0; i < 29; i++ {
			mp.Info()
			time.Sleep(time.Second)
		}
	}()

	var g sync.WaitGroup
	for i := 0; i < 50; i++ {
		g.Add(1)
		go func() {
			defer g.Done()
			time.Sleep(1000)
			c := mp.PopByAddr(addr)
			if c == nil {
				t.Error("c==nil....................")
				return
			}
			// c.PipeSend("set", strconv.Itoa(i), strconv.Itoa(i))
			// c.PipeExec()
			// fmt.Println("PING")
			n := rand.Intn(10)
			c.CallN(3, "PING")
			time.Sleep(time.Duration(n * 1e9))
			// mp.PushByAddr(addr, c)
			mp.Push(c)
		}()
	}
	g.Wait()
	time.Sleep(30e9)
}

func TestPushAndPop(t *testing.T) {
	addresses := []string{"10.16.15.121:9731", "10.16.15.121:9991:1234567890"}
	addr := "10.16.15.121:9991:1234567890"
	mp := NewMultiPool(addresses, 20, 10, 20)

	start := time.Now()
	for i := 0; i < 10000; i++ {
		c := mp.PopByAddr(addr)
		if c == nil {
			t.Error("c==nil....................")
			return
		}
		c.SET("a", "a")
		mp.Push(c)
	}
	fmt.Println("push and pop costs=", time.Now().Sub(start).String())

	start = time.Now()
	c := mp.PopByAddr(addr)
	for i := 0; i < 10000; i++ {
		if c == nil {
			t.Error("c==nil....................")
			return
		}
		c.SET("a", "a")
	}
	fmt.Println("no push=", time.Now().Sub(start).String())

}
