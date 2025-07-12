package ips

import (
	"fmt"
	"testing"
)

func TestGetIp(t *testing.T) {
	ip := GetIp()
	fmt.Println(ip)
}
