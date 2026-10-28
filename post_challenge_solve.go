package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func parseChallengeSolveRequest(r *http.Request) (ChallengeSolveRequest, error) {
	q := r.URL.Query()
	req := ChallengeSolveRequest{
		SiteKey:         q.Get("sitekey"),
		K:               q.Get("k"),
		Action:          q.Get("action"),
		ChallengeAnswer: q.Get("challenge_answer"),
		Hostname:        q.Get("hostname"),
	}
	if scoreStr := q.Get("score"); scoreStr != "" {
		if s, err := strconv.ParseFloat(scoreStr, 64); err == nil {
			req.Score = &s
		}
	}

	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.Contains(contentType, "application/json") {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return req, err
		}
		if len(strings.TrimSpace(string(bodyBytes))) > 0 {
			var jsonReq ChallengeSolveRequest
			if err := json.Unmarshal(bodyBytes, &jsonReq); err != nil {
				return req, err
			}
			if jsonReq.SiteKey != "" {
				req.SiteKey = jsonReq.SiteKey
			}
			if jsonReq.K != "" {
				req.K = jsonReq.K
			}
			if jsonReq.Action != "" {
				req.Action = jsonReq.Action
			}
			if jsonReq.ChallengeAnswer != "" {
				req.ChallengeAnswer = jsonReq.ChallengeAnswer
			}
			if jsonReq.Score != nil {
				req.Score = jsonReq.Score
			}
			if jsonReq.Hostname != "" {
				req.Hostname = jsonReq.Hostname
			}
		}
	} else {
		_ = r.ParseForm()
		if v := r.FormValue("sitekey"); v != "" {
			req.SiteKey = v
		}
		if v := r.FormValue("k"); v != "" {
			req.K = v
		}
		if v := r.FormValue("action"); v != "" {
			req.Action = v
		}
		if v := r.FormValue("challenge_answer"); v != "" {
			req.ChallengeAnswer = v
		}
		if v := r.FormValue("hostname"); v != "" {
			req.Hostname = v
		}
		if v := r.FormValue("score"); v != "" {
			if s, err := strconv.ParseFloat(v, 64); err == nil {
				req.Score = &s
			}
		}
	}

	return req, nil
}

func POST_challenge_solve(w http.ResponseWriter, r *http.Request) {
	req, err := parseChallengeSolveRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success":     false,
			"error":       "bad-request",
			"error-codes": []string{"bad-request"},
		})
		return
	}

	AppContext.mu.RLock()
	cfg := CloneConfig(AppContext.Config)
	AppContext.mu.RUnlock()

	if cfg.ResponseDelayMs > 0 {
		time.Sleep(time.Duration(cfg.ResponseDelayMs) * time.Millisecond)
	}

	siteKey := strings.TrimSpace(req.SiteKey)
	if siteKey == "" {
		siteKey = strings.TrimSpace(req.K)
	}

	if siteKey == "" {
		if !cfg.AllowAnySite {
			writeJSON(w, http.StatusBadRequest, ChallengeSolveResponse{
				Success:    false,
				Error:      "invalid-sitekey",
				ErrorCodes: []string{"invalid-sitekey"},
			})
			return
		}
		siteKey = "default-sitekey"
	}

	if !isSiteKeyAllowed(cfg, siteKey) {
		writeJSON(w, http.StatusBadRequest, ChallengeSolveResponse{
			Success:    false,
			Error:      "invalid-sitekey",
			ErrorCodes: []string{"invalid-sitekey"},
		})
		return
	}

	site := findSiteBySiteKey(cfg, siteKey)

	// Check failure conditions: force_error_codes, always_fail mode, or challenge_answer == "fail"
	if len(cfg.ForceErrorCodes) > 0 {
		writeJSON(w, http.StatusOK, ChallengeSolveResponse{
			Success:    false,
			ErrorCodes: cfg.ForceErrorCodes,
		})
		return
	}

	effectiveMode := resolveEffectiveMode(cfg, site)
	if effectiveMode == ModeAlwaysFail {
		errCodes := []string{"invalid-input-response"}
		if site != nil && len(site.ErrorCodes) > 0 {
			errCodes = site.ErrorCodes
		}
		writeJSON(w, http.StatusOK, ChallengeSolveResponse{
			Success:    false,
			ErrorCodes: errCodes,
		})
		return
	}

	if strings.EqualFold(strings.TrimSpace(req.ChallengeAnswer), "fail") {
		writeJSON(w, http.StatusOK, ChallengeSolveResponse{
			Success:    false,
			ErrorCodes: []string{"invalid-input-response"},
		})
		return
	}

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

	now := time.Now().UTC()
	ttl := cfg.TokenTtlSeconds
	if ttl <= 0 {
		ttl = 300
	}
	expiresAt := now.Add(time.Duration(ttl) * time.Second)
	tokenStr := generateTokenString()

	tokenObj := &CaptchaToken{
		Token:         tokenStr,
		SiteKey:       siteKey,
		Action:        action,
		Score:         score,
		Hostname:      hostname,
		CreatedAt:     now.Format(time.RFC3339),
		ExpiresAt:     expiresAt.Format(time.RFC3339),
		Used:          false,
		Expired:       false,
		CreatedAtTime: now,
		ExpiresAtTime: expiresAt,
	}

	AppContext.mu.Lock()
	if _, exists := AppContext.Tokens[tokenStr]; !exists {
		AppContext.TokenOrder = append(AppContext.TokenOrder, tokenStr)
	}
	AppContext.Tokens[tokenStr] = tokenObj
	AppContext.mu.Unlock()

	scoreVal := score
	ttlVal := ttl
	writeJSON(w, http.StatusOK, ChallengeSolveResponse{
		Success:     true,
		Token:       tokenStr,
		SiteKey:     siteKey,
		Action:      action,
		Score:       &scoreVal,
		Hostname:    hostname,
		ChallengeTs: tokenObj.CreatedAt,
		ExpiresIn:   &ttlVal,
	})
}
