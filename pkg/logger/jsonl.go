package logger

import (
	"encoding/json"
	"hachimi/pkg/types"
	"os"
	"sync"
	"time"
)

type JSONLLogger struct {
	logChan  chan Loggable
	writer   *os.File
	wg       sync.WaitGroup
	buffer   []Loggable
	maxSize  int
	mu       sync.Mutex
	nodeName string
}

// NewJSONLLogger 创建 JSONLLogger
func NewJSONLLogger(output string, bufferSize int, nodeName string) (*JSONLLogger, error) {
	var file *os.File
	var err error

	switch output {
	case "stdout":
		file = os.Stdout
	case "stderr":
		file = os.Stderr
	default:
		file, err = os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
	}

	logger := &JSONLLogger{
		logChan:  make(chan Loggable, 100),
		writer:   file,
		maxSize:  bufferSize,
		buffer:   make([]Loggable, 0, bufferSize),
		nodeName: nodeName,
	}
	logger.wg.Add(1)
	go logger.processLogs()
	return logger, nil
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
				j.flush()
				j.mu.Unlock()
				return
			}
			// 收到新日志，加入缓冲区
			j.mu.Lock()
			j.buffer = append(j.buffer, log)
			// 如果缓冲区已满，触发写入
			if len(j.buffer) >= j.maxSize {
				j.flush()
			}
			j.mu.Unlock()
		case <-ticker.C:
			// 定时器触发，写入缓冲区中的日志
			j.mu.Lock()
			j.flush()
			j.mu.Unlock()
		}
	}
}

func (j *JSONLLogger) flush() {
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
	if j.writer != os.Stdout && j.writer != os.Stderr {
		return j.writer.Close()
	}
	return nil
}
