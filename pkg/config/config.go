package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"gorm.io/gorm"
	"hachimi/pkg/logger"
	"log"
	"math/big"
	"sync"
	"time"
)

// 全局数据库实例
var (
	DB        *gorm.DB
	once      sync.Once // 确保初始化只运行一次
	TimeOut   = 60
	TlsConfig *tls.Config
)

// 全局日志处理器
var (
	Logger logger.Logger
)
var SshPrivateKey *rsa.PrivateKey

func init() {
	once.Do(func() {
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

	})
}

func genCert() tls.Certificate {
	//生成 TLS 证书
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate private key: %s\n", err)
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.Fatalf("Failed to generate serial number: %s\n", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Hachimi"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
	}

	cer, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s\n", err)
	}
	return tls.Certificate{
		Certificate: [][]byte{cer},
		PrivateKey:  priv,
	}
}
