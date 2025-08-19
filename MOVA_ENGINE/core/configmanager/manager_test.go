package configmanager

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockStorage implements ConfigStorage for testing
type MockStorage struct {
	policies      map[string]*PolicyConfig
	budgets       map[string]*BudgetConfig
	retryProfiles map[string]*RetryProfileConfig
}

// NewMockStorage creates a new mock storage
func NewMockStorage() *MockStorage {
	return &MockStorage{
		policies:      make(map[string]*PolicyConfig),
		budgets:       make(map[string]*BudgetConfig),
		retryProfiles: make(map[string]*RetryProfileConfig),
	}
}

// Mock methods for ConfigStorage interface
func (m *MockStorage) SavePolicy(ctx context.Context, policy *PolicyConfig) error {
	m.policies[policy.ID] = policy
	return nil
}

func (m *MockStorage) SaveBudget(ctx context.Context, budget *BudgetConfig) error {
	m.budgets[budget.ID] = budget
	return nil
}

func (m *MockStorage) SaveRetryProfile(ctx context.Context, profile *RetryProfileConfig) error {
	m.retryProfiles[profile.Name] = profile
	return nil
}

func (m *MockStorage) GetPolicy(ctx context.Context, id string) (*PolicyConfig, error) {
	if policy, exists := m.policies[id]; exists {
		return policy, nil
	}
	return nil, fmt.Errorf("policy not found: %s", id)
}

func (m *MockStorage) GetBudget(ctx context.Context, id string) (*BudgetConfig, error) {
	if budget, exists := m.budgets[id]; exists {
		return budget, nil
	}
	return nil, fmt.Errorf("budget not found: %s", id)
}

func (m *MockStorage) GetRetryProfile(ctx context.Context, name string) (*RetryProfileConfig, error) {
	if profile, exists := m.retryProfiles[name]; exists {
		return profile, nil
	}
	return nil, fmt.Errorf("retry profile not found: %s", name)
}

func (m *MockStorage) ListPolicies(ctx context.Context) ([]*PolicyConfig, error) {
	policies := make([]*PolicyConfig, 0, len(m.policies))
	for _, policy := range m.policies {
		policies = append(policies, policy)
	}
	return policies, nil
}

func (m *MockStorage) ListBudgets(ctx context.Context) ([]*BudgetConfig, error) {
	budgets := make([]*BudgetConfig, 0, len(m.budgets))
	for _, budget := range m.budgets {
		budgets = append(budgets, budget)
	}
	return budgets, nil
}

