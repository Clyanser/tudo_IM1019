package ips

import (
	"fmt"
	"net"
)

func GetIp() (addr string) {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("获取网卡信息出错:", err)
		return
	}

	// 遍历所有网络接口
	for _, iface := range interfaces {
		// 排除一些特殊接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// 获取接口的地址信息
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("获取ip地址出错", err)
			continue
		}

		// 遍历接口的地址
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					return ipNet.IP.String()
				}
			}
		}
	}
	return
}
