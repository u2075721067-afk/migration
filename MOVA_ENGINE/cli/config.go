package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mova-engine/mova-engine/core/configmanager"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage MOVA configurations",
	Long: `Manage MOVA configurations including export, import, and validation.

Examples:
  mova config export --format=yaml --out=policies.yaml
  mova config import policies.yaml --merge
  mova config validate policies.yaml --format=yaml`,
}

// configExportCmd represents the config export command
var configExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export configurations to file",
	Long: `Export MOVA configurations to a file in the specified format.

Supported formats: json, yaml, hcl
Default format: yaml

Examples:
  mova config export --format=yaml --out=policies.yaml
  mova config export --format=json --out=config.json --include-dlq
  mova config export --format=hcl --out=terraform.hcl --include-workflows`,
	RunE: runConfigExport,
}

// configImportCmd represents the config import command
var configImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import configurations from file",
	Long: `Import MOVA configurations from a file in the specified format.

Supported formats: json, yaml, hcl
Import modes: merge, overwrite, validate

Examples:
  mova config import policies.yaml --merge
  mova config import config.json --overwrite
  mova config import policies.yaml --validate-only
  mova config import policies.yaml --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigImport,
}

// configValidateCmd represents the config validate command
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long: `Validate a configuration file without importing it.

Supported formats: json, yaml, hcl

Examples:
  mova config validate policies.yaml --format=yaml
  mova config validate config.json --format=json`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigValidate,
}

// configInfoCmd represents the config info command
var configInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show configuration system information",
	Long: `Show information about the configuration system including supported formats and features.

