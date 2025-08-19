package configmanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Manager implements the ConfigManager interface
type Manager struct {
	storage   ConfigStorage
	validator ConfigValidator
	exporters map[ConfigFormat]ConfigExporter
	importers map[ConfigFormat]ConfigImporter
	version   string
}

// NewManager creates a new ConfigManager instance
func NewManager(storage ConfigStorage, validator ConfigValidator) *Manager {
	m := &Manager{
		storage:   storage,
		validator: validator,
		exporters: make(map[ConfigFormat]ConfigExporter),
		importers: make(map[ConfigFormat]ConfigImporter),
		version:   "1.0.0",
	}

	// Register default exporters and importers
	m.registerDefaultFormats()

	return m
}

// registerDefaultFormats registers the default format handlers
func (m *Manager) registerDefaultFormats() {
	// JSON format
	jsonExporter := &JSONExporter{}
	jsonImporter := &JSONImporter{}
	m.exporters[FormatJSON] = jsonExporter
	m.importers[FormatJSON] = jsonImporter

	// YAML format
	yamlExporter := &YAMLExporter{}
	yamlImporter := &YAMLImporter{}
	m.exporters[FormatYAML] = yamlExporter
	m.importers[FormatYAML] = yamlImporter

	// HCL format
	hclExporter := &HCLExporter{}
	hclImporter := &HCLImporter{}
	m.exporters[FormatHCL] = hclExporter
	m.importers[FormatHCL] = hclImporter
}

// RegisterExporter registers a custom exporter for a specific format
func (m *Manager) RegisterExporter(exporter ConfigExporter) {
	m.exporters[exporter.GetFormat()] = exporter
}

// RegisterImporter registers a custom importer for a specific format
func (m *Manager) RegisterImporter(importer ConfigImporter) {
	m.importers[importer.GetFormat()] = importer
}

// Export exports configurations in the specified format
func (m *Manager) Export(ctx context.Context, opts ExportOptions) (*ConfigBundle, error) {
	// Validate export options
	if err := m.validateExportOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid export options: %w", err)
	}

	// Create config bundle
	bundle, err := m.createConfigBundle(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create config bundle: %w", err)
	}

	return bundle, nil
}

// ExportToWriter exports configurations to an io.Writer in the specified format
func (m *Manager) ExportToWriter(ctx context.Context, opts ExportOptions, w io.Writer) error {
	bundle, err := m.Export(ctx, opts)
	if err != nil {
		return err
	}

	exporter, exists := m.exporters[opts.Format]
	if !exists {
		return fmt.Errorf("unsupported export format: %s", opts.Format)
	}

	return exporter.ExportToWriter(bundle, w)
}

// Import imports configurations from the specified format
func (m *Manager) Import(ctx context.Context, data []byte, opts ImportOptions) (*ImportResult, error) {
	// Validate import options
	if err := m.validateImportOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid import options: %w", err)
	}

	// Validate format support
	importer, exists := m.importers[opts.Format]
	if !exists {
		return nil, fmt.Errorf("unsupported import format: %s", opts.Format)
	}

	// Import configuration data
	bundle, err := importer.Import(data)
	if err != nil {
		return nil, fmt.Errorf("failed to import configuration: %w", err)
	}

	// Validate imported configuration
	if m.validator != nil {
		errors, err := m.validator.Validate(bundle)
		if err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
		if len(errors) > 0 {
			return &ImportResult{
				Success: false,
				Errors:  errors,
				Summary: ImportSummary{},
			}, nil
		}
	}

	// If validate-only mode, return without importing
	if opts.ValidateOnly {
		return &ImportResult{
			Success: true,
			Summary: m.createImportSummary(bundle),
		}, nil
	}

	// Perform dry-run if requested
	if opts.DryRun {
		return &ImportResult{
			Success: true,
			Summary: m.createImportSummary(bundle),
		}, nil
	}

	// Import configurations based on mode
	result, err := m.performImport(ctx, bundle, opts)
	if err != nil {
		return nil, fmt.Errorf("import operation failed: %w", err)
	}

	return result, nil
}

// ImportFromReader imports configurations from an io.Reader in the specified format
func (m *Manager) ImportFromReader(ctx context.Context, r io.Reader, opts ImportOptions) (*ImportResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	return m.Import(ctx, data, opts)
}

