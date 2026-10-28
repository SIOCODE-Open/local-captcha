package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func GET_testing_config(w http.ResponseWriter, r *http.Request) {
	AppContext.mu.RLock()
	cfg := CloneConfig(AppContext.Config)
	AppContext.mu.RUnlock()

	writeJSON(w, http.StatusOK, cfg)
}

func PUT_testing_config(w http.ResponseWriter, r *http.Request) {
	var req UpdateConfigRequest
	if r.Body == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Request body is required"})
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to read request body"})
		return
	}

	if len(strings.TrimSpace(string(bodyBytes))) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Empty JSON body"})
		return
	}

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	if req.Mode != nil && !isValidMode(*req.Mode) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid mode: must be auto_pass, challenge, or always_fail"})
		return
	}

	if req.Sites != nil {
		for _, s := range *req.Sites {
			if s.Mode != "" && !isValidMode(s.Mode) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid site mode: must be auto_pass, challenge, or always_fail"})
				return
			}
		}
	}

	AppContext.mu.Lock()
	cfg := AppContext.Config

	if req.Mode != nil {
		cfg.Mode = *req.Mode
	}
	if req.DefaultScore != nil {
		cfg.DefaultScore = *req.DefaultScore
	}
	if req.TokenTtlSeconds != nil && *req.TokenTtlSeconds > 0 {
		cfg.TokenTtlSeconds = *req.TokenTtlSeconds
	}
	if req.VerifyOnce != nil {
		cfg.VerifyOnce = *req.VerifyOnce
	}
	if req.AllowAnySite != nil {
		cfg.AllowAnySite = *req.AllowAnySite
	}
	if req.AllowAnySecret != nil {
		cfg.AllowAnySecret = *req.AllowAnySecret
	}
	if req.ForceErrorCodes != nil {
		if *req.ForceErrorCodes == nil {
			cfg.ForceErrorCodes = []string{}
		} else {
			cfg.ForceErrorCodes = make([]string, len(*req.ForceErrorCodes))
			copy(cfg.ForceErrorCodes, *req.ForceErrorCodes)
		}
	}
	if req.ResponseDelayMs != nil {
		if *req.ResponseDelayMs >= 0 {
			cfg.ResponseDelayMs = *req.ResponseDelayMs
		} else {
			cfg.ResponseDelayMs = 0
		}
	}
	if req.Hostname != nil && strings.TrimSpace(*req.Hostname) != "" {
		cfg.Hostname = strings.TrimSpace(*req.Hostname)
	}
	if req.AllowedOrigins != nil && strings.TrimSpace(*req.AllowedOrigins) != "" {
		cfg.AllowedOrigins = strings.TrimSpace(*req.AllowedOrigins)
	}
	if req.BaseUrl != nil && strings.TrimSpace(*req.BaseUrl) != "" {
		cfg.BaseUrl = strings.TrimSpace(*req.BaseUrl)
	}
	if req.Sites != nil {
		if *req.Sites == nil {
			cfg.Sites = []SiteConfig{}
		} else {
			cfg.Sites = make([]SiteConfig, len(*req.Sites))
			copy(cfg.Sites, *req.Sites)
		}
	}
	if req.MagicTokens != nil {
		if *req.MagicTokens == nil {
			cfg.MagicTokens = nil
		} else {
			cfg.MagicTokens = make([]MagicTokenConfig, len(*req.MagicTokens))
			copy(cfg.MagicTokens, *req.MagicTokens)
		}
	}

	updated := CloneConfig(cfg)
	AppContext.mu.Unlock()

	writeJSON(w, http.StatusOK, updated)
}
