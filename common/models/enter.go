package models

import "time"

type Model struct {
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
