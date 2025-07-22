package etcd

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/netx"
	"strings"
	"tudo_IM1019/core"
)

// DeliveryAddress 上送服务地址
func DeliveryAddress(etcdAddr string, serviceName string, addr string) {
	list := strings.Split(addr, ":")
	if len(list) < 2 {
		logx.Errorf("地址出错 %s", addr)
		return
	}
	if list[0] == "0.0.0.0" {
		ip := netx.InternalIp()
		addr = strings.ReplaceAll(addr, "0.0.0.0", ip)
	}
	client := core.InitEtcd(etcdAddr)
	_, err := client.Put(context.Background(), serviceName, addr)
	if err != nil {
		logx.Errorf("地址上传失败 %s", err)
		return
	}
	logx.Infof("地址上传成功 %s %s", serviceName, addr)

}

// GetServiceAddr 地址获取
func GetServiceAddr(etcdAddr string, serviceName string) (addr string) {
	client := core.InitEtcd(etcdAddr)
	res, err := client.Get(context.Background(), serviceName)
	if err == nil && len(res.Kvs) > 0 {
		return string(res.Kvs[0].Value)
	}
	return ""
}
