package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/johntech-o/idgen"
	"git.huajiao.com/qmessenger/pepperbus/utility"
)

var ServiceId string

func GetInnerIP() string {
	ips := utility.GetLocalIP()
	prefixList := netGlobalConf().InnerIpPrefix

	for _, prefix := range prefixList {
		for _, ip := range ips {
			if ip[:len(prefix)] == prefix {
				return ip
			}
		}
	}

	return ""
}

func updateServiceId() {
	id := IP2Num(GetInnerIP())
	ServiceId = fmt.Sprintf("-%s-%s", strconv.FormatUint(id, 10), netGlobalConf().GatewayPort)
}

func GenTraceId() string {
	return strconv.FormatUint(idgen.GenId(), 10) + ServiceId
}

func GenId() string {
	return strconv.FormatUint(idgen.GenId(), 10)
}

type BusId struct {
	CreateTime  int64  `json:"createTime"`
	GatewayIp   string `json:"gatewayIp"`
	GatewayPort string `json:"gatewayPort"`
}

func ParseTraceId(id string) (*BusId, error) {
	busId := new(BusId)
	fields := strings.Split(id, "-")
	if len(fields) != 3 {
		return nil, errors.New("invalid id format: " + id)
	}
	idInt64, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return nil, err
	}
	busId.CreateTime = idgen.GetTimeUnixMill(idInt64)
	ipNum, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, err
	}
	busId.GatewayIp = Num2IP(ipNum)
	busId.GatewayPort = fields[2]
	return busId, nil
}

func IP2Num(requestip string) uint64 {
	//获取客户端地址的long
	nowip := strings.Split(requestip, ".")
	if len(nowip) != 4 {
		return 0
	}
	a, _ := strconv.ParseUint(nowip[0], 10, 64)
	b, _ := strconv.ParseUint(nowip[1], 10, 64)
	c, _ := strconv.ParseUint(nowip[2], 10, 64)
	d, _ := strconv.ParseUint(nowip[3], 10, 64)
	ipNum := a<<24 | b<<16 | c<<8 | d
	return ipNum
}

func Num2IP(ipnum uint64) string {
	byte1 := ipnum & 0xff
	byte2 := ipnum & 0xff00
	byte2 >>= 8
	byte3 := ipnum & 0xff0000
	byte3 >>= 16
	byte4 := ipnum & 0xff000000
	byte4 >>= 24
	result := strconv.FormatUint(byte4, 10) + "." +
		strconv.FormatUint(byte3, 10) + "." +
		strconv.FormatUint(byte2, 10) + "." +
		strconv.FormatUint(byte1, 10)
	return result
}
