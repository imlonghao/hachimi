package types

import (
	"bytes"
	"time"
)

//TODO JA3

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
	Duration  int `json:"duration"`
	inBuffer  *bytes.Buffer
	outBuffer *bytes.Buffer
}

func (h Http) Type() string {
	return "http_session"
}
