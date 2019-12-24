package main

import (
	"fmt"

	"github.com/huajiao-tv/pepperbus/utility/msgRedis"
)

func main() {
	p := msgRedis.NewPool("127.0.0.1:10000", "", 10, 10, 10)
	c := p.Pop()
	fmt.Println(c.Call("HMGET", "a", "a"))

}
