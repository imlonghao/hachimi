package main

import (
	"context"
	"flag"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/oschwald/geoip2-golang"
	"github.com/pelletier/go-toml"
	"hachimi/pkg/analysis"
	"hachimi/pkg/analysis/model"
	"hachimi/pkg/config"
	"hachimi/pkg/mq"
	"log"
	"os"
)

var cityDb *geoip2.Reader
var countryDb *geoip2.Reader
var asnDb *geoip2.Reader

type HunterConfig struct {
	// MQ 消息队列配置
	MQ config.MQConfig `toml:"mq"`
	//数据库
	DBHost     string `toml:"db_host"`
	DBUser     string `toml:"db_user"`
	DBPassword string `toml:"db_password"`
	DBName     string `toml:"db_name"`
	// GeoLite
	GeoLiteCountryPath string `toml:"geo_country"`
	GeoLiteASNPath     string `toml:"geo_asn"`
}

// flag
var (
	configPath = flag.String("c", "hunter.toml", "config path")
)
var hunterConfig HunterConfig

func main() {
	flag.Parse()
	//configPath 文件是否存在
	var err error
	file, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("open config file error: %v", err)
	}

	defer file.Close()
	decoder := toml.NewDecoder(file)
	err = decoder.Decode(&hunterConfig)
	if err != nil {
		log.Fatalf("load config file error: %v", err)
	}
	if hunterConfig.DBHost == "" {
		log.Fatalf("db_host is empty")
	}

	if hunterConfig.GeoLiteCountryPath == "" {
		hunterConfig.GeoLiteCountryPath = "GeoLite2-Country.mmdb"
	}
	if hunterConfig.GeoLiteASNPath == "" {
		hunterConfig.GeoLiteASNPath = "GeoLite2-ASN.mmdb"
	}
	if hunterConfig.DBName == "" {
		hunterConfig.DBName = "default"
	}

	countryDb, err = geoip2.Open(hunterConfig.GeoLiteCountryPath)
	if err != nil {
		log.Fatal(err)
	}
	asnDb, err = geoip2.Open(hunterConfig.GeoLiteASNPath)
	if err != nil {
		log.Fatal(err)
	}
	defer countryDb.Close()
	defer asnDb.Close()

	// 创建 ClickHouse 客户端
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{hunterConfig.DBHost},
		Auth: clickhouse.Auth{
			Database: hunterConfig.DBName,
			Username: hunterConfig.DBUser,
			Password: hunterConfig.DBPassword,
		},
		Debug: false,
	})
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	if err := conn.Exec(context.Background(), model.CreateTablehttp_session()); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	if err := conn.Exec(context.Background(), model.CreateTablesession()); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	consumer, err := mq.NewNsqConsumer(hunterConfig.MQ.Topic, "channel", hunterConfig.MQ.AuthSecret, hunterConfig.MQ.Compression, hunterConfig.MQ.CompressionLevel, hunterConfig.MQ.Tls, hunterConfig.MQ.EnableTlsVerify, hunterConfig.MQ.ClientCertPath, hunterConfig.MQ.ClientKeyPath, hunterConfig.MQ.CaCertPath)
	if err != nil {
		log.Println(err)
		return
	}
	hander, err := analysis.NewPotMessageHandler(1000, conn, countryDb, asnDb)

	consumer.AddHandler(hander)
	err = consumer.ConnectToNSQD(hunterConfig.MQ.Host)
	if err != nil {
		log.Fatalf("Failed to connect to NSQ: %v", err)
		return
	}
	hander.Wait()

}
