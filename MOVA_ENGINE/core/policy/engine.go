package policy

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

// Engine manages retry policies and profiles
type Engine struct {
	profiles    map[string]*RetryProfile
	policies    map[string]*Policy
	profilesDir string
}

// NewEngine creates a new policy engine
func NewEngine(profilesDir string) *Engine {
	return &Engine{
		profiles:    make(map[string]*RetryProfile),
		policies:    make(map[string]*Policy),
		profilesDir: profilesDir,
	}
}

// LoadProfiles loads retry profiles from the profiles directory
func (e *Engine) LoadProfiles() error {
	files, err := filepath.Glob(filepath.Join(e.profilesDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to glob profiles: %w", err)
	}

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read profile %s: %w", file, err)
		}

		var profile RetryProfile
		if err := yaml.Unmarshal(data, &profile); err != nil {
			return fmt.Errorf("failed to unmarshal profile %s: %w", file, err)
		}

		e.profiles[profile.Name] = &profile
	}

	return nil
}

// LoadPolicy loads a policy from YAML
func (e *Engine) LoadPolicy(data []byte, format string) (*Policy, error) {
	var policy Policy

	if format != "yaml" && format != "yml" {
		return nil, fmt.Errorf("unsupported format: %s, only YAML is supported", format)
	}

	if err := yaml.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Validate policy
	if err := e.validatePolicy(&policy); err != nil {
		return nil, fmt.Errorf("invalid policy: %w", err)
	}

	// Set timestamps
	now := time.Now()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = now
	}
	policy.UpdatedAt = now

	return &policy, nil
}

// validatePolicy validates policy configuration
func (e *Engine) validatePolicy(policy *Policy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if policy.RetryProfile == "" {
		return fmt.Errorf("retry profile is required")
	}

	if _, exists := e.profiles[policy.RetryProfile]; !exists {
		return fmt.Errorf("retry profile '%s' not found", policy.RetryProfile)
	}

	return nil
}

// FindMatchingPolicy finds the best matching policy for given context
func (e *Engine) FindMatchingPolicy(ctx *PolicyContext) *PolicyMatch {
	var bestMatch *PolicyMatch
	bestScore := -1

	for _, policy := range e.policies {
		if !policy.Enabled {
			continue
		}

		score := e.calculateMatchScore(policy, ctx)
		if score > 0 && score > bestScore {
			bestScore = score
			bestMatch = &PolicyMatch{
				Policy:       policy,
				RetryProfile: e.profiles[policy.RetryProfile],
				Score:        score,
			}
		}
	}

	return bestMatch
}

// PolicyContext represents the context for policy matching
type PolicyContext struct {
	ErrorType    string
	HTTPStatus   int
	ErrorMessage string
	ActionType   string
	WorkflowID   string
	SessionID    string
}

// calculateMatchScore calculates how well a policy matches the context
func (e *Engine) calculateMatchScore(policy *Policy, ctx *PolicyContext) int {
	score := 0

	for _, condition := range policy.Conditions {
		if condition.ErrorType != "" && condition.ErrorType == ctx.ErrorType {
			score += 10
		}

		if condition.HTTPStatus != 0 && condition.HTTPStatus == ctx.HTTPStatus {
			score += 8
		}

		if condition.ActionType != "" && condition.ActionType == ctx.ActionType {
			score += 6
		}

		if condition.ErrorMessagePattern != "" {
			if matched, _ := regexp.MatchString(condition.ErrorMessagePattern, ctx.ErrorMessage); matched {
				score += 5
			}
		}
	}

	return score
}

// GetRetryPolicy returns the retry policy for the given context
func (e *Engine) GetRetryPolicy(ctx *PolicyContext) (*RetryProfile, error) {
	match := e.FindMatchingPolicy(ctx)
	if match == nil {
		// Return default balanced profile
		if profile, exists := e.profiles["balanced"]; exists {
			return profile, nil
		}
		return nil, fmt.Errorf("no matching policy and no default profile found")
	}

	return match.RetryProfile, nil
}

// AddPolicy adds a policy to the engine
func (e *Engine) AddPolicy(policy *Policy) error {
	if err := e.validatePolicy(policy); err != nil {
		return err
	}

	e.policies[policy.ID] = policy
	return nil
}

// RemovePolicy removes a policy from the engine
func (e *Engine) RemovePolicy(id string) error {
	if _, exists := e.policies[id]; !exists {
		return fmt.Errorf("policy not found: %s", id)
	}

	delete(e.policies, id)
	return nil
}

// ListPolicies returns all policies
func (e *Engine) ListPolicies() []*Policy {
	policies := make([]*Policy, 0, len(e.policies))
	for _, policy := range e.policies {
		policies = append(policies, policy)
	}
	return policies
}

// ListProfiles returns all retry profiles
func (e *Engine) ListProfiles() []*RetryProfile {
	profiles := make([]*RetryProfile, 0, len(e.profiles))
	for _, profile := range e.profiles {
		profiles = append(profiles, profile)
	}
	return profiles
}
