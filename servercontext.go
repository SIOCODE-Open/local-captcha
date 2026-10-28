package main

import (
	"strings"
	"sync"

	"github.com/google/uuid"
)

type AppServerContext struct {
	mu            sync.RWMutex
	InitialConfig *CaptchaConfig
	Config        *CaptchaConfig
	Tokens        map[string]*CaptchaToken
	TokenOrder    []string
	Verifications []VerificationRecord
}

var AppContext *AppServerContext

func NewAppContext() *AppServerContext {
	return &AppServerContext{
		InitialConfig: CloneConfig(AppConfig),
		Config:        CloneConfig(AppConfig),
		Tokens:        make(map[string]*CaptchaToken),
		TokenOrder:    make([]string, 0),
		Verifications: make([]VerificationRecord, 0),
	}
}

func generateTokenString() string {
	return "tok_" + strings.ReplaceAll(uuid.New().String(), "-", "")
}

// findSiteBySiteKey returns the matching SiteConfig for a sitekey, or nil if not found.
// Must be called with AppContext.mu held (at least RLock) or on a cloned config.
func findSiteBySiteKey(cfg *CaptchaConfig, siteKey string) *SiteConfig {
	if siteKey == "" {
		return nil
	}
	for i := range cfg.Sites {
		if cfg.Sites[i].SiteKey == siteKey {
			return &cfg.Sites[i]
		}
	}
	return nil
}

// findSiteBySecret returns the matching SiteConfig for a secret, or nil if not found.
func findSiteBySecret(cfg *CaptchaConfig, secret string) *SiteConfig {
	if secret == "" {
		return nil
	}
	for i := range cfg.Sites {
		if cfg.Sites[i].Secret != "" && cfg.Sites[i].Secret == secret {
			return &cfg.Sites[i]
		}
	}
	return nil
}

// isSiteKeyAllowed checks if a sitekey is permitted under the current config.
func isSiteKeyAllowed(cfg *CaptchaConfig, siteKey string) bool {
	if siteKey == GoogleTestSiteKey {
		return true
	}
	if cfg.AllowAnySite {
		return true
	}
	return findSiteBySiteKey(cfg, siteKey) != nil
}

// resolveEffectiveMode returns the effective mode for a sitekey ("auto_pass", "challenge", or "always_fail").
func resolveEffectiveMode(cfg *CaptchaConfig, site *SiteConfig) string {
	if len(cfg.ForceErrorCodes) > 0 || cfg.Mode == ModeAlwaysFail {
		return ModeAlwaysFail
	}
	if site != nil && isValidMode(site.Mode) {
		return site.Mode
	}
	if isValidMode(cfg.Mode) {
		return cfg.Mode
	}
	return ModeAutoPass
}

// resolveDefaultScore returns the default score for a site or falls back to global default_score.
func resolveDefaultScore(cfg *CaptchaConfig, site *SiteConfig) float64 {
	if site != nil && site.DefaultScore != nil {
		return *site.DefaultScore
	}
	return cfg.DefaultScore
}

// resolveDefaultHostname returns the hostname for a site or falls back to global hostname.
func resolveDefaultHostname(cfg *CaptchaConfig, site *SiteConfig) string {
	if site != nil && site.Hostname != "" {
		return site.Hostname
	}
	if cfg.Hostname != "" {
		return cfg.Hostname
	}
	return "localhost"
}
