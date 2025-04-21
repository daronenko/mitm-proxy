package config

import (
	"github.com/daronenko/https-proxy/pkg/logger"
)

type Spec struct {
	ProxyServer HttpServerSpec `mapstructure:"proxyServer"`
	ApiServer   HttpServerSpec `mapstructure:"apiServer"`
	Logger      logger.Config  `mapstructure:"logger"`
	Mongo       MongoSpec      `mapstructure:"mongo"`
}

type HttpServerSpec struct {
	Address string  `mapstructure:"address"`
	TLS     TLSSpec `mapstructure:"tls"`
}

type TLSSpec struct {
	CertPath string `mapstructure:"certPath"`
	KeyPath  string `mapstructure:"keyPath"`
}

type MongoSpec struct {
	URI         string               `mapstructure:"uri"`
	Username    string               `mapstructure:"username"`
	Password    string               `mapstructure:"password"`
	Database    string               `mapstructure:"database"`
	Collections MongoCollectionsSpec `mapstructure:"collections"`
}

type MongoCollectionsSpec struct {
	Transactions string `mapstructure:"transactions"`
}
