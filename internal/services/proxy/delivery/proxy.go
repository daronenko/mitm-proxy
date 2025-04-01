package httpdelivery

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/daronenko/https-proxy/internal/app/config"
	"github.com/rs/zerolog/log"
)

type Proxy struct {
	conf      *config.Config
	key       []byte
	certCache sync.Map
}

func New(conf *config.Config) (*Proxy, error) {
	key, err := os.ReadFile(conf.App.ProxyServer.TLS.KeyPath)
	if err != nil {
		log.Err(err).Msg("failed to read tls key")
		return nil, err
	}

	return &Proxy{
		conf: conf,
		key:  key,
	}, nil
}

func (d *Proxy) Proxy(clientConn net.Conn, req *http.Request) {
	if req.Method == http.MethodConnect {
		d.httpsStrategy(clientConn, req)
	} else {
		d.httpStrategy(clientConn, req)
	}
}

func (d *Proxy) httpsStrategy(clientConn net.Conn, req *http.Request) {
	if _, err := clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n")); err != nil {
		return
	}

	tlsConfig, err := d.getTLSConfig(req.URL.Hostname())
	if err != nil {
		log.Err(err).Msg("failed to get tls config")
		return
	}

	tlsClientConn := tls.Server(clientConn, tlsConfig)
	defer tlsClientConn.Close()

	for {
		req, err := http.ReadRequest(bufio.NewReader(tlsClientConn))
		if err == io.EOF {
			break
		} else if err != nil {
			log.Err(err).Msg("failed to read request")
			return
		}

		targetConn, err := d.secureConn(net.JoinHostPort(req.Host, "443"), tlsConfig)
		if err != nil {
			log.Err(err).Msg("failed to establish tls connection")
			return
		}
		defer targetConn.Close()

		if err := d.forwardRequest(tlsClientConn, targetConn, req); err != nil {
			log.Err(err).Msg("failed to forward request from client to target connection over tls")
			return
		}
	}
}

func (d *Proxy) httpStrategy(clientConn net.Conn, req *http.Request) {
	targetConn, err := d.tcpConn(net.JoinHostPort(req.Host, getPort(req.URL)))
	if err != nil {
		log.Err(err).Msg("failed to establish tcp connection")
		return
	}
	defer targetConn.Close()

	if err := d.forwardRequest(clientConn, targetConn, req); err != nil {
		log.Err(err).Msg("failed to forward request from client to target connection")
		return
	}
}

func (d *Proxy) forwardRequest(clientConn, targetConn net.Conn, req *http.Request) error {
	hideProxy(req)

	resp, err := d.sendRequest(targetConn, req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := resp.Write(clientConn); err != nil {
		log.Err(err).Msg("failed to write response from target to client connection")
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func (d *Proxy) sendRequest(targetConn net.Conn, req *http.Request) (*http.Response, error) {
	if err := req.Write(targetConn); err != nil {
		log.Err(err).Msg("failed to write request from client to target connection")
		return nil, fmt.Errorf("write request: %w", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(targetConn), req)
	if err != nil {
		log.Err(err).Msg("failed to read response from target connection")
		return nil, fmt.Errorf("read response: %w", err)
	}

	return resp, nil
}

func (d *Proxy) tcpConn(address string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		log.Err(err).Msg("failed to dial tcp connection")
		return nil, fmt.Errorf("tcp dial: %w", err)
	}

	return conn, nil
}

func (d *Proxy) secureConn(address string, tlsConfig *tls.Config) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	if err != nil {
		log.Err(err).Msg("failed to dial tls connection")
		return nil, fmt.Errorf("tls dial: %w", err)
	}

	return conn, nil
}

func (d *Proxy) getTLSConfig(host string) (*tls.Config, error) {
	if certData, exists := d.certCache.Load(host); exists {
		certBytes := certData.([]byte)
		return d.createTLSConfig(certBytes)
	}

	certPath := fmt.Sprintf("%s/%s.crt", d.conf.App.ProxyServer.TLS.CertPath, host)
	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		log.Err(err).Msg("failed to read certificate")
		return nil, fmt.Errorf("read host cert: %w", err)
	}

	d.certCache.Store(host, certBytes)

	return d.createTLSConfig(certBytes)
}

func (d *Proxy) createTLSConfig(certBytes []byte) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(certBytes, d.key)
	if err != nil {
		log.Err(err).Msg("failed to parse key pair from pem encoded data")
		return nil, fmt.Errorf("tls x509 key pair: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}
