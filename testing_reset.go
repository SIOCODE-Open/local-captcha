package main

import (
	"net/http"
)

func POST_testing_reset(w http.ResponseWriter, r *http.Request) {
	AppContext.mu.Lock()
	AppContext.Config = CloneConfig(AppContext.InitialConfig)
	AppContext.Tokens = make(map[string]*CaptchaToken)
	AppContext.TokenOrder = make([]string, 0)
	AppContext.Verifications = make([]VerificationRecord, 0)
	AppContext.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "OK",
		"message": "State reset",
	})
}
