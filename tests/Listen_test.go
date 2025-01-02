package tests

import (
	"context"
	"hachimi/pkg/ingress"
	"testing"
)

func TestListen(t *testing.T) {
	//ListenerManager
	lm := ingress.ListenerManager{}
	//TCPListener
	tcpListener := ingress.NewTCPListener("0.0.0.0", 8777)
	//NewListenerManager
	lm = *ingress.NewListenerManager()
	//AddTCPListener
	lm.AddTCPListener(tcpListener)
	tcpListener = ingress.NewTCPListener("0.0.0.0", 6379)
	//NewListenerManager
	lm = *ingress.NewListenerManager()
	//AddTCPListener
	lm.AddTCPListener(tcpListener)
	lm.StartAll(context.Background())

	select {}

}
