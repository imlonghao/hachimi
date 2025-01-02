package types

import (
	"net"
)

type ProtocolHandler interface {
	Handle(conn net.Conn, session *Session)
}
