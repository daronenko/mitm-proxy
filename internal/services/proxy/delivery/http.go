package httpdelivery

import (
	"net/http"
	"net/url"
)

var proxyHeaders = []string{
	"Proxy-Authenticate",
	"Proxy-Authorization",
}

func clearProxyHeaders(r *http.Request) {
	for _, header := range proxyHeaders {
		r.Header.Del(header)
	}
}

func updateHost(r *http.Request) {
	r.Header.Set("Host", r.Host)
	r.RequestURI = ""
}

func hideProxy(r *http.Request) {
	clearProxyHeaders(r)
	updateHost(r)
}

func getPort(url *url.URL) string {
	if port := url.Port(); port != "" {
		return port
	}

	if url.Scheme == "https" {
		return "443"
	} else {
		return "80"
	}
}
