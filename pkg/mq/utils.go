package mq

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/nsqio/go-nsq"
)

func NewNsqConsumer(Topic string, Channel string, AuthSecret string, Compression bool, CompressionLevel int, Tls bool, EnableTlsVerify bool, ClientCertPath string, ClientKeyPath string, CaCertPath string) (*nsq.Consumer, error) {
	config, err := NewNsqConfig(AuthSecret, Compression, CompressionLevel, Tls, EnableTlsVerify, ClientCertPath, ClientKeyPath, CaCertPath)
	if err != nil {
		return nil, err
	}
	c, err := nsq.NewConsumer(Topic, Channel, config)
	if err != nil {
		return nil, err
	}
	return c, nil
}
func NewNsqProducer(Host string, AuthSecret string, Compression bool, CompressionLevel int, Tls bool, EnableTlsVerify bool, ClientCertPath string, ClientKeyPath string, CaCertPath string) (*nsq.Producer, error) {
	config, err := NewNsqConfig(AuthSecret, Compression, CompressionLevel, Tls, EnableTlsVerify, ClientCertPath, ClientKeyPath, CaCertPath)
	if err != nil {
		return nil, err
	}
	p, err := nsq.NewProducer(Host, config)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func NewNsqConfig(AuthSecret string, Compression bool, CompressionLevel int, Tls bool, EnableTlsVerify bool, ClientCertPath string, ClientKeyPath string, CaCertPath string) (*nsq.Config, error) {
	config := nsq.NewConfig()
	//config.UserAgent = "hachimi"
	//config.TlsConfig //TODO
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

	return config, nil
}
