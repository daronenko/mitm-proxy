package httpserver

import (
	"github.com/gorilla/mux"
)

type ApiRouter struct {
	*mux.Router
}

func NewApiRouter() *ApiRouter {
	return &ApiRouter{
		mux.NewRouter(),
	}
}
