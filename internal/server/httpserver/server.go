package httpserver

import (
	"bufio"
	"net"
	"net/http"

	"github.com/daronenko/https-proxy/internal/app/config"
	"github.com/daronenko/https-proxy/internal/server/httprouter"
	httpdelivery "github.com/daronenko/https-proxy/internal/services/proxy/delivery/v1"
	"github.com/rs/zerolog/log"
)

type ProxyServer struct {
	proxy *httpdelivery.Proxy
}

func NewProxyServer(proxy *httpdelivery.Proxy, config *config.Config) *ProxyServer {
	return &ProxyServer{proxy}
}

func (s *ProxyServer) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Err(err).Msg("failed to accept connection")
			continue
		}
		go s.handleConnection(conn)
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

func NewApiServer(api *httprouter.ApiRouter, config *config.Config) *ApiServer {
	return &ApiServer{
		http.Server{
			Handler: api,
		},
	}
}
