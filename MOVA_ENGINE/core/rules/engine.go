package rules

import (
	"fmt"
	"sort"
	"time"
)

// Engine implements the RuleEngine interface
type Engine struct {
	operators map[string]OperatorFunc
	actions   map[string]ActionFunc
}

// NewEngine creates a new rule engine instance
func NewEngine() *Engine {
	engine := &Engine{
		operators: make(map[string]OperatorFunc),
		actions:   make(map[string]ActionFunc),
	}

	// Register default operators
	engine.registerDefaultOperators()

	// Register default actions
	engine.registerDefaultActions()

	return engine
}

// Evaluate evaluates a single rule against the context
func (e *Engine) Evaluate(rule Rule, ctx Context) (Result, error) {
	result := Result{
		RuleID:       rule.ID,
		Matched:      false,
		Actions:      []Action{},
		Variables:    make(map[string]interface{}),
		ExecutedAt:   time.Now(),
		ExecutionLog: []string{},
	}

	if !rule.Enabled {
		result.ExecutionLog = append(result.ExecutionLog, fmt.Sprintf("Rule %s is disabled", rule.ID))
		return result, nil
	}

	// Evaluate all conditions
	matched, err := e.evaluateConditions(rule.Conditions, ctx)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Matched = matched
	result.ExecutionLog = append(result.ExecutionLog, fmt.Sprintf("Rule %s conditions evaluated: %v", rule.ID, matched))

	if matched {
		result.Actions = rule.Actions
		result.ExecutionLog = append(result.ExecutionLog, fmt.Sprintf("Rule %s matched, %d actions to execute", rule.ID, len(rule.Actions)))
	}

	// Copy context variables to result
	for k, v := range ctx.Variables {
		result.Variables[k] = v
	}

	return result, nil
}

// Execute executes a list of actions in the context
func (e *Engine) Execute(actions []Action, ctx Context) ([]ExecutionResult, error) {
	var results []ExecutionResult

	for _, action := range actions {
		startTime := time.Now()

		result := ExecutionResult{
			ActionType: action.Type,
			Success:    false,
			Output:     make(map[string]interface{}),
			Duration:   0,
		}

		actionFunc, exists := e.actions[action.Type]
		if !exists {
			result.Error = fmt.Sprintf("Unknown action type: %s", action.Type)
			result.Duration = time.Since(startTime)
			results = append(results, result)
			continue
		}

		output, err := actionFunc(action.Params, ctx)
		result.Duration = time.Since(startTime)

		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
			result.Output = output
		}

		results = append(results, result)
	}

	return results, nil
}

// Run runs a complete ruleset against the context
func (e *Engine) Run(ruleset RuleSet, ctx Context) ([]Result, error) {
	var results []Result

	// Sort rules by priority (higher priority first)
	sortedRules := make([]Rule, len(ruleset.Rules))
	copy(sortedRules, ruleset.Rules)
	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].Priority > sortedRules[j].Priority
	})

	for _, rule := range sortedRules {
		result, err := e.Evaluate(rule, ctx)
		if err != nil {
			result.Error = err.Error()
		}

		results = append(results, result)

		// If rule matched and has actions, execute them
		if result.Matched && len(result.Actions) > 0 {
			execResults, execErr := e.Execute(result.Actions, ctx)
			if execErr != nil {
				result.Error = fmt.Sprintf("%s; execution error: %s", result.Error, execErr.Error())
			}

			// Update context variables based on action results
			for _, execResult := range execResults {
				if execResult.Success && execResult.ActionType == ActionSetVar {
					if varName, ok := execResult.Output["variable"].(string); ok {
						if varValue, ok := execResult.Output["value"]; ok {
							ctx.Variables[varName] = varValue
						}
					}
				}
			}
		}
	}

	return results, nil
}

// ValidateRule validates rule syntax and structure
func (e *Engine) ValidateRule(rule Rule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}

	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if len(rule.Conditions) == 0 {
		return fmt.Errorf("rule must have at least one condition")
	}

	if len(rule.Actions) == 0 {
		return fmt.Errorf("rule must have at least one action")
	}

	// Validate conditions
	for i, condition := range rule.Conditions {
		if condition.Field == "" {
			return fmt.Errorf("condition %d: field is required", i)
		}

		if condition.Operator == "" {
			return fmt.Errorf("condition %d: operator is required", i)
		}

		if _, exists := e.operators[condition.Operator]; !exists {
			return fmt.Errorf("condition %d: unknown operator '%s'", i, condition.Operator)
		}
	}

	// Validate actions
	for i, action := range rule.Actions {
		if action.Type == "" {
			return fmt.Errorf("action %d: type is required", i)
		}

		if _, exists := e.actions[action.Type]; !exists {
			return fmt.Errorf("action %d: unknown action type '%s'", i, action.Type)
		}
	}

	return nil
}

// ValidateRuleSet validates ruleset syntax and structure
func (e *Engine) ValidateRuleSet(ruleset RuleSet) error {
	if ruleset.Name == "" {
		return fmt.Errorf("ruleset name is required")
	}

	if ruleset.Version == "" {
		return fmt.Errorf("ruleset version is required")
	}

	if len(ruleset.Rules) == 0 {
		return fmt.Errorf("ruleset must contain at least one rule")
	}

	// Check for duplicate rule IDs
	ruleIDs := make(map[string]bool)
	for i, rule := range ruleset.Rules {
		if ruleIDs[rule.ID] {
			return fmt.Errorf("duplicate rule ID '%s' at index %d", rule.ID, i)
		}
		ruleIDs[rule.ID] = true

		// Validate each rule
		if err := e.ValidateRule(rule); err != nil {
			return fmt.Errorf("rule %d (%s): %s", i, rule.ID, err.Error())
		}
	}

	return nil
}

// evaluateConditions evaluates all conditions with AND logic
func (e *Engine) evaluateConditions(conditions []Condition, ctx Context) (bool, error) {
	for _, condition := range conditions {
		matched, err := e.evaluateCondition(condition, ctx)
		if err != nil {
			return false, err
		}

		// Apply negation if specified
		if condition.Negate {
			matched = !matched
		}

		// AND logic - if any condition fails, return false
		if !matched {
			return false, nil
		}
	}

	return true, nil
}

// evaluateCondition evaluates a single condition
func (e *Engine) evaluateCondition(condition Condition, ctx Context) (bool, error) {
	operatorFunc, exists := e.operators[condition.Operator]
	if !exists {
		return false, fmt.Errorf("unknown operator: %s", condition.Operator)
	}

	// Get field value from context
	fieldValue := e.getFieldValue(condition.Field, ctx)

	return operatorFunc(fieldValue, condition.Value)
}

// getFieldValue extracts field value from context using dot notation
func (e *Engine) getFieldValue(field string, ctx Context) interface{} {
	// Simple implementation - can be extended for nested field access
	if val, ok := ctx.Variables[field]; ok {
		return val
	}

	if val, ok := ctx.Request[field]; ok {
		return val
	}

	if val, ok := ctx.Response[field]; ok {
		return val
	}

	if val, ok := ctx.Metadata[field]; ok {
		return val
	}

	return nil
}

// RegisterOperator registers a custom operator
func (e *Engine) RegisterOperator(name string, fn OperatorFunc) {
	e.operators[name] = fn
}

// RegisterAction registers a custom action
func (e *Engine) RegisterAction(name string, fn ActionFunc) {
	e.actions[name] = fn
}
