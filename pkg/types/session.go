package types

import (
	"bytes"
	"encoding/json"
	"time"
)

type Session struct {
	ID         string `gorm:"primaryKey" json:"id"`
	Protocol   string `gorm:"index" json:"protocol"`
	connection interface{}
	addr       interface{}
	StartTime  time.Time `gorm:"index" json:"start_time"`
	EndTime    time.Time `gorm:"index" json:"end_time"`
	SrcIP      string    `gorm:"index" json:"src_ip"`
	SrcPort    int       `gorm:"index" json:"src_port"`
	DstIP      string    `gorm:"index" json:"dst_ip"`
	DstPort    int       `gorm:"index" json:"dst_port"`
	IsTls      bool      `gorm:"index" json:"is_tls"`
	IsGmTls    bool      `gorm:"index" json:"is_gm_tls"`
	IsHandled  bool      `gorm:"index" json:"is_handled"`
	IsHttp     bool      `gorm:"index" json:"is_http"`
	Data       string    `gorm:"index" json:"data"`
	Service    string    `gorm:"index" json:"service"`
	//RawData    string    `gorm:"index" json:"raw_data"`
	//经过的时间 ms
	Duration  int
	inBuffer  *bytes.Buffer
	outBuffer *bytes.Buffer
}

// TableName 设置表名
func (s *Session) TableName() string {
	return "sessions"
}
func (s *Session) SetConnection(conn interface{}) {
	s.connection = conn

}
func (s *Session) SetAddr(addr interface{}) {
	s.addr = addr
}
func (s *Session) SetInBuffer(buffer *bytes.Buffer) {
	s.inBuffer = buffer
}
func (s *Session) SetOutBuffer(buffer *bytes.Buffer) {
	s.outBuffer = buffer
}
func (s *Session) GetOutBuffer() *bytes.Buffer {
	return s.outBuffer
}
func (s *Session) Close() {
	if s.connection != nil {
		if conn, ok := s.connection.(interface{ Close() error }); ok {
			conn.Close()
		}
	}
}

func (s *Session) ToMap() (map[string]interface{}, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}
