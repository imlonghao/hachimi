package ingress

import (
	"context"
	"log"
	"net"
	"strconv"
	"sync"
)

// TCPListener represents a managed TCP listener.
type TCPListener struct {
	Host     string
	Port     int
	listener *net.TCPListener
	wg       *sync.WaitGroup
}

// NewTCPListener creates a new TCPListener instance.
func NewTCPListener(address string) *TCPListener {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		log.Fatalf("Failed to split TCP listener address: %s\n", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Failed to parse TCP listener port: %s\n", err)
	}
	return &TCPListener{
		Host: host,
		Port: port,
		wg:   &sync.WaitGroup{},
	}
}

// Start begins listening for TCP connections on the specified address.
func (t *TCPListener) Start(ctx context.Context, handler func(conn *net.TCPConn)) error {
	addr := &net.TCPAddr{
		IP:   net.ParseIP(t.Host),
		Port: t.Port,
	}
	var err error
	t.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	file, err := t.listener.File()
	if err == nil {
		defer file.Close()
		// TransparentProxy
		err = TransparentProxy(file)
		if err != nil {
			log.Printf("Warning: Failed to set socket option (IP_TRANSPARENT/IP_RECVORIGDSTADDR): %s\n", err)
			log.Printf("Fallback to normal TCP listener. Full Port Forwarding is not available.\n")
		}

	} else {
		log.Printf("Warning: Failed to get socket file descriptor: %s\n", err)
		log.Printf("Fallback to normal TCP listener. Full Port Forwarding is not available.\n")
	}

	log.Printf("TCP listener started on %s", t.listener.Addr().String())
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Printf("TCP listener on %s is stopping", t.listener.Addr().String())
				return
			default:
				conn, err := t.listener.AcceptTCP()
				if err != nil {
					if ctx.Err() != nil {
						// Context canceled, listener stopped
						return
					}
					log.Printf("Error accepting TCP connection: %v", err)
					continue
				}
				// Handle the connection in a separate goroutine
				go handler(conn)
			}
		}
	}()
	return nil
}

// Stop gracefully stops the TCP listener.
func (t *TCPListener) Stop() error {
	if t.listener != nil {
		err := t.listener.Close()
		t.wg.Wait()
		return err
	}
	return nil
}
