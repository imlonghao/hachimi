package analysis

import (
	"encoding/json"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"hachimi/pkg/analysis/model"
	"hachimi/pkg/types"
	"hachimi/pkg/utils"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/mmcloughlin/geohash"
	"github.com/nsqio/go-nsq"
	"github.com/oschwald/geoip2-golang"
)

type PotMessageHandler struct {
	logChan   chan *types.HoneyData
	wg        sync.WaitGroup
	mu        sync.Mutex
	buffer    []*types.HoneyData
	count     int64
	conn      clickhouse.Conn
	cityDb    *geoip2.Reader
	countryDb *geoip2.Reader
	asnDb     *geoip2.Reader
	maxSize   int
}

func NewPotMessageHandler(bufferSize int, conn clickhouse.Conn, countryDb *geoip2.Reader, cityDb *geoip2.Reader, asnDb *geoip2.Reader) (*PotMessageHandler, error) {
	Handler := &PotMessageHandler{
		logChan:   make(chan *types.HoneyData, 100),
		buffer:    make([]*types.HoneyData, 0, bufferSize),
		maxSize:   bufferSize,
		conn:      conn,
		countryDb: countryDb,
		cityDb:    cityDb,
		asnDb:     asnDb,
	}
	Handler.wg.Add(1)
	go Handler.processLogs()
	return Handler, nil
}
func (h *PotMessageHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		return nil
	}
	var data types.HoneyData
	err := json.Unmarshal(m.Body, &data)
	if err != nil {
		log.Println(err)
		return err
	}
	atomic.AddInt64(&h.count, 1)
	h.logChan <- &data
	return nil
}

func (h *PotMessageHandler) processLogs() {
	defer h.wg.Done()

	ticker := time.NewTicker(1 * time.Second) // 每 1 秒强制写入一次
	defer ticker.Stop()

	for {
		select {
		case log, ok := <-h.logChan:
			if !ok {
				// 通道关闭，写入剩余日志
				h.mu.Lock()
				h.flush()
				h.mu.Unlock()
				return
			}
			// 收到新日志，加入缓冲区
			h.mu.Lock()
			h.buffer = append(h.buffer, log)
			// 如果缓冲区已满，触发写入
			if len(h.buffer) >= h.maxSize {
				h.flush()
			}
			h.mu.Unlock()
		case <-ticker.C:
			// 定时器触发，写入缓冲区中的日志
			if h.count > 0 {
				h.mu.Lock()
				h.flush()
				h.mu.Unlock()
			}
		}
	}
}

