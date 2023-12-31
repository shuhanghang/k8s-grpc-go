package utils

import (
	"log"
	"net"
)

func GetIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalln(err.Error())
		return ""
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// fmt.Println(ipnet.IP.String())
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
