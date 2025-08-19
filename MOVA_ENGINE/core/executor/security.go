package executor

import (
	"regexp"
	"strings"
)

// SensitiveKeys contains patterns for keys that should be redacted
var SensitiveKeys = []string{
	"authorization",
	"x-api-key",
	"password",
	"secret",
	"token",
	"key",
	"credential",
	"auth",
	"bearer",
	"jwt",
	"api_key",
	"apikey",
	"client_secret",
	"access_token",
	"refresh_token",
}

// RedactSecrets masks sensitive values in a map[string]string
func RedactSecrets(data map[string]string) map[string]string {
	if data == nil {
		return nil
	}

	redacted := make(map[string]string)
	for key, value := range data {
		if isSensitiveKey(key) {
			redacted[key] = maskValue(value)
		} else {
			redacted[key] = value
		}
	}
	return redacted
}

// RedactSecretsInterface masks sensitive values in a map[string]interface{}
func RedactSecretsInterface(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	redacted := make(map[string]interface{})
	for key, value := range data {
		if isSensitiveKey(key) {
			if strValue, ok := value.(string); ok {
				redacted[key] = maskValue(strValue)
			} else {
				redacted[key] = "*****"
			}
		} else {
			// Recursively redact nested maps
			if nestedMap, ok := value.(map[string]interface{}); ok {
				redacted[key] = RedactSecretsInterface(nestedMap)
			} else if nestedStringMap, ok := value.(map[string]string); ok {
				redacted[key] = RedactSecrets(nestedStringMap)
			} else {
				redacted[key] = value
			}
		}
	}
	return redacted
}

// isSensitiveKey checks if a key should be redacted
func isSensitiveKey(key string) bool {
	lowerKey := strings.ToLower(key)

	// Check exact matches and patterns
	for _, sensitive := range SensitiveKeys {
		if strings.Contains(lowerKey, sensitive) {
			return true
		}
	}

	// Check regex patterns for common secret formats
	patterns := []string{
		`.*password.*`,
		`.*secret.*`,
		`.*token.*`,
		`.*key.*`,
		`.*auth.*`,
		`.*credential.*`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, lowerKey)
		if matched {
			return true
		}
	}

	return false
}

// maskValue masks a sensitive value, showing only first/last characters for longer values
func maskValue(value string) string {
	if value == "" {
		return ""
	}

	if len(value) <= 4 {
		return "*****"
	}

	if len(value) <= 8 {
		return value[:1] + "*****" + value[len(value)-1:]
	}

	// For longer values like "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
	// Show "Bearer *****"
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 2 && (strings.ToLower(parts[0]) == "bearer" || strings.ToLower(parts[0]) == "basic") {
		return parts[0] + " *****"
	}

	// For other long values, show first 2 and last 2 characters
	return value[:2] + "*****" + value[len(value)-2:]
}

// RedactSecretsInString redacts secrets in string values (for URLs, JSON strings, etc.)
func RedactSecretsInString(input string) string {
	// Patterns for common secret formats in strings
	patterns := []struct {
		regex       *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`(password=)[^&\s]+`), "$1*****"},
		{regexp.MustCompile(`(token=)[^&\s]+`), "$1*****"},
		{regexp.MustCompile(`(key=)[^&\s]+`), "$1*****"},
		{regexp.MustCompile(`(secret=)[^&\s]+`), "$1*****"},
		{regexp.MustCompile(`(api_key=)[^&\s]+`), "$1*****"},
		{regexp.MustCompile(`(Authorization:\s*Bearer\s+)[^\s]+`), "$1*****"},
		{regexp.MustCompile(`(Authorization:\s*Basic\s+)[^\s]+`), "$1*****"},
		{regexp.MustCompile(`("password"\s*:\s*")[^"]+"`), "$1*****\""},
		{regexp.MustCompile(`("token"\s*:\s*")[^"]+"`), "$1*****\""},
		{regexp.MustCompile(`("secret"\s*:\s*")[^"]+"`), "$1*****\""},
		{regexp.MustCompile(`("api_key"\s*:\s*")[^"]+"`), "$1*****\""},
	}

	result := input
	for _, pattern := range patterns {
		result = pattern.regex.ReplaceAllString(result, pattern.replacement)
	}

	return result
}