func (h *PotMessageHandler) flush() {
	//return
	var logs map[string][]interface{}
	logs = make(map[string][]interface{})
	for _, msg := range h.buffer {
		data := msg.Data.(map[string]interface{})

		switch msg.Type {
		case "session":
			// map[string]interface {}
			var session model.Session
			session.ID = data["id"].(string)
			session.Protocol = data["protocol"].(string)
			session.StartTime, _ = utils.StringToTime(data["start_time"].(string))
			session.EndTime, _ = utils.StringToTime(data["end_time"].(string))
			session.SrcIP = data["src_ip"].(string)
			session.SrcPort = int(data["src_port"].(float64))
			session.DstIP = data["dst_ip"].(string)
			session.DstPort = int(data["dst_port"].(float64))
			session.NodeName = msg.NodeName
			session.IsTls = data["is_tls"].(bool)
			//session.IsIpV6 = data["is_ipv6"].(bool)
			//session.IsGmTls = data["is_gm_tls"].(bool)
			session.IsHandled = data["is_handled"].(bool)
			session.IsHttp = data["is_http"].(bool)
			session.Data = data["data"].(string)
			session.Service = data["service"].(string)
			session.Duration = int(data["duration"].(float64))

			ip := net.ParseIP(session.SrcIP)

			if h.cityDb != nil {
				record, err := h.cityDb.City(ip)
				if err != nil {
					log.Println(err)
				} else {
					// Country
					CountryName := record.Country.Names["zh-CN"]
					IsoCode := record.Country.IsoCode
					if CountryName == "" {
						CountryName = record.RegisteredCountry.Names["zh-CN"]
						IsoCode = record.Country.IsoCode
					}
					session.CountryName = CountryName
					session.IsoCode = IsoCode
					// City
					session.CityName = record.City.Names["en"]
					session.GeoHash = geohash.Encode(record.Location.Latitude, record.Location.Longitude)
				}
			} else if h.countryDb != nil {
				record, err := h.countryDb.Country(ip)
				if err != nil {
					log.Println(err)
				} else {
					CountryName := record.Country.Names["zh-CN"]
					IsoCode := record.Country.IsoCode
					if CountryName == "" {
						CountryName = record.RegisteredCountry.Names["zh-CN"]
						IsoCode = record.Country.IsoCode
					}
					session.CountryName = CountryName
					session.IsoCode = IsoCode
				}
			}

			if h.asnDb != nil {
				record, err := h.asnDb.ASN(ip)
				if err != nil {
					log.Println(err)
				} else {
					session.AsnOrg = record.AutonomousSystemOrganization
					session.AsnNumber = record.AutonomousSystemNumber
				}
			}

			session.DataHash = utils.SHA1(session.Data)
			logs["session"] = append(logs["session"], session)

		case "http_session":
			var http model.HttpSession
			http.ID = data["id"].(string)
			http.SessionID = data["session_id"].(string)
			http.StartTime, _ = utils.StringToTime(data["start_time"].(string))
			http.EndTime, _ = utils.StringToTime(data["end_time"].(string))
			http.Header = utils.MapInterfaceToString(data["header"].(map[string]interface{}))
			http.UriParam = utils.MapInterfaceToString(data["uri_param"].(map[string]interface{}))
			http.BodyParam = utils.MapInterfaceToString(data["body_param"].(map[string]interface{}))
			http.SrcIP = data["src_ip"].(string)
			http.SrcPort = int(data["src_port"].(float64))
			http.DstIP = data["dst_ip"].(string)
			http.DstPort = int(data["dst_port"].(float64))
			http.NodeName = msg.NodeName
			http.IsTls = data["is_tls"].(bool)
			//session.IsIpV6 = data["is_ipv6"].(bool)
			//session.IsGmTls = data["is_gm_tls"].(bool)
			http.IsHandled = data["is_handled"].(bool)
			http.IsHttp = data["is_http"].(bool)
			http.Data = data["data"].(string)
			http.Method = data["method"].(string)
			http.Path = data["path"].(string)
			http.UA = data["ua"].(string)
			http.Host = data["host"].(string)
			http.RawHeader = data["raw_header"].(string)
			http.Body = data["body"].(string)
			http.Service = data["service"].(string)
			http.Duration = int(data["duration"].(float64))

			ip := net.ParseIP(http.SrcIP)

			if h.cityDb != nil {
				record, err := h.cityDb.City(ip)
				if err != nil {
					log.Println(err)
				} else {
					// Country
					CountryName := record.Country.Names["zh-CN"]
					IsoCode := record.Country.IsoCode
					if CountryName == "" {
						CountryName = record.RegisteredCountry.Names["zh-CN"]
						IsoCode = record.Country.IsoCode
					}
					http.CountryName = CountryName
					http.IsoCode = IsoCode
					// City
					http.CityName = record.City.Names["en"]
					http.GeoHash = geohash.Encode(record.Location.Latitude, record.Location.Longitude)
				}
			} else if h.countryDb != nil {
				record, err := h.countryDb.Country(ip)
				if err != nil {
					log.Println(err)
				} else {
					CountryName := record.Country.Names["zh-CN"]
					IsoCode := record.Country.IsoCode
					if CountryName == "" {
						CountryName = record.RegisteredCountry.Names["zh-CN"]
						IsoCode = record.Country.IsoCode
					}
					http.CountryName = CountryName
					http.IsoCode = IsoCode
				}
			}

			if h.asnDb != nil {
				record, err := h.asnDb.ASN(ip)
				if err != nil {
					log.Println(err)
				} else {
					http.AsnOrg = record.AutonomousSystemOrganization
					http.AsnNumber = record.AutonomousSystemNumber
				}
			}

			http.BodyHash = utils.SHA1(http.Body)
			http.HeaderHash = utils.SHA1(http.RawHeader)
			http.PathHash = utils.SHA1(http.Path)
			http.DataHash = utils.SHA1(http.Data)
			logs["http"] = append(logs["http"], http)

		}
	}
	for key, value := range logs {
		if len(value) == 0 {
			continue
		}
		switch key {
		case "session":
			// 通过反射获取结构体字段的列名
			if len(value) == 0 {
				continue
			}
			var sessions []model.Session
			for _, v := range value {
				sessions = append(sessions, v.(model.Session))
			}

			// 执行批量插入
			if err := model.InsertSession(h.conn, sessions); err != nil {
				log.Fatalf("failed to send batch: %v", err)
			}
		case "http":
			if len(value) == 0 {
				continue
			}
			var httpsessions []model.HttpSession
			for _, v := range value {
				httpsessions = append(httpsessions, v.(model.HttpSession))
			}
			if err := model.InsertHttpSession(h.conn, httpsessions); err != nil {
				log.Fatalf("failed to send batch: %v", err)
			}

		}
		h.buffer = h.buffer[:0]
	}
}
func (h *PotMessageHandler) Close() {
	close(h.logChan)
	h.Wait()
}
func (h *PotMessageHandler) Wait() {
	h.wg.Wait()
}
