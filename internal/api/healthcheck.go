package api

import (
	"net/http"
)

func (a *API) healthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"status": "ok"}
		a.successResponse(w, resp, http.StatusOK)
	}
}
