package configmanager

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mova-engine/mova-engine/core/rules"
)

// ExportRules exports rules to a configuration bundle
func (m *Manager) ExportRules(rules []rules.Rule) ([]RuleConfig, error) {
	var ruleConfigs []RuleConfig

	for _, rule := range rules {
		ruleConfig := RuleConfig{
			ID:          rule.ID,
			Name:        rule.Name,
			Description: rule.Description,
			Priority:    rule.Priority,
			Enabled:     rule.Enabled,
			CreatedAt:   rule.CreatedAt,
			UpdatedAt:   rule.UpdatedAt,
		}

		// Convert conditions
		for _, condition := range rule.Conditions {
			ruleConfig.Conditions = append(ruleConfig.Conditions, RuleConditionConfig{
				Field:    condition.Field,
				Operator: condition.Operator,
				Value:    condition.Value,
				Negate:   condition.Negate,
			})
		}

		// Convert actions
		for _, action := range rule.Actions {
			ruleConfig.Actions = append(ruleConfig.Actions, RuleActionConfig{
				Type:   action.Type,
				Params: action.Params,
			})
		}

		ruleConfigs = append(ruleConfigs, ruleConfig)
	}

	return ruleConfigs, nil
}

// ExportRuleSets exports rulesets to a configuration bundle
func (m *Manager) ExportRuleSets(rulesets []rules.RuleSet) ([]RuleSetConfig, error) {
	var rulesetConfigs []RuleSetConfig

	for _, ruleset := range rulesets {
		rulesetConfig := RuleSetConfig{
			Version:     ruleset.Version,
			Name:        ruleset.Name,
			Description: ruleset.Description,
			Metadata:    ruleset.Metadata,
			CreatedAt:   ruleset.CreatedAt,
			UpdatedAt:   ruleset.UpdatedAt,
		}

		// Export rules in the ruleset
		ruleConfigs, err := m.ExportRules(ruleset.Rules)
		if err != nil {
			return nil, fmt.Errorf("failed to export rules for ruleset %s: %w", ruleset.Name, err)
		}

		rulesetConfig.Rules = ruleConfigs
		rulesetConfigs = append(rulesetConfigs, rulesetConfig)
	}

	return rulesetConfigs, nil
}

// ImportRules imports rules from a configuration bundle
func (m *Manager) ImportRules(ruleConfigs []RuleConfig) ([]rules.Rule, []ImportError) {
	var importedRules []rules.Rule
	var errors []ImportError

	for i, ruleConfig := range ruleConfigs {
		rule := rules.Rule{
			ID:          ruleConfig.ID,
			Name:        ruleConfig.Name,
			Description: ruleConfig.Description,
			Priority:    ruleConfig.Priority,
			Enabled:     ruleConfig.Enabled,
			CreatedAt:   ruleConfig.CreatedAt,
			UpdatedAt:   ruleConfig.UpdatedAt,
		}

		// Convert conditions
		for _, conditionConfig := range ruleConfig.Conditions {
			rule.Conditions = append(rule.Conditions, rules.Condition{
				Field:    conditionConfig.Field,
				Operator: conditionConfig.Operator,
				Value:    conditionConfig.Value,
				Negate:   conditionConfig.Negate,
			})
		}

		// Convert actions
		for _, actionConfig := range ruleConfig.Actions {
			rule.Actions = append(rule.Actions, rules.Action{
				Type:   actionConfig.Type,
				Params: actionConfig.Params,
			})
		}

		// Validate rule
		engine := rules.NewEngine()
		if err := engine.ValidateRule(rule); err != nil {
			errors = append(errors, ImportError{
				Type:    "validation",
				Message: fmt.Sprintf("Rule validation failed: %s", err.Error()),
				Context: fmt.Sprintf("rule[%d]: %s", i, rule.ID),
			})
			continue
		}

		importedRules = append(importedRules, rule)
	}

	return importedRules, errors
}

