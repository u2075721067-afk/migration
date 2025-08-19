package rules

import (
	"testing"
)

func TestEngine_Evaluate(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		rule     Rule
		context  Context
		expected bool
		hasError bool
	}{
		{
			name: "simple equality condition matches",
			rule: Rule{
				ID:      "test-1",
				Name:    "Test Rule",
				Enabled: true,
				Conditions: []Condition{
					{Field: "status", Operator: "==", Value: "error"},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{"message": "test"}},
				},
			},
			context: Context{
				Variables: map[string]interface{}{"status": "error"},
				Request:   make(map[string]interface{}),
				Response:  make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "simple equality condition does not match",
			rule: Rule{
				ID:      "test-2",
				Name:    "Test Rule",
				Enabled: true,
				Conditions: []Condition{
					{Field: "status", Operator: "==", Value: "error"},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{"message": "test"}},
				},
			},
			context: Context{
				Variables: map[string]interface{}{"status": "success"},
				Request:   make(map[string]interface{}),
				Response:  make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "disabled rule should not match",
			rule: Rule{
				ID:      "test-3",
				Name:    "Test Rule",
				Enabled: false,
				Conditions: []Condition{
					{Field: "status", Operator: "==", Value: "error"},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{"message": "test"}},
				},
			},
			context: Context{
				Variables: map[string]interface{}{"status": "error"},
				Request:   make(map[string]interface{}),
				Response:  make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "multiple conditions all match",
			rule: Rule{
				ID:      "test-4",
				Name:    "Test Rule",
				Enabled: true,
				Conditions: []Condition{
					{Field: "status", Operator: "==", Value: "error"},
					{Field: "retry_count", Operator: ">", Value: 2},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{"message": "test"}},
				},
			},
			context: Context{
				Variables: map[string]interface{}{
					"status":      "error",
					"retry_count": 3,
				},
				Request:  make(map[string]interface{}),
				Response: make(map[string]interface{}),
				Metadata: make(map[string]interface{}),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "multiple conditions one fails",
			rule: Rule{
				ID:      "test-5",
				Name:    "Test Rule",
				Enabled: true,
				Conditions: []Condition{
					{Field: "status", Operator: "==", Value: "error"},
					{Field: "retry_count", Operator: ">", Value: 5},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{"message": "test"}},
				},
			},
			context: Context{
				Variables: map[string]interface{}{
					"status":      "error",
					"retry_count": 3,
				},
				Request:  make(map[string]interface{}),
				Response: make(map[string]interface{}),
				Metadata: make(map[string]interface{}),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "negated condition",
			rule: Rule{
				ID:      "test-6",
				Name:    "Test Rule",
				Enabled: true,
				Conditions: []Condition{
					{Field: "status", Operator: "==", Value: "success", Negate: true},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{"message": "test"}},
				},
			},
			context: Context{
				Variables: map[string]interface{}{"status": "error"},
				Request:   make(map[string]interface{}),
				Response:  make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
			},
			expected: true,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Evaluate(tt.rule, tt.context)

			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result.Matched != tt.expected {
				t.Errorf("expected matched=%v, got matched=%v", tt.expected, result.Matched)
			}

			if result.RuleID != tt.rule.ID {
				t.Errorf("expected rule_id=%s, got rule_id=%s", tt.rule.ID, result.RuleID)
			}
		})
	}
}

