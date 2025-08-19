package configmanager

import (
	"context"
	"io"
)

// ConfigManager defines the interface for configuration export/import operations
type ConfigManager interface {
	// Export exports configurations in the specified format
	Export(ctx context.Context, opts ExportOptions) (*ConfigBundle, error)

	// ExportToWriter exports configurations to an io.Writer in the specified format
	ExportToWriter(ctx context.Context, opts ExportOptions, w io.Writer) error

	// Import imports configurations from the specified format
	Import(ctx context.Context, data []byte, opts ImportOptions) (*ImportResult, error)

	// ImportFromReader imports configurations from an io.Reader in the specified format
	ImportFromReader(ctx context.Context, r io.Reader, opts ImportOptions) (*ImportResult, error)

	// Validate validates configuration data without importing
	Validate(ctx context.Context, data []byte, format ConfigFormat) ([]ImportError, error)

	// GetSupportedFormats returns list of supported export/import formats
	GetSupportedFormats() []ConfigFormat

	// GetVersion returns the current configuration version
	GetVersion() string
}

// ConfigExporter defines the interface for format-specific exporters
type ConfigExporter interface {
	// Export exports a ConfigBundle to the specific format
	Export(bundle *ConfigBundle) ([]byte, error)

	// ExportToWriter exports a ConfigBundle to an io.Writer
	ExportToWriter(bundle *ConfigBundle, w io.Writer) error

	// GetFormat returns the format this exporter handles
	GetFormat() ConfigFormat
}

// ConfigImporter defines the interface for format-specific importers
type ConfigImporter interface {
	// Import imports configuration data to a ConfigBundle
	Import(data []byte) (*ConfigBundle, error)

	// ImportFromReader imports configuration data from an io.Reader
	ImportFromReader(r io.Reader) (*ConfigBundle, error)

	// GetFormat returns the format this importer handles
	GetFormat() ConfigFormat
}

// ConfigValidator defines the interface for configuration validation
type ConfigValidator interface {
	// Validate validates a ConfigBundle
	Validate(bundle *ConfigBundle) ([]ImportError, error)

	// ValidatePolicy validates a single policy configuration
	ValidatePolicy(policy *PolicyConfig) []ImportError

	// ValidateBudget validates a single budget configuration
	ValidateBudget(budget *BudgetConfig) []ImportError

	// ValidateRetryProfile validates a single retry profile configuration
	ValidateRetryProfile(profile *RetryProfileConfig) []ImportError
}

// ConfigStorage defines the interface for configuration persistence
type ConfigStorage interface {
	// SavePolicy saves a policy configuration
	SavePolicy(ctx context.Context, policy *PolicyConfig) error

	// SaveBudget saves a budget configuration
	SaveBudget(ctx context.Context, budget *BudgetConfig) error

	// SaveRetryProfile saves a retry profile configuration
	SaveRetryProfile(ctx context.Context, profile *RetryProfileConfig) error

	// GetPolicy retrieves a policy configuration by ID
	GetPolicy(ctx context.Context, id string) (*PolicyConfig, error)

	// GetBudget retrieves a budget configuration by ID
	GetBudget(ctx context.Context, id string) (*BudgetConfig, error)

	// GetRetryProfile retrieves a retry profile configuration by name
	GetRetryProfile(ctx context.Context, name string) (*RetryProfileConfig, error)

	// ListPolicies lists all policy configurations
	ListPolicies(ctx context.Context) ([]*PolicyConfig, error)

	// ListBudgets lists all budget configurations
	ListBudgets(ctx context.Context) ([]*BudgetConfig, error)

	// ListRetryProfiles lists all retry profile configurations
	ListRetryProfiles(ctx context.Context) ([]*RetryProfileConfig, error)

	// DeletePolicy deletes a policy configuration
	DeletePolicy(ctx context.Context, id string) error

	// DeleteBudget deletes a budget configuration
	DeleteBudget(ctx context.Context, id string) error

	// DeleteRetryProfile deletes a retry profile configuration
	DeleteRetryProfile(ctx context.Context, name string) error
}
