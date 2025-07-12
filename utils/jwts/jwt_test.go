package jwts

import (
	"fmt"
	"testing"
)

func TestGenToken(t *testing.T) {
	token, err := GenToken(JwtPayLoad{
		UserID:   1,
		Role:     1,
		Username: "jaory",
	}, "123456", 5)
	fmt.Println(token, err)
}

func TestParseToken(t *testing.T) {
	payload, err := ParseToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6Imphb3J5Iiwicm9sZSI6MSwiZXhwIjoxNzUxNTczODQ2fQ.Ij483KU3Pcq-ncvegjA5QtAwwcWukvTz6EB1d6RW5l4", "123456")
	fmt.Println(payload, err)
}