func TestEngine_Execute(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name           string
		actions        []Action
		context        Context
		expectedCount  int
		expectedErrors int
	}{
		{
			name: "log action executes successfully",
			actions: []Action{
				{Type: "log", Params: map[string]interface{}{
					"message": "test message",
					"level":   "info",
				}},
			},
			context: Context{
				Variables: make(map[string]interface{}),
				Request:   make(map[string]interface{}),
				Response:  make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
			},
			expectedCount:  1,
			expectedErrors: 0,
		},
		{
			name: "set_var action executes successfully",
			actions: []Action{
				{Type: "set_var", Params: map[string]interface{}{
					"variable": "test_var",
					"value":    "test_value",
				}},
			},
			context: Context{
				Variables: make(map[string]interface{}),
				Request:   make(map[string]interface{}),
				Response:  make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
			},
			expectedCount:  1,
			expectedErrors: 0,
		},
		{
			name: "unknown action type fails",
			actions: []Action{
				{Type: "unknown_action", Params: map[string]interface{}{}},
			},
			context: Context{
				Variables: make(map[string]interface{}),
				Request:   make(map[string]interface{}),
				Response:  make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
			},
			expectedCount:  1,
			expectedErrors: 1,
		},
		{
			name: "multiple actions execute",
			actions: []Action{
				{Type: "log", Params: map[string]interface{}{
					"message": "first message",
					"level":   "info",
				}},
				{Type: "set_var", Params: map[string]interface{}{
					"variable": "test_var",
					"value":    "test_value",
				}},
				{Type: "log", Params: map[string]interface{}{
					"message": "second message",
					"level":   "info",
				}},
			},
			context: Context{
				Variables: make(map[string]interface{}),
				Request:   make(map[string]interface{}),
				Response:  make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
			},
			expectedCount:  3,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := engine.Execute(tt.actions, tt.context)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(results))
			}

			errorCount := 0
			for _, result := range results {
				if result.Error != "" {
					errorCount++
				}
			}

			if errorCount != tt.expectedErrors {
				t.Errorf("expected %d errors, got %d", tt.expectedErrors, errorCount)
			}
		})
	}
}

func TestEngine_Run(t *testing.T) {
	engine := NewEngine()

	ruleset := RuleSet{
		Version: "1.0.0",
		Name:    "Test RuleSet",
		Rules: []Rule{
			{
				ID:       "rule-1",
				Name:     "High Priority Rule",
				Priority: 200,
				Enabled:  true,
				Conditions: []Condition{
					{Field: "priority", Operator: "==", Value: "high"},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{
						"message": "high priority detected",
						"level":   "warning",
					}},
				},
			},
			{
				ID:       "rule-2",
				Name:     "Low Priority Rule",
				Priority: 100,
				Enabled:  true,
				Conditions: []Condition{
					{Field: "priority", Operator: "==", Value: "low"},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{
						"message": "low priority detected",
						"level":   "info",
					}},
				},
			},
			{
				ID:       "rule-3",
				Name:     "Disabled Rule",
				Priority: 300,
				Enabled:  false,
				Conditions: []Condition{
					{Field: "priority", Operator: "==", Value: "critical"},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{
						"message": "critical priority detected",
						"level":   "error",
					}},
				},
			},
		},
	}

	context := Context{
		Variables: map[string]interface{}{
			"priority": "high",
		},
		Request:  make(map[string]interface{}),
		Response: make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
	}

	results, err := engine.Run(ruleset, context)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Check that high priority rule matched
	var highPriorityResult *Result
	for i := range results {
		if results[i].RuleID == "rule-1" {
			highPriorityResult = &results[i]
			break
		}
	}

	if highPriorityResult == nil {
		t.Fatal("high priority rule result not found")
	}

	if !highPriorityResult.Matched {
		t.Error("high priority rule should have matched")
	}

	// Check that low priority rule did not match
	var lowPriorityResult *Result
	for i := range results {
		if results[i].RuleID == "rule-2" {
			lowPriorityResult = &results[i]
			break
		}
	}

	if lowPriorityResult == nil {
		t.Fatal("low priority rule result not found")
	}

	if lowPriorityResult.Matched {
		t.Error("low priority rule should not have matched")
	}

	// Check that disabled rule did not match
	var disabledResult *Result
	for i := range results {
		if results[i].RuleID == "rule-3" {
			disabledResult = &results[i]
			break
		}
	}

	if disabledResult == nil {
		t.Fatal("disabled rule result not found")
	}

	if disabledResult.Matched {
		t.Error("disabled rule should not have matched")
	}
}

