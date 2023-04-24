package main

import (
	"egressproxy"
	"flag"
	"sync"
)

var (
	httpListenAddr string
	tlsListenAddr  string
)

func init() {
	flag.StringVar(&httpListenAddr, "httpListenAddr", ":38000", "Address to listen on for HTTP_PROXY connections")
	flag.StringVar(&tlsListenAddr, "tlsListenAddr", "127.0.0.1:38443", "Address to listen on for internal CONNECT egressproxy connections")
	flag.Parse()
}

func main() {

	wg := sync.WaitGroup{}
	wg.Add(2)

	httpProxy := egressproxy.NewHTTPProxy(&wg, tlsListenAddr)
	tlsProxy := egressproxy.NewTLSProxy(&wg)

	go runServer(httpProxy, httpListenAddr)
	go runServer(tlsProxy, tlsListenAddr)

	wg.Wait()
}

func runServer(l egressproxy.Listener, addr string) {
	l.Listen(addr)
}
