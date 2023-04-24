package egressproxy

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

// NewTLSProxy returns a TLSProxy that implements a Listener
func NewTLSProxy(wg *sync.WaitGroup) Listener {
	return &TLSProxy{
		Conn: nil,
		Wg:   wg,
	}
}

func NewTLSConfig() *tls.Config {
	return &tls.Config{
		GetCertificate: func(hInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
			certPEM, certPrivKeyPEM := makeCert(hInfo.ServerName)
			serverCert, err := tls.X509KeyPair(certPEM, certPrivKeyPEM)
			if err != nil {
				return nil, err
			}
			return &serverCert, nil
		},
	}
}

// Listen waits for incoming connects for anything using HTTPS_PROXY
// It generates certificates and signs them on demand using the Academy CA key pair
func (s *TLSProxy) Listen(laddr string) error {
	listener, err := tls.Listen("tcp", laddr, NewTLSConfig())
	if err != nil {
		return err

	}

	for {
		s.Conn, err = listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}

		go s.handleTLSConn()
	}

	return nil
}

func (s *TLSProxy) handleTLSConn() {
	defer s.Conn.Close()

	var ok bool
	s.TLSConn, ok = s.Conn.(*tls.Conn)
	if !ok {
		fmt.Printf("error establishing TLS session to %v\n", s.Conn.RemoteAddr().String())
		return
	}

	// docs say low level connection manipulation needs these
	err := s.TLSConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		fmt.Printf("tls setdeadline error: %v\n", err)
		return
	}
	err = s.TLSConn.Handshake()
	if err != nil {
		fmt.Printf("tls handshake error: %v\n", err)
		return
	}

	reader := bufio.NewReader(s.TLSConn)

	req, err := http.ReadRequest(reader)
	if err != nil {
		fmt.Printf("error reading request: %v\n", err)
		return
	}

	req.URL.Scheme = "https"
	req.URL.Host = req.Host
	req.RequestURI = "" // unset this for client requests

	log.Printf("Processing CONNECT request to: %v for %v\n", req.URL.String(), req.Header.Get("X-Forwarded-For"))
	client := http.Client{
		Transport: &http.Transport{
			TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()
	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		fmt.Printf("error dumping response: %v\n", err)
		return
	}

	_, _ = s.TLSConn.Write(b)
	s.TLSConn.Close()
}