func TestEngine_ValidateRule(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		rule     Rule
		hasError bool
	}{
		{
			name: "valid rule",
			rule: Rule{
				ID:   "valid-rule",
				Name: "Valid Rule",
				Conditions: []Condition{
					{Field: "status", Operator: "==", Value: "error"},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{"message": "test"}},
				},
			},
			hasError: false,
		},
		{
			name: "missing ID",
			rule: Rule{
				Name: "Rule Without ID",
				Conditions: []Condition{
					{Field: "status", Operator: "==", Value: "error"},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{"message": "test"}},
				},
			},
			hasError: true,
		},
		{
			name: "missing name",
			rule: Rule{
				ID: "rule-without-name",
				Conditions: []Condition{
					{Field: "status", Operator: "==", Value: "error"},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{"message": "test"}},
				},
			},
			hasError: true,
		},
		{
			name: "no conditions",
			rule: Rule{
				ID:         "rule-no-conditions",
				Name:       "Rule Without Conditions",
				Conditions: []Condition{},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{"message": "test"}},
				},
			},
			hasError: true,
		},
		{
			name: "no actions",
			rule: Rule{
				ID:   "rule-no-actions",
				Name: "Rule Without Actions",
				Conditions: []Condition{
					{Field: "status", Operator: "==", Value: "error"},
				},
				Actions: []Action{},
			},
			hasError: true,
		},
		{
			name: "invalid operator",
			rule: Rule{
				ID:   "rule-invalid-operator",
				Name: "Rule With Invalid Operator",
				Conditions: []Condition{
					{Field: "status", Operator: "invalid_op", Value: "error"},
				},
				Actions: []Action{
					{Type: "log", Params: map[string]interface{}{"message": "test"}},
				},
			},
			hasError: true,
		},
		{
			name: "invalid action type",
			rule: Rule{
				ID:   "rule-invalid-action",
				Name: "Rule With Invalid Action",
				Conditions: []Condition{
					{Field: "status", Operator: "==", Value: "error"},
				},
				Actions: []Action{
					{Type: "invalid_action", Params: map[string]interface{}{}},
				},
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateRule(tt.rule)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestEngine_ValidateRuleSet(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		ruleset  RuleSet
		hasError bool
	}{
		{
			name: "valid ruleset",
			ruleset: RuleSet{
				Name:    "Valid RuleSet",
				Version: "1.0.0",
				Rules: []Rule{
					{
						ID:   "rule-1",
						Name: "Rule 1",
						Conditions: []Condition{
							{Field: "status", Operator: "==", Value: "error"},
						},
						Actions: []Action{
							{Type: "log", Params: map[string]interface{}{"message": "test"}},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name: "missing name",
			ruleset: RuleSet{
				Version: "1.0.0",
				Rules: []Rule{
					{
						ID:   "rule-1",
						Name: "Rule 1",
						Conditions: []Condition{
							{Field: "status", Operator: "==", Value: "error"},
						},
						Actions: []Action{
							{Type: "log", Params: map[string]interface{}{"message": "test"}},
						},
					},
				},
			},
			hasError: true,
		},
		{
			name: "missing version",
			ruleset: RuleSet{
				Name: "RuleSet Without Version",
				Rules: []Rule{
					{
						ID:   "rule-1",
						Name: "Rule 1",
						Conditions: []Condition{
							{Field: "status", Operator: "==", Value: "error"},
						},
						Actions: []Action{
							{Type: "log", Params: map[string]interface{}{"message": "test"}},
						},
					},
				},
			},
			hasError: true,
		},
		{
			name: "no rules",
			ruleset: RuleSet{
				Name:    "Empty RuleSet",
				Version: "1.0.0",
				Rules:   []Rule{},
			},
			hasError: true,
		},
		{
			name: "duplicate rule IDs",
			ruleset: RuleSet{
				Name:    "RuleSet With Duplicates",
				Version: "1.0.0",
				Rules: []Rule{
					{
						ID:   "duplicate-id",
						Name: "Rule 1",
						Conditions: []Condition{
							{Field: "status", Operator: "==", Value: "error"},
						},
						Actions: []Action{
							{Type: "log", Params: map[string]interface{}{"message": "test"}},
						},
					},
					{
						ID:   "duplicate-id",
						Name: "Rule 2",
						Conditions: []Condition{
							{Field: "status", Operator: "==", Value: "success"},
						},
						Actions: []Action{
							{Type: "log", Params: map[string]interface{}{"message": "test"}},
						},
					},
				},
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateRuleSet(tt.ruleset)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
