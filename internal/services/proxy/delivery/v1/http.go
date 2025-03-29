package httpdelivery

import (
	"net/http"
	"net/url"
)

func copyHeaders(w http.ResponseWriter, src http.Header) {
	headers := w.Header()
	for key, values := range src {
		for _, value := range values {
			headers.Add(key, value)
		}
	}
}

func copyCookies(w http.ResponseWriter, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		http.SetCookie(w, cookie)
	}
}

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
