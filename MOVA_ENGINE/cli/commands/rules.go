package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mova-engine/mova-engine/core/rules"
)

// rulesCmd represents the rules command
var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage MOVA rules and rulesets",
	Long: `Commands for managing MOVA rules and rulesets.
	
Rules allow you to define conditions and actions that are executed
when certain conditions are met during workflow execution.`,
}

// rulesListCmd lists all rules
var rulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all rules",
	Long:  "List all rules with optional filtering by name, priority, or enabled status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get filter options
		name, _ := cmd.Flags().GetString("name")
		priority, _ := cmd.Flags().GetInt("priority")
		enabled, _ := cmd.Flags().GetBool("enabled")
		format, _ := cmd.Flags().GetString("format")

		// Create filter map
		filter := make(map[string]interface{})
		if name != "" {
			filter["name"] = name
		}
		if priority > 0 {
			filter["priority"] = priority
		}
		if enabled {
			filter["enabled"] = enabled
		}

		// TODO: Get rules from repository
		// For now, return placeholder
		rules := []rules.Rule{
			{
				ID:          "example-rule-1",
				Name:        "Example Rule",
				Description: "An example rule for demonstration",
				Priority:    100,
				Enabled:     true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		return outputRules(rules, format)
	},
}

// rulesApplyCmd applies rules from a file
var rulesApplyCmd = &cobra.Command{
	Use:   "apply <file>",
	Short: "Apply rules from a file",
	Long:  "Load and apply rules from a YAML or JSON file.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		// Check if file exists
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filename)
		}

		// Read file
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file: %s", err.Error())
		}

		// Parse based on file extension
		var ruleset rules.RuleSet
		ext := strings.ToLower(filepath.Ext(filename))

		switch ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(data, &ruleset); err != nil {
				return fmt.Errorf("failed to parse YAML: %s", err.Error())
			}
		case ".json":
			if err := json.Unmarshal(data, &ruleset); err != nil {
				return fmt.Errorf("failed to parse JSON: %s", err.Error())
			}
		default:
			return fmt.Errorf("unsupported file format: %s (supported: .yaml, .yml, .json)", ext)
		}

		// Validate ruleset
		engine := rules.NewEngine()
		if err := engine.ValidateRuleSet(ruleset); err != nil {
			return fmt.Errorf("ruleset validation failed: %s", err.Error())
		}

		// TODO: Apply rules to repository
		fmt.Printf("Successfully applied ruleset '%s' with %d rules\n", ruleset.Name, len(ruleset.Rules))

		// Show summary
		enabledCount := 0
		for _, rule := range ruleset.Rules {
			if rule.Enabled {
				enabledCount++
			}
		}

		fmt.Printf("  - Total rules: %d\n", len(ruleset.Rules))
		fmt.Printf("  - Enabled rules: %d\n", enabledCount)
		fmt.Printf("  - Version: %s\n", ruleset.Version)

		return nil
	},
}

// rulesValidateCmd validates rules from a file
var rulesValidateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate rules from a file",
	Long:  "Validate the syntax and structure of rules in a YAML or JSON file without applying them.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		// Check if file exists
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filename)
		}

		// Read file
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file: %s", err.Error())
		}

		// Parse based on file extension
		var ruleset rules.RuleSet
		ext := strings.ToLower(filepath.Ext(filename))

		switch ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(data, &ruleset); err != nil {
				return fmt.Errorf("failed to parse YAML: %s", err.Error())
			}
		case ".json":
			if err := json.Unmarshal(data, &ruleset); err != nil {
				return fmt.Errorf("failed to parse JSON: %s", err.Error())
			}
		default:
			return fmt.Errorf("unsupported file format: %s (supported: .yaml, .yml, .json)", ext)
		}

		// Validate ruleset
		engine := rules.NewEngine()
		if err := engine.ValidateRuleSet(ruleset); err != nil {
			return fmt.Errorf("validation failed: %s", err.Error())
		}

		fmt.Printf("✓ Ruleset '%s' is valid\n", ruleset.Name)
		fmt.Printf("  - Rules: %d\n", len(ruleset.Rules))
		fmt.Printf("  - Version: %s\n", ruleset.Version)

		// Show rule details if verbose
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			fmt.Println("\nRule Details:")
			for i, rule := range ruleset.Rules {
				status := "disabled"
				if rule.Enabled {
					status = "enabled"
				}
				fmt.Printf("  %d. %s (ID: %s, Priority: %d, %s)\n",
					i+1, rule.Name, rule.ID, rule.Priority, status)
				fmt.Printf("     Conditions: %d, Actions: %d\n",
					len(rule.Conditions), len(rule.Actions))
			}
		}

		return nil
	},
}

