package logger

import (
	"encoding/json"
	"hachimi/pkg/types"
	"io"
	"sync"
	"time"
)

type JSONLLogger struct {
	logChan  chan Loggable
	writer   io.Writer
	wg       sync.WaitGroup
	buffer   []Loggable
	maxSize  int
	mu       sync.Mutex
	nodeName string
}

// NewJSONLLogger 创建 JSONLLogger
func NewJSONLLogger(output io.Writer, bufferSize int, nodeName string) *JSONLLogger {
	logger := &JSONLLogger{
		logChan:  make(chan Loggable, 100),
		writer:   output,
		maxSize:  bufferSize,
		buffer:   make([]Loggable, 0, bufferSize),
		nodeName: nodeName,
	}
	logger.wg.Add(1)
	go logger.processLogs()
	return logger
}

func (j *JSONLLogger) processLogs() {
	defer j.wg.Done()

	ticker := time.NewTicker(1 * time.Second) // 每 1 秒强制写入一次
	defer ticker.Stop()

	for {
		select {
		case log, ok := <-j.logChan:
			if !ok {
				// 通道关闭，写入剩余日志
				j.mu.Lock()
				j.Flush()
				j.mu.Unlock()
				return
			}
			// 收到新日志，加入缓冲区
			j.mu.Lock()
			j.buffer = append(j.buffer, log)
			// 如果缓冲区已满，触发写入
			if len(j.buffer) >= j.maxSize {
				j.Flush()
			}
			j.mu.Unlock()
		case <-ticker.C:
			// 定时器触发，写入缓冲区中的日志
			j.mu.Lock()
			j.Flush()
			j.mu.Unlock()
		}
	}
}

func (j *JSONLLogger) Flush() {
	for _, log := range j.buffer {
		jsonData, _ := json.Marshal(types.HoneyData{Type: log.Type(), Data: log, Time: time.Now().Unix(), NodeName: j.nodeName})
		j.writer.Write(append(jsonData, '\n'))
	}
	j.buffer = j.buffer[:0]
}

func (j *JSONLLogger) Log(data Loggable) error {
	j.logChan <- data
	return nil
}

func (j *JSONLLogger) Close() error {
	close(j.logChan)
	j.wg.Wait()
	// 判断 writer 是否是文件
	if closer, ok := j.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
