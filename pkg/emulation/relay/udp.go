package relay

import (
	"hachimi/pkg/types"
	"net"
	"time"
)

func HandleUDPRelay(src net.Conn, session *types.Session, buf []byte, config map[string]string) []byte {
	targetAddr := config["targetAddr"] //目标地址
	dst, err := net.Dial("udp", targetAddr)
	if err != nil {
		return nil
	}
	defer dst.Close()
	dst.Write(buf)
	// 读取 等待2s
	dst.SetReadDeadline(time.Now().Add(time.Second * 2))
	var buffer = make([]byte, 1024)
	n, err := dst.Read(buffer)
	if n == 0 {
		return nil
	}
	return buffer[:n]

}
