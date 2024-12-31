package ingress

import (
	"context"
	"hachimi/pkg/config"
	"hachimi/pkg/emulation/session"
	"hachimi/pkg/utils"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// ListenerManager manages multiple TCP and UDP listeners.
type ListenerManager struct {
	tcpListeners []*TCPListener
	udpListeners []*UDPListener
	wg           *sync.WaitGroup
}

// NewListenerManager creates a new ListenerManager instance.
func NewListenerManager() *ListenerManager {
	return &ListenerManager{
		wg: &sync.WaitGroup{},
	}
}

// AddTCPListener adds a new TCP listener to the manager.
func (m *ListenerManager) AddTCPListener(listener *TCPListener) {
	m.tcpListeners = append(m.tcpListeners, listener)
}

// AddUDPListener adds a new UDP listener to the manager.
func (m *ListenerManager) AddUDPListener(listener *UDPListener) {
	m.udpListeners = append(m.udpListeners, listener)
}

// StartAll starts all managed listeners.
func (m *ListenerManager) StartAll(ctx context.Context) error {
	for _, tcpListener := range m.tcpListeners {
		m.wg.Add(1)
		go func(listener *TCPListener) {
			defer m.wg.Done()
			err := listener.Start(ctx, DefaultTCPHandler)
			if err != nil {
				// Log the error but continue starting other listeners
				DefaultErrorHandler(err)
			}
		}(tcpListener)
	}

	for _, udpListener := range m.udpListeners {
		m.wg.Add(1)
		go func(listener *UDPListener) {
			defer m.wg.Done()
			err := listener.Start(ctx, DefaultUDPHandler)
			if err != nil {
				DefaultErrorHandler(err)
			}
		}(udpListener)
	}
	return nil
}

// StopAll stops all managed listeners.
func (m *ListenerManager) StopAll() {
	for _, tcpListener := range m.tcpListeners {
		tcpListener.Stop()
	}

	for _, udpListener := range m.udpListeners {
		udpListener.Stop()
	}

	m.wg.Wait()
}
func (m *ListenerManager) Wait() {
	m.wg.Wait()
}
func DefaultTCPHandler(conn *net.TCPConn) {
	sess := session.NewSession(conn, nil)
	session.Distributor(conn, sess)
	sess.EndTime = time.Now()
	sess.Duration = int(sess.EndTime.Sub(sess.StartTime).Milliseconds())
	sess.Data = strings.Trim(utils.EscapeBytes(sess.GetOutBuffer().Bytes()), `"`)
	sess.GetOutBuffer().Reset()
	config.Logger.Log(sess)
	sess.Close()

}
func DefaultUDPHandler(conn *net.UDPConn, src *net.UDPAddr, buf []byte) {
	sess := session.NewSession(conn, src)
	session.Distributor(conn, sess)
	sess.EndTime = time.Now()
	sess.Duration = int(sess.EndTime.Sub(sess.StartTime).Milliseconds())
	sess.Data = strings.Trim(utils.EscapeBytes(sess.GetOutBuffer().Bytes()), `"`)
	sess.GetOutBuffer().Reset()
	config.Logger.Log(sess)
	sess.Close()

}

// DefaultErrorHandler handles errors during listener startup.
func DefaultErrorHandler(err error) {
	if err != nil {
		log.Printf("Error starting listener: %v", err)
	}
}
