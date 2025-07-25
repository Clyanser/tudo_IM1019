package utils

import (
	"fmt"
	"testing"
)

func TestMD5(t *testing.T) {
	md := MD5([]byte("123456"))
	fmt.Println(md)
}