// Validate validates configuration data without importing
func (m *Manager) Validate(ctx context.Context, data []byte, format ConfigFormat) ([]ImportError, error) {
	importer, exists := m.importers[format]
	if !exists {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	bundle, err := importer.Import(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	if m.validator == nil {
		return nil, nil
	}

	return m.validator.Validate(bundle)
}

// GetSupportedFormats returns list of supported export/import formats
func (m *Manager) GetSupportedFormats() []ConfigFormat {
	formats := make([]ConfigFormat, 0, len(m.exporters))
	for format := range m.exporters {
		formats = append(formats, format)
	}
	return formats
}

// GetVersion returns the current configuration version
func (m *Manager) GetVersion() string {
	return m.version
}

// validateExportOptions validates export options
func (m *Manager) validateExportOptions(opts ExportOptions) error {
	if opts.Format == "" {
		return fmt.Errorf("format is required")
	}

	if _, exists := m.exporters[opts.Format]; !exists {
		return fmt.Errorf("unsupported export format: %s", opts.Format)
	}

	return nil
}

// validateImportOptions validates import options
func (m *Manager) validateImportOptions(opts ImportOptions) error {
	if opts.Format == "" {
		return fmt.Errorf("format is required")
	}

	if opts.Mode == "" {
		opts.Mode = ModeMerge
	}

	if opts.Mode != ModeOverwrite && opts.Mode != ModeMerge && opts.Mode != ModeValidate {
		return fmt.Errorf("invalid import mode: %s", opts.Mode)
	}

	return nil
}

// createConfigBundle creates a configuration bundle from storage
func (m *Manager) createConfigBundle(ctx context.Context, opts ExportOptions) (*ConfigBundle, error) {
	bundle := &ConfigBundle{
		Metadata: ConfigMetadata{
			Version:     m.version,
			Format:      string(opts.Format),
			GeneratedAt: time.Now().UTC(),
			Source:      "mova-engine",
		},
	}

	// Add policies
	policies, err := m.storage.ListPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	bundle.Policies = make([]PolicyConfig, len(policies))
	for i, policy := range policies {
		bundle.Policies[i] = *policy
	}

	// Add budgets
	budgets, err := m.storage.ListBudgets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list budgets: %w", err)
	}
	bundle.Budgets = make([]BudgetConfig, len(budgets))
	for i, budget := range budgets {
		bundle.Budgets[i] = *budget
	}

	// Add retry profiles
	profiles, err := m.storage.ListRetryProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list retry profiles: %w", err)
	}
	bundle.RetryProfiles = make([]RetryProfileConfig, len(profiles))
	for i, profile := range profiles {
		bundle.RetryProfiles[i] = *profile
	}

	// Add DLQ entries if requested (sandbox mode only)
	if opts.IncludeDLQ {
		// Note: DLQ entries are typically not stored permanently
		// This would need to be implemented based on the specific DLQ implementation
		bundle.DLQEntries = []DLQEntryConfig{}
	}

	// Add workflows if requested
	if opts.IncludeWorkflows {
		// Note: Workflows are typically not stored permanently
		// This would need to be implemented based on the specific workflow storage
		bundle.Workflows = []WorkflowConfig{}
	}

	// Calculate checksum
	checksum := m.calculateChecksum(bundle)
	bundle.Metadata.Checksum = checksum

	return bundle, nil
}

// performImport performs the actual import operation
func (m *Manager) performImport(ctx context.Context, bundle *ConfigBundle, opts ImportOptions) (*ImportResult, error) {
	result := &ImportResult{
		Success: true,
		Summary: m.createImportSummary(bundle),
	}

	// Import based on mode
	switch opts.Mode {
	case ModeOverwrite:
		return m.importOverwrite(ctx, bundle)
	case ModeMerge:
		return m.importMerge(ctx, bundle)
	default:
		return result, nil
	}
}

// importOverwrite imports configurations by overwriting existing ones
func (m *Manager) importOverwrite(ctx context.Context, bundle *ConfigBundle) (*ImportResult, error) {
	result := &ImportResult{
		Success: true,
		Summary: m.createImportSummary(bundle),
	}

	// Import policies
	for _, policy := range bundle.Policies {
		if err := m.storage.SavePolicy(ctx, &policy); err != nil {
			result.Errors = append(result.Errors, ImportError{
				Type:    "policy_save",
				Message: fmt.Sprintf("failed to save policy %s: %v", policy.ID, err),
			})
			result.Success = false
		} else {
			result.Imported++
		}
	}

	// Import budgets
	for _, budget := range bundle.Budgets {
		if err := m.storage.SaveBudget(ctx, &budget); err != nil {
			result.Errors = append(result.Errors, ImportError{
				Type:    "budget_save",
				Message: fmt.Sprintf("failed to save budget %s: %v", budget.ID, err),
			})
			result.Success = false
		} else {
			result.Imported++
		}
	}

	// Import retry profiles
	for _, profile := range bundle.RetryProfiles {
		if err := m.storage.SaveRetryProfile(ctx, &profile); err != nil {
			result.Errors = append(result.Errors, ImportError{
				Type:    "profile_save",
				Message: fmt.Sprintf("failed to save retry profile %s: %v", profile.Name, err),
			})
			result.Success = false
		} else {
			result.Imported++
		}
	}

	return result, nil
}

