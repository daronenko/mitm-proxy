package httpdelivery

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/daronenko/https-proxy/internal/app/config"
	"github.com/daronenko/https-proxy/internal/httpserver"
	"github.com/daronenko/https-proxy/internal/model"
	"github.com/daronenko/https-proxy/internal/services/api/repo"
	"github.com/daronenko/https-proxy/pkg/httpctl"
	"github.com/daronenko/https-proxy/pkg/scanner"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/fx"
)

const RequestID = "request_id"

type Api struct {
	fx.In
	Conf *config.Config
	Repo *repo.Request
}

func Init(d Api, api *httpserver.ApiRouter) {
	api.HandleFunc("/ping", d.Ping).Methods("GET")

	api.HandleFunc("/requests", d.RequestsList).Methods("GET")
	api.HandleFunc("/request/{request_id}", d.GetRequestByID).Methods("GET")
	api.HandleFunc("/repeat/{request_id}", d.RepeatRequestByID).Methods("POST")
	api.HandleFunc("/scan/{request_id}", d.ScanRequestByID).Methods("POST")
}

func (d *Api) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func (d *Api) RequestsList(w http.ResponseWriter, r *http.Request) {
	requests, err := d.Repo.GetTransactionsList(context.Background())
	if err != nil {
		httpctl.ErrorResponse(w, http.StatusInternalServerError, "failed to get requests")
		return
	}

	if len(requests) == 0 {
		httpctl.ErrorResponse(w, http.StatusNotFound, "requests not found")
		return
	}

	httpctl.JsonResponse(w, http.StatusOK, requests)
}

func (d *Api) GetRequestByID(w http.ResponseWriter, r *http.Request) {
	requestIDStr, present := mux.Vars(r)["request_id"]
	if !present {
		httpctl.ErrorResponse(w, http.StatusNotFound, "request id not found")
		return
	}

	requestID, err := bson.ObjectIDFromHex(requestIDStr)
	if err != nil {
		httpctl.ErrorResponse(w, http.StatusBadRequest, "invalid request id format")
		return
	}

	request, err := d.Repo.GetTransactionByID(context.Background(), requestID)
	if err != nil {
		httpctl.ErrorResponse(w, http.StatusNotFound, "request not found")
		return
	}

	httpctl.JsonResponse(w, http.StatusOK, request)
}

func (d *Api) RepeatRequestByID(w http.ResponseWriter, r *http.Request) {
	requestIDStr, present := mux.Vars(r)["request_id"]
	if !present {
		httpctl.ErrorResponse(w, http.StatusNotFound, "request id not found")
		return
	}

	requestID, err := bson.ObjectIDFromHex(requestIDStr)
	if err != nil {
		httpctl.ErrorResponse(w, http.StatusBadRequest, "invalid request id format")
		return
	}

	transaction, err := d.Repo.GetTransactionByID(context.Background(), requestID)
	if err != nil {
		httpctl.ErrorResponse(w, http.StatusNotFound, "original request not found")
		return
	}

	originalReq := transaction.Request
	url := model.BuildURL(originalReq)

	req, err := http.NewRequest(originalReq.Method, url, bytes.NewReader(originalReq.Body))
	if err != nil {
		httpctl.ErrorResponse(w, http.StatusInternalServerError, "failed to construct repeated request")
		return
	}

	for k, v := range originalReq.Headers {
		req.Header.Set(k, v)
	}

	for name, val := range originalReq.Cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: val})
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		httpctl.ErrorResponse(w, http.StatusBadGateway, "failed to perform repeated request")
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)

	for k, v := range resp.Header {
		for _, hv := range v {
			w.Header().Add(k, hv)
		}
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("failed to write response body")
	}
}

type VulnerabilityScanner interface {
	Name() string
	Scan(original model.Request, try func(*http.Request) bool) []string
}

func (d *Api) ScanRequestByID(w http.ResponseWriter, r *http.Request) {
	requestIDStr, present := mux.Vars(r)["request_id"]
	if !present {
		httpctl.ErrorResponse(w, http.StatusNotFound, "request id not found")
		return
	}

	requestID, err := bson.ObjectIDFromHex(requestIDStr)
	if err != nil {
		httpctl.ErrorResponse(w, http.StatusBadRequest, "invalid request id format")
		return
	}

	transaction, err := d.Repo.GetTransactionByID(context.Background(), requestID)
	if err != nil {
		httpctl.ErrorResponse(w, http.StatusNotFound, "original request not found")
		return
	}
	originalReq := transaction.Request

	try := func(modifiedReq *http.Request) bool {
		resp, err := http.DefaultClient.Do(modifiedReq)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return bytes.Contains(body, []byte("root:"))
	}

	scanners := []VulnerabilityScanner{
		scanner.CmdInjection{
			Payloads: []string{";cat /etc/passwd;", "|cat /etc/passwd|", "`cat /etc/passwd`"},
		},
	}

	found := map[string][]string{}
	for _, scanner := range scanners {
		vuln := scanner.Scan(originalReq, try)
		if len(vuln) > 0 {
			found[scanner.Name()] = vuln
		}
	}

	if len(found) == 0 {
		httpctl.JsonResponse(w, http.StatusOK, map[string]any{
			"result": "no vulnerabilities found",
		})
	} else {
		httpctl.JsonResponse(w, http.StatusOK, map[string]any{
			"result":          "vulnerabilities found",
			"vulnerabilities": found,
		})
	}
}
