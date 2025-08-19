package config

import (
	"testing"
	"time"
)

func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	// Test HTTP config
	if len(config.HTTP.AllowedHosts) == 0 {
		t.Error("expected default allowed hosts to be set")
	}

	if len(config.HTTP.DeniedHosts) == 0 {
		t.Error("expected default denied hosts to be set")
	}

	if len(config.HTTP.AllowedPorts) == 0 {
		t.Error("expected default allowed ports to be set")
	}

	if config.HTTP.MaxRequestSize <= 0 {
		t.Error("expected max request size to be positive")
	}

	if config.HTTP.MaxResponseSize <= 0 {
		t.Error("expected max response size to be positive")
	}

	if config.HTTP.UserAgent == "" {
		t.Error("expected user agent to be set")
	}

	// Test logging config
	if !config.Logging.RedactSecrets {
		t.Error("expected redact secrets to be enabled by default")
	}

	// Test timeout config
	if config.Timeouts.HTTPTimeout <= 0 {
		t.Error("expected HTTP timeout to be positive")
	}

	if config.Timeouts.ActionTimeout <= 0 {
		t.Error("expected action timeout to be positive")
	}

	if config.Timeouts.WorkflowTimeout <= 0 {
		t.Error("expected workflow timeout to be positive")
	}
}

func TestValidateURL(t *testing.T) {
	config := DefaultSecurityConfig()

	tests := []struct {
		name      string
		url       string
		shouldErr bool
	}{
		// Valid URLs
		{"valid HTTPS", "https://api.github.com/repos", false},
		{"valid HTTP", "http://httpbin.org/get", false},
		{"allowed host with port", "https://api.github.com:443/data", false},

		// Invalid schemes
		{"FTP scheme", "ftp://example.com/file", true},
		{"file scheme", "file:///etc/passwd", true},

		// Denied hosts
		{"localhost", "http://localhost:8080/api", true},
		{"127.0.0.1", "http://127.0.0.1/api", true},
		{"internal host", "http://service.internal/api", true},
		{"metadata service", "http://169.254.169.254/metadata", true},

		// Denied ports
		{"SSH port", "http://example.com:22/", true},
		{"MySQL port", "http://example.com:3306/", true},
		{"Redis port", "http://example.com:6379/", true},

		// Invalid URLs
		{"malformed URL", "not-a-url", true},
		{"empty URL", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.HTTP.ValidateURL(tt.url)

			if tt.shouldErr && err == nil {
				t.Errorf("expected error for URL %s, got nil", tt.url)
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error for URL %s, got %v", tt.url, err)
			}
		})
	}
}

func TestValidateURLWithCustomConfig(t *testing.T) {
	// Create custom config that only allows specific hosts
	config := &HTTPSecurityConfig{
		AllowedHosts:   []string{"api.example.com", "*.trusted.com"},
		DeniedHosts:    []string{"blocked.com"},
		AllowedPorts:   []int{80, 443, 8080},
		DeniedPorts:    []int{22, 3306},
		AllowedSchemes: []string{"https"},
		DeniedNetworks: []string{"192.168.0.0/16"},
	}

	tests := []struct {
		name      string
		url       string
		shouldErr bool
	}{
		// Allowed hosts
		{"exact allowed host", "https://api.example.com/data", false},
		{"wildcard allowed host", "https://sub.trusted.com/api", false},

		// Not in allowed list
		{"not allowed host", "https://other.com/api", true},

		// Explicitly denied
		{"denied host", "https://blocked.com/api", true},

		// Scheme restrictions
		{"HTTP not allowed", "http://api.example.com/data", true},

		// Port restrictions
		{"denied port", "https://api.example.com:22/data", true},
		{"allowed port", "https://api.example.com:8080/data", false},

		// Network restrictions
		{"denied network", "https://192.168.1.1/api", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateURL(tt.url)

			if tt.shouldErr && err == nil {
				t.Errorf("expected error for URL %s, got nil", tt.url)
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error for URL %s, got %v", tt.url, err)
			}
		})
	}
}

