package httpdelivery

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/daronenko/https-proxy/internal/app/config"
	"github.com/rs/zerolog/log"
)

type Proxy struct {
	conf *config.Config

	certCache map[string][]byte
	mutex     sync.Mutex
	key       []byte
}

func New(conf *config.Config) *Proxy {
	key, _ := os.ReadFile(conf.App.ProxyServer.TLS.KeyPath)

	return &Proxy{
		conf:      conf,
		certCache: make(map[string][]byte),
		key:       key,
	}
}

// go run cmd/proxy/main.go --config config/config.yaml
// curl -x http://localhost:8080 https://mail.ru -vv
func (p *Proxy) Proxy(clientConn net.Conn, req *http.Request) {
	if req.Method == http.MethodConnect {
		p.httpsStrategy(clientConn, req)
	} else {
		p.httpStrategy(clientConn, req)
	}
}

func (p *Proxy) httpsStrategy(clientConn net.Conn, req *http.Request) {
	if _, err := clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n")); err != nil {
		return
	}

	tlsConfig, err := p.getTLSConfig(req.URL.Hostname())
	if err != nil {
		log.Err(err).Msg("failed to get tls config")
		return
	}

	tlsClientConn := tls.Server(clientConn, tlsConfig)
	defer tlsClientConn.Close()

	if err := tlsClientConn.Handshake(); err != nil {
		log.Err(err).Msg("failed to complete tls handshake")
		return
	}

	request, err := http.ReadRequest(bufio.NewReader(tlsClientConn))
	if err != nil {
		log.Err(err).Msg("failed to read request")
		return
	}

	targetConn, err := p.secureConn(net.JoinHostPort(request.Host, getPort(req.URL)), tlsConfig)
	if err != nil {
		return
	}
	defer targetConn.Close()

	hideProxy(req)
	p.forwardRequest(tlsClientConn, targetConn, request)
}

func (p *Proxy) httpStrategy(clientConn net.Conn, req *http.Request) {
	targetConn, err := p.tcpConn(net.JoinHostPort(req.Host, getPort(req.URL)))
	if err != nil {
		return
	}
	defer targetConn.Close()

	hideProxy(req)
	p.forwardRequest(clientConn, targetConn, req)
}

func (p *Proxy) forwardRequest(clientConn, targetConn net.Conn, req *http.Request) {
	if err := req.Write(targetConn); err != nil {
		return
	}

	resp, err := http.ReadResponse(bufio.NewReader(targetConn), req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	_ = resp.Write(clientConn)
}

func (d *Proxy) tcpConn(address string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", address, http.DefaultClient.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to init a tcp connection: %w", err)
	}

	return conn, nil
}

func (d *Proxy) secureConn(address string, tlsConfig *tls.Config) (net.Conn, error) {
	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to init a secure tcp connection: %w", err)
	}

	return conn, nil
}

func (p *Proxy) getTLSConfig(host string) (*tls.Config, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, exists := p.certCache[host]; !exists {
		cert, err := os.ReadFile(fmt.Sprintf("certs/%s.crt", host))
		if err != nil {
			return nil, errors.New("failed to read certificate")
		}
		p.certCache[host] = cert
	}

	cert, err := tls.X509KeyPair(p.certCache[host], p.key)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}
