package ingress

import (
	"context"
	"hachimi/pkg/config"
	"log"
	"net"
	"sync"
)

// UDPListener represents a managed UDP listener.
type UDPListener struct {
	Host      string
	Port      int
	conn      *net.UDPConn
	wg        *sync.WaitGroup
	transport bool
}

// NewUDPListener creates a new UDPListener instance.
func NewUDPListener(host string, port int) *UDPListener {
	return &UDPListener{
		Host: host,
		Port: port,
		wg:   &sync.WaitGroup{},
	}
}

// Start begins listening for UDP packets on the specified address.
func (u *UDPListener) Start(ctx context.Context, handler func(*net.UDPConn, *net.UDPAddr, *net.UDPAddr, []byte)) error {
	addr := &net.UDPAddr{
		IP:   net.ParseIP(u.Host),
		Port: u.Port,
	}
	var err error
	u.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	file, err := u.conn.File()
	if err == nil {
		defer file.Close()
		// TransparentProxy
		err = TransparentProxy(file)
		if err != nil {
			log.Printf("Warning: Failed to set socket option (IP_TRANSPARENT/IP_RECVORIGDSTADDR): %s\n", err)
			log.Printf("Fallback to normal UDP listener. Full Port Forwarding is not available.\n")
		} else {
			u.transport = true
		}
	} else {
		log.Printf("Warning: Failed to get socket file descriptor: %s\n", err)
		log.Printf("Fallback to normal UDP listener. Full Port Forwarding is not available.\n")
	}

	u.wg.Add(1)
	go func() {
		defer u.wg.Done()
		buf := make([]byte, 65535)
		oob := make([]byte, 2048)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, oobN, _, src, err := u.conn.ReadMsgUDP(buf, oob)
				if err != nil {
					log.Printf("Failed to read UDP packet: %s\n", err)
					continue
				}
				var dst *net.UDPAddr
				if u.transport {
					origDst, err := getOrigDst(oob, oobN)
					if err == nil {
						dst = origDst
					} else {
						if config.GetPotConfig().Debug {
							log.Printf("Failed to get original destination: %s\n", err)
						}
					}
				}

				handler(u.conn, src, dst, buf[:n])
			}
		}
	}()
	return nil

}

// Stop gracefully stops the UDP listener.
func (u *UDPListener) Stop() error {
	if u.conn != nil {
		err := u.conn.Close()
		u.wg.Wait()
		return err
	}
	return nil
}
