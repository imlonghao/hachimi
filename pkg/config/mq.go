package config

// MQConfig  NSQ 消息队列配置
type MQConfig struct {
	// Host 服务监听地址
	Host string `toml:"host"`
	// Port 服务监听端口
	Port int `toml:"port"`
	// User 用户名
	User string `toml:"user"`
	// Password 密码
	Password string `toml:"password"`
	// Topic 主题
	Topic string `toml:"topic"`
}
