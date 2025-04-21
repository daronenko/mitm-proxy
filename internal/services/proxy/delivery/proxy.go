package httpdelivery

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/daronenko/https-proxy/internal/app/config"
	"github.com/daronenko/https-proxy/internal/model"
	"github.com/daronenko/https-proxy/internal/services/api/repo"
	"github.com/rs/zerolog/log"
)

type Proxy struct {
	repo      *repo.Request
	conf      *config.Config
	key       []byte
	certCache sync.Map
}

func New(repo *repo.Request, conf *config.Config) (*Proxy, error) {
	key, err := os.ReadFile(conf.App.ProxyServer.TLS.KeyPath)
	if err != nil {
		log.Err(err).Msg("failed to read tls key")
		return nil, err
	}

	return &Proxy{
		repo: repo,
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

		if err := d.forwardRequest(tlsClientConn, targetConn, req); err != nil {
			log.Err(err).Msg("failed to forward request from client to target connection over tls")
			return
		}

		targetConn.Close()
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
	originalBody := resp.Body

	bodyBytes, err := io.ReadAll(originalBody)
	if err != nil {
		originalBody.Close()
		return fmt.Errorf("read response body: %w", err)
	}
	originalBody.Close()

	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	respCopy := *resp // shallow copy
	respCopy.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	go func(respCopy *http.Response) {
		transaction := &model.Transaction{
			Request:   model.NewRequest(req),
			Response:  model.NewResponse(respCopy),
			CreatedAt: time.Now(),
		}
		if _, err := d.repo.CreateTransaction(context.Background(), transaction); err != nil {
			log.Err(err).Msg("failed to store transaction")
		}
	}(&respCopy)

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
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		log.Err(err).Msg("failed to dial tcp connection")
		return nil, fmt.Errorf("tcp dial: %w", err)
	}

	return conn, nil
}

func (d *Proxy) secureConn(address string, tlsConfig *tls.Config) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
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
		log.Warn().Err(err).Msgf("certificate not found for host '%s', generating...", host)

		certBytes, err = genCert(host, big.NewInt(time.Now().UnixNano()))
		if err != nil {
			return nil, fmt.Errorf("failed to generate certificate: %w", err)
		}

		if err := os.WriteFile(certPath, certBytes, 0644); err != nil {
			log.Err(err).Msg("failed to write generated certificate to file")
			return nil, fmt.Errorf("failed to save certificate: %w", err)
		}
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
