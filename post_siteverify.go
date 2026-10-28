package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

func parseSiteVerifyRequest(r *http.Request) SiteVerifyRequest {
	q := r.URL.Query()
	req := SiteVerifyRequest{
		Secret:   q.Get("secret"),
		Response: q.Get("response"),
		RemoteIp: q.Get("remoteip"),
	}

	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.Contains(contentType, "application/json") {
		if bodyBytes, err := io.ReadAll(r.Body); err == nil && len(strings.TrimSpace(string(bodyBytes))) > 0 {
			var jsonReq SiteVerifyRequest
			if err := json.Unmarshal(bodyBytes, &jsonReq); err == nil {
				if jsonReq.Secret != "" {
					req.Secret = jsonReq.Secret
				}
				if jsonReq.Response != "" {
					req.Response = jsonReq.Response
				}
				if jsonReq.RemoteIp != "" {
					req.RemoteIp = jsonReq.RemoteIp
				}
			}
		}
	} else {
		_ = r.ParseForm()
		if v := r.FormValue("secret"); v != "" {
			req.Secret = v
		}
		if v := r.FormValue("response"); v != "" {
			req.Response = v
		}
		if v := r.FormValue("remoteip"); v != "" {
			req.RemoteIp = v
		}
	}

	return req
}

func recordAndRespondSiteVerify(w http.ResponseWriter, req SiteVerifyRequest, resp SiteVerifyResponse) {
	errCodes := make([]string, 0)
	if len(resp.ErrorCodes) > 0 {
		errCodes = append(errCodes, resp.ErrorCodes...)
	}

	var scorePtr *float64
	if resp.Score != nil {
		v := *resp.Score
		scorePtr = &v
	}

	rec := VerificationRecord{
		Secret:     req.Secret,
		Response:   req.Response,
		RemoteIp:   req.RemoteIp,
		Success:    resp.Success,
		Score:      scorePtr,
		Action:     resp.Action,
		Hostname:   resp.Hostname,
		ErrorCodes: errCodes,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	AppContext.mu.Lock()
	AppContext.Verifications = append(AppContext.Verifications, rec)
	AppContext.mu.Unlock()

	writeJSON(w, http.StatusOK, resp)
}

func evaluateMagicToken(cfg *CaptchaConfig, matchedSite *SiteConfig, token string, now time.Time) (*SiteVerifyResponse, bool) {
	defaultScore := resolveDefaultScore(cfg, matchedSite)
	defaultHostname := resolveDefaultHostname(cfg, matchedSite)
	ts := now.Format(time.RFC3339)

	// Check custom configured magic tokens first
	for _, mt := range cfg.MagicTokens {
		if mt.Token == token {
			if mt.Success {
				s := defaultScore
				if mt.Score != nil {
					s = *mt.Score
				}
				act := mt.Action
				if act == "" {
					act = "submit"
				}
				host := mt.Hostname
				if host == "" {
					host = defaultHostname
				}
				return &SiteVerifyResponse{
					Success:     true,
					ChallengeTs: ts,
					Hostname:    host,
					Score:       &s,
					Action:      act,
				}, true
			}
			errCodes := mt.ErrorCodes
			if len(errCodes) == 0 {
				errCodes = []string{"invalid-input-response"}
			}
			return &SiteVerifyResponse{
				Success:    false,
				ErrorCodes: errCodes,
			}, true
		}
	}

	// Built-in magic tokens
	switch token {
	case "always-pass-token", "valid-token":
		s := defaultScore
		return &SiteVerifyResponse{
			Success:     true,
			ChallengeTs: ts,
			Hostname:    defaultHostname,
			Score:       &s,
			Action:      "submit",
		}, true
	case "low-score-token":
		s := 0.1
		return &SiteVerifyResponse{
			Success:     true,
			ChallengeTs: ts,
			Hostname:    defaultHostname,
			Score:       &s,
			Action:      "submit",
		}, true
	case "high-score-token":
		s := 1.0
		return &SiteVerifyResponse{
			Success:     true,
			ChallengeTs: ts,
			Hostname:    defaultHostname,
			Score:       &s,
			Action:      "submit",
		}, true
	case "always-fail-token", "invalid-token":
		return &SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"invalid-input-response"},
		}, true
	case "expired-token", "duplicate-token":
		return &SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"timeout-or-duplicate"},
		}, true
	case "bad-request-token":
		return &SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"bad-request"},
		}, true
	}

	return nil, false
}

