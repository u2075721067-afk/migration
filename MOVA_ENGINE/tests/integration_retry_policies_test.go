package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mova-engine/mova-engine/core/policy"
	"gopkg.in/yaml.v3"
)

func TestRetryPolicyIntegration(t *testing.T) {
	// Create temporary directory for profiles
	tempDir, err := os.MkdirTemp("", "retry-profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test profiles
	profiles := map[string]policy.RetryProfile{
		"aggressive": {
			Name:              "aggressive",
			Description:       "Fast retry with minimal backoff",
			MaxRetries:        3,
			InitialDelay:      100 * time.Millisecond,
			MaxDelay:          1 * time.Second,
			BackoffMultiplier: 1.5,
			Jitter:            0.1,
			Timeout:           5 * time.Second,
		},
		"balanced": {
			Name:              "balanced",
			Description:       "Balanced retry with exponential backoff",
			MaxRetries:        5,
			InitialDelay:      500 * time.Millisecond,
			MaxDelay:          10 * time.Second,
			BackoffMultiplier: 2.0,
			Jitter:            0.2,
			Timeout:           30 * time.Second,
		},
		"conservative": {
			Name:              "conservative",
			Description:       "Conservative retry with long intervals",
			MaxRetries:        10,
			InitialDelay:      2 * time.Second,
			MaxDelay:          60 * time.Second,
			BackoffMultiplier: 2.5,
			Jitter:            0.3,
			Timeout:           300 * time.Second,
		},
	}

	// Write profiles to files
	for name, profile := range profiles {
		profileData, err := yaml.Marshal(profile)
		if err != nil {
			t.Fatalf("Failed to marshal profile %s: %v", name, err)
		}

		profilePath := filepath.Join(tempDir, name+".yaml")
		if err := os.WriteFile(profilePath, profileData, 0644); err != nil {
			t.Fatalf("Failed to write profile %s: %v", name, err)
		}
	}

	// Create policy engine
	engine := policy.NewEngine(tempDir)

	// Load profiles
	if err := engine.LoadProfiles(); err != nil {
		t.Fatalf("Failed to load profiles: %v", err)
	}

	// Verify profiles were loaded
	loadedProfiles := engine.ListProfiles()
	if len(loadedProfiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(loadedProfiles))
	}

	// Test policy creation and matching
	t.Run("TimeoutPolicy", func(t *testing.T) {
		testTimeoutPolicy(t, engine)
	})

	t.Run("RateLimitPolicy", func(t *testing.T) {
		testRateLimitPolicy(t, engine)
	})

	t.Run("NetworkErrorPolicy", func(t *testing.T) {
		testNetworkErrorPolicy(t, engine)
	})

	t.Run("PolicyPriority", func(t *testing.T) {
		testPolicyPriority(t, engine)
	})
}

func testTimeoutPolicy(t *testing.T, engine *policy.Engine) {
	// Create timeout policy
	timeoutPolicy := &policy.Policy{
		ID:           "timeout-policy",
		Name:         "Timeout Handling Policy",
		Description:  "Aggressive retry for timeout errors",
		RetryProfile: "aggressive",
		Enabled:      true,
		Conditions: []policy.Condition{
			{
				ErrorType:  "timeout",
				HTTPStatus: 408,
			},
		},
	}

	if err := engine.AddPolicy(timeoutPolicy); err != nil {
		t.Fatalf("Failed to add timeout policy: %v", err)
	}

	// Test matching timeout context
	timeoutContext := &policy.PolicyContext{
		ErrorType:  "timeout",
		HTTPStatus: 408,
	}

	match := engine.FindMatchingPolicy(timeoutContext)
	if match == nil {
		t.Fatal("Expected to find matching timeout policy")
	}

	if match.Policy.ID != "timeout-policy" {
		t.Errorf("Expected timeout-policy, got %s", match.Policy.ID)
	}

	if match.RetryProfile.Name != "aggressive" {
		t.Errorf("Expected aggressive profile, got %s", match.RetryProfile.Name)
	}

	// Verify score calculation
	if match.Score != 18 { // 10 (errorType) + 8 (httpStatus)
		t.Errorf("Expected score 18, got %d", match.Score)
	}
}