// importMerge imports configurations by merging with existing ones
func (m *Manager) importMerge(ctx context.Context, bundle *ConfigBundle) (*ImportResult, error) {
	result := &ImportResult{
		Success: true,
		Summary: m.createImportSummary(bundle),
	}

	// Import policies with merge logic
	for _, policy := range bundle.Policies {
		existing, err := m.storage.GetPolicy(ctx, policy.ID)
		if err != nil {
			// Policy doesn't exist, create new
			if err := m.storage.SavePolicy(ctx, &policy); err != nil {
				result.Errors = append(result.Errors, ImportError{
					Type:    "policy_save",
					Message: fmt.Sprintf("failed to save policy %s: %v", policy.ID, err),
				})
				result.Success = false
			} else {
				result.Imported++
			}
		} else {
			// Policy exists, update if newer
			if policy.UpdatedAt.After(existing.UpdatedAt) {
				if err := m.storage.SavePolicy(ctx, &policy); err != nil {
					result.Errors = append(result.Errors, ImportError{
						Type:    "policy_update",
						Message: fmt.Sprintf("failed to update policy %s: %v", policy.ID, err),
					})
					result.Success = false
				} else {
					result.Imported++
				}
			} else {
				result.Skipped++
			}
		}
	}

	// Import budgets with merge logic
	for _, budget := range bundle.Budgets {
		existing, err := m.storage.GetBudget(ctx, budget.ID)
		if err != nil {
			// Budget doesn't exist, create new
			if err := m.storage.SaveBudget(ctx, &budget); err != nil {
				result.Errors = append(result.Errors, ImportError{
					Type:    "budget_save",
					Message: fmt.Sprintf("failed to save budget %s: %v", budget.ID, err),
				})
				result.Success = false
			} else {
				result.Imported++
			}
		} else {
			// Budget exists, update if newer
			if budget.UpdatedAt.After(existing.UpdatedAt) {
				if err := m.storage.SaveBudget(ctx, &budget); err != nil {
					result.Errors = append(result.Errors, ImportError{
						Type:    "budget_update",
						Message: fmt.Sprintf("failed to update budget %s: %v", budget.ID, err),
					})
					result.Success = false
				} else {
					result.Imported++
				}
			} else {
				result.Skipped++
			}
		}
	}

	// Import retry profiles with merge logic
	for _, profile := range bundle.RetryProfiles {
		existing, err := m.storage.GetRetryProfile(ctx, profile.Name)
		if err != nil {
			// Profile doesn't exist, create new
			if err := m.storage.SaveRetryProfile(ctx, &profile); err != nil {
				result.Errors = append(result.Errors, ImportError{
					Type:    "profile_save",
					Message: fmt.Sprintf("failed to save retry profile %s: %v", profile.Name, err),
				})
				result.Success = false
			} else {
				result.Imported++
			}
		} else {
			// Profile exists, update if newer
			if profile.UpdatedAt.After(existing.UpdatedAt) {
				if err := m.storage.SaveRetryProfile(ctx, &profile); err != nil {
					result.Errors = append(result.Errors, ImportError{
						Type:    "profile_update",
						Message: fmt.Sprintf("failed to update retry profile %s: %v", profile.Name, err),
					})
					result.Success = false
				} else {
					result.Imported++
				}
			} else {
				result.Skipped++
			}
		}
	}

	return result, nil
}

// createImportSummary creates a summary of import results
func (m *Manager) createImportSummary(bundle *ConfigBundle) ImportSummary {
	return ImportSummary{
		Policies:      len(bundle.Policies),
		Budgets:       len(bundle.Budgets),
		DLQEntries:    len(bundle.DLQEntries),
		Workflows:     len(bundle.Workflows),
		RetryProfiles: len(bundle.RetryProfiles),
	}
}

// calculateChecksum calculates a SHA256 checksum for the configuration bundle
func (m *Manager) calculateChecksum(bundle *ConfigBundle) string {
	data, err := json.Marshal(bundle)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
