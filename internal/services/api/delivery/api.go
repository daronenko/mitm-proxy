package httpdelivery

import (
	"net/http"

	"github.com/daronenko/https-proxy/internal/app/config"
	"github.com/daronenko/https-proxy/internal/httpserver"
	"go.uber.org/fx"
)

type Api struct {
	fx.In
	Conf *config.Config
}

func Init(d Api, api *httpserver.ApiRouter) {
	api.HandleFunc("/ping", d.Ping).Methods("GET")
}

func (d *Api) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