func testRateLimitPolicy(t *testing.T, engine *policy.Engine) {
	// Create rate limit policy
	rateLimitPolicy := &policy.Policy{
		ID:           "rate-limit-policy",
		Name:         "Rate Limit Policy",
		Description:  "Conservative retry for rate limits",
		RetryProfile: "conservative",
		Enabled:      true,
		Conditions: []policy.Condition{
			{
				HTTPStatus:          429,
				ErrorMessagePattern: ".*rate limit.*",
			},
		},
	}

	if err := engine.AddPolicy(rateLimitPolicy); err != nil {
		t.Fatalf("Failed to add rate limit policy: %v", err)
	}

	// Test matching rate limit context
	rateLimitContext := &policy.PolicyContext{
		HTTPStatus:   429,
		ErrorMessage: "API rate limit exceeded",
	}

	match := engine.FindMatchingPolicy(rateLimitContext)
	if match == nil {
		t.Fatal("Expected to find matching rate limit policy")
	}

	if match.Policy.ID != "rate-limit-policy" {
		t.Errorf("Expected rate-limit-policy, got %s", match.Policy.ID)
	}

	if match.RetryProfile.Name != "conservative" {
		t.Errorf("Expected conservative profile, got %s", match.RetryProfile.Name)
	}
}

func testNetworkErrorPolicy(t *testing.T, engine *policy.Engine) {
	// Create network error policy
	networkPolicy := &policy.Policy{
		ID:           "network-policy",
		Name:         "Network Error Policy",
		Description:  "Balanced retry for network errors",
		RetryProfile: "balanced",
		Enabled:      true,
		Conditions: []policy.Condition{
			{
				ErrorType:  "network",
				HTTPStatus: 502,
			},
			{
				HTTPStatus: 503,
			},
		},
	}

	if err := engine.AddPolicy(networkPolicy); err != nil {
		t.Fatalf("Failed to add network policy: %v", err)
	}

	// Test matching network context
	networkContext := &policy.PolicyContext{
		ErrorType:  "network",
		HTTPStatus: 502,
	}

	match := engine.FindMatchingPolicy(networkContext)
	if match == nil {
		t.Fatal("Expected to find matching network policy")
	}

	if match.Policy.ID != "network-policy" {
		t.Errorf("Expected network-policy, got %s", match.Policy.ID)
	}

	// Test 503 status also matches
	serviceUnavailableContext := &policy.PolicyContext{
		HTTPStatus: 503,
	}

	match = engine.FindMatchingPolicy(serviceUnavailableContext)
	if match == nil {
		t.Fatal("Expected to find matching policy for 503")
	}

	if match.Policy.ID != "network-policy" {
		t.Errorf("Expected network-policy for 503, got %s", match.Policy.ID)
	}
}

func testPolicyPriority(t *testing.T, engine *policy.Engine) {
	// Create high-priority policy with specific conditions
	highPriorityPolicy := &policy.Policy{
		ID:           "high-priority-policy",
		Name:         "High Priority Policy",
		Description:  "Specific timeout + action policy",
		RetryProfile: "aggressive",
		Enabled:      true,
		Conditions: []policy.Condition{
			{
				ErrorType:  "timeout",
				HTTPStatus: 408,
				ActionType: "http_fetch",
			},
		},
	}

	if err := engine.AddPolicy(highPriorityPolicy); err != nil {
		t.Fatalf("Failed to add high priority policy: %v", err)
	}

	// Test with context that matches both timeout and high-priority policies
	specificContext := &policy.PolicyContext{
		ErrorType:  "timeout",
		HTTPStatus: 408,
		ActionType: "http_fetch",
	}

	match := engine.FindMatchingPolicy(specificContext)
	if match == nil {
		t.Fatal("Expected to find matching policy")
	}

	// Should match high-priority policy due to higher score
	if match.Policy.ID != "high-priority-policy" {
		t.Errorf("Expected high-priority-policy, got %s", match.Policy.ID)
	}

	// Verify score is higher (10 + 8 + 6 = 24)
	if match.Score != 24 {
		t.Errorf("Expected score 24, got %d", match.Score)
	}
}

