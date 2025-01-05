package types

import (
	"time"
)

type RedisSession struct {
	Session
	ID        string    `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"index" json:"session_id"`
	Protocol  string    `gorm:"index" json:"protocol"`
	StartTime time.Time `gorm:"index" json:"start_time"`
	EndTime   time.Time `gorm:"index" json:"end_time"`
	Duration  int       `json:"duration"`
	Error     bool      `json:"error"`
	Service   string    `json:"service"`
	Data      string    `json:"data"`
	User      string    `json:"user"`
	PassWord  string    `json:"password"`
}

func (r RedisSession) Type() string {
	return "redis_session"
}
