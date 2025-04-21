package scanner

import (
	"bytes"
	"net/http"

	"github.com/daronenko/https-proxy/internal/model"
)

type CmdInjection struct {
	Payloads []string
}

func (s CmdInjection) Name() string {
	return "Command Injection"
}

func (s CmdInjection) Scan(original model.Request, try func(*http.Request) bool) []string {
	var vulnerable []string

	injections := []struct {
		Name string
		Iter func(payload string)
	}{
		{
			Name: "GET param",
			Iter: func(payload string) {
				for k, v := range original.QueryParams {
					mod := original.Clone()
					mod.QueryParams[k] = v + payload
					req, _ := http.NewRequest(mod.Method, model.BuildURL(mod), bytes.NewReader(mod.Body))
					copyHeadersCookies(req, &mod)
					if try(req) {
						vulnerable = append(vulnerable, "GET param: "+k)
						break
					}
				}
			},
		},
		{
			Name: "POST param",
			Iter: func(payload string) {
				if original.Method != "POST" {
					return
				}
				for k, v := range original.FormParams {
					mod := original.Clone()
					mod.FormParams[k] = v + payload
					body := model.BuildFormBody(mod.FormParams)
					req, _ := http.NewRequest("POST", model.BuildURL(mod), bytes.NewReader([]byte(body)))
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					copyHeadersCookies(req, &mod)
					if try(req) {
						vulnerable = append(vulnerable, "POST param: "+k)
						break
					}
				}
			},
		},
		{
			Name: "Header",
			Iter: func(payload string) {
				for k, v := range original.Headers {
					mod := original.Clone()
					mod.Headers[k] = v + payload
					req, _ := http.NewRequest(mod.Method, model.BuildURL(mod), bytes.NewReader(mod.Body))
					copyHeadersCookies(req, &mod)
					if try(req) {
						vulnerable = append(vulnerable, "Header: "+k)
						break
					}
				}
			},
		},
		{
			Name: "Cookie",
			Iter: func(payload string) {
				for k, v := range original.Cookies {
					mod := original.Clone()
					mod.Cookies[k] = v + payload
					req, _ := http.NewRequest(mod.Method, model.BuildURL(mod), bytes.NewReader(mod.Body))
					copyHeadersCookies(req, &mod)
					if try(req) {
						vulnerable = append(vulnerable, "Cookie: "+k)
						break
					}
				}
			},
		},
	}

	for _, payload := range s.Payloads {
		for _, inj := range injections {
			inj.Iter(payload)
		}
	}

	return vulnerable
}

func copyHeadersCookies(req *http.Request, r *model.Request) {
	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	for name, val := range r.Cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: val})
	}
}
