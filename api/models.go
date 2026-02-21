package api

import (
	"time"
)

type Message struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
}
