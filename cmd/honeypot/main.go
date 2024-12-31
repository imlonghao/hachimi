// 蜜罐 蜜网的最小组成部分 可单机独立运行
package main

import (
	"context"
	"flag"
	"hachimi/pkg/ingress"
)

var (
	address = flag.String("a", "0.0.0.0:12345", "Address to listen on")
)

func main() {
	flag.Parse()
	//ListenerManager
	lm := ingress.ListenerManager{}
	//TCPListener
	tcpListener := ingress.NewTCPListener(*address)
	//NewListenerManager
	lm = *ingress.NewListenerManager()
	//AddTCPListener
	lm.AddTCPListener(tcpListener)
	udpListener := ingress.NewUDPListener(*address)
	//AddUDPListener
	lm.AddUDPListener(udpListener)
	lm.StartAll(context.Background())

	select {}
}
