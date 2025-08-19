package config

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// SecurityConfig contains security settings for MOVA Engine
type SecurityConfig struct {
	HTTP     HTTPSecurityConfig    `json:"http" yaml:"http"`
	Logging  LoggingSecurityConfig `json:"logging" yaml:"logging"`
	Timeouts TimeoutSecurityConfig `json:"timeouts" yaml:"timeouts"`
}

// HTTPSecurityConfig contains HTTP-specific security settings
type HTTPSecurityConfig struct {
	// AllowedHosts contains patterns of allowed hostnames/IPs
	AllowedHosts []string `json:"allowed_hosts" yaml:"allowed_hosts"`

	// DeniedHosts contains patterns of explicitly denied hostnames/IPs
	DeniedHosts []string `json:"denied_hosts" yaml:"denied_hosts"`

	// AllowedPorts contains allowed port numbers (empty means all allowed)
	AllowedPorts []int `json:"allowed_ports" yaml:"allowed_ports"`

	// DeniedPorts contains explicitly denied port numbers
	DeniedPorts []int `json:"denied_ports" yaml:"denied_ports"`

	// AllowedSchemes contains allowed URL schemes (http, https)
	AllowedSchemes []string `json:"allowed_schemes" yaml:"allowed_schemes"`

	// DeniedNetworks contains CIDR blocks that are denied (e.g., private networks)
	DeniedNetworks []string `json:"denied_networks" yaml:"denied_networks"`

	// MaxRequestSize limits the size of HTTP request body in bytes
	MaxRequestSize int64 `json:"max_request_size" yaml:"max_request_size"`

	// MaxResponseSize limits the size of HTTP response body in bytes
	MaxResponseSize int64 `json:"max_response_size" yaml:"max_response_size"`

	// UserAgent to use for HTTP requests
	UserAgent string `json:"user_agent" yaml:"user_agent"`

	// FollowRedirects controls whether to follow HTTP redirects
	FollowRedirects bool `json:"follow_redirects" yaml:"follow_redirects"`

	// MaxRedirects limits the number of redirects to follow
	MaxRedirects int `json:"max_redirects" yaml:"max_redirects"`
}

// LoggingSecurityConfig contains logging security settings
type LoggingSecurityConfig struct {
	// RedactSecrets controls whether to redact sensitive information in logs
	RedactSecrets bool `json:"redact_secrets" yaml:"redact_secrets"`

	// SensitiveKeys contains additional keys to redact beyond defaults
	SensitiveKeys []string `json:"sensitive_keys" yaml:"sensitive_keys"`

	// MaxLogEntrySize limits the size of individual log entries
	MaxLogEntrySize int `json:"max_log_entry_size" yaml:"max_log_entry_size"`
}

// TimeoutSecurityConfig contains timeout settings
type TimeoutSecurityConfig struct {
	// HTTPTimeout is the timeout for HTTP requests
	HTTPTimeout time.Duration `json:"http_timeout" yaml:"http_timeout"`

	// ActionTimeout is the maximum time an action can run
	ActionTimeout time.Duration `json:"action_timeout" yaml:"action_timeout"`

	// WorkflowTimeout is the maximum time a workflow can run
	WorkflowTimeout time.Duration `json:"workflow_timeout" yaml:"workflow_timeout"`
}