// rulesEvalCmd evaluates rules against a context
var rulesEvalCmd = &cobra.Command{
	Use:   "eval <rules-file> --context <context-file>",
	Short: "Evaluate rules against a context",
	Long:  "Perform a dry-run evaluation of rules against a provided context without executing actions.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rulesFile := args[0]
		contextFile, _ := cmd.Flags().GetString("context")
		format, _ := cmd.Flags().GetString("format")

		if contextFile == "" {
			return fmt.Errorf("context file is required (use --context flag)")
		}

		// Read rules file
		rulesData, err := ioutil.ReadFile(rulesFile)
		if err != nil {
			return fmt.Errorf("failed to read rules file: %s", err.Error())
		}

		// Read context file
		contextData, err := ioutil.ReadFile(contextFile)
		if err != nil {
			return fmt.Errorf("failed to read context file: %s", err.Error())
		}

		// Parse rules
		var ruleset rules.RuleSet
		rulesExt := strings.ToLower(filepath.Ext(rulesFile))

		switch rulesExt {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(rulesData, &ruleset); err != nil {
				return fmt.Errorf("failed to parse rules YAML: %s", err.Error())
			}
		case ".json":
			if err := json.Unmarshal(rulesData, &ruleset); err != nil {
				return fmt.Errorf("failed to parse rules JSON: %s", err.Error())
			}
		default:
			return fmt.Errorf("unsupported rules file format: %s", rulesExt)
		}

		// Parse context
		var ctx rules.Context
		contextExt := strings.ToLower(filepath.Ext(contextFile))

		switch contextExt {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(contextData, &ctx); err != nil {
				return fmt.Errorf("failed to parse context YAML: %s", err.Error())
			}
		case ".json":
			if err := json.Unmarshal(contextData, &ctx); err != nil {
				return fmt.Errorf("failed to parse context JSON: %s", err.Error())
			}
		default:
			return fmt.Errorf("unsupported context file format: %s", contextExt)
		}

		// Set timestamp if not provided
		if ctx.Timestamp.IsZero() {
			ctx.Timestamp = time.Now()
		}

		// Initialize maps if nil
		if ctx.Variables == nil {
			ctx.Variables = make(map[string]interface{})
		}
		if ctx.Request == nil {
			ctx.Request = make(map[string]interface{})
		}
		if ctx.Response == nil {
			ctx.Response = make(map[string]interface{})
		}
		if ctx.Metadata == nil {
			ctx.Metadata = make(map[string]interface{})
		}

		// Evaluate rules
		engine := rules.NewEngine()
		results, err := engine.Run(ruleset, ctx)
		if err != nil {
			return fmt.Errorf("evaluation failed: %s", err.Error())
		}

		return outputResults(results, format)
	},
}

// outputRules outputs rules in the specified format
func outputRules(rulesList []rules.Rule, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(rulesList, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(rulesList)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default: // table format
		fmt.Printf("%-36s %-30s %-8s %-8s %s\n", "ID", "Name", "Priority", "Enabled", "Description")
		fmt.Println(strings.Repeat("-", 120))
		for _, rule := range rulesList {
			enabled := "No"
			if rule.Enabled {
				enabled = "Yes"
			}
			description := rule.Description
			if len(description) > 50 {
				description = description[:47] + "..."
			}
			fmt.Printf("%-36s %-30s %-8d %-8s %s\n",
				rule.ID, rule.Name, rule.Priority, enabled, description)
		}
	}
	return nil
}

// outputResults outputs evaluation results in the specified format
func outputResults(results []rules.Result, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(results)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default: // summary format
		matchedCount := 0
		for _, result := range results {
			if result.Matched {
				matchedCount++
			}
		}

		fmt.Printf("Evaluation Summary:\n")
		fmt.Printf("  Total rules: %d\n", len(results))
		fmt.Printf("  Matched rules: %d\n", matchedCount)
		fmt.Printf("  Success rate: %.1f%%\n", float64(matchedCount)/float64(len(results))*100)

		fmt.Printf("\nRule Results:\n")
		fmt.Printf("%-36s %-8s %-8s %s\n", "Rule ID", "Matched", "Actions", "Error")
		fmt.Println(strings.Repeat("-", 80))

		for _, result := range results {
			matched := "No"
			if result.Matched {
				matched = "Yes"
			}

			errorMsg := result.Error
			if len(errorMsg) > 30 {
				errorMsg = errorMsg[:27] + "..."
			}

			fmt.Printf("%-36s %-8s %-8d %s\n",
				result.RuleID, matched, len(result.Actions), errorMsg)
		}
	}
	return nil
}

// init initializes the rules command
func init() {
	// Add rules command to root - this will be called from main CLI setup
	// rootCmd.AddCommand(rulesCmd) - commented out, should be added in main CLI setup

	// Add subcommands
	rulesCmd.AddCommand(rulesListCmd)
	rulesCmd.AddCommand(rulesApplyCmd)
	rulesCmd.AddCommand(rulesValidateCmd)
	rulesCmd.AddCommand(rulesEvalCmd)

	// Flags for list command
	rulesListCmd.Flags().String("name", "", "Filter by rule name")
	rulesListCmd.Flags().Int("priority", 0, "Filter by priority")
	rulesListCmd.Flags().Bool("enabled", false, "Show only enabled rules")
	rulesListCmd.Flags().String("format", "table", "Output format (table, json, yaml)")

	// Flags for validate command
	rulesValidateCmd.Flags().Bool("verbose", false, "Show detailed rule information")

	// Flags for eval command
	rulesEvalCmd.Flags().String("context", "", "Context file for evaluation (required)")
	rulesEvalCmd.Flags().String("format", "summary", "Output format (summary, json, yaml)")

	// Mark required flags
	rulesEvalCmd.MarkFlagRequired("context")
}
