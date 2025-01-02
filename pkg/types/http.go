package types

import (
	"bytes"
	"encoding/json"
	"time"
)

type Http struct {
	Session
	ID        string            `json:"id"`
	SessionID string            `json:"session_id"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Header    map[string]string `json:"header"`
	UriParam  map[string]string `json:"uri_param"`
	BodyParam map[string]string `json:"body_param"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	UA        string            `json:"ua"`
	Host      string            `json:"host"`
	RawHeader string            `json:"raw_header"`
	Body      string            `json:"body"`
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
