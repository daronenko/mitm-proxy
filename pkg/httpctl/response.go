package httpctl

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Error string `json:"error"`
}

func JsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error": "something went wrong"}`, http.StatusInternalServerError)
		return
	}
}

func ErrorResponse(w http.ResponseWriter, status int, msg string) {
	JsonResponse(w, status, errorBody{msg})
}
