package session

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"hachimi/pkg/config"
	"hachimi/pkg/emulation/http"
	"hachimi/pkg/emulation/redis"
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
	conn.SetDeadline(time.Now().Add(time.Duration(config.TimeOut) * time.Second))
	session.SetOutBuffer(new(bytes.Buffer))
	if !PortDistributor(conn, session) {
		return MagicDistributor(conn, session)
	} else {
		return true
	}
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
		return true
	case 6379:
		conn2 := utils.NewLoggedConn(
			conn,
			conn,
			session.GetOutBuffer(),
		)
		redis.HandleRedis(conn2, session)
		return true

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
	conn = utils.NewLoggedConn(
		conn,
		io.MultiReader(bytes.NewReader(firstByte[0:n]), conn), //由于之前已经读取了一部分数据 需要覆盖 reader 还原buffer
		session.GetOutBuffer(),
	)
	if bytes.Equal(firstByte[0:2], []uint8{22, 3}) { //TLS 首包特征 0x1603  TODO 其他协议模拟
		tlsServer := tls.NewTlsServer(conn, session)
		err := tlsServer.Handle()
		if err != nil {
			io.ReadAll(conn) //出错继续读取
			return false
		}

		session.IsTls = true
		session.SetOutBuffer(new(bytes.Buffer))
		conn = utils.NewLoggedConn(
			tlsServer.Conn,
			tlsServer.Conn, //由于之前已经读取了一部分数据 需要覆盖 reader 还原buffer
			session.GetOutBuffer(),
		)
		firstByte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		n, err = conn.Read(firstByte)
		if err != nil {
			io.ReadAll(conn) //出错继续读取
			return false
		}
	}
	if string(firstByte[0:5]) == "POST " || string(firstByte[0:4]) == "GET " || string(firstByte[0:5]) == "HEAD " || string(firstByte[0:8]) == "OPTIONS " || string(firstByte[0:7]) == "DELETE " || string(firstByte[0:4]) == "PUT " || string(firstByte[0:6]) == "TRACE " || string(firstByte[0:8]) == "CONNECT " {
		session.Service = "http"
		session.SetOutBuffer(new(bytes.Buffer))
		conn = utils.NewLoggedConn(
			conn,
			io.MultiReader(bytes.NewReader(firstByte[0:n]), conn), //由于之前已经读取了一部分数据 需要覆盖 reader 还原buffer
			session.GetOutBuffer(),
		)
		httpLog := &types.Http{}
		httpLog.StartTime = session.StartTime
		httpLog.ID = uuid.New().String()
		httpLog.SessionID = session.ID
		httpLog.IsTls = session.IsTls
		httpLog.IsGmTls = session.IsGmTls
		httpLog.SrcIP = session.SrcIP
		httpLog.SrcPort = session.SrcPort
		httpLog.DstIP = session.DstIP
		httpLog.DstPort = session.DstPort
		httpLog.Header = make(map[string]string)
		httpLog.BodyParam = make(map[string]string)
		httpLog.UriParam = make(map[string]string)
		httpLog.IsHandled = true
		err := http.ServeHttp(conn, func(fasthttpCtx *fasthttp.RequestCtx) {
			// 在 requestHandlerFunc 中传递 ctx
			http.RequestHandlerFunc(httpLog, fasthttpCtx)
		})
		httpLog.EndTime = time.Now()
		httpLog.Duration = int(httpLog.EndTime.Sub(httpLog.StartTime).Milliseconds())
		if err != nil {
			httpLog.IsHandled = false
			io.ReadAll(conn) //出错继续读取
		}
		session.IsHandled = true
		session.IsHttp = true

		config.Logger.Log(httpLog)
		return true
	}
	if strings.HasPrefix(string(firstByte), "SSH-") { //SSH
		session.Service = "ssh"
		session.IsHandled = true
		ssh.HandleSsh(conn, session)
		return true
	} else if string(firstByte) == "*1\r\n" || string(firstByte) == "*2\r\n" || strings.ToLower(string(firstByte)) == "info" || strings.ToLower(string(firstByte)) == "ping" {
		session.Service = "redis"
		session.IsHandled = true
		redis.HandleRedis(conn, session)
		return true
	} else {
		session.IsHandled = false
		return false
	}
}
