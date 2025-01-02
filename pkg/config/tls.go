package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"math/big"
	"time"
)

// TLSConfig  TLS 配置
type TLSConfig struct {
	// CertKey 证书路径
	CertKey string `toml:"certKey"`
	// CertFile 证书文件
	CertFile string `toml:"certFile"`
	// TODO 证书自动生成配置
}

func LoadCert(certFile, keyFile string) (tls.Certificate, error) {
	return tls.LoadX509KeyPair(certFile, keyFile)
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
