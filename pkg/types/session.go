package types

import (
	"bytes"
	"net"
	"time"
)

// TODO JA3
type Session struct {
	ID         string `json:"id"`
	Protocol   string `json:"protocol"`
	connection interface{}
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
	Duration  int `json:"duration"`
	inBuffer  *bytes.Buffer
	outBuffer *bytes.Buffer
}

func (s *Session) SetConnection(conn interface{}) {
	s.connection = conn

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

// Close 关闭原始连接
func (s *Session) Close() {
	if s.connection != nil {
		//只关闭TCP  UDP不需要关闭 UDP是无状态的 关闭就会停止监听
		if conn, ok := s.connection.(*net.TCPConn); ok {
			conn.Close()
		}
	}
}
func (s Session) Type() string {
	return "session"
}