// ImportRuleSets imports rulesets from a configuration bundle
func (m *Manager) ImportRuleSets(rulesetConfigs []RuleSetConfig) ([]rules.RuleSet, []ImportError) {
	var importedRuleSets []rules.RuleSet
	var errors []ImportError

	for i, rulesetConfig := range rulesetConfigs {
		ruleset := rules.RuleSet{
			Version:     rulesetConfig.Version,
			Name:        rulesetConfig.Name,
			Description: rulesetConfig.Description,
			Metadata:    rulesetConfig.Metadata,
			CreatedAt:   rulesetConfig.CreatedAt,
			UpdatedAt:   rulesetConfig.UpdatedAt,
		}

		// Import rules in the ruleset
		importedRules, ruleErrors := m.ImportRules(rulesetConfig.Rules)
		if len(ruleErrors) > 0 {
			for _, ruleError := range ruleErrors {
				ruleError.Context = fmt.Sprintf("ruleset[%d]: %s -> %s", i, ruleset.Name, ruleError.Context)
				errors = append(errors, ruleError)
			}
		}

		ruleset.Rules = importedRules

		// Validate ruleset
		engine := rules.NewEngine()
		if err := engine.ValidateRuleSet(ruleset); err != nil {
			errors = append(errors, ImportError{
				Type:    "validation",
				Message: fmt.Sprintf("RuleSet validation failed: %s", err.Error()),
				Context: fmt.Sprintf("ruleset[%d]: %s", i, ruleset.Name),
			})
			continue
		}

		importedRuleSets = append(importedRuleSets, ruleset)
	}

	return importedRuleSets, errors
}