// DefaultSecurityConfig returns a secure default configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		HTTP: HTTPSecurityConfig{
			AllowedHosts: []string{
				"api.github.com",
				"httpbin.org",
				"jsonplaceholder.typicode.com",
				"*.googleapis.com",
				"*.amazonaws.com",
			},
			DeniedHosts: []string{
				"localhost",
				"127.0.0.1",
				"0.0.0.0",
				"*.internal",
				"*.local",
				"metadata.google.internal",
				"169.254.169.254", // AWS metadata service
			},
			AllowedPorts:   []int{80, 443, 8080, 8443},
			DeniedPorts:    []int{22, 23, 25, 53, 135, 139, 445, 1433, 1521, 3306, 3389, 5432, 6379},
			AllowedSchemes: []string{"http", "https"},
			DeniedNetworks: []string{
				"10.0.0.0/8", // Private networks
				"172.16.0.0/12",
				"192.168.0.0/16",
				"127.0.0.0/8",    // Loopback
				"169.254.0.0/16", // Link-local
				"::1/128",        // IPv6 loopback
				"fc00::/7",       // IPv6 unique local
				"fe80::/10",      // IPv6 link-local
			},
			MaxRequestSize:  10 * 1024 * 1024, // 10MB
			MaxResponseSize: 50 * 1024 * 1024, // 50MB
			UserAgent:       "MOVA-Engine/1.0",
			FollowRedirects: false,
			MaxRedirects:    0,
		},
		Logging: LoggingSecurityConfig{
			RedactSecrets:   true,
			SensitiveKeys:   []string{},
			MaxLogEntrySize: 1024 * 1024, // 1MB
		},
		Timeouts: TimeoutSecurityConfig{
			HTTPTimeout:     30 * time.Second,
			ActionTimeout:   5 * time.Minute,
			WorkflowTimeout: 30 * time.Minute,
		},
	}
}

// ValidateURL checks if a URL is allowed by the security configuration
func (c *HTTPSecurityConfig) ValidateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme
	if !c.isSchemeAllowed(parsedURL.Scheme) {
		return fmt.Errorf("scheme %s is not allowed", parsedURL.Scheme)
	}

	// Check host
	if err := c.validateHost(parsedURL.Hostname()); err != nil {
		return err
	}

	// Check port
	if err := c.validatePort(parsedURL.Port(), parsedURL.Scheme); err != nil {
		return err
	}

	return nil
}

// isSchemeAllowed checks if the URL scheme is allowed
func (c *HTTPSecurityConfig) isSchemeAllowed(scheme string) bool {
	if len(c.AllowedSchemes) == 0 {
		return true // If no schemes specified, allow all
	}

	for _, allowed := range c.AllowedSchemes {
		if strings.EqualFold(scheme, allowed) {
			return true
		}
	}
	return false
}

// validateHost checks if the hostname is allowed
func (c *HTTPSecurityConfig) validateHost(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("empty hostname")
	}

	// Check denied hosts first
	for _, denied := range c.DeniedHosts {
		if c.matchesPattern(hostname, denied) {
			return fmt.Errorf("host %s is explicitly denied", hostname)
		}
	}

	// Check denied networks (IP addresses)
	if ip := net.ParseIP(hostname); ip != nil {
		for _, network := range c.DeniedNetworks {
			if c.isIPInNetwork(ip, network) {
				return fmt.Errorf("IP %s is in denied network %s", hostname, network)
			}
		}
	}

	// If allowed hosts are specified, check if host is in the list
	if len(c.AllowedHosts) > 0 {
		for _, allowed := range c.AllowedHosts {
			if c.matchesPattern(hostname, allowed) {
				return nil // Host is allowed
			}
		}
		return fmt.Errorf("host %s is not in allowed hosts list", hostname)
	}

	return nil // No restrictions, allow
}

// validatePort checks if the port is allowed
func (c *HTTPSecurityConfig) validatePort(portStr, scheme string) error {
	var port int
	var err error

	if portStr == "" {
		// Use default ports
		switch strings.ToLower(scheme) {
		case "http":
			port = 80
		case "https":
			port = 443
		default:
			return fmt.Errorf("cannot determine default port for scheme %s", scheme)
		}
	} else {
		port, err = parsePort(portStr)
		if err != nil {
			return fmt.Errorf("invalid port %s: %w", portStr, err)
		}
	}

	// Check denied ports first
	for _, denied := range c.DeniedPorts {
		if port == denied {
			return fmt.Errorf("port %d is explicitly denied", port)
		}
	}

	// If allowed ports are specified, check if port is in the list
	if len(c.AllowedPorts) > 0 {
		for _, allowed := range c.AllowedPorts {
			if port == allowed {
				return nil // Port is allowed
			}
		}
		return fmt.Errorf("port %d is not in allowed ports list", port)
	}

	return nil // No restrictions, allow
}

