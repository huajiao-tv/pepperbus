package msgRedis

import (
	"errors"
)

type OnceConn struct {
	*Conn
}

func (mp *MultiPool) Call(address string) *OnceConn {
	oc := &OnceConn{}
	c := mp.PopByAddr(address)
	if c == nil {
		oc.Conn = &Conn{err: errors.New("get a nil conn address=" + address)}
		return oc
	}
	oc.Conn = c
	oc.Conn.isOnce = true
	return oc
}