func (m *MockStorage) ListRetryProfiles(ctx context.Context) ([]*RetryProfileConfig, error) {
	profiles := make([]*RetryProfileConfig, 0, len(m.retryProfiles))
	for _, profile := range m.retryProfiles {
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func (m *MockStorage) DeletePolicy(ctx context.Context, id string) error {
	delete(m.policies, id)
	return nil
}

func (m *MockStorage) DeleteBudget(ctx context.Context, id string) error {
	delete(m.budgets, id)
	return nil
}

func (m *MockStorage) DeleteRetryProfile(ctx context.Context, name string) error {
	delete(m.retryProfiles, name)
	return nil
}

// TestNewManager tests ConfigManager creation
func TestNewManager(t *testing.T) {
	storage := NewMockStorage()
	validator := NewValidator()

	manager := NewManager(storage, validator)

	if manager == nil {
		t.Fatal("expected manager to be created")
	}

	if manager.GetVersion() != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", manager.GetVersion())
	}

	formats := manager.GetSupportedFormats()
	expectedFormats := []ConfigFormat{FormatJSON, FormatYAML, FormatHCL}

	if len(formats) != len(expectedFormats) {
		t.Errorf("expected %d formats, got %d", len(expectedFormats), len(formats))
	}
}

// TestExport tests configuration export
func TestExport(t *testing.T) {
	storage := NewMockStorage()
	validator := NewValidator()
	manager := NewManager(storage, validator)

	// Add test data
	policy := &PolicyConfig{
		ID:           "test-policy",
		Name:         "Test Policy",
		Description:  "Test policy for testing",
		RetryProfile: "balanced",
		Enabled:      true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	storage.SavePolicy(context.Background(), policy)

	budget := &BudgetConfig{
		ID:          "test-budget",
		Name:        "Test Budget",
		Description: "Test budget for testing",
		Type:        "retries",
		Scope:       "global",
		MaxCount:    100,
		TimeWindow:  time.Hour,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	storage.SaveBudget(context.Background(), budget)

	profile := &RetryProfileConfig{
		Name:              "balanced",
		Description:       "Balanced retry profile",
		MaxRetries:        5,
		InitialDelay:      time.Second,
		MaxDelay:          10 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            0.2,
		Timeout:           30 * time.Second,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	storage.SaveRetryProfile(context.Background(), profile)

	// Test export
	opts := ExportOptions{
		Format:           FormatYAML,
		IncludeDLQ:       false,
		IncludeWorkflows: false,
		Compress:         false,
	}

	bundle, err := manager.Export(context.Background(), opts)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	if bundle == nil {
		t.Fatal("expected bundle to be created")
	}

	if bundle.Metadata.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", bundle.Metadata.Version)
	}

	if len(bundle.Policies) != 1 {
		t.Errorf("expected 1 policy, got %d", len(bundle.Policies))
	}

	if len(bundle.Budgets) != 1 {
		t.Errorf("expected 1 budget, got %d", len(bundle.Budgets))
	}

	if len(bundle.RetryProfiles) != 1 {
		t.Errorf("expected 1 retry profile, got %d", len(bundle.RetryProfiles))
	}
}

// TestImport tests configuration import
func TestImport(t *testing.T) {
	storage := NewMockStorage()
	validator := NewValidator()
	manager := NewManager(storage, validator)

	// Create test configuration data
	testConfig := &ConfigBundle{
		Metadata: ConfigMetadata{
			Version:     "1.0.0",
			Format:      "yaml",
			GeneratedAt: time.Now().UTC(),
			Source:      "test",
		},
		Policies: []PolicyConfig{
			{
				ID:           "imported-policy",
				Name:         "Imported Policy",
				Description:  "Policy imported from test",
				RetryProfile: "balanced",
				Enabled:      true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
		Budgets: []BudgetConfig{
			{
				ID:          "imported-budget",
				Name:        "Imported Budget",
				Description: "Budget imported from test",
				Type:        "retries",
				Scope:       "global",
				MaxCount:    200,
				TimeWindow:  time.Hour,
				Enabled:     true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	// Convert to YAML for import test
	yamlExporter := &YAMLExporter{}
	yamlData, err := yamlExporter.Export(testConfig)
	if err != nil {
		t.Fatalf("failed to create YAML data: %v", err)
	}

	// Test import with merge mode
	opts := ImportOptions{
		Format:       FormatYAML,
		Mode:         ModeMerge,
		ValidateOnly: false,
		DryRun:       false,
		Overwrite:    false,
	}

	result, err := manager.Import(context.Background(), yamlData, opts)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	if !result.Success {
		t.Errorf("import should succeed, got errors: %v", result.Errors)
	}

	if result.Imported != 2 {
		t.Errorf("expected 2 items imported, got %d", result.Imported)
	}

	// Verify imported data
	importedPolicy, err := storage.GetPolicy(context.Background(), "imported-policy")
	if err != nil {
		t.Errorf("imported policy not found: %v", err)
	}

	if importedPolicy.Name != "Imported Policy" {
		t.Errorf("expected policy name 'Imported Policy', got %s", importedPolicy.Name)
	}

	importedBudget, err := storage.GetBudget(context.Background(), "imported-budget")
	if err != nil {
		t.Errorf("imported budget not found: %v", err)
	}

	if importedBudget.Name != "Imported Budget" {
		t.Errorf("expected budget name 'Imported Budget', got %s", importedBudget.Name)
	}
}

// TestValidation tests configuration validation
func TestValidation(t *testing.T) {
	validator := NewValidator()

	// Test valid configuration
	validBundle := &ConfigBundle{
		Metadata: ConfigMetadata{
			Version:     "1.0.0",
			Format:      "yaml",
			GeneratedAt: time.Now().UTC(),
			Source:      "test",
		},
		Policies: []PolicyConfig{
			{
				ID:           "valid-policy",
				Name:         "Valid Policy",
				Description:  "Valid policy for testing",
				RetryProfile: "balanced",
				Enabled:      true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
	}

	errors, err := validator.Validate(validBundle)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	if len(errors) > 0 {
		t.Errorf("expected no validation errors, got %d: %v", len(errors), errors)
	}

	// Test invalid configuration
	invalidBundle := &ConfigBundle{
		Metadata: ConfigMetadata{
			Format:      "yaml",
			GeneratedAt: time.Now().UTC(),
			Source:      "test",
			// Missing version
		},
		Policies: []PolicyConfig{
			{
				// Missing required fields
				Description: "Invalid policy for testing",
				Enabled:     true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	errors, err = validator.Validate(invalidBundle)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	if len(errors) == 0 {
		t.Error("expected validation errors, got none")
	}

	// Check for specific errors
	foundVersionError := false
	foundPolicyError := false

	for _, err := range errors {
		if err.Type == "metadata" && err.Message == "version is required" {
			foundVersionError = true
		}
		if err.Type == "policy" && err.Message == "ID is required" {
			foundPolicyError = true
		}
	}

	if !foundVersionError {
		t.Error("expected version validation error")
	}

	if !foundPolicyError {
		t.Error("expected policy ID validation error")
	}
}

// TestFormatSupport tests format support
func TestFormatSupport(t *testing.T) {
	storage := NewMockStorage()
	validator := NewValidator()
	manager := NewManager(storage, validator)

	// Test JSON export
	jsonData, err := manager.Export(context.Background(), ExportOptions{Format: FormatJSON})
	if err != nil {
		t.Fatalf("JSON export failed: %v", err)
	}

	if jsonData == nil {
		t.Fatal("expected JSON export data")
	}

	// Test YAML export
	yamlData, err := manager.Export(context.Background(), ExportOptions{Format: FormatYAML})
	if err != nil {
		t.Fatalf("YAML export failed: %v", err)
	}

	if yamlData == nil {
		t.Fatal("expected YAML export data")
	}

	// Test HCL export
	hclData, err := manager.Export(context.Background(), ExportOptions{Format: FormatHCL})
	if err != nil {
		t.Fatalf("HCL export failed: %v", err)
	}

	if hclData == nil {
		t.Fatal("expected HCL export data")
	}
}

// TestImportModes tests different import modes
func TestImportModes(t *testing.T) {
	storage := NewMockStorage()
	validator := NewValidator()
	manager := NewManager(storage, validator)

	// Create test configuration
	testConfig := &ConfigBundle{
		Metadata: ConfigMetadata{
			Version:     "1.0.0",
			Format:      "yaml",
			GeneratedAt: time.Now().UTC(),
			Source:      "test",
		},
		Policies: []PolicyConfig{
			{
				ID:           "test-policy",
				Name:         "Test Policy",
				Description:  "Test policy",
				RetryProfile: "balanced",
				Enabled:      true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
	}

	// Test validate-only mode
	yamlExporter := &YAMLExporter{}
	yamlData, err := yamlExporter.Export(testConfig)
	if err != nil {
		t.Fatalf("failed to create YAML data: %v", err)
	}

	validateOpts := ImportOptions{
		Format:       FormatYAML,
		Mode:         ModeValidate,
		ValidateOnly: true,
	}

	result, err := manager.Import(context.Background(), yamlData, validateOpts)
	if err != nil {
		t.Fatalf("validate-only import failed: %v", err)
	}

	if !result.Success {
		t.Errorf("validation should succeed, got errors: %v", result.Errors)
	}

	// Verify nothing was actually imported
	policies, err := storage.ListPolicies(context.Background())
	if err != nil {
		t.Fatalf("failed to list policies: %v", err)
	}

	if len(policies) != 0 {
		t.Errorf("expected no policies to be imported in validate-only mode, got %d", len(policies))
	}

	// Test dry-run mode
	dryRunOpts := ImportOptions{
		Format: FormatYAML,
		Mode:   ModeMerge,
		DryRun: true,
	}

	result, err = manager.Import(context.Background(), yamlData, dryRunOpts)
	if err != nil {
		t.Fatalf("dry-run import failed: %v", err)
	}

	if !result.Success {
		t.Errorf("dry-run should succeed, got errors: %v", result.Errors)
	}

	// Verify nothing was actually imported
	policies, err = storage.ListPolicies(context.Background())
	if err != nil {
		t.Fatalf("failed to list policies: %v", err)
	}

	if len(policies) != 0 {
		t.Errorf("expected no policies to be imported in dry-run mode, got %d", len(policies))
	}
}
