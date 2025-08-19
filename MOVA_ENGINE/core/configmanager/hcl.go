package configmanager

import (
	"fmt"
	"io"
	"strings"
)

// HCLExporter implements ConfigExporter for HCL format
type HCLExporter struct{}

// Export exports a ConfigBundle to HCL format
func (e *HCLExporter) Export(bundle *ConfigBundle) ([]byte, error) {
	var builder strings.Builder
	if err := e.writeHCL(&builder, bundle); err != nil {
		return nil, err
	}
	return []byte(builder.String()), nil
}

// ExportToWriter exports a ConfigBundle to HCL format to an io.Writer
func (e *HCLExporter) ExportToWriter(bundle *ConfigBundle, w io.Writer) error {
	return e.writeHCL(w, bundle)
}

// writeHCL writes the configuration bundle in HCL format
func (e *HCLExporter) writeHCL(w io.Writer, bundle *ConfigBundle) error {
	// Write metadata
	if _, err := fmt.Fprintf(w, "# MOVA Configuration Export\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "# Generated at: %s\n", bundle.Metadata.GeneratedAt.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "# Version: %s\n", bundle.Metadata.Version); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "# Source: %s\n\n", bundle.Metadata.Source); err != nil {
		return err
	}

	// Write retry profiles
	if len(bundle.RetryProfiles) > 0 {
		if _, err := fmt.Fprintf(w, "# Retry Profiles\n"); err != nil {
			return err
		}
		for _, profile := range bundle.RetryProfiles {
			if err := e.writeRetryProfileHCL(w, profile); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "\n"); err != nil {
			return err
		}
	}

	// Write budgets
	if len(bundle.Budgets) > 0 {
		if _, err := fmt.Fprintf(w, "# Budget Constraints\n"); err != nil {
			return err
		}
		for _, budget := range bundle.Budgets {
			if err := e.writeBudgetHCL(w, budget); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "\n"); err != nil {
			return err
		}
	}

	// Write policies
	if len(bundle.Policies) > 0 {
		if _, err := fmt.Fprintf(w, "# Policies\n"); err != nil {
			return err
		}
		for _, policy := range bundle.Policies {
			if err := e.writePolicyHCL(w, policy); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "\n"); err != nil {
			return err
		}
	}

	return nil
}

// writeRetryProfileHCL writes a retry profile in HCL format
func (e *HCLExporter) writeRetryProfileHCL(w io.Writer, profile RetryProfileConfig) error {
	if _, err := fmt.Fprintf(w, "retry_profile \"%s\" {\n", profile.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  description = \"%s\"\n", profile.Description); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  max_retries = %d\n", profile.MaxRetries); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  initial_delay = \"%s\"\n", profile.InitialDelay); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  max_delay = \"%s\"\n", profile.MaxDelay); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  backoff_multiplier = %.2f\n", profile.BackoffMultiplier); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  jitter = %.2f\n", profile.Jitter); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  timeout = \"%s\"\n", profile.Timeout); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "}\n\n"); err != nil {
		return err
	}
	return nil
}

// writeBudgetHCL writes a budget configuration in HCL format
func (e *HCLExporter) writeBudgetHCL(w io.Writer, budget BudgetConfig) error {
	if _, err := fmt.Fprintf(w, "budget \"%s\" {\n", budget.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  name = \"%s\"\n", budget.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  description = \"%s\"\n", budget.Description); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  type = \"%s\"\n", budget.Type); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  scope = \"%s\"\n", budget.Scope); err != nil {
		return err
	}
	if budget.MaxCount > 0 {
		if _, err := fmt.Fprintf(w, "  max_count = %d\n", budget.MaxCount); err != nil {
			return err
		}
	}
	if budget.MaxMemory > 0 {
		if _, err := fmt.Fprintf(w, "  max_memory = %d\n", budget.MaxMemory); err != nil {
			return err
		}
	}
	if budget.MaxCPU > 0 {
		if _, err := fmt.Fprintf(w, "  max_cpu = %.2f\n", budget.MaxCPU); err != nil {
			return err
		}
	}
	if budget.MaxDuration > 0 {
		if _, err := fmt.Fprintf(w, "  max_duration = \"%s\"\n", budget.MaxDuration); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "  time_window = \"%s\"\n", budget.TimeWindow); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  enabled = %t\n", budget.Enabled); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "}\n\n"); err != nil {
		return err
	}
	return nil
}