// ExportRulesBundle creates a complete configuration bundle with rules
func (m *Manager) ExportRulesBundle(rules []rules.Rule, rulesets []rules.RuleSet, format ConfigFormat) ([]byte, error) {
	// Export rules and rulesets
	ruleConfigs, err := m.ExportRules(rules)
	if err != nil {
		return nil, fmt.Errorf("failed to export rules: %w", err)
	}

	rulesetConfigs, err := m.ExportRuleSets(rulesets)
	if err != nil {
		return nil, fmt.Errorf("failed to export rulesets: %w", err)
	}

	// Create configuration bundle
	bundle := ConfigBundle{
		Metadata: ConfigMetadata{
			Version:     "1.0.0",
			Format:      string(format),
			GeneratedAt: time.Now(),
			Source:      "MOVA Rule Engine",
		},
		Rules:    ruleConfigs,
		RuleSets: rulesetConfigs,
	}

	// Marshal based on format
	switch format {
	case FormatJSON:
		return json.MarshalIndent(bundle, "", "  ")
	case FormatYAML:
		if exporter, exists := m.exporters[FormatYAML]; exists {
			return exporter.Export(&bundle)
		}
		return nil, fmt.Errorf("YAML exporter not available")
	case FormatHCL:
		if exporter, exists := m.exporters[FormatHCL]; exists {
			return exporter.Export(&bundle)
		}
		return nil, fmt.Errorf("HCL exporter not available")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ImportRulesBundle imports rules and rulesets from a configuration bundle
func (m *Manager) ImportRulesBundle(data []byte, format ConfigFormat, options ImportOptions) (ImportResult, error) {
	var bundle ConfigBundle
	var err error

	// Unmarshal based on format
	switch format {
	case FormatJSON:
		err = json.Unmarshal(data, &bundle)
	case FormatYAML:
		if importer, exists := m.importers[FormatYAML]; exists {
			bundlePtr, importErr := importer.Import(data)
			if importErr != nil {
				err = importErr
			} else {
				bundle = *bundlePtr
			}
		} else {
			err = fmt.Errorf("YAML importer not available")
		}
	case FormatHCL:
		if importer, exists := m.importers[FormatHCL]; exists {
			bundlePtr, importErr := importer.Import(data)
			if importErr != nil {
				err = importErr
			} else {
				bundle = *bundlePtr
			}
		} else {
			err = fmt.Errorf("HCL importer not available")
		}
	default:
		return ImportResult{}, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return ImportResult{}, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	result := ImportResult{
		Success: true,
		Summary: ImportSummary{},
	}

	// Import rules
	if len(bundle.Rules) > 0 {
		importedRules, errors := m.ImportRules(bundle.Rules)
		result.Imported += len(importedRules)
		result.Summary.Rules = len(importedRules)
		result.Errors = append(result.Errors, errors...)

		if len(errors) > 0 {
			result.Success = false
		}

		// If not in dry run mode, save the rules
		if !options.DryRun && !options.ValidateOnly {
			// TODO: Save rules to repository
			// This would require access to a rule repository
		}
	}

	// Import rulesets
	if len(bundle.RuleSets) > 0 {
		importedRuleSets, errors := m.ImportRuleSets(bundle.RuleSets)
		result.Imported += len(importedRuleSets)
		result.Summary.RuleSets = len(importedRuleSets)
		result.Errors = append(result.Errors, errors...)

		if len(errors) > 0 {
			result.Success = false
		}

		// If not in dry run mode, save the rulesets
		if !options.DryRun && !options.ValidateOnly {
			// TODO: Save rulesets to repository
			// This would require access to a rule repository
		}
	}

	return result, nil
}

// ValidateRulesConfiguration validates rules configuration without importing
func (m *Manager) ValidateRulesConfiguration(data []byte, format ConfigFormat) ([]ImportError, []ImportWarning) {
	var bundle ConfigBundle
	var err error
	var errors []ImportError
	var warnings []ImportWarning

	// Unmarshal based on format
	switch format {
	case FormatJSON:
		err = json.Unmarshal(data, &bundle)
	case FormatYAML:
		if importer, exists := m.importers[FormatYAML]; exists {
			bundlePtr, importErr := importer.Import(data)
			if importErr != nil {
				err = importErr
			} else {
				bundle = *bundlePtr
			}
		} else {
			err = fmt.Errorf("YAML importer not available")
		}
	case FormatHCL:
		if importer, exists := m.importers[FormatHCL]; exists {
			bundlePtr, importErr := importer.Import(data)
			if importErr != nil {
				err = importErr
			} else {
				bundle = *bundlePtr
			}
		} else {
			err = fmt.Errorf("HCL importer not available")
		}
	default:
		errors = append(errors, ImportError{
			Type:    "format",
			Message: fmt.Sprintf("unsupported format: %s", format),
		})
		return errors, warnings
	}

	if err != nil {
		errors = append(errors, ImportError{
			Type:    "parse",
			Message: fmt.Sprintf("failed to parse configuration: %s", err.Error()),
		})
		return errors, warnings
	}

	// Validate rules
	if len(bundle.Rules) > 0 {
		_, ruleErrors := m.ImportRules(bundle.Rules)
		errors = append(errors, ruleErrors...)

		if len(bundle.Rules) > 100 {
			warnings = append(warnings, ImportWarning{
				Type:    "performance",
				Message: fmt.Sprintf("Large number of rules (%d) may impact performance", len(bundle.Rules)),
			})
		}
	}

	// Validate rulesets
	if len(bundle.RuleSets) > 0 {
		_, rulesetErrors := m.ImportRuleSets(bundle.RuleSets)
		errors = append(errors, rulesetErrors...)

		// Check for duplicate ruleset names
		rulesetNames := make(map[string]bool)
		for i, ruleset := range bundle.RuleSets {
			if rulesetNames[ruleset.Name] {
				errors = append(errors, ImportError{
					Type:    "duplicate",
					Message: fmt.Sprintf("Duplicate ruleset name: %s", ruleset.Name),
					Context: fmt.Sprintf("ruleset[%d]", i),
				})
			}
			rulesetNames[ruleset.Name] = true
		}
	}

	return errors, warnings
}
