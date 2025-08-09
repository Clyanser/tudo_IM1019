package utils

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/zeromicro/go-zero/core/logx"
	"regexp"
	"strings"
)

func InList(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
func InlistByRegx(list []string, s string) (ok bool) {
	for _, item := range list {
		regex, err := regexp.Compile(item)
		if err != nil {
			logx.Error(err.Error())
			return
		}
		if regex.MatchString(s) {
			return true
		}
	}
	return false
}

func MD5(data []byte) string {
	hash := md5.New()
	hash.Write(data)
	cipherStr := hash.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// name.zjy.jpg
func GetFilePrefix(fileName string) (prefix string) {
	nameList := strings.Split(fileName, ".")
	for i := 0; i < len(nameList)-1; i++ {
		if i == len(nameList)-2 {
			prefix += nameList[i]
			continue
		} else {
			prefix += nameList[i] + "."
		}
	}
	return
}

// Unique 去除切片中的重复元素，保持原始顺序
// T 必须是可比较的类型（int, string, struct{comparable fields} 等）
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{}) // 用 struct{} 节省空间
	var result []T

	for _, v := range slice {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}