func HANDLE_siteverify(w http.ResponseWriter, r *http.Request) {
	req := parseSiteVerifyRequest(r)

	AppContext.mu.RLock()
	cfg := CloneConfig(AppContext.Config)
	AppContext.mu.RUnlock()

	if cfg.ResponseDelayMs > 0 {
		time.Sleep(time.Duration(cfg.ResponseDelayMs) * time.Millisecond)
	}

	secret := strings.TrimSpace(req.Secret)
	responseToken := strings.TrimSpace(req.Response)

	// 1. Missing Parameters
	if secret == "" && responseToken == "" {
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"missing-input-secret", "missing-input-response"},
		})
		return
	}
	if secret == "" {
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"missing-input-secret"},
		})
		return
	}
	if responseToken == "" {
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"missing-input-response"},
		})
		return
	}

	// 2. Global Forced Errors / Always-Fail Mode
	if len(cfg.ForceErrorCodes) > 0 {
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: cfg.ForceErrorCodes,
		})
		return
	}
	if cfg.Mode == ModeAlwaysFail {
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"invalid-input-response"},
		})
		return
	}

	// 3. Secret Validation
	matchedSiteBySecret := findSiteBySecret(cfg, secret)
	isGoogleTestSecret := secret == GoogleTestSecret

	if !cfg.AllowAnySecret && matchedSiteBySecret == nil && !isGoogleTestSecret {
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"invalid-input-secret"},
		})
		return
	}

	if matchedSiteBySecret != nil && matchedSiteBySecret.Mode == ModeAlwaysFail {
		errCodes := []string{"invalid-input-response"}
		if len(matchedSiteBySecret.ErrorCodes) > 0 {
			errCodes = matchedSiteBySecret.ErrorCodes
		}
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: errCodes,
		})
		return
	}

	now := time.Now().UTC()

	// 4. Built-in & Configured Magic Tokens
	if magicResp, ok := evaluateMagicToken(cfg, matchedSiteBySecret, responseToken, now); ok {
		recordAndRespondSiteVerify(w, req, *magicResp)
		return
	}

	// 5. Issued / Pre-seeded Token Validation
	AppContext.mu.Lock()
	storedToken, exists := AppContext.Tokens[responseToken]
	if !exists {
		AppContext.mu.Unlock()
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"invalid-input-response"},
		})
		return
	}

	tokenSite := findSiteBySiteKey(cfg, storedToken.SiteKey)
	if tokenSite != nil && tokenSite.Secret != "" && secret != tokenSite.Secret && !isGoogleTestSecret {
		AppContext.mu.Unlock()
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"invalid-input-secret"},
		})
		return
	}

	if tokenSite != nil && tokenSite.Mode == ModeAlwaysFail {
		errCodes := []string{"invalid-input-response"}
		if len(tokenSite.ErrorCodes) > 0 {
			errCodes = tokenSite.ErrorCodes
		}
		AppContext.mu.Unlock()
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: errCodes,
		})
		return
	}

	if len(storedToken.ErrorCodes) > 0 {
		errCodes := make([]string, len(storedToken.ErrorCodes))
		copy(errCodes, storedToken.ErrorCodes)
		AppContext.mu.Unlock()
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: errCodes,
		})
		return
	}

	if storedToken.Expired || now.After(storedToken.ExpiresAtTime) {
		storedToken.Expired = true
		AppContext.mu.Unlock()
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"timeout-or-duplicate"},
		})
		return
	}

	if cfg.VerifyOnce && storedToken.Used {
		AppContext.mu.Unlock()
		recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
			Success:    false,
			ErrorCodes: []string{"timeout-or-duplicate"},
		})
		return
	}

	storedToken.Used = true
	scoreVal := storedToken.Score
	challengeTs := storedToken.CreatedAt
	hostname := storedToken.Hostname
	action := storedToken.Action
	AppContext.mu.Unlock()

	recordAndRespondSiteVerify(w, req, SiteVerifyResponse{
		Success:     true,
		ChallengeTs: challengeTs,
		Hostname:    hostname,
		Score:       &scoreVal,
		Action:      action,
	})
}
