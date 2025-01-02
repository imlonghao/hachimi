package types

import (
	"bytes"
	"encoding/json"
	"time"
)

type Session struct {
	ID         string `json:"id"`
	Protocol   string `json:"protocol"`
	connection interface{}
	addr       interface{}
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	SrcIP      string    `json:"src_ip"`
	SrcPort    int       `json:"src_port"`
	DstIP      string    `json:"dst_ip"`
	DstPort    int       `json:"dst_port"`
	IsTls      bool      `json:"is_tls"`
	IsGmTls    bool      `json:"is_gm_tls"`
	IsHandled  bool      `json:"is_handled"`
	IsHttp     bool      `json:"is_http"`
	Data       string    `json:"data"`
	Service    string    `json:"service"`
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
