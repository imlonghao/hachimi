package logger

import (
	"encoding/json"
	"github.com/nsqio/go-nsq"
	"hachimi/pkg/types"
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
	nodeName string
}

// NewNSQLogger 创建 NSQLogger
func NewNSQLogger(producer *nsq.Producer, topic string, bufSize int, nodeName string) (*NSQLogger, error) {
	logger := &NSQLogger{
		logChan:  make(chan Loggable, 100),
		producer: producer,
		topic:    topic,
		bufSize:  bufSize,
		buffer:   make([]Loggable, 0, bufSize),
		nodeName: nodeName,
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
				o.Flush()
				o.mu.Unlock()
				return
			}
			// 收到新日志，加入缓冲区
			o.mu.Lock()
			o.buffer = append(o.buffer, log)
			// 如果缓冲区已满，触发写入
			if len(o.buffer) >= o.bufSize {
				o.Flush()
			}
			o.mu.Unlock()
		case <-ticker.C:
			// 定时器触发，写入缓冲区中的日志
			o.mu.Lock()
			o.Flush()
			o.mu.Unlock()
		}
	}

	// Flush remaining logs
	o.mu.Lock()
	o.Flush()
	o.mu.Unlock()
}

func (o *NSQLogger) Flush() {
	if len(o.buffer) == 0 {
		return
	}
	var buf [][]byte
	for _, logData := range o.buffer {
		jsonData, _ := json.Marshal(types.HoneyData{Type: logData.Type(), Data: logData, Time: time.Now().Unix(), NodeName: o.nodeName})
		buf = append(buf, jsonData)
	}
	o.buffer = o.buffer[:0]
	err := o.producer.MultiPublish(o.topic, buf)
	//高延迟时 消息队列可能阻塞 开一个新线程避免阻塞请求 一直失败可能会爆内存?
	//发送失败 nsq客户端会一直存放在内存中 等待重试
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
