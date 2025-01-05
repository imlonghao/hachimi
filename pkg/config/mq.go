package config

// MQConfig  NSQ 消息队列配置
type MQConfig struct {
	// Host 服务监听地址
	Host string `toml:"host"`
	// AuthSecret 密码
	AuthSecret string `toml:"secret"`
	// Topic 主题
	Topic            string `toml:"topic"`
	Compression      bool   `toml:"compression"`
	CompressionLevel int    `toml:"compressionLevel"`
	Tls              bool   `toml:"tls"`
	EnableTlsVerify  bool   `toml:"enableTlsVerify"`
	ClientCertPath   string `toml:"clientCertPath"`
	ClientKeyPath    string `toml:"clientKeyPath"`
	CaCertPath       string `toml:"caCertPath"`
}
