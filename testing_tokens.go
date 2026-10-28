package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

func POST_testing_tokens(w http.ResponseWriter, r *http.Request) {
	var req CreateTestingTokenRequest
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to read request body"})
			return
		}
		if len(strings.TrimSpace(string(bodyBytes))) > 0 {
			if err := json.Unmarshal(bodyBytes, &req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
				return
			}
		}
	}

	AppContext.mu.RLock()
	cfg := CloneConfig(AppContext.Config)
	AppContext.mu.RUnlock()

	tokenStr := strings.TrimSpace(req.Token)
	if tokenStr == "" {
		tokenStr = generateTokenString()
	}

	siteKey := strings.TrimSpace(req.SiteKey)
	if siteKey == "" {
		siteKey = "default-sitekey"
	}

	site := findSiteBySiteKey(cfg, siteKey)

	action := strings.TrimSpace(req.Action)
	if action == "" {
		action = "submit"
	}

	score := resolveDefaultScore(cfg, site)
	if req.Score != nil {
		score = *req.Score
	}

	hostname := resolveDefaultHostname(cfg, site)
	if strings.TrimSpace(req.Hostname) != "" {
		hostname = strings.TrimSpace(req.Hostname)
	}

	ttl := cfg.TokenTtlSeconds
	if req.TtlSeconds != nil {
		ttl = *req.TtlSeconds
	}

	expired := false
	if req.Expired != nil {
		expired = *req.Expired
	}

	used := false
	if req.Used != nil {
		used = *req.Used
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(ttl) * time.Second)
	if expired {
		expiresAt = now.Add(-1 * time.Second)
	}

	tokenObj := &CaptchaToken{
		Token:         tokenStr,
		SiteKey:       siteKey,
		Action:        action,
		Score:         score,
		Hostname:      hostname,
		CreatedAt:     now.Format(time.RFC3339),
		ExpiresAt:     expiresAt.Format(time.RFC3339),
		Used:          used,
		Expired:       expired,
		ErrorCodes:    req.ErrorCodes,
		CreatedAtTime: now,
		ExpiresAtTime: expiresAt,
	}

	AppContext.mu.Lock()
	if _, exists := AppContext.Tokens[tokenStr]; !exists {
		AppContext.TokenOrder = append(AppContext.TokenOrder, tokenStr)
	}
	AppContext.Tokens[tokenStr] = tokenObj
	AppContext.mu.Unlock()

	writeJSON(w, http.StatusCreated, tokenObj)
}

func GET_testing_tokens(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filterSiteKey := strings.TrimSpace(q.Get("sitekey"))
	filterUsed := strings.TrimSpace(q.Get("used"))

	now := time.Now().UTC()
	tokens := make([]CaptchaToken, 0)

	AppContext.mu.Lock()
	for _, tokenKey := range AppContext.TokenOrder {
		tok, ok := AppContext.Tokens[tokenKey]
		if !ok {
			continue
		}
		if !tok.Expired && now.After(tok.ExpiresAtTime) {
			tok.Expired = true
		}

		if filterSiteKey != "" && tok.SiteKey != filterSiteKey {
			continue
		}
		if filterUsed == "true" && !tok.Used {
			continue
		}
		if filterUsed == "false" && tok.Used {
			continue
		}

		tokens = append(tokens, *tok)
	}
	AppContext.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tokens": tokens,
	})
}