Examples:
  mova config info`,
	RunE: runConfigInfo,
}

// configExportFlags holds flags for config export command
type configExportFlags struct {
	format           string
	output           string
	includeDLQ       bool
	includeWorkflows bool
	compress         bool
}

// configImportFlags holds flags for config import command
type configImportFlags struct {
	format       string
	mode         string
	validateOnly bool
	dryRun       bool
	overwrite    bool
}

// configValidateFlags holds flags for config validate command
type configValidateFlags struct {
	format string
}

var (
	exportFlags   = &configExportFlags{}
	importFlags   = &configImportFlags{}
	validateFlags = &configValidateFlags{}
)

func init() {
	// Add config command to root
	rootCmd.AddCommand(configCmd)

	// Add subcommands to config
	configCmd.AddCommand(configExportCmd)
	configCmd.AddCommand(configImportCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configInfoCmd)

	// Export flags
	configExportCmd.Flags().StringVarP(&exportFlags.format, "format", "f", "yaml", "Export format (json, yaml, hcl)")
	configExportCmd.Flags().StringVarP(&exportFlags.output, "out", "o", "", "Output file path (required)")
	configExportCmd.Flags().BoolVar(&exportFlags.includeDLQ, "include-dlq", false, "Include DLQ entries in export")
	configExportCmd.Flags().BoolVar(&exportFlags.includeWorkflows, "include-workflows", false, "Include workflows in export")
	configExportCmd.Flags().BoolVar(&exportFlags.compress, "compress", false, "Compress output (if supported)")

	// Import flags
	configImportCmd.Flags().StringVarP(&importFlags.format, "format", "f", "", "Import format (json, yaml, hcl) - auto-detected if not specified")
	configImportCmd.Flags().StringVarP(&importFlags.mode, "mode", "m", "merge", "Import mode (merge, overwrite, validate)")
	configImportCmd.Flags().BoolVar(&importFlags.validateOnly, "validate-only", false, "Only validate, don't import")
	configImportCmd.Flags().BoolVar(&importFlags.dryRun, "dry-run", false, "Show what would be imported without actually importing")
	configImportCmd.Flags().BoolVar(&importFlags.overwrite, "overwrite", false, "Overwrite existing configurations")

	// Validate flags
	configValidateCmd.Flags().StringVarP(&validateFlags.format, "format", "f", "", "File format (json, yaml, hcl) - auto-detected if not specified")

	// Mark required flags
	configExportCmd.MarkFlagRequired("out")
}

// runConfigExport executes the config export command
func runConfigExport(cmd *cobra.Command, args []string) error {
	// Validate format
	format := configmanager.ConfigFormat(exportFlags.format)
	if format != configmanager.FormatJSON &&
		format != configmanager.FormatYAML &&
		format != configmanager.FormatHCL {
		return fmt.Errorf("unsupported format: %s. supported formats: json, yaml, hcl", exportFlags.format)
	}

	// Create output file
	outputFile, err := os.Create(exportFlags.output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create config manager (this would need to be initialized with actual storage and validator)
	// For now, we'll create a mock implementation
	configManager := createMockConfigManager()

	// Create export options
	opts := configmanager.ExportOptions{
		Format:           format,
		IncludeDLQ:       exportFlags.includeDLQ,
		IncludeWorkflows: exportFlags.includeWorkflows,
		Compress:         exportFlags.compress,
	}

	// Export configuration
	ctx := context.Background()
	if err := configManager.ExportToWriter(ctx, opts, outputFile); err != nil {
		return fmt.Errorf("failed to export configuration: %w", err)
	}

	fmt.Printf("Configuration exported successfully to %s\n", exportFlags.output)
	return nil
}

// runConfigImport executes the config import command
func runConfigImport(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Check if input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// Auto-detect format if not specified
	format := importFlags.format
	if format == "" {
		ext := filepath.Ext(inputFile)
		switch strings.ToLower(ext) {
		case ".json":
			format = "json"
		case ".yaml", ".yml":
			format = "yaml"
		case ".hcl":
			format = "hcl"
		default:
			return fmt.Errorf("cannot auto-detect format from file extension: %s. please specify --format", ext)
		}
	}

	// Validate format
	configFormat := configmanager.ConfigFormat(format)
	if configFormat != configmanager.FormatJSON &&
		configFormat != configmanager.FormatYAML &&
		configFormat != configmanager.FormatHCL {
		return fmt.Errorf("unsupported format: %s. supported formats: json, yaml, hcl", format)
	}

	// Validate mode
	mode := configmanager.ImportMode(importFlags.mode)
	if mode != configmanager.ModeOverwrite &&
		mode != configmanager.ModeMerge &&
		mode != configmanager.ModeValidate {
		return fmt.Errorf("invalid import mode: %s. valid modes: merge, overwrite, validate", importFlags.mode)
	}

	// Read input file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Create config manager
	configManager := createMockConfigManager()

	// Create import options
	opts := configmanager.ImportOptions{
		Format:       configFormat,
		Mode:         mode,
		ValidateOnly: importFlags.validateOnly,
		DryRun:       importFlags.dryRun,
		Overwrite:    importFlags.overwrite,
	}

	// Import configuration
	ctx := context.Background()
	result, err := configManager.Import(ctx, inputData, opts)
	if err != nil {
		return fmt.Errorf("failed to import configuration: %w", err)
	}

	// Display results
	fmt.Printf("Import completed: %s\n", inputFile)
	fmt.Printf("Success: %t\n", result.Success)
	fmt.Printf("Imported: %d\n", result.Imported)
	fmt.Printf("Skipped: %d\n", result.Skipped)

	if len(result.Errors) > 0 {
		fmt.Printf("Errors: %d\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %s: %s\n", err.Type, err.Message)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("Warnings: %d\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s: %s\n", warning.Type, warning.Message)
		}
	}

	if !result.Success {
		return fmt.Errorf("import failed with %d errors", len(result.Errors))
	}

	return nil
}

// runConfigValidate executes the config validate command
func runConfigValidate(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Check if input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// Auto-detect format if not specified
	format := validateFlags.format
	if format == "" {
		ext := filepath.Ext(inputFile)
		switch strings.ToLower(ext) {
		case ".json":
			format = "json"
		case ".yaml", ".yml":
			format = "yaml"
		case ".hcl":
			format = "hcl"
		default:
			return fmt.Errorf("cannot auto-detect format from file extension: %s. please specify --format", ext)
		}
	}

	// Validate format
	configFormat := configmanager.ConfigFormat(format)
	if configFormat != configmanager.FormatJSON &&
		configFormat != configmanager.FormatYAML &&
		configFormat != configmanager.FormatHCL {
		return fmt.Errorf("unsupported format: %s. supported formats: json, yaml, hcl", format)
	}

	// Read input file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Create config manager
	configManager := createMockConfigManager()

	// Validate configuration
	ctx := context.Background()
	errors, err := configManager.Validate(ctx, inputData, configFormat)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Display results
	if len(errors) == 0 {
		fmt.Printf("✓ Configuration file %s is valid\n", inputFile)
		return nil
	}

	fmt.Printf("✗ Configuration file %s has %d validation errors:\n", inputFile, len(errors))
	for i, err := range errors {
		fmt.Printf("  %d. %s: %s", i+1, err.Type, err.Message)
		if err.Context != "" {
			fmt.Printf(" (context: %s)", err.Context)
		}
		if err.Line > 0 {
			fmt.Printf(" (line: %d", err.Line)
			if err.Column > 0 {
				fmt.Printf(", column: %d", err.Column)
			}
			fmt.Printf(")")
		}
		fmt.Println()
	}

	return fmt.Errorf("validation failed with %d errors", len(errors))
}

// runConfigInfo executes the config info command
func runConfigInfo(cmd *cobra.Command, args []string) error {
	// Create config manager
	configManager := createMockConfigManager()

	// Get information
	version := configManager.GetVersion()
	supportedFormats := configManager.GetSupportedFormats()

	// Display information
	fmt.Println("MOVA Configuration System")
	fmt.Println("=========================")
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Supported formats: %v\n", supportedFormats)
	fmt.Printf("Default format: yaml\n")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  - Export policies, budgets, retry profiles")
	fmt.Println("  - Export DLQ entries (sandbox mode)")
	fmt.Println("  - Export workflows")
	fmt.Println("  - Import with merge/overwrite modes")
	fmt.Println("  - Validation with detailed error reporting")
	fmt.Println("  - Dry-run import mode")
	fmt.Println("  - HCL export (Terraform-compatible)")

	return nil
}

// createMockConfigManager creates a mock config manager for CLI testing
// In a real implementation, this would be replaced with actual storage and validator
func createMockConfigManager() configmanager.ConfigManager {
	// This is a placeholder - in real implementation, you would:
	// 1. Initialize actual storage (database, file system, etc.)
	// 2. Initialize actual validator
	// 3. Return real ConfigManager instance
	return nil
}
