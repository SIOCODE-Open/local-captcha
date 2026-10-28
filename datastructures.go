package main

import "time"

const (
	GoogleTestSiteKey = "6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI"
	GoogleTestSecret  = "6LeIxAcTAAAAAGG-vFI1TnRWxMZNFuojJ4WifJWe"

	ModeAutoPass   = "auto_pass"
	ModeChallenge  = "challenge"
	ModeAlwaysFail = "always_fail"
)

type SiteConfig struct {
	SiteKey      string   `json:"sitekey" yaml:"sitekey"`
	Secret       string   `json:"secret" yaml:"secret"`
	Hostname     string   `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Mode         string   `json:"mode,omitempty" yaml:"mode,omitempty"`
	DefaultScore *float64 `json:"default_score,omitempty" yaml:"default_score,omitempty"`
	ErrorCodes   []string `json:"error_codes,omitempty" yaml:"error_codes,omitempty"`
}

type MagicTokenConfig struct {
	Token      string   `json:"token" yaml:"token"`
	Success    bool     `json:"success" yaml:"success"`
	Score      *float64 `json:"score,omitempty" yaml:"score,omitempty"`
	Action     string   `json:"action,omitempty" yaml:"action,omitempty"`
	Hostname   string   `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	ErrorCodes []string `json:"error_codes,omitempty" yaml:"error_codes,omitempty"`
}

type CaptchaConfig struct {
	Port            int                `json:"port" yaml:"port"`
	BaseUrl         string             `json:"base_url" yaml:"base_url"`
	Hostname        string             `json:"hostname" yaml:"hostname"`
	AllowedOrigins  string             `json:"allowed_origins" yaml:"allowed_origins"`
	Mode            string             `json:"mode" yaml:"mode"`
	DefaultScore    float64            `json:"default_score" yaml:"default_score"`
	TokenTtlSeconds int                `json:"token_ttl_seconds" yaml:"token_ttl_seconds"`
	VerifyOnce      bool               `json:"verify_once" yaml:"verify_once"`
	AllowAnySite    bool               `json:"allow_any_site" yaml:"allow_any_site"`
	AllowAnySecret  bool               `json:"allow_any_secret" yaml:"allow_any_secret"`
	ForceErrorCodes []string           `json:"force_error_codes" yaml:"force_error_codes"`
	ResponseDelayMs int                `json:"response_delay_ms" yaml:"response_delay_ms"`
	Sites           []SiteConfig       `json:"sites" yaml:"sites"`
	MagicTokens     []MagicTokenConfig `json:"magic_tokens,omitempty" yaml:"magic_tokens,omitempty"`
}

type CaptchaToken struct {
	Token         string    `json:"token"`
	SiteKey       string    `json:"sitekey"`
	Action        string    `json:"action"`
	Score         float64   `json:"score"`
	Hostname      string    `json:"hostname"`
	CreatedAt     string    `json:"created_at"`
	ExpiresAt     string    `json:"expires_at"`
	Used          bool      `json:"used"`
	Expired       bool      `json:"expired"`
	ErrorCodes    []string  `json:"error_codes,omitempty"`
	CreatedAtTime time.Time `json:"-"`
	ExpiresAtTime time.Time `json:"-"`
}

type VerificationRecord struct {
	Secret     string   `json:"secret"`
	Response   string   `json:"response"`
	RemoteIp   string   `json:"remoteip"`
	Success    bool     `json:"success"`
	Score      *float64 `json:"score,omitempty"`
	Action     string   `json:"action,omitempty"`
	Hostname   string   `json:"hostname,omitempty"`
	ErrorCodes []string `json:"error_codes"`
	Timestamp  string   `json:"timestamp"`
}

type ChallengeSolveRequest struct {
	SiteKey         string   `json:"sitekey"`
	K               string   `json:"k"`
	Action          string   `json:"action"`
	ChallengeAnswer string   `json:"challenge_answer"`
	Score           *float64 `json:"score"`
	Hostname        string   `json:"hostname"`
}

type ChallengeSolveResponse struct {
	Success     bool     `json:"success"`
	Token       string   `json:"token,omitempty"`
	SiteKey     string   `json:"sitekey,omitempty"`
	Action      string   `json:"action,omitempty"`
	Score       *float64 `json:"score,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	ChallengeTs string   `json:"challenge_ts,omitempty"`
	ExpiresIn   *int     `json:"expires_in,omitempty"`
	Error       string   `json:"error,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

type SiteVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIp string `json:"remoteip"`
}

type SiteVerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTs string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	Score       *float64 `json:"score,omitempty"`
	Action      string   `json:"action,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

type CreateTestingTokenRequest struct {
	Token      string   `json:"token"`
	SiteKey    string   `json:"sitekey"`
	Action     string   `json:"action"`
	Score      *float64 `json:"score"`
	Hostname   string   `json:"hostname"`
	TtlSeconds *int     `json:"ttl_seconds"`
	Expired    *bool    `json:"expired"`
	Used       *bool    `json:"used"`
	ErrorCodes []string `json:"error_codes"`
}

type UpdateConfigRequest struct {
	Mode            *string             `json:"mode"`
	DefaultScore    *float64            `json:"default_score"`
	TokenTtlSeconds *int                `json:"token_ttl_seconds"`
	VerifyOnce      *bool               `json:"verify_once"`
	AllowAnySite    *bool               `json:"allow_any_site"`
	AllowAnySecret  *bool               `json:"allow_any_secret"`
	ForceErrorCodes *[]string           `json:"force_error_codes"`
	ResponseDelayMs *int                `json:"response_delay_ms"`
	Hostname        *string             `json:"hostname"`
	AllowedOrigins  *string             `json:"allowed_origins"`
	BaseUrl         *string             `json:"base_url"`
	Sites           *[]SiteConfig       `json:"sites"`
	MagicTokens     *[]MagicTokenConfig `json:"magic_tokens"`
}
