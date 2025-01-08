package logger

// Logger 通用日志接口
type Logger interface {
	Log(data Loggable) error
	Flush()
	Close() error
}
type Loggable interface {
	Type() string // 返回数据库表名
}
