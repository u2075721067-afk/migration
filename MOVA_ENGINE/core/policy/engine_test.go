package policy

import (
	"testing"
	"time"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine("/test/profiles")

	if engine == nil {
		t.Fatal("Engine should not be nil")
	}

	if engine.profilesDir != "/test/profiles" {
		t.Errorf("Expected profilesDir to be '/test/profiles', got '%s'", engine.profilesDir)
	}

	if len(engine.profiles) != 0 {
		t.Errorf("Expected empty profiles map, got %d profiles", len(engine.profiles))
	}

	if len(engine.policies) != 0 {
		t.Errorf("Expected empty policies map, got %d policies", len(engine.policies))
	}
}

func TestLoadProfiles(t *testing.T) {
	// Test with non-existent directory
	engine := NewEngine("/non/existent/path")
	err := engine.LoadProfiles()

	// Should not fail, just return empty profiles
	if err != nil {
		t.Errorf("Expected no error for non-existent path, got: %v", err)
	}

	if len(engine.profiles) != 0 {
		t.Errorf("Expected no profiles loaded, got %d", len(engine.profiles))
	}
}

func TestLoadPolicy(t *testing.T) {
	engine := NewEngine("/test/profiles")

	// Add test profile first
	engine.profiles["balanced"] = &RetryProfile{Name: "balanced"}

	// Test YAML loading
	yamlData := []byte(`
name: "Test Policy"
description: "Test description"
retryProfile: "balanced"
enabled: true
`)

	policy, err := engine.LoadPolicy(yamlData, "yaml")
	if err != nil {
		t.Fatalf("Failed to load YAML policy: %v", err)
	}

	if policy.Name != "Test Policy" {
		t.Errorf("Expected name 'Test Policy', got '%s'", policy.Name)
	}

	if policy.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", policy.Description)
	}

	if policy.RetryProfile != "balanced" {
		t.Errorf("Expected retryProfile 'balanced', got '%s'", policy.RetryProfile)
	}

	if !policy.Enabled {
		t.Error("Expected policy to be enabled")
	}
}

