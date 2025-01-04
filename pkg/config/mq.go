package config

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/nsqio/go-nsq"
)

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

func NewNsqProducer(Host string, AuthSecret string, Compression bool, CompressionLevel int, Tls bool, EnableTlsVerify bool, ClientCertPath string, ClientKeyPath string, CaCertPath string) (*nsq.Producer, error) {
	config := nsq.NewConfig()
	//config.UserAgent = "hachimi"
	//config.TlsConfig //TODO
	//AuthSecret requires nsqd 0.2.29+
	config.AuthSecret = AuthSecret
	config.Deflate = Compression
	var tlsConfig *tls.Config
	if Tls {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		if EnableTlsVerify {
			tlsConfig.InsecureSkipVerify = false
		}
		if ClientCertPath != "" && ClientKeyPath != "" {
			cert, err := tls.LoadX509KeyPair(ClientCertPath, ClientKeyPath)
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		if CaCertPath != "" {
			caCert, err := tls.LoadX509KeyPair(CaCertPath, CaCertPath)
			if err != nil {
				return nil, err
			}
			caCertPool, err := x509.SystemCertPool()
			if err != nil {
				caCertPool = x509.NewCertPool()
			}
			for i := range caCert.Certificate {
				caCertPool.AppendCertsFromPEM(caCert.Certificate[i])
			}
			tlsConfig.RootCAs = caCertPool
		}
	}

	if Compression && CompressionLevel > 0 && CompressionLevel <= 9 {
		config.DeflateLevel = CompressionLevel
	}
	config.TlsConfig = tlsConfig
	p, err := nsq.NewProducer(Host, config)
	if err != nil {
		return nil, err
	}
	return p, nil
}
