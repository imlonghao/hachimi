package types

import (
	"encoding/json"
	"time"
)

type RedisSession struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"index" json:"session_id"`
	Protocol  string    `gorm:"index" json:"protocol"`
	StartTime time.Time `gorm:"index" json:"start_time"`
	EndTime   time.Time `gorm:"index" json:"end_time"`
	Duration  int       `json:"duration"`
	SrcIP     string    `gorm:"index" json:"src_ip"`
	SrcPort   int       `gorm:"index" json:"src_port"`
	DstIP     string    `gorm:"index" json:"dst_ip"`
	DstPort   int       `gorm:"index" json:"dst_port"`
	Error     bool      `json:"error"`
	Service   string    `json:"service"`
	Data      string    `json:"data"`
	User      string    `json:"user"`
	PassWord  string    `json:"password"`
}

// TableName 设置表名
func (s *RedisSession) TableName() string {
	return "redis_logs"
}
func (s *RedisSession) ToMap() (map[string]interface{}, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}
