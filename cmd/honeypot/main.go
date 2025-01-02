// 蜜罐 蜜网的最小组成部分 可单机独立运行
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"hachimi/pkg/config"
	"hachimi/pkg/ingress"
	"log"
	"os"
)

var (
	configPath   = flag.String("configPath", "config.toml", "config path if not set will use cli flags")
	host         = flag.String("host", "0.0.0.0", "listen host")
	port         = flag.Int("port", 12345, "listen port")
	logPath      = flag.String("logPath", "stdout", "system log path")
	honeyLogPath = flag.String("honeyLogPath", "stdout", "honey log path")
	keyPath      = flag.String("keyPath", "", "ssl key path")
	certPath     = flag.String("certPath", "", "ssl cert path")
	timeOut      = flag.Int("timeOut", 60, "timeout for honeypot session Default 60")
)

func main() {
	flag.Parse()
	//configPath 文件是否存在
	if *configPath != "" {
		if _, err := os.Stat(*configPath); err == nil {
			err := config.LoadConfig(*configPath)
			if err != nil {
				log.Fatalf("load config file error: %v", err)
			}
		}

	}
	if *keyPath != "" && *certPath != "" {
		cert, err := config.LoadCert(*certPath, *keyPath)
		if err != nil {
			log.Fatalf("Failed to load certificate: %s\n", err)
		}
		config.TlsConfig.Certificates = []tls.Certificate{cert}
	}
	if *timeOut != 60 {
		config.GetConfig().TimeOut = *timeOut
	}
	if config.GetConfig().Host == "" {
		config.GetConfig().Host = *host
	}
	if config.GetConfig().Port == 0 {
		config.GetConfig().Port = *port
	}
	if config.GetConfig().LogPath == "" {
		config.GetConfig().LogPath = *logPath
		if config.GetConfig().LogPath != "stdout" {
			logFile, err := os.OpenFile(config.GetConfig().LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				log.Fatalf("Failed to open log file: %s\n", err)
			}
			defer logFile.Close()
			log.SetOutput(logFile)
		}
	}
	if config.GetConfig().HoneyLogPath == "" {
		//TODO MQ
		config.GetConfig().HoneyLogPath = *honeyLogPath
		if config.GetConfig().HoneyLogPath != "stdout" {
			err := config.SetLogger(config.GetConfig().HoneyLogPath)
			if err != nil {
				log.Fatalf("Failed to set honey log file: %s\n", err)
			}
		}
	}

	lm := ingress.ListenerManager{}
	tcpListener := ingress.NewTCPListener(config.GetConfig().Host, config.GetConfig().Port)
	lm = *ingress.NewListenerManager()
	lm.AddTCPListener(tcpListener)
	udpListener := ingress.NewUDPListener(config.GetConfig().Host, config.GetConfig().Port)
	lm.AddUDPListener(udpListener)
	lm.StartAll(context.Background())
	select {}
}
