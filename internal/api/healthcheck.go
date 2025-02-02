package api

import (
	"encoding/json"
	"net/http"
)

func (a *API) healthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"status": "ok"}
		_ = json.NewEncoder(w).Encode(resp)
	}
}
