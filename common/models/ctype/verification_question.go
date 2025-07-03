package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type VerificationQuestion struct {
	Problem1 *string `json:"problem1"`
	Problem2 *string `json:"problem2"`
	Problem3 *string `json:"problem3"`
	Answer1  *string `json:"answer1"`
	Answer2  *string `json:"answer2"`
	Answer3  *string `json:"answer3"`
}

// Value 实现 driver.Valuer 接口
func (v VerificationQuestion) Value() (driver.Value, error) {
	return json.Marshal(v)
}

// Scan 实现 sql.Scanner 接口
func (v *VerificationQuestion) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}
	return json.Unmarshal(b, v)
}
