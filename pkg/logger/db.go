package logger

import (
	"gorm.io/gorm"
	"sync"
	"time"
)

type DBLogger struct {
	logChan chan Loggable
	db      *gorm.DB
	wg      sync.WaitGroup
	buffer  []Loggable
	maxSize int
	mu      sync.Mutex
}

// NewORMLogger 创建 ORMLogger
func NewORMLogger(db *gorm.DB, maxSize int) (*DBLogger, error) {
	logger := &DBLogger{
		logChan: make(chan Loggable, 100),
		db:      db,
		maxSize: maxSize,
		buffer:  make([]Loggable, 0, maxSize),
	}
	logger.wg.Add(1)
	go logger.processLogs()
	return logger, nil
}

func (o *DBLogger) processLogs() {
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
			if len(o.buffer) >= o.maxSize {
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

func (o *DBLogger) flush() {
	//TODO
	panic("待迁移")
}

func (o *DBLogger) Log(data Loggable) error {
	o.logChan <- data
	return nil
}

func (o *DBLogger) Close() error {
	close(o.logChan)
	o.wg.Wait()
	sqlDB, _ := o.db.DB()
	return sqlDB.Close()
}