func TestLoadPolicyUnsupportedFormat(t *testing.T) {
	engine := NewEngine("/test/profiles")

	_, err := engine.LoadPolicy([]byte("test"), "json")
	if err == nil {
		t.Error("Expected error for unsupported format")
	}

	if err.Error() != "unsupported format: json, only YAML is supported" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestValidatePolicy(t *testing.T) {
	engine := NewEngine("/test/profiles")

	// Add a test profile
	engine.profiles["balanced"] = &RetryProfile{
		Name: "balanced",
	}

	tests := []struct {
		name    string
		policy  *Policy
		wantErr bool
	}{
		{
			name: "valid policy",
			policy: &Policy{
				Name:         "Test Policy",
				RetryProfile: "balanced",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			policy: &Policy{
				RetryProfile: "balanced",
			},
			wantErr: true,
		},
		{
			name: "missing retry profile",
			policy: &Policy{
				Name: "Test Policy",
			},
			wantErr: true,
		},
		{
			name: "non-existent retry profile",
			policy: &Policy{
				Name:         "Test Policy",
				RetryProfile: "non-existent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.validatePolicy(tt.policy)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddPolicy(t *testing.T) {
	engine := NewEngine("/test/profiles")

	// Add a test profile
	engine.profiles["balanced"] = &RetryProfile{
		Name: "balanced",
	}

	policy := &Policy{
		ID:           "test-1",
		Name:         "Test Policy",
		RetryProfile: "balanced",
	}

	err := engine.AddPolicy(policy)
	if err != nil {
		t.Errorf("Failed to add policy: %v", err)
	}

	if len(engine.policies) != 1 {
		t.Errorf("Expected 1 policy, got %d", len(engine.policies))
	}

	if engine.policies["test-1"] != policy {
		t.Error("Policy not stored correctly")
	}
}

func TestRemovePolicy(t *testing.T) {
	engine := NewEngine("/test/profiles")

	// Add a test profile
	engine.profiles["balanced"] = &RetryProfile{
		Name: "balanced",
	}

	policy := &Policy{
		ID:           "test-1",
		Name:         "Test Policy",
		RetryProfile: "balanced",
	}

	engine.AddPolicy(policy)

	// Remove policy
	err := engine.RemovePolicy("test-1")
	if err != nil {
		t.Errorf("Failed to remove policy: %v", err)
	}

	if len(engine.policies) != 0 {
		t.Errorf("Expected 0 policies, got %d", len(engine.policies))
	}
}

func TestRemovePolicyNotFound(t *testing.T) {
	engine := NewEngine("/test/profiles")

	err := engine.RemovePolicy("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent policy")
	}

	if err.Error() != "policy not found: non-existent" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestListPolicies(t *testing.T) {
	engine := NewEngine("/test/profiles")

	// Add test profiles
	engine.profiles["balanced"] = &RetryProfile{Name: "balanced"}
	engine.profiles["aggressive"] = &RetryProfile{Name: "aggressive"}

	// Add test policies
	policy1 := &Policy{ID: "test-1", Name: "Policy 1", RetryProfile: "balanced"}
	policy2 := &Policy{ID: "test-2", Name: "Policy 2", RetryProfile: "aggressive"}

	engine.AddPolicy(policy1)
	engine.AddPolicy(policy2)

	policies := engine.ListPolicies()
	if len(policies) != 2 {
		t.Errorf("Expected 2 policies, got %d", len(policies))
	}
}

func TestListProfiles(t *testing.T) {
	engine := NewEngine("/test/profiles")

	// Add test profiles
	engine.profiles["balanced"] = &RetryProfile{Name: "balanced"}
	engine.profiles["aggressive"] = &RetryProfile{Name: "aggressive"}

	profiles := engine.ListProfiles()
	if len(profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profiles))
	}
}

func TestFindMatchingPolicy(t *testing.T) {
	engine := NewEngine("/test/profiles")

	// Add test profiles
	engine.profiles["balanced"] = &RetryProfile{Name: "balanced"}
	engine.profiles["aggressive"] = &RetryProfile{Name: "aggressive"}

	// Add test policies
	policy1 := &Policy{
		ID:           "timeout-policy",
		Name:         "Timeout Policy",
		RetryProfile: "aggressive",
		Enabled:      true,
		Conditions: []Condition{
			{ErrorType: "timeout", HTTPStatus: 408},
		},
	}

	policy2 := &Policy{
		ID:           "network-policy",
		Name:         "Network Policy",
		RetryProfile: "balanced",
		Enabled:      true,
		Conditions: []Condition{
			{ErrorType: "network", HTTPStatus: 502},
		},
	}

	engine.AddPolicy(policy1)
	engine.AddPolicy(policy2)

	tests := []struct {
		name     string
		context  *PolicyContext
		expected string
	}{
		{
			name: "timeout error should match timeout policy",
			context: &PolicyContext{
				ErrorType:  "timeout",
				HTTPStatus: 408,
			},
			expected: "timeout-policy",
		},
		{
			name: "network error should match network policy",
			context: &PolicyContext{
				ErrorType:  "network",
				HTTPStatus: 502,
			},
			expected: "network-policy",
		},
		{
			name: "unknown error should not match any policy",
			context: &PolicyContext{
				ErrorType: "unknown",
			},
			expected: "", // No match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := engine.FindMatchingPolicy(tt.context)

			if tt.expected == "" {
				if match != nil {
					t.Errorf("Expected no match, got policy: %s", match.Policy.ID)
				}
			} else {
				if match == nil {
					t.Errorf("Expected match for policy: %s, got none", tt.expected)
				} else if match.Policy.ID != tt.expected {
					t.Errorf("Expected policy %s, got %s", tt.expected, match.Policy.ID)
				}
			}
		})
	}
}

func TestCalculateMatchScore(t *testing.T) {
	engine := NewEngine("/test/profiles")

	policy := &Policy{
		Conditions: []Condition{
			{ErrorType: "timeout", HTTPStatus: 408},
			{ActionType: "http_fetch"},
		},
	}

	context := &PolicyContext{
		ErrorType:  "timeout",
		HTTPStatus: 408,
		ActionType: "http_fetch",
	}

	score := engine.calculateMatchScore(policy, context)
	expectedScore := 10 + 8 + 6 // errorType + httpStatus + actionType

	if score != expectedScore {
		t.Errorf("Expected score %d, got %d", expectedScore, score)
	}
}

func TestGetRetryPolicy(t *testing.T) {
	engine := NewEngine("/test/profiles")

	// Add test profiles
	engine.profiles["balanced"] = &RetryProfile{Name: "balanced"}
	engine.profiles["aggressive"] = &RetryProfile{Name: "aggressive"}

	// Add test policy
	policy := &Policy{
		ID:           "test-policy",
		Name:         "Test Timeout Policy",
		RetryProfile: "aggressive",
		Enabled:      true,
		Conditions: []Condition{
			{ErrorType: "timeout"},
		},
	}

	err := engine.AddPolicy(policy)
	if err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}

	// Test with matching context
	context := &PolicyContext{ErrorType: "timeout"}
	retryProfile, err := engine.GetRetryPolicy(context)

	if err != nil {
		t.Errorf("Failed to get retry policy: %v", err)
	}

	if retryProfile == nil {
		t.Fatal("Expected retry profile, got nil")
	}

	if retryProfile.Name != "aggressive" {
		t.Errorf("Expected aggressive profile, got %s", retryProfile.Name)
	}

	// Test with non-matching context (should return default balanced)
	context = &PolicyContext{ErrorType: "unknown"}
	retryProfile, err = engine.GetRetryPolicy(context)

	if err != nil {
		t.Errorf("Failed to get default retry policy: %v", err)
	}

	if retryProfile == nil {
		t.Fatal("Expected default retry profile, got nil")
	}

	if retryProfile.Name != "balanced" {
		t.Errorf("Expected balanced profile, got %s", retryProfile.Name)
	}
}

func TestPolicyTimestamps(t *testing.T) {
	engine := NewEngine("/test/profiles")

	// Add a test profile
	engine.profiles["balanced"] = &RetryProfile{Name: "balanced"}

	yamlData := []byte(`
name: "Test Policy"
retryProfile: "balanced"
`)

	policy, err := engine.LoadPolicy(yamlData, "yaml")
	if err != nil {
		t.Fatalf("Failed to load policy: %v", err)
	}

	// Check that timestamps are set
	if policy.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if policy.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	// Check that timestamps are recent (within last minute)
	now := time.Now()
	if now.Sub(policy.CreatedAt) > time.Minute {
		t.Error("CreatedAt should be recent")
	}

	if now.Sub(policy.UpdatedAt) > time.Minute {
		t.Error("UpdatedAt should be recent")
	}
}
