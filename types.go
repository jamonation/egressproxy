package egressproxy

import (
	"crypto/tls"
	"net"
	"net/http"
	"regexp"
	"sync"
)

// Listener gives a Listen() method and ACLs to anything that wants it
type Listener interface {
	Listen(string) error
}

// HTTPProxy implements Listener
type HTTPProxy struct {
	Server           http.Server
	Wg               *sync.WaitGroup
	allowedHosts     []*regexp.Regexp
	allowedURLs      []*regexp.Regexp
	upstreamTLSProxy string // used for CONNECT tunnels
}

// TLSProxy implements Listener
type TLSProxy struct {
	Conn    net.Conn
	TLSConn *tls.Conn
	Wg      *sync.WaitGroup
}