// matchesPattern checks if a hostname matches a pattern (supports wildcards)
func (c *HTTPSecurityConfig) matchesPattern(hostname, pattern string) bool {
	// Convert wildcard pattern to regex
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = "^" + pattern + "$"

	matched, err := regexp.MatchString(pattern, hostname)
	return err == nil && matched
}

// isIPInNetwork checks if an IP address is in a CIDR network
func (c *HTTPSecurityConfig) isIPInNetwork(ip net.IP, network string) bool {
	_, ipNet, err := net.ParseCIDR(network)
	if err != nil {
		return false
	}
	return ipNet.Contains(ip)
}

// parsePort parses a port string to integer
func parsePort(portStr string) (int, error) {
	if portStr == "" {
		return 0, fmt.Errorf("empty port")
	}

	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	if err != nil {
		return 0, err
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port %d out of valid range (1-65535)", port)
	}

	return port, nil
}

// Merge merges another security config into this one, with the other config taking precedence
func (c *SecurityConfig) Merge(other *SecurityConfig) {
	if other == nil {
		return
	}

	// Merge HTTP config
	if len(other.HTTP.AllowedHosts) > 0 {
		c.HTTP.AllowedHosts = other.HTTP.AllowedHosts
	}
	if len(other.HTTP.DeniedHosts) > 0 {
		c.HTTP.DeniedHosts = other.HTTP.DeniedHosts
	}
	if len(other.HTTP.AllowedPorts) > 0 {
		c.HTTP.AllowedPorts = other.HTTP.AllowedPorts
	}
	if len(other.HTTP.DeniedPorts) > 0 {
		c.HTTP.DeniedPorts = other.HTTP.DeniedPorts
	}
	if len(other.HTTP.AllowedSchemes) > 0 {
		c.HTTP.AllowedSchemes = other.HTTP.AllowedSchemes
	}
	if len(other.HTTP.DeniedNetworks) > 0 {
		c.HTTP.DeniedNetworks = other.HTTP.DeniedNetworks
	}
	if other.HTTP.MaxRequestSize > 0 {
		c.HTTP.MaxRequestSize = other.HTTP.MaxRequestSize
	}
	if other.HTTP.MaxResponseSize > 0 {
		c.HTTP.MaxResponseSize = other.HTTP.MaxResponseSize
	}
	if other.HTTP.UserAgent != "" {
		c.HTTP.UserAgent = other.HTTP.UserAgent
	}
	if other.HTTP.MaxRedirects >= 0 {
		c.HTTP.MaxRedirects = other.HTTP.MaxRedirects
		c.HTTP.FollowRedirects = other.HTTP.FollowRedirects
	}

	// Merge logging config
	c.Logging.RedactSecrets = other.Logging.RedactSecrets
	if len(other.Logging.SensitiveKeys) > 0 {
		c.Logging.SensitiveKeys = other.Logging.SensitiveKeys
	}
	if other.Logging.MaxLogEntrySize > 0 {
		c.Logging.MaxLogEntrySize = other.Logging.MaxLogEntrySize
	}

	// Merge timeout config
	if other.Timeouts.HTTPTimeout > 0 {
		c.Timeouts.HTTPTimeout = other.Timeouts.HTTPTimeout
	}
	if other.Timeouts.ActionTimeout > 0 {
		c.Timeouts.ActionTimeout = other.Timeouts.ActionTimeout
	}
	if other.Timeouts.WorkflowTimeout > 0 {
		c.Timeouts.WorkflowTimeout = other.Timeouts.WorkflowTimeout
	}
}
