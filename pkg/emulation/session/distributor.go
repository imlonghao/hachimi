package session

import (
	"bytes"
	"hachimi/pkg/config"
	"hachimi/pkg/emulation/http"
	"hachimi/pkg/emulation/redis"
	"hachimi/pkg/emulation/relay"
	"hachimi/pkg/emulation/ssh"
	"hachimi/pkg/emulation/tls"
	"hachimi/pkg/types"
	"hachimi/pkg/utils"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

func Distributor(conn net.Conn, session *types.Session) bool {
	conn.SetDeadline(time.Now().Add(time.Duration(config.GetPotConfig().TimeOut) * time.Second))
	session.SetOutBuffer(new(bytes.Buffer))
	//if !PortDistributor(conn, session) {
	//if session.GetOutBuffer().Len() > 0 {
	//	conn = utils.NewLoggedConn(
	//		conn,
	//		io.MultiReader(session.GetOutBuffer(), conn), //由于之前已经读取了一部分数据 需要覆盖 reader 还原buffer
	//		session.GetOutBuffer(),
	//	)
	//}
	//return MagicDistributor(conn, session)
	//} else {
	//	return true
	//}
	return MagicDistributor(conn, session)
}
func PortDistributor(conn net.Conn, session *types.Session) bool {
	switch conn.LocalAddr().(*net.TCPAddr).Port {
	/*
			case 22:
				ssh.HandleSsh(conn, session)
				return true


		case 23:
			relay.HandleTCPRelay(conn, session, map[string]string{"targetAddr": "10.13.1.2:23", "sendSession": "false"})
			return true
	*/
	case 6379:
		return redis.HandleRedis(conn, session)
	}

	return false
}
func MagicDistributor(conn net.Conn, session *types.Session) bool {
	// 读取magicByte 长度 10
	magicByte := make([]byte, 10)
	n, err := conn.Read(magicByte)
	if err != nil {
		conn.Close()
		return false
	}
	session.IsHandled = true
	var conn2 net.Conn
	//已知问题 生成的的conn2 第一次读取只能读到第一个reader的数据 并不是紧连着的 之后需要再次读取才能读到原始连接的数据
	conn2 = utils.NewLoggedConn(
		conn,
		io.MultiReader(bytes.NewReader(magicByte[0:n]), conn), //由于之前已经读取了一部分数据 需要覆盖 reader 还原buffer
		session.GetOutBuffer(),
		config.GetLimitSize(),
	)
	if config.GetPotConfig().ForwardingRules != nil {
		for _, rule := range *config.GetPotConfig().ForwardingRules {
			if rule.Handler == "relay_tcp" && (rule.Port == 0 || rule.Port == conn.LocalAddr().(*net.TCPAddr).Port) {
				session.IsHandled = true
				session.IsHttp = true
				if rule.Config["service"] == "" {
					session.Service = "relay_tcp"
				} else {
					session.Service = rule.Config["service"]
				}
				relay.HandleTCPRelay(conn2, session, rule.Config)
				return true
			}
		}
	}
	/* TLS */
	if bytes.Equal(magicByte[0:2], []uint8{22, 3}) { //SSL3.0+ (TLS 1.0 1.1 1.2 1.3) ClientHello 0x16 0x03  TODO 其他协议模拟
		tlsServer := tls.NewTlsServer(conn2, session)
		err := tlsServer.Handle()
		if err != nil {
			if config.GetPotConfig().Debug {
				log.Printf("SESSION %s SRC %s:%d DST %s:%d TLS ERROR %s\n", session.ID, session.SrcIP, session.SrcPort, session.DstIP, session.DstPort, err)
			}
			io.ReadAll(conn) //出错继续读取
			return false
		}

		session.IsTls = true
		session.SetOutBuffer(new(bytes.Buffer)) // 重置buffer 不记录tls原始数据
		// 重新读取magicByte 长度 10
		magicByte = make([]byte, 10)
		n, err = tlsServer.Conn.Read(magicByte)
		if err != nil {
			io.ReadAll(conn2) //出错继续读取
			return false
		}
		conn2 = utils.NewLoggedConn(
			tlsServer.Conn,
			io.MultiReader(bytes.NewReader(magicByte[0:n]), tlsServer.Conn),
			session.GetOutBuffer(),
			config.GetLimitSize(),
		)
	}
	/* gmTLS */
	/*deleted*/
	/* HTTP */
	// 通过开头几个字节快速判断是否是HTTP请求
	//CONNECT 为 HTTP 代理请求
	if string(magicByte[0:5]) == "POST " || string(magicByte[0:4]) == "GET " || string(magicByte[0:5]) == "HEAD " || string(magicByte[0:8]) == "OPTIONS " || string(magicByte[0:7]) == "DELETE " || string(magicByte[0:4]) == "PUT " || string(magicByte[0:6]) == "TRACE " || string(magicByte[0:8]) == "CONNECT " || string(magicByte[0:6]) == "PATCH " {
		if config.GetPotConfig().ForwardingRules != nil {
			for _, rule := range *config.GetPotConfig().ForwardingRules {
				if rule.Handler == "relay_http" && (rule.Port == 0 || rule.Port == conn.LocalAddr().(*net.TCPAddr).Port) {
					session.IsHandled = true
					session.IsHttp = true
					relay.HandleHttpRelay(conn2, session, rule.Config)
					return true
				}
			}
		}
		session.Service = "http"
		session.IsHandled = true
		session.IsHttp = true
		http.HandleHttp(conn2, session)
		return true
	}
	/* SSH */
	if strings.HasPrefix(string(magicByte), "SSH-") { //SSH
		session.Service = "ssh"
		session.IsHandled = true
		ssh.HandleSsh(conn2, session)
		return true
	}
	/* Redis */
	//简单匹配
	if string(magicByte) == "*1\r\n" || string(magicByte) == "*2\r\n" || strings.ToLower(string(magicByte)) == "info" || strings.ToLower(string(magicByte)) == "ping" {
		session.Service = "redis"
		session.IsHandled = true
		if redis.HandleRedis(conn2, session) {
			return true
		}
	} else {
		/* Other */
		//TODO BUFFER POOL
		var buffer = make([]byte, 1024*1024)
		n, err = conn2.Read(buffer)
		if err != nil {
			conn.Close()
			return false
		}
		buffer = nil
	}

	//读取第一行 限制大小1K
	buf := make([]byte, 1024)
	_, err = conn2.Read(buf)
	if err != nil {
		conn.Close()
		return false
	}
	firstLine := strings.Split(string(session.GetOutBuffer().Bytes()[0:utils.Min(session.GetOutBuffer().Len(), 1024)]), "\n")[0]
	//判断是遗漏的非标准HTTP请求
	conn2 = utils.NewLoggedConn(
		conn2,
		io.MultiReader(bytes.NewReader(session.GetOutBuffer().Bytes()), conn2),
		nil, //日志在上一层记录
		0,
	)
	if isHTTPRequestLine(firstLine) {
		session.Service = "http"
		session.IsHandled = true
		session.IsHttp = true
		http.HandleHttp(conn2, session)
		return true
	}
	if PortDistributor(conn2, session) {
		return true
	} else {
		io.ReadAll(conn2)
		session.IsHandled = false
		session.Service = "raw"
		return false
	}

}

// 判断是否为 HTTP 请求行
func isHTTPRequestLine(line string) bool {
	// 去掉首尾空格
	line = strings.TrimSpace(line)

	// 拆分字符串为三部分：METHOD PATH PROTOCOL
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return false // 如果不是三部分，直接返回 false
	}

	// 检查协议部分是否以 "HTTP/" 开头，并且后面跟一个版本号
	protocol := parts[2]
	if !strings.HasPrefix(protocol, "HTTP/") {
		return false
	}

	version := strings.TrimPrefix(protocol, "HTTP/")
	if version != "1.0" && version != "1.1" && version != "2" && version != "3" {
		return false
	}

	// 检查路径是否以 "/" 开头
	//path := parts[1]
	//if !strings.HasPrefix(path, "/") {
	//	return false
	//}

	return true
}