func TestPolicyYAMLSerialization(t *testing.T) {
	// Test complete policy serialization/deserialization
	originalPolicy := policy.Policy{
		ID:           "test-policy",
		Name:         "Test Policy",
		Description:  "Test policy for serialization",
		RetryProfile: "balanced",
		Enabled:      true,
		Conditions: []policy.Condition{
			{
				ErrorType:           "timeout",
				HTTPStatus:          408,
				ErrorMessagePattern: ".*timeout.*",
				ActionType:          "http_fetch",
			},
		},
		BudgetConstraints: policy.BudgetConstraint{
			MaxRetriesPerWorkflow: 10,
			MaxRetriesPerSession:  50,
			MaxTotalRetryTime:     5 * time.Minute,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Serialize to YAML
	yamlData, err := yaml.Marshal(originalPolicy)
	if err != nil {
		t.Fatalf("Failed to marshal policy to YAML: %v", err)
	}

	// Deserialize from YAML
	var deserializedPolicy policy.Policy
	if err := yaml.Unmarshal(yamlData, &deserializedPolicy); err != nil {
		t.Fatalf("Failed to unmarshal policy from YAML: %v", err)
	}

	// Verify fields
	if deserializedPolicy.ID != originalPolicy.ID {
		t.Errorf("ID mismatch: expected %s, got %s", originalPolicy.ID, deserializedPolicy.ID)
	}

	if deserializedPolicy.Name != originalPolicy.Name {
		t.Errorf("Name mismatch: expected %s, got %s", originalPolicy.Name, deserializedPolicy.Name)
	}

	if deserializedPolicy.RetryProfile != originalPolicy.RetryProfile {
		t.Errorf("RetryProfile mismatch: expected %s, got %s", originalPolicy.RetryProfile, deserializedPolicy.RetryProfile)
	}

	if len(deserializedPolicy.Conditions) != len(originalPolicy.Conditions) {
		t.Errorf("Conditions length mismatch: expected %d, got %d", len(originalPolicy.Conditions), len(deserializedPolicy.Conditions))
	}
}

func TestPolicyEnginePerformance(t *testing.T) {
	// Create engine with multiple policies
	engine := policy.NewEngine("/test/profiles")

	// Create and add test profiles via YAML
	balancedProfile := policy.RetryProfile{Name: "balanced"}
	aggressiveProfile := policy.RetryProfile{Name: "aggressive"}

	// We need to use LoadProfiles or create a temporary directory
	tempProfileDir, _ := os.MkdirTemp("", "test-profiles")
	defer os.RemoveAll(tempProfileDir)

	balancedData, _ := yaml.Marshal(balancedProfile)
	aggressiveData, _ := yaml.Marshal(aggressiveProfile)

	os.WriteFile(filepath.Join(tempProfileDir, "balanced.yaml"), balancedData, 0644)
	os.WriteFile(filepath.Join(tempProfileDir, "aggressive.yaml"), aggressiveData, 0644)

	engine = policy.NewEngine(tempProfileDir)
	engine.LoadProfiles()

	// Add many policies
	for i := 0; i < 100; i++ {
		testPolicy := &policy.Policy{
			ID:           fmt.Sprintf("policy-%d", i),
			Name:         fmt.Sprintf("Policy %d", i),
			RetryProfile: "balanced",
			Enabled:      true,
			Conditions: []policy.Condition{
				{
					ErrorType:  fmt.Sprintf("error-%d", i%10),
					HTTPStatus: 400 + (i % 100),
				},
			},
		}

		if err := engine.AddPolicy(testPolicy); err != nil {
			t.Fatalf("Failed to add policy %d: %v", i, err)
		}
	}

	// Benchmark policy matching
	context := &policy.PolicyContext{
		ErrorType:  "error-5",
		HTTPStatus: 450,
	}

	start := time.Now()
	for i := 0; i < 1000; i++ {
		match := engine.FindMatchingPolicy(context)
		if match == nil {
			t.Fatal("Expected to find matching policy")
		}
	}
	duration := time.Since(start)

	// Should complete 1000 matches in reasonable time (< 100ms)
	if duration > 100*time.Millisecond {
		t.Errorf("Policy matching too slow: %v", duration)
	}

	t.Logf("1000 policy matches completed in %v", duration)
}

// Helper function to format duration as needed
func init() {
	// This is a placeholder for any initialization needed
}
