package logger

// Logger 通用日志接口
type Logger interface {
	Log(data Loggable) error
	Close() error
}
type Loggable interface {
	ToMap() (map[string]interface{}, error) // 将结构体转换为通用 map
	TableName() string                      // 返回数据库表名
}
