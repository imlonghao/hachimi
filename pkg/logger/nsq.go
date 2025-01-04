package logger

import (
	"encoding/json"
	"github.com/nsqio/go-nsq"
	"log"
	"sync"
	"time"
)

type NSQLogger struct {
	logChan  chan Loggable
	producer *nsq.Producer
	topic    string
	wg       sync.WaitGroup
	buffer   []Loggable
	bufSize  int
	mu       sync.Mutex
}

// NewNSQLogger 创建 NSQLogger
func NewNSQLogger(producer *nsq.Producer, topic string, bufSize int) (*NSQLogger, error) {
	logger := &NSQLogger{
		logChan:  make(chan Loggable, 100),
		producer: producer,
		topic:    topic,
		bufSize:  bufSize,
		buffer:   make([]Loggable, 0, bufSize),
	}
	logger.wg.Add(1)
	go logger.processLogs()
	return logger, nil
}

func (o *NSQLogger) processLogs() {
	defer o.wg.Done()
	ticker := time.NewTicker(1 * time.Second) // 每 1 秒强制写入一次
	defer ticker.Stop()
	for {
		select {
		case log, ok := <-o.logChan:
			if !ok {
				// 通道关闭，写入剩余日志
				o.mu.Lock()
				o.flush()
				o.mu.Unlock()
				return
			}
			// 收到新日志，加入缓冲区
			o.mu.Lock()
			o.buffer = append(o.buffer, log)
			// 如果缓冲区已满，触发写入
			if len(o.buffer) >= o.bufSize {
				o.flush()
			}
			o.mu.Unlock()
		case <-ticker.C:
			// 定时器触发，写入缓冲区中的日志
			o.mu.Lock()
			o.flush()
			o.mu.Unlock()
		}
	}

	// Flush remaining logs
	o.mu.Lock()
	o.flush()
	o.mu.Unlock()
}

func (o *NSQLogger) flush() {
	if len(o.buffer) == 0 {
		return
	}
	var buf [][]byte
	for _, logData := range o.buffer {
		dataMap, _ := logData.ToMap()
		dataMap["table_name"] = logData.TableName()
		jsonData, _ := json.Marshal(dataMap)
		buf = append(buf, jsonData)
	}
	o.buffer = o.buffer[:0]
	err := o.producer.MultiPublish(o.topic, buf)
	//高延迟时 消息队列可能阻塞 开一个新线程避免阻塞请求 一直失败可能会爆内存?
	if err != nil {
		log.Println(err)
	}
}

func (o *NSQLogger) Log(data Loggable) error {
	o.logChan <- data
	return nil
}

func (o *NSQLogger) Close() error {
	close(o.logChan)
	o.wg.Wait()
	o.producer.Stop()
	return nil
}
