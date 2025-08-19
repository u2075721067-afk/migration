package commands

import (
	"strings"
	"testing"
)

func TestPoliciesCmd(t *testing.T) {
	if PoliciesCmd == nil {
		t.Fatal("PoliciesCmd should not be nil")
	}

	if PoliciesCmd.Use != "policies" {
		t.Errorf("Expected Use to be 'policies', got '%s'", PoliciesCmd.Use)
	}

	if PoliciesCmd.Short != "Manage retry policies" {
		t.Errorf("Expected Short to be 'Manage retry policies', got '%s'", PoliciesCmd.Short)
	}

	// Check that we have some subcommands
	if len(PoliciesCmd.Commands()) == 0 {
		t.Error("Expected PoliciesCmd to have subcommands")
	}
}

func TestListPoliciesCmd(t *testing.T) {
	if ListPoliciesCmd == nil {
		t.Fatal("ListPoliciesCmd should not be nil")
	}

	if ListPoliciesCmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got '%s'", ListPoliciesCmd.Use)
	}

	if ListPoliciesCmd.Short != "List all retry policies" {
		t.Errorf("Expected Short description, got '%s'", ListPoliciesCmd.Short)
	}

	if ListPoliciesCmd.Run == nil {
		t.Error("Expected Run function to be set")
	}
}

func TestApplyPolicyCmd(t *testing.T) {
	if ApplyPolicyCmd == nil {
		t.Fatal("ApplyPolicyCmd should not be nil")
	}

	if ApplyPolicyCmd.Use != "apply [policy-file]" {
		t.Errorf("Expected Use to be 'apply [policy-file]', got '%s'", ApplyPolicyCmd.Use)
	}

	if ApplyPolicyCmd.Args == nil {
		t.Error("Expected Args to be set for exact args validation")
	}

	if ApplyPolicyCmd.Run == nil {
		t.Error("Expected Run function to be set")
	}
}

func TestExportPolicyCmd(t *testing.T) {
	if ExportPolicyCmd == nil {
		t.Fatal("ExportPolicyCmd should not be nil")
	}

	if ExportPolicyCmd.Use != "export [policy-id] [output-file]" {
		t.Errorf("Expected Use to be 'export [policy-id] [output-file]', got '%s'", ExportPolicyCmd.Use)
	}

	if ExportPolicyCmd.Args == nil {
		t.Error("Expected Args to be set for exact args validation")
	}

	if ExportPolicyCmd.Run == nil {
		t.Error("Expected Run function to be set")
	}
}

func TestDeletePolicyCmd(t *testing.T) {
	if DeletePolicyCmd == nil {
		t.Fatal("DeletePolicyCmd should not be nil")
	}

	if DeletePolicyCmd.Use != "delete [policy-id]" {
		t.Errorf("Expected Use to be 'delete [policy-id]', got '%s'", DeletePolicyCmd.Use)
	}

	if DeletePolicyCmd.Args == nil {
		t.Error("Expected Args to be set for exact args validation")
	}

	if DeletePolicyCmd.Run == nil {
		t.Error("Expected Run function to be set")
	}
}

func TestShowProfilesCmd(t *testing.T) {
	if ShowProfilesCmd == nil {
		t.Fatal("ShowProfilesCmd should not be nil")
	}

	if ShowProfilesCmd.Use != "profiles" {
		t.Errorf("Expected Use to be 'profiles', got '%s'", ShowProfilesCmd.Use)
	}

	if ShowProfilesCmd.Short != "Show available retry profiles" {
		t.Errorf("Expected Short description, got '%s'", ShowProfilesCmd.Short)
	}

	if ShowProfilesCmd.Run == nil {
		t.Error("Expected Run function to be set")
	}
}

func TestCommandInitialization(t *testing.T) {
	// Test that init() function properly adds subcommands
	expectedSubcommands := []string{"list", "apply", "export", "delete", "profiles"}

	actualCommands := PoliciesCmd.Commands()
	if len(actualCommands) != len(expectedSubcommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedSubcommands), len(actualCommands))
	}

	// Create a map of actual command names for easier lookup
	actualNames := make(map[string]bool)
	for _, cmd := range actualCommands {
		// Extract just the command name (first word) from Use field
		parts := strings.Fields(cmd.Use)
		if len(parts) > 0 {
			actualNames[parts[0]] = true
		}
	}

	// Check each expected command exists
	for _, expected := range expectedSubcommands {
		if !actualNames[expected] {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}
