package session

import (
	"sync"
	"time"
)

const (
	blockDuration  = 10 * time.Minute // 拉黑时长
	windowDuration = 10 * time.Minute // 统计时间窗口
)

type ConnectionLimiter struct {
	maxConnections int
	mu             sync.Mutex
	connections    map[string][]time.Time // 每个 IP 的连接时间记录
	blacklist      map[string]time.Time   // 黑名单列表
	cleanupTicker  *time.Ticker           // 定期清理任务
}

// NewConnectionLimiter 创建一个新的连接限制器
func NewConnectionLimiter(maxConnections int) *ConnectionLimiter {
	limiter := &ConnectionLimiter{
		maxConnections: maxConnections,
		connections:    make(map[string][]time.Time),
		blacklist:      make(map[string]time.Time),
		cleanupTicker:  time.NewTicker(1 * time.Minute),
	}

	// 启动清理任务
	go limiter.cleanupExpiredEntries()
	return limiter
}

// AllowConnection 判断是否允许来自指定 IP 的连接
func (cl *ConnectionLimiter) AllowConnection(ip string) bool {
	if cl.maxConnections == 0 || ip == "" {
		return true
	}
	cl.mu.Lock()
	defer cl.mu.Unlock()

	// 检查黑名单
	if unblockTime, blacklisted := cl.blacklist[ip]; blacklisted {
		// 如果还在封禁时间内，拒绝连接
		if time.Now().Before(unblockTime) {
			return false
		}
		// 否则移除黑名单
		delete(cl.blacklist, ip)
	}

	// 记录当前连接时间
	now := time.Now()
	cl.connections[ip] = append(cl.connections[ip], now)

	// 移除超过时间窗口的记录
	windowStart := now.Add(-windowDuration)
	validConnections := []time.Time{}
	for _, t := range cl.connections[ip] {
		if t.After(windowStart) {
			validConnections = append(validConnections, t)
		}
	}
	cl.connections[ip] = validConnections

	// 检查是否超过阈值
	if len(validConnections) > cl.maxConnections {
		cl.blacklist[ip] = now.Add(blockDuration) // 加入黑名单
		delete(cl.connections, ip)                // 清空历史记录
		return false
	}

	return true
}

// cleanupExpiredEntries 定期清理过期的连接记录和黑名单
func (cl *ConnectionLimiter) cleanupExpiredEntries() {
	for range cl.cleanupTicker.C {
		cl.mu.Lock()

		now := time.Now()

		// 清理过期的连接记录
		for ip, times := range cl.connections {
			validTimes := []time.Time{}
			for _, t := range times {
				if t.After(now.Add(-windowDuration)) {
					validTimes = append(validTimes, t)
				}
			}
			if len(validTimes) > 0 {
				cl.connections[ip] = validTimes
			} else {
				delete(cl.connections, ip)
			}
		}

		// 清理过期的黑名单
		for ip, unblockTime := range cl.blacklist {
			if now.After(unblockTime) {
				delete(cl.blacklist, ip)
			}
		}

		cl.mu.Unlock()
	}
}

// Close 停止清理任务
func (cl *ConnectionLimiter) Close() {
	cl.cleanupTicker.Stop()
}
