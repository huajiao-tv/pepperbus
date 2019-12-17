package utility

import (
	"fmt"
	"net"
	"os"
)

func InArray(v string, s []string) bool {
	for _, i := range s {
		if v == i {
			return true
		}
	}

	return false
}

func GetLocalIP() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ips := make([]string, 0, len(addrs))
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips
}
