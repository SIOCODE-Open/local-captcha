package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/goccy/go-yaml"
)

var AppConfig *CaptchaConfig
var ConfigPathFlag string

func init() {
	flag.StringVar(&ConfigPathFlag, "config-path", "", "Path to configuration file")
	flag.StringVar(&ConfigPathFlag, "c", "", "Path to configuration file (shorthand)")
}

func isValidMode(mode string) bool {
	return mode == ModeAutoPass || mode == ModeChallenge || mode == ModeAlwaysFail
}

func LoadConfig() *CaptchaConfig {
	config := &CaptchaConfig{
		Port:            8080,
		Hostname:        "localhost",
		AllowedOrigins:  "*",
		Mode:            ModeAutoPass,
		DefaultScore:    0.9,
		TokenTtlSeconds: 300,
		VerifyOnce:      true,
		AllowAnySite:    true,
		AllowAnySecret:  true,
		ForceErrorCodes: []string{},
		ResponseDelayMs: 0,
		Sites:           []SiteConfig{},
	}

	configPath := ConfigPathFlag
	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	if configPath == "" {
		configPath = "/config.yaml"
	}

	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			log.Printf("Error parsing config file: %v", err)
		}
	} else {
		log.Printf("Error reading config file: %v", err)
	}

	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			config.Port = p
		}
	}

	if config.Port == 0 {
		config.Port = 8080
	}

	if config.BaseUrl == "" {
		config.BaseUrl = fmt.Sprintf("http://localhost:%d", config.Port)
	}

	if config.Hostname == "" {
		config.Hostname = "localhost"
	}

	if config.AllowedOrigins == "" {
		config.AllowedOrigins = "*"
	}

	if !isValidMode(config.Mode) {
		config.Mode = ModeAutoPass
	}

	if config.TokenTtlSeconds <= 0 {
		config.TokenTtlSeconds = 300
	}

	if config.ForceErrorCodes == nil {
		config.ForceErrorCodes = []string{}
	}

	if config.Sites == nil {
		config.Sites = []SiteConfig{}
	}

	return config
}

func CloneConfig(src *CaptchaConfig) *CaptchaConfig {
	if src == nil {
		return nil
	}
	dst := *src

	dst.ForceErrorCodes = make([]string, len(src.ForceErrorCodes))
	copy(dst.ForceErrorCodes, src.ForceErrorCodes)

	dst.Sites = make([]SiteConfig, len(src.Sites))
	for i, s := range src.Sites {
		siteCopy := s
		if s.DefaultScore != nil {
			val := *s.DefaultScore
			siteCopy.DefaultScore = &val
		}
		if s.ErrorCodes != nil {
			siteCopy.ErrorCodes = make([]string, len(s.ErrorCodes))
			copy(siteCopy.ErrorCodes, s.ErrorCodes)
		}
		dst.Sites[i] = siteCopy
	}

	if src.MagicTokens != nil {
		dst.MagicTokens = make([]MagicTokenConfig, len(src.MagicTokens))
		for i, m := range src.MagicTokens {
			mCopy := m
			if m.Score != nil {
				val := *m.Score
				mCopy.Score = &val
			}
			if m.ErrorCodes != nil {
				mCopy.ErrorCodes = make([]string, len(m.ErrorCodes))
				copy(mCopy.ErrorCodes, m.ErrorCodes)
			}
			dst.MagicTokens[i] = mCopy
		}
	}

	return &dst
}
