package model

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"time"
)

type Session struct {
	ID        string    `ch_name:"id" ch_type:"String"`
	SrcIP     string    `ch_name:"src_ip" ch_type:"String"` //TODO IPV6
	SrcPort   int       `ch_name:"src_port" ch_type:"UInt16"`
	DstIP     string    `ch_name:"dst_ip" ch_type:"String"` //TODO IPV6
	DstPort   int       `ch_name:"dst_port" ch_type:"UInt16"`
	NodeName  string    `ch_name:"node_name" ch_type:"String"`
	IsIpV6    bool      `ch_name:"is_ipv6" ch_type:"UInt8"`
	IsTls     bool      `ch_name:"is_tls" ch_type:"UInt8"`
	IsGmTls   bool      `ch_name:"is_gm_tls" ch_type:"UInt8"`
	IsHttp    bool      `ch_name:"is_http" ch_type:"UInt8"`
	IsHandled bool      `ch_name:"is_handled" ch_type:"UInt8"`
	Protocol  string    `ch_name:"protocol" ch_type:"String"`
	Data      string    `ch_name:"data" ch_type:"String"`
	Service   string    `ch_name:"service" ch_type:"String"`
	StartTime time.Time `ch_name:"start_time" ch_type:"DateTime('UTC')" ch_order:"true"`
	EndTime   time.Time `ch_name:"end_time" ch_type:"DateTime('UTC')"`
	//经过的时间 ms
	Duration int `ch_name:"duration" ch_type:"UInt32"`
	/* ip info */
	IsoCode     string `ch_name:"iso_code" ch_type:"String"`
	CountryName string `ch_name:"country_name" ch_type:"String"`
	AsnNumber   uint   `ch_name:"asn_number" ch_type:"UInt32"`
	AsnOrg      string `ch_name:"asn_org" ch_type:"String"`
	/* hash */
	DataHash string `ch_name:"data_hash" ch_type:"String"`
}

// Code generated by gen_clickhouse.go DO NOT EDIT.

func CreateTablesession() string {
	query := `CREATE TABLE IF NOT EXISTS session (
	id String,
	src_ip String,
	src_port UInt16,
	dst_ip String,
	dst_port UInt16,
	node_name String,
	is_ipv6 UInt8,
	is_tls UInt8,
	is_gm_tls UInt8,
	is_http UInt8,
	is_handled UInt8,
	protocol String,
	data String,
	service String,
	start_time DateTime('UTC'),
	end_time DateTime('UTC'),
	duration UInt32,
	iso_code String,
	country_name String,
	asn_number UInt32,
	asn_org String,
	data_hash String
) ENGINE = MergeTree() ORDER BY start_time`

	return query
}

func InsertSession(conn clickhouse.Conn, Sessions []Session) error {
	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO session (id, src_ip, src_port, dst_ip, dst_port, node_name, is_ipv6, is_tls, is_gm_tls, is_http, is_handled, protocol, data, service, start_time, end_time, duration, iso_code, country_name, asn_number, asn_org, data_hash)")
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, session := range Sessions {
		if err := batch.Append(session.ID, session.SrcIP, session.SrcPort, session.DstIP, session.DstPort, session.NodeName, session.IsIpV6, session.IsTls, session.IsGmTls, session.IsHttp, session.IsHandled, session.Protocol, session.Data, session.Service, session.StartTime, session.EndTime, session.Duration, session.IsoCode, session.CountryName, session.AsnNumber, session.AsnOrg, session.DataHash); err != nil {
			return fmt.Errorf("failed to append data: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}
	return nil
}

// End of generated code
