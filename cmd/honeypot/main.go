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
	beekeeper    = flag.String("beekeeper", "", "beekeeper address")
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
		config.GetPotConfig().TimeOut = *timeOut
	}
	if config.GetPotConfig().Host == "" {
		config.GetPotConfig().Host = *host
	}
	if config.GetPotConfig().Port == 0 {
		config.GetPotConfig().Port = *port
	}
	if config.GetPotConfig().LogPath == "" {
		config.GetPotConfig().LogPath = *logPath
		if config.GetPotConfig().LogPath != "stdout" {
			logFile, err := os.OpenFile(config.GetPotConfig().LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				log.Fatalf("Failed to open log file: %s\n", err)
			}
			defer logFile.Close()
			log.SetOutput(logFile)
		}
	}
	if config.GetPotConfig().HoneyLogPath == "" {
		//TODO MQ
		config.GetPotConfig().HoneyLogPath = *honeyLogPath
		if config.GetPotConfig().HoneyLogPath != "stdout" {
			err := config.SetLogger(config.GetPotConfig().HoneyLogPath)
			if err != nil {
				log.Fatalf("Failed to set honey log file: %s\n", err)
			}
		}
	}

	lm := ingress.ListenerManager{}
	tcpListener := ingress.NewTCPListener(config.GetPotConfig().Host, config.GetPotConfig().Port)
	lm = *ingress.NewListenerManager()
	lm.AddTCPListener(tcpListener)
	udpListener := ingress.NewUDPListener(config.GetPotConfig().Host, config.GetPotConfig().Port)
	lm.AddUDPListener(udpListener)
	lm.StartAll(context.Background())
	select {}
}
