package pwd

import (
	"fmt"
	"testing"
)

func TestCheckPwd(t *testing.T) {

}

func TestHashPwd(t *testing.T) {
	hash := HashPwd("123456")
	fmt.Println(hash)
}
