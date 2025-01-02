package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"github.com/pelletier/go-toml"
	"gorm.io/gorm"
	"hachimi/pkg/logger"
	"log"
	"os"
)

// 全局数据库实例
var (
	DB        *gorm.DB
	TlsConfig *tls.Config
)

type Config struct {
	// Host 服务监听地址
	Host string `toml:"host"`
	// Port 服务监听端口
	Port int `toml:"port"`
	// Forward 转发规则
	ForwardingRules *[]ForwardingRule `toml:"forward"`
	// TLS TLS 配置 默认 随机生成一个硬编码主体的证书 用于临时使用
	TLS *TLSConfig `toml:"tls"`
	// LogPath 系统日志路径 默认为stderr
	LogPath string `toml:"logPath"`
	// HoneyLogPath 蜜罐会话日志路径 默认为stdout 如果启用 mq 会话日志将会被发送到 nsq 中 不会写入文件
	HoneyLogPath string `toml:"honeyLogPath"`
	TimeOut      int    `toml:"timeOut"`
	// MQ 消息队列配置
	MQ *MQConfig `toml:"mq"`
}

var honeyConfig *Config

func GetConfig() *Config {
	if honeyConfig == nil {
		log.Fatalln("Config is not loaded")
	}
	return honeyConfig
}

// Logger 全局会话日志处理器
var Logger logger.Logger

var SshPrivateKey *rsa.PrivateKey

func init() {
	honeyConfig = &Config{
		TimeOut: 60,
	}
	TlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		//CipherSuites:       AllCiphers,
		Certificates: []tls.Certificate{genCert()},
		//MaxVersion:   tls.VersionTLS12,
		//MinVersion:   tls.VersionSSL30,
	}
	var err error
	SshPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	jsonlLogger, err := logger.NewJSONLLogger("stdout", 100)
	if err != nil {
		log.Fatalln("Failed to create JSONL logger:", err)
	}
	Logger = jsonlLogger
}

func LoadConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var config Config
	decoder := toml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}
	if config.LogPath != "" {
		logFile, err := os.OpenFile(config.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %s\n", err)
		}
		defer logFile.Close()
		// 设置日志输出
		log.Println("Log Will be written to ", config.LogPath)
		log.SetOutput(logFile)
	}
	// 设置全局日志处理器
	if config.MQ == nil {
		if config.HoneyLogPath != "" {
			err := SetLogger(config.HoneyLogPath)
			if err != nil {
				return err
			}
		}
	} else {
		//TODO 使用NSQ 作为日志处理器
	}
	if config.TLS != nil {
		if config.TLS.CertFile != "" && config.TLS.CertKey != "" {
			cert, err := LoadCert(config.TLS.CertFile, config.TLS.CertKey)
			if err != nil {
				log.Fatalf("Failed to load certificate: %s\n", err)
			}
			TlsConfig.Certificates = []tls.Certificate{cert}
		} else {
			log.Println("TLS config is not complete, using default TLS config")
		}
	}
	if config.TimeOut == 0 {
		config.TimeOut = 60
	}
	if config.ForwardingRules != nil {
		err := validateForwardingRule(config.ForwardingRules)
		if err != nil {
			return err
		}
	}
	honeyConfig = &config
	return nil
}
func SetLogger(path string) error {
	honeyLogger, err := logger.NewJSONLLogger(path, 100)
	if err != nil {
		return err
	}
	Logger = honeyLogger
	return nil
}
func LoadConfigFromString(data string) (*Config, error) {
	var config Config
	err := toml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
