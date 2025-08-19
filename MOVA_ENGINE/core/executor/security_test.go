package executor

import (
	"testing"
)

func TestRedactSecrets(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected map[string]string
	}{
		{
			name: "redact authorization header",
			input: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
				"Content-Type":  "application/json",
			},
			expected: map[string]string{
				"Authorization": "Bearer *****",
				"Content-Type":  "application/json",
			},
		},
		{
			name: "redact password field",
			input: map[string]string{
				"username": "admin",
				"password": "secret123",
			},
			expected: map[string]string{
				"username": "admin",
				"password": "se*****23",
			},
		},
		{
			name: "redact API key",
			input: map[string]string{
				"X-API-KEY":  "abc123def456",
				"User-Agent": "MOVA/1.0",
			},
			expected: map[string]string{
				"X-API-KEY":  "ab*****56",
				"User-Agent": "MOVA/1.0",
			},
		},
		{
			name:     "empty input",
			input:    nil,
			expected: nil,
		},
		{
			name: "no sensitive keys",
			input: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "MOVA/1.0",
			},
			expected: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "MOVA/1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactSecrets(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d keys, got %d", len(tt.expected), len(result))
				return
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("key %s missing from result", key)
				} else if actualValue != expectedValue {
					t.Errorf("key %s: expected %s, got %s", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestRedactSecretsInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "redact nested secrets",
			input: map[string]interface{}{
				"config": map[string]interface{}{
					"password": "secret123",
					"timeout":  30,
				},
				"headers": map[string]string{
					"Authorization": "Bearer token123",
					"Content-Type":  "application/json",
				},
			},
			expected: map[string]interface{}{
				"config": map[string]interface{}{
					"password": "se*****23",
					"timeout":  30,
				},
				"headers": map[string]string{
					"Authorization": "Bearer *****",
					"Content-Type":  "application/json",
				},
			},
		},
		{
			name: "redact string secrets",
			input: map[string]interface{}{
				"api_key": "abc123",
				"count":   42,
			},
			expected: map[string]interface{}{
				"api_key": "ab*****23",
				"count":   42,
			},
		},
		{
			name: "redact non-string secrets",
			input: map[string]interface{}{
				"secret": 12345,
				"data":   "normal",
			},
			expected: map[string]interface{}{
				"secret": "*****",
				"data":   "normal",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactSecretsInterface(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			// Compare keys
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d keys, got %d", len(tt.expected), len(result))
				return
			}

			for key := range tt.expected {
				if _, exists := result[key]; !exists {
					t.Errorf("key %s missing from result", key)
				}
			}
		})
	}
}

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"authorization header", "Authorization", true},
		{"password field", "password", true},
		{"api key", "X-API-KEY", true},
		{"secret config", "client_secret", true},
		{"token field", "access_token", true},
		{"auth header", "auth", true},
		{"credential field", "credential", true},
		{"normal field", "Content-Type", false},
		{"user agent", "User-Agent", false},
		{"timeout", "timeout", false},
		{"data field", "data", false},
		{"case insensitive", "PASSWORD", true},
		{"mixed case", "Api-Key", true},
		{"partial match", "my_password_field", true},
		{"bearer token", "bearer_token", true},
		{"jwt token", "jwt_token", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSensitiveKey(tt.key)
			if result != tt.expected {
				t.Errorf("key %s: expected %v, got %v", tt.key, tt.expected, result)
			}
		})
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"short value", "abc", "*****"},
		{"medium value", "password", "p*****d"},
		{"bearer token", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "Bearer *****"},
		{"basic auth", "Basic dXNlcjpwYXNz", "Basic *****"},
		{"long value", "verylongpassword123", "ve*****23"},
		{"api key", "sk-1234567890abcdef", "sk*****ef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskValue(tt.input)
			if result != tt.expected {
				t.Errorf("input %s: expected %s, got %s", tt.input, tt.expected, result)
			}
		})
	}
}

func TestRedactSecretsInString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL with password parameter",
			input:    "https://api.example.com/data?user=admin&password=secret123&format=json",
			expected: "https://api.example.com/data?user=admin&password=*****&format=json",
		},
		{
			name:     "JSON with password field",
			input:    `{"username": "admin", "password": "secret123", "timeout": 30}`,
			expected: `{"username": "admin", "password": "*****", "timeout": 30}`,
		},
		{
			name:     "Authorization header in log",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "Authorization: Bearer *****",
		},
		{
			name:     "Multiple secrets",
			input:    "token=abc123&api_key=def456&user=admin",
			expected: "token=*****&api_key=*****&user=admin",
		},
		{
			name:     "No secrets",
			input:    "https://api.example.com/data?user=admin&format=json",
			expected: "https://api.example.com/data?user=admin&format=json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactSecretsInString(tt.input)
			if result != tt.expected {
				t.Errorf("input %s: expected %s, got %s", tt.input, tt.expected, result)
			}
		})
	}
}
