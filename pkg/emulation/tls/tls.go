package tls

import (
	"crypto/tls"
	"hachimi/pkg/config"
	"hachimi/pkg/types"
	"net"
)

type TlsServer struct {
	Conn    net.Conn
	session *types.Session
}

func NewTlsServer(conn net.Conn, session *types.Session) *TlsServer {
	return &TlsServer{Conn: tls.Server(conn, config.TlsConfig), session: session}
}

func (t *TlsServer) Handle() error {
	return t.Conn.(*tls.Conn).Handshake()
}
