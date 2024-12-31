package types

import (
	"bytes"
	"encoding/json"
	"time"
)

type Http struct {
	ID        string            `gorm:"primaryKey" json:"id"`
	SessionID string            `gorm:"index" json:"session_id"`
	Protocol  string            `gorm:"index" json:"protocol"`
	StartTime time.Time         `gorm:"index" json:"start_time"`
	EndTime   time.Time         `gorm:"index" json:"end_time"`
	SrcIP     string            `gorm:"index" json:"src_ip"`
	SrcPort   int               `gorm:"index" json:"src_port"`
	DstIP     string            `gorm:"index" json:"dst_ip"`
	DstPort   int               `gorm:"index" json:"dst_port"`
	IsTls     bool              `gorm:"index" json:"is_tls"`
	IsGmTls   bool              `gorm:"index" json:"is_gm_tls"`
	IsHandled bool              `gorm:"index" json:"is_handled"`
	Header    map[string]string `gorm:"type:string" json:"header"`
	UriParam  map[string]string `gorm:"type:string" json:"uri_param"`
	BodyParam map[string]string `gorm:"type:string" json:"body_param"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	UA        string            `json:"ua"`
	Host      string            `json:"host"`
	Body      string            `json:"body"`
	RawBody   string            `json:"raw_body"`
	Service   string            `json:"service"`
	//经过的时间 ms
	Duration  int
	inBuffer  *bytes.Buffer
	outBuffer *bytes.Buffer
}

// TableName 设置表名
func (s *Http) TableName() string {
	return "http_logs"
}
func (s *Http) ToMap() (map[string]interface{}, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}