func TestMatchesPattern(t *testing.T) {
	config := &HTTPSecurityConfig{}

	tests := []struct {
		name     string
		hostname string
		pattern  string
		expected bool
	}{
		{"exact match", "api.github.com", "api.github.com", true},
		{"wildcard subdomain", "api.github.com", "*.github.com", true},
		{"wildcard prefix", "api.github.com", "api.*", true},
		{"no match", "api.github.com", "api.gitlab.com", false},
		{"partial match should fail", "api.github.com", "github.com", false},
		{"multiple wildcards", "api.v1.github.com", "*.*.github.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.matchesPattern(tt.hostname, tt.pattern)
			if result != tt.expected {
				t.Errorf("hostname %s, pattern %s: expected %v, got %v",
					tt.hostname, tt.pattern, tt.expected, result)
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		name      string
		portStr   string
		expected  int
		shouldErr bool
	}{
		{"valid port", "8080", 8080, false},
		{"HTTPS port", "443", 443, false},
		{"HTTP port", "80", 80, false},
		{"high port", "65535", 65535, false},
		{"invalid port - too high", "65536", 0, true},
		{"invalid port - zero", "0", 0, true},
		{"invalid port - negative", "-1", 0, true},
		{"invalid port - text", "abc", 0, true},
		{"empty port", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePort(tt.portStr)

			if tt.shouldErr && err == nil {
				t.Errorf("expected error for port %s, got nil", tt.portStr)
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error for port %s, got %v", tt.portStr, err)
			}

			if !tt.shouldErr && result != tt.expected {
				t.Errorf("port %s: expected %d, got %d", tt.portStr, tt.expected, result)
			}
		})
	}
}

func TestSecurityConfigMerge(t *testing.T) {
	base := DefaultSecurityConfig()

	// Create override config
	override := &SecurityConfig{
		HTTP: HTTPSecurityConfig{
			AllowedHosts:    []string{"custom.com"},
			MaxRequestSize:  1024,
			MaxResponseSize: 2048,
			UserAgent:       "Custom-Agent/1.0",
		},
		Logging: LoggingSecurityConfig{
			RedactSecrets:   false,
			SensitiveKeys:   []string{"custom_key"},
			MaxLogEntrySize: 512,
		},
		Timeouts: TimeoutSecurityConfig{
			HTTPTimeout:     10 * time.Second,
			ActionTimeout:   2 * time.Minute,
			WorkflowTimeout: 10 * time.Minute,
		},
	}

	// Merge configs
	base.Merge(override)

	// Check that values were overridden
	if len(base.HTTP.AllowedHosts) != 1 || base.HTTP.AllowedHosts[0] != "custom.com" {
		t.Error("allowed hosts not merged correctly")
	}

	if base.HTTP.MaxRequestSize != 1024 {
		t.Error("max request size not merged correctly")
	}

	if base.HTTP.MaxResponseSize != 2048 {
		t.Error("max response size not merged correctly")
	}

	if base.HTTP.UserAgent != "Custom-Agent/1.0" {
		t.Error("user agent not merged correctly")
	}

	if base.Logging.RedactSecrets != false {
		t.Error("redact secrets not merged correctly")
	}

	if base.Logging.MaxLogEntrySize != 512 {
		t.Error("max log entry size not merged correctly")
	}

	if base.Timeouts.HTTPTimeout != 10*time.Second {
		t.Error("HTTP timeout not merged correctly")
	}

	if base.Timeouts.ActionTimeout != 2*time.Minute {
		t.Error("action timeout not merged correctly")
	}

	if base.Timeouts.WorkflowTimeout != 10*time.Minute {
		t.Error("workflow timeout not merged correctly")
	}
}

func TestSecurityConfigMergeNil(t *testing.T) {
	config := DefaultSecurityConfig()
	originalHosts := make([]string, len(config.HTTP.AllowedHosts))
	copy(originalHosts, config.HTTP.AllowedHosts)

	// Merge with nil should not change anything
	config.Merge(nil)

	if len(config.HTTP.AllowedHosts) != len(originalHosts) {
		t.Error("config changed after merging with nil")
	}

	for i, host := range originalHosts {
		if config.HTTP.AllowedHosts[i] != host {
			t.Error("config changed after merging with nil")
		}
	}
}
