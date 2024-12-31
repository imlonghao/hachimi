package session

import (
	"github.com/google/uuid"
	"hachimi/pkg/types"
	"net"
	"strconv"
	"time"
)

func NewSession(conn interface{}, addr interface{}) *types.Session {
	s := &types.Session{}
	s.SetConnection(conn)
	s.ID = uuid.New().String()
	// conn 是否是 *net.UDPConn
	if _, ok := conn.(*net.TCPConn); ok {
		s.Protocol = "TCP"
		//兼容IPV6  net.SplitHostPort(address)
		var port string
		s.SrcIP, port, _ = net.SplitHostPort(conn.(*net.TCPConn).RemoteAddr().String())
		s.SrcPort, _ = strconv.Atoi(port)
		s.DstIP, port, _ = net.SplitHostPort(conn.(*net.TCPConn).LocalAddr().String())
		s.DstPort, _ = strconv.Atoi(port)
	} else if _, ok := conn.(*net.UDPConn); ok {
		s.Protocol = "UDP"
		var port string
		s.SrcIP, port, _ = net.SplitHostPort(conn.(*net.UDPConn).RemoteAddr().String())
		s.SrcPort, _ = strconv.Atoi(port)
		s.DstIP, port, _ = net.SplitHostPort(conn.(*net.UDPConn).LocalAddr().String())
		s.DstPort, _ = strconv.Atoi(port)
		if addr != nil {
			s.DstIP, port, _ = net.SplitHostPort(addr.(*net.UDPAddr).String())
			s.DstPort, _ = strconv.Atoi(port)
		}
	}
	s.SetAddr(addr)
	s.StartTime = time.Now()
	return s
}
