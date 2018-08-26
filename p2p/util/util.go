package util

import (
	"cherrychain/common/clogging"
	"net"
)

var p2pUtilLogger = clogging.MustGetLogger("P2P")

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		p2pUtilLogger.Error("Can not get interfaceAddrs")
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "0.0.0.0", err
}
