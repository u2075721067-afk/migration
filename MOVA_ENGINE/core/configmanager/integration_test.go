package configmanager

import (
	"context"
	"testing"
	"time"
)

// TestIntegration tests the complete ConfigManager workflow
func TestIntegration(t *testing.T) {
	// Create storage and validator
	storage := NewMockStorage()
	validator := NewValidator()

	// Create config manager
	manager := NewManager(storage, validator)

	// Test data
	policy := &PolicyConfig{
		ID:           "integration-policy",
		Name:         "Integration Test Policy",
		Description:  "Policy for integration testing",
		RetryProfile: "balanced",
		Enabled:      true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Conditions: []ConditionConfig{
			{
				ErrorType:  "http_error",
				HTTPStatus: 500,
			},
		},
		BudgetConstraints: BudgetConstraintConfig{
			MaxRetriesPerWorkflow: 10,
			MaxRetriesPerSession:  5,
		},
	}

	budget := &BudgetConfig{
		ID:          "integration-budget",
		Name:        "Integration Test Budget",
		Description: "Budget for integration testing",
		Type:        "retries",
		Scope:       "global",
		MaxCount:    1000,
		TimeWindow:  time.Hour,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	profile := &RetryProfileConfig{
		Name:              "balanced",
		Description:       "Balanced retry profile for integration",
		MaxRetries:        5,
		InitialDelay:      time.Second,
		MaxDelay:          10 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            0.2,
		Timeout:           30 * time.Second,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Save test data
	if err := storage.SavePolicy(context.Background(), policy); err != nil {
		t.Fatalf("failed to save policy: %v", err)
	}

	if err := storage.SaveBudget(context.Background(), budget); err != nil {
		t.Fatalf("failed to save budget: %v", err)
	}

	if err := storage.SaveRetryProfile(context.Background(), profile); err != nil {
		t.Fatalf("failed to save retry profile: %v", err)
	}

	// Test export in all formats
	formats := []ConfigFormat{FormatJSON, FormatYAML, FormatHCL}

	for _, format := range formats {
		t.Run("Export_"+string(format), func(t *testing.T) {
			opts := ExportOptions{
				Format:           format,
				IncludeDLQ:       false,
				IncludeWorkflows: false,
				Compress:         false,
			}

			bundle, err := manager.Export(context.Background(), opts)
			if err != nil {
				t.Fatalf("export failed for format %s: %v", format, err)
			}

			if bundle == nil {
				t.Fatal("expected bundle to be created")
			}

			if bundle.Metadata.Format != string(format) {
				t.Errorf("expected format %s, got %s", format, bundle.Metadata.Format)
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
		})
	}

	// Test import with different modes
	t.Run("Import_Merge", func(t *testing.T) {
		// Create new configuration to import
		newPolicy := &PolicyConfig{
			ID:           "new-policy",
			Name:         "New Policy",
			Description:  "New policy for testing",
			RetryProfile: "balanced",
			Enabled:      true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		newBundle := &ConfigBundle{
			Metadata: ConfigMetadata{
				Version:     "1.0.0",
				Format:      "yaml",
				GeneratedAt: time.Now().UTC(),
				Source:      "test",
			},
			Policies: []PolicyConfig{*newPolicy},
		}

		// Export to YAML
		yamlExporter := &YAMLExporter{}
		yamlData, err := yamlExporter.Export(newBundle)
		if err != nil {
			t.Fatalf("failed to create YAML data: %v", err)
		}

		// Import with merge mode
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

		if result.Imported != 1 {
			t.Errorf("expected 1 item imported, got %d", result.Imported)
		}

		// Verify imported data
		importedPolicy, err := storage.GetPolicy(context.Background(), "new-policy")
		if err != nil {
			t.Errorf("imported policy not found: %v", err)
		}

		if importedPolicy.Name != "New Policy" {
			t.Errorf("expected policy name 'New Policy', got %s", importedPolicy.Name)
		}
	})

	// Test validation
	t.Run("Validation", func(t *testing.T) {
		// Create valid configuration
		validBundle := &ConfigBundle{
			Metadata: ConfigMetadata{
				Version:     "1.0.0",
				Format:      "yaml",
				GeneratedAt: time.Now().UTC(),
				Source:      "test",
			},
			Policies: []PolicyConfig{*policy},
		}

		// Export to YAML
		yamlExporter := &YAMLExporter{}
		yamlData, err := yamlExporter.Export(validBundle)
		if err != nil {
			t.Fatalf("failed to create YAML data: %v", err)
		}

		// Validate
		errors, err := manager.Validate(context.Background(), yamlData, FormatYAML)
		if err != nil {
			t.Fatalf("validation failed: %v", err)
		}

		if len(errors) > 0 {
			t.Errorf("expected no validation errors, got %d: %v", len(errors), errors)
		}
	})

	// Test dry-run import
	t.Run("Import_DryRun", func(t *testing.T) {
		// Create configuration to import
		dryRunBundle := &ConfigBundle{
			Metadata: ConfigMetadata{
				Version:     "1.0.0",
				Format:      "yaml",
				GeneratedAt: time.Now().UTC(),
				Source:      "test",
			},
			Policies: []PolicyConfig{
				{
					ID:           "dry-run-policy",
					Name:         "Dry Run Policy",
					Description:  "Policy for dry run testing",
					RetryProfile: "balanced",
					Enabled:      true,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
			},
		}

		// Export to YAML
		yamlExporter := &YAMLExporter{}
		yamlData, err := yamlExporter.Export(dryRunBundle)
		if err != nil {
			t.Fatalf("failed to create YAML data: %v", err)
		}

		// Import with dry-run mode
		opts := ImportOptions{
			Format:       FormatYAML,
			Mode:         ModeMerge,
			ValidateOnly: false,
			DryRun:       true,
			Overwrite:    false,
		}

		result, err := manager.Import(context.Background(), yamlData, opts)
		if err != nil {
			t.Fatalf("dry-run import failed: %v", err)
		}

		if !result.Success {
			t.Errorf("dry-run should succeed, got errors: %v", result.Errors)
		}

		// Verify nothing was actually imported
		policies, err := storage.ListPolicies(context.Background())
		if err != nil {
			t.Fatalf("failed to list policies: %v", err)
		}

		// Should only have the original policies (integration-policy and new-policy)
		if len(policies) != 2 {
			t.Errorf("expected 2 policies after dry-run, got %d", len(policies))
		}
	})
}
