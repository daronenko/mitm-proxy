package model

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"
)

type Transaction struct {
	ID        interface{} `bson:"_id,omitempty" json:"id,omitempty"`
	Request   Request     `bson:"request" json:"request"`
	Response  Response    `bson:"response" json:"response"`
	CreatedAt time.Time   `bson:"created_at" json:"created_at"`
}

type Request struct {
	Method      string            `bson:"method" json:"method"`
	Version     string            `bson:"version" json:"version"`
	Host        string            `bson:"host" json:"host"`
	Path        string            `bson:"path" json:"path"`
	Protocol    string            `bson:"protocol" json:"protocol"`
	Headers     map[string]string `bson:"headers" json:"headers"`
	Cookies     map[string]string `bson:"cookies" json:"cookies"`
	QueryParams map[string]string `bson:"query_params" json:"query_params"`
	FormParams  map[string]string `bson:"form_params" json:"form_params"`
	Body        []byte            `bson:"body" json:"body"`
}

type Response struct {
	Status  int               `bson:"status" json:"status"`
	Headers map[string]string `bson:"headers" json:"headers"`
	Body    []byte            `bson:"body" json:"-"`
}

func NewRequest(req *http.Request) Request {
	var bodyBytes []byte
	if req.Body != nil {
		encoding := req.Header.Get("Content-Encoding")
		switch encoding {
		case "gzip":
			gzReader, _ := gzip.NewReader(req.Body)
			defer gzReader.Close()
			bodyBytes, _ = io.ReadAll(gzReader)
		default:
			bodyBytes, _ = io.ReadAll(req.Body)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	headers := make(map[string]string)
	for key, values := range req.Header {
		headers[key] = strings.Join(values, ", ")
	}

	cookies := make(map[string]string)
	for _, c := range req.Cookies() {
		cookies[c.Name] = c.Value
	}

	queryParams := make(map[string]string)
	for key, values := range req.URL.Query() {
		queryParams[key] = strings.Join(values, ", ")
	}

	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	formParams := make(map[string]string)
	if err := req.ParseForm(); err == nil {
		for key, values := range req.PostForm {
			formParams[key] = strings.Join(values, ", ")
		}
	}

	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}

	return Request{
		Method:      req.Method,
		Version:     req.Proto,
		Host:        req.Host,
		Path:        req.URL.Path,
		Protocol:    scheme,
		Headers:     headers,
		Cookies:     cookies,
		QueryParams: queryParams,
		FormParams:  formParams,
		Body:        bodyBytes,
	}
}

func (r Request) Clone() Request {
	cloneMap := func(src map[string]string) map[string]string {
		dst := make(map[string]string, len(src))
		maps.Copy(dst, src)
		return dst
	}

	return Request{
		Method:      r.Method,
		Version:     r.Version,
		Host:        r.Host,
		Path:        r.Path,
		Protocol:    r.Protocol,
		Headers:     cloneMap(r.Headers),
		Cookies:     cloneMap(r.Cookies),
		QueryParams: cloneMap(r.QueryParams),
		FormParams:  cloneMap(r.FormParams),
		Body:        slices.Clone(r.Body),
	}
}

func NewResponse(resp *http.Response) Response {
	var bodyBytes []byte
	if resp.Body != nil {
		bodyBytes, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	return Response{
		Status:  resp.StatusCode,
		Headers: headers,
		Body:    bodyBytes,
	}
}

func BuildURL(req Request) string {
	var builder strings.Builder

	scheme := req.Protocol
	if scheme == "" {
		scheme = "http"
	}
	builder.WriteString(scheme)
	builder.WriteString("://")

	builder.WriteString(req.Host)
	builder.WriteString(req.Path)

	if len(req.QueryParams) > 0 {
		builder.WriteString("?")
		queries := make([]string, 0, len(req.QueryParams))
		for k, v := range req.QueryParams {
			queries = append(queries, fmt.Sprintf("%s=%s", k, v))
		}
		builder.WriteString(strings.Join(queries, "&"))
	}

	return builder.String()
}

func BuildFormBody(form map[string]string) string {
	parts := make([]string, 0, len(form))
	for k, v := range form {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, "&")
}
