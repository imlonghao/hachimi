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
	"net"
	"strings"
	"time"
)

func Distributor(conn net.Conn, session *types.Session) bool {
	conn.SetDeadline(time.Now().Add(time.Duration(config.GetConfig().TimeOut) * time.Second))
	var buf bytes.Buffer
	session.SetOutBuffer(&buf)
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
	case 22:
		conn2 := utils.NewLoggedConn(
			conn,
			conn, //由于之前已经读取了一部分数据 需要覆盖 reader 还原buffer
			session.GetOutBuffer(),
		)
		ssh.HandleSsh(conn2, session)
		return true
	case 23:
		conn2 := utils.NewLoggedConn(
			conn,
			conn, //由于之前已经读取了一部分数据 需要覆盖 reader 还原buffer
			session.GetOutBuffer(),
		)
		relay.HandleTCPRelay(conn2, session, map[string]string{"targetAddr": "127.0.0.1:4524", "sendSession": "false"})
		return true
	case 6379:
		conn2 := utils.NewLoggedConn(
			conn,
			conn,
			session.GetOutBuffer(),
		)
		return redis.HandleRedis(conn2, session)
	}

	return false
}
func MagicDistributor(conn net.Conn, session *types.Session) bool {
	firstByte := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	n, err := conn.Read(firstByte)
	if err != nil {
		conn.Close()
		return false
	}
	session.IsHandled = true
	var conn2 net.Conn
	conn2 = utils.NewLoggedConn(
		conn,
		io.MultiReader(bytes.NewReader(firstByte[0:n]), conn), //由于之前已经读取了一部分数据 需要覆盖 reader 还原buffer
		session.GetOutBuffer(),
	)
	if config.GetConfig().ForwardingRules != nil {
		for _, rule := range *config.GetConfig().ForwardingRules {
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
	if bytes.Equal(firstByte[0:2], []uint8{22, 3}) { //TLS 首包特征 0x1603  TODO 其他协议模拟
		tlsServer := tls.NewTlsServer(conn2, session)
		err := tlsServer.Handle()
		if err != nil {
			io.ReadAll(conn) //出错继续读取
			return false
		}

		session.IsTls = true
		session.SetOutBuffer(new(bytes.Buffer)) // 重置buffer 不记录tls原始数据
		conn2 = utils.NewLoggedConn(
			tlsServer.Conn,
			tlsServer.Conn, //由于之前已经读取了一部分数据 需要覆盖 reader 还原buffer
			session.GetOutBuffer(),
		)
		firstByte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		n, err = conn2.Read(firstByte)
		if err != nil {
			io.ReadAll(conn2) //出错继续读取
			return false
		}
	}
	/* gmTLS */

	/* HTTP */
	if string(firstByte[0:5]) == "POST " || string(firstByte[0:4]) == "GET " || string(firstByte[0:5]) == "HEAD " || string(firstByte[0:8]) == "OPTIONS " || string(firstByte[0:7]) == "DELETE " || string(firstByte[0:4]) == "PUT " || string(firstByte[0:6]) == "TRACE " || string(firstByte[0:8]) == "CONNECT " {
		if config.GetConfig().ForwardingRules != nil {
			for _, rule := range *config.GetConfig().ForwardingRules {
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
	if strings.HasPrefix(string(firstByte), "SSH-") { //SSH
		session.Service = "ssh"
		session.IsHandled = true
		ssh.HandleSsh(conn2, session)
		return true
	}
	/* Redis */
	if string(firstByte) == "*1\r\n" || string(firstByte) == "*2\r\n" || strings.ToLower(string(firstByte)) == "info" || strings.ToLower(string(firstByte)) == "ping" {
		session.Service = "redis"
		session.IsHandled = true
		redis.HandleRedis(conn2, session)
		return true
	}

	/* Other */

	if PortDistributor(conn2, session) {
		return true
	} else {
		session.IsHandled = false
		session.Service = "raw"
		return false
	}

}