// writePolicyHCL writes a policy configuration in HCL format
func (e *HCLExporter) writePolicyHCL(w io.Writer, policy PolicyConfig) error {
	if _, err := fmt.Fprintf(w, "policy \"%s\" {\n", policy.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  name = \"%s\"\n", policy.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  description = \"%s\"\n", policy.Description); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  retry_profile = \"%s\"\n", policy.RetryProfile); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  enabled = %t\n", policy.Enabled); err != nil {
		return err
	}

	// Write conditions if any
	if len(policy.Conditions) > 0 {
		if _, err := fmt.Fprintf(w, "  conditions {\n"); err != nil {
			return err
		}
		for _, condition := range policy.Conditions {
			if err := e.writeConditionHCL(w, condition); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "  }\n"); err != nil {
			return err
		}
	}

	// Write budget constraints if any
	if policy.BudgetConstraints.MaxRetriesPerWorkflow > 0 ||
		policy.BudgetConstraints.MaxRetriesPerSession > 0 ||
		policy.BudgetConstraints.MaxTotalRetryTime > 0 {
		if _, err := fmt.Fprintf(w, "  budget_constraints {\n"); err != nil {
			return err
		}
		if policy.BudgetConstraints.MaxRetriesPerWorkflow > 0 {
			if _, err := fmt.Fprintf(w, "    max_retries_per_workflow = %d\n", policy.BudgetConstraints.MaxRetriesPerWorkflow); err != nil {
				return err
			}
		}
		if policy.BudgetConstraints.MaxRetriesPerSession > 0 {
			if _, err := fmt.Fprintf(w, "    max_retries_per_session = %d\n", policy.BudgetConstraints.MaxRetriesPerSession); err != nil {
				return err
			}
		}
		if policy.BudgetConstraints.MaxTotalRetryTime > 0 {
			if _, err := fmt.Fprintf(w, "    max_total_retry_time = \"%s\"\n", policy.BudgetConstraints.MaxTotalRetryTime); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "  }\n"); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "}\n\n"); err != nil {
		return err
	}
	return nil
}

// writeConditionHCL writes a condition configuration in HCL format
func (e *HCLExporter) writeConditionHCL(w io.Writer, condition ConditionConfig) error {
	if _, err := fmt.Fprintf(w, "    condition {\n"); err != nil {
		return err
	}
	if condition.ErrorType != "" {
		if _, err := fmt.Fprintf(w, "      error_type = \"%s\"\n", condition.ErrorType); err != nil {
			return err
		}
	}
	if condition.HTTPStatus > 0 {
		if _, err := fmt.Fprintf(w, "      http_status = %d\n", condition.HTTPStatus); err != nil {
			return err
		}
	}
	if condition.ErrorMessagePattern != "" {
		if _, err := fmt.Fprintf(w, "      error_message_pattern = \"%s\"\n", condition.ErrorMessagePattern); err != nil {
			return err
		}
	}
	if condition.ActionType != "" {
		if _, err := fmt.Fprintf(w, "      action_type = \"%s\"\n", condition.ActionType); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "    }\n"); err != nil {
		return err
	}
	return nil
}

// GetFormat returns the format this exporter handles
func (e *HCLExporter) GetFormat() ConfigFormat {
	return FormatHCL
}

// HCLImporter implements ConfigImporter for HCL format
type HCLImporter struct{}

// Import imports configuration data from HCL format to a ConfigBundle
func (i *HCLImporter) Import(data []byte) (*ConfigBundle, error) {
	// HCL parsing is complex and would require a proper HCL parser
	// For now, we'll return an error indicating HCL import is not yet implemented
	return nil, fmt.Errorf("HCL import is not yet implemented - please use JSON or YAML format")
}

// ImportFromReader imports configuration data from HCL format from an io.Reader
func (i *HCLImporter) ImportFromReader(r io.Reader) (*ConfigBundle, error) {
	// HCL parsing is complex and would require a proper HCL parser
	// For now, we'll return an error indicating HCL import is not yet implemented
	return nil, fmt.Errorf("HCL import is not yet implemented - please use JSON or YAML format")
}

// GetFormat returns the format this importer handles
func (i *HCLImporter) GetFormat() ConfigFormat {
	return FormatHCL
}
