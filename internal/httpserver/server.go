package httpserver

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/daronenko/https-proxy/internal/app/config"
	httpdelivery "github.com/daronenko/https-proxy/internal/services/proxy/delivery"
	"github.com/rs/zerolog/log"
)

type ProxyServer struct {
	proxy    *httpdelivery.Proxy
	listener net.Listener
	wg       sync.WaitGroup
	shutdown chan struct{}
}

func NewProxyServer(proxy *httpdelivery.Proxy, config *config.Config) *ProxyServer {
	return &ProxyServer{
		proxy:    proxy,
		shutdown: make(chan struct{}),
	}
}

func (s *ProxyServer) Serve(listener net.Listener) error {
	s.listener = listener

	for {
		select {
		case <-s.shutdown:
			log.Info().Msg("gracefully shutting down proxy server...")
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Err(err).Msg("failed to accept connection")
				continue
			}

			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				s.handleConnection(conn)
			}()
		}
	}
}

func (s *ProxyServer) Shutdown(ctx context.Context) error {
	close(s.shutdown)

	if s.listener != nil {
		s.listener.Close()
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("proxy server gracefully stopped")
		return nil
	case <-ctx.Done():
		log.Warn().Msg("proxy server shutdown timed out, forcing exit")
		return errors.New("proxy server shutdown timed out")
	}
}

func (s *ProxyServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		log.Err(err).Msg("failed to read http request")
		return
	}

	s.proxy.Proxy(conn, request)
}

type ApiServer struct {
	http.Server
}

func NewApiServer(api *ApiRouter, config *config.Config) *ApiServer {
	return &ApiServer{
		http.Server{
			Handler: api,
		},
	}
}
