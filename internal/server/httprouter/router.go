package httprouter

import (
	"github.com/gorilla/mux"
)

type ApiRouter struct {
	*mux.Router
}

func New() *ApiRouter {
	return &ApiRouter{
		mux.NewRouter(),
	}
}
