package main

import (
	"net/http"
	"strings"
)

func GET_testing_verifications(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filterResponse := strings.TrimSpace(q.Get("response"))
	filterSuccess := strings.TrimSpace(q.Get("success"))

	results := make([]VerificationRecord, 0)

	AppContext.mu.RLock()
	for _, v := range AppContext.Verifications {
		if filterResponse != "" && v.Response != filterResponse {
			continue
		}
		if filterSuccess == "true" && !v.Success {
			continue
		}
		if filterSuccess == "false" && v.Success {
			continue
		}
		results = append(results, v)
	}
	AppContext.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"verifications": results,
	})
}
