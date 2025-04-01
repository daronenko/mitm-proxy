package config

import (
	"github.com/daronenko/https-proxy/pkg/logger"
)

type Spec struct {
	ProxyServer HttpServerSpec `mapstructure:"proxyServer"`
	ApiServer   HttpServerSpec `mapstructure:"apiServer"`
	Logger      logger.Config  `mapstructure:"logger"`
}

type HttpServerSpec struct {
	Address string  `mapstructure:"address"`
	TLS     TLSSpec `mapstructure:"tls"`
}

type TLSSpec struct {
	CertPath string `mapstructure:"certPath"`
	KeyPath  string `mapstructure:"keyPath"`
}
