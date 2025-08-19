import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { RuleBuilder } from '../RuleBuilder'

// Mock the CodeHighlight component
jest.mock('../ui/code-highlight', () => ({
  CodeHighlight: ({ code }: { code: string }) => (
    <pre data-testid="code-highlight">{code}</pre>
  ),
}))

describe('RuleBuilder', () => {
  it('renders rule builder interface', () => {
    render(<RuleBuilder />)
    
    expect(screen.getByText('Rule Builder')).toBeInTheDocument()
    expect(screen.getByText('Add Rule')).toBeInTheDocument()
    expect(screen.getByText('Save All')).toBeInTheDocument()
  })

  it('displays rules table with correct headers', () => {
    render(<RuleBuilder />)
    
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Priority')).toBeInTheDocument()
    expect(screen.getByText('Status')).toBeInTheDocument()
    expect(screen.getByText('Conditions')).toBeInTheDocument()
    expect(screen.getByText('Actions')).toBeInTheDocument()
    expect(screen.getByText('Operations')).toBeInTheDocument()
  })

  it('shows empty state when no rules are present', () => {
    render(<RuleBuilder />)
    
    expect(screen.getByText('No rules defined. Click "Add Rule" to create your first rule.')).toBeInTheDocument()
  })

  it('creates new rule when Add Rule button is clicked', () => {
    render(<RuleBuilder />)
    
    const addButton = screen.getByText('Add Rule')
    fireEvent.click(addButton)
    
    // Should switch to edit mode
    expect(screen.getByText('Edit Rule: New Rule')).toBeInTheDocument()
    expect(screen.getByText('Back to List')).toBeInTheDocument()
  })

  it('renders rule editor with form fields', () => {
    render(<RuleBuilder />)
    
    const addButton = screen.getByText('Add Rule')
    fireEvent.click(addButton)
    
    // Check for form fields
    expect(screen.getByText('Rule Details')).toBeInTheDocument()
    expect(screen.getByText('Conditions')).toBeInTheDocument()
    expect(screen.getByText('Actions')).toBeInTheDocument()
    
    // Check for input fields
    expect(screen.getByDisplayValue('New Rule')).toBeInTheDocument()
    expect(screen.getByDisplayValue('A new rule')).toBeInTheDocument()
    expect(screen.getByDisplayValue('100')).toBeInTheDocument()
  })

  it('allows editing rule name', () => {
    render(<RuleBuilder />)
    
    const addButton = screen.getByText('Add Rule')
    fireEvent.click(addButton)
    
    const nameInput = screen.getByDisplayValue('New Rule')
    fireEvent.change(nameInput, { target: { value: 'My Test Rule' } })
    
    expect(screen.getByDisplayValue('My Test Rule')).toBeInTheDocument()
  })

  it('allows adding conditions', () => {
    render(<RuleBuilder />)
    
    const addButton = screen.getByText('Add Rule')
    fireEvent.click(addButton)
    
    const addConditionButton = screen.getByText('Add Condition')
    fireEvent.click(addConditionButton)
    
    // Should have 2 conditions now (1 default + 1 added)
    const fieldInputs = screen.getAllByPlaceholderText('Field')
    expect(fieldInputs).toHaveLength(2)
  })

  it('allows adding actions', () => {
    render(<RuleBuilder />)
    
    const addButton = screen.getByText('Add Rule')
    fireEvent.click(addButton)
    
    const addActionButton = screen.getByText('Add Action')
    fireEvent.click(addActionButton)
    
    // Should have 2 actions now (1 default + 1 added)
    const removeButtons = screen.getAllByText('Remove')
    expect(removeButtons.length).toBeGreaterThanOrEqual(2)
  })

  it('switches to YAML view when YAML View button is clicked', () => {
    render(<RuleBuilder />)
    
    const yamlButton = screen.getByText('YAML View')
    fireEvent.click(yamlButton)
    
    expect(screen.getByText('Rule YAML Editor')).toBeInTheDocument()
    expect(screen.getByText('Back to List')).toBeInTheDocument()
    expect(screen.getByTestId('code-highlight')).toBeInTheDocument()
  })

  it('calls onSave callback when Save All is clicked', () => {
    const mockOnSave = jest.fn()
    render(<RuleBuilder onSave={mockOnSave} />)
    
    const saveButton = screen.getByText('Save All')
    fireEvent.click(saveButton)
    
    expect(mockOnSave).toHaveBeenCalledWith([])
  })

  it('renders evaluation section when onEvaluate is provided', () => {
    const mockOnEvaluate = jest.fn()
    render(<RuleBuilder onEvaluate={mockOnEvaluate} />)
    
    expect(screen.getByText('Rule Evaluation')).toBeInTheDocument()
    expect(screen.getByText('Evaluation Context (JSON)')).toBeInTheDocument()
    expect(screen.getByText('Evaluate Rules')).toBeInTheDocument()
  })

  it('calls onEvaluate when Evaluate Rules button is clicked with valid context', async () => {
    const mockOnEvaluate = jest.fn().mockResolvedValue({ results: [] })
    render(<RuleBuilder onEvaluate={mockOnEvaluate} />)
    
    const contextTextarea = screen.getByPlaceholderText(/{"variables"/)
    fireEvent.change(contextTextarea, { 
      target: { value: '{"variables": {"status": "test"}}' } 
    })
    
    const evaluateButton = screen.getByText('Evaluate Rules')
    fireEvent.click(evaluateButton)
    
    await waitFor(() => {
      expect(mockOnEvaluate).toHaveBeenCalledWith([], { variables: { status: 'test' } })
    })
  })

  it('renders with initial rules', () => {
    const initialRules = [
      {
        id: 'test-rule-1',
        name: 'Test Rule',
        description: 'A test rule',
        priority: 100,
        enabled: true,
        conditions: [{ field: 'status', operator: '==', value: 'error' }],
        actions: [{ type: 'log', params: { message: 'test', level: 'info' } }],
      },
    ]
    
    render(<RuleBuilder initialRules={initialRules} />)
    
    expect(screen.getByText('Test Rule')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument()
    expect(screen.getByText('Enabled')).toBeInTheDocument()
    // Should have 1 condition and 1 action
    const ones = screen.getAllByText('1')
    expect(ones).toHaveLength(2) // conditions count and actions count
  })

  it('allows duplicating rules', () => {
    const initialRules = [
      {
        id: 'test-rule-1',
        name: 'Test Rule',
        description: 'A test rule',
        priority: 100,
        enabled: true,
        conditions: [{ field: 'status', operator: '==', value: 'error' }],
        actions: [{ type: 'log', params: { message: 'test', level: 'info' } }],
      },
    ]
    
    render(<RuleBuilder initialRules={initialRules} />)
    
    const copyButton = screen.getByText('Copy')
    fireEvent.click(copyButton)
    
    // Should now have 2 rules
    expect(screen.getAllByText('Test Rule')).toHaveLength(1)
    expect(screen.getByText('Test Rule (Copy)')).toBeInTheDocument()
  })

  it('allows deleting rules', () => {
    const initialRules = [
      {
        id: 'test-rule-1',
        name: 'Test Rule',
        description: 'A test rule',
        priority: 100,
        enabled: true,
        conditions: [{ field: 'status', operator: '==', value: 'error' }],
        actions: [{ type: 'log', params: { message: 'test', level: 'info' } }],
      },
    ]
    
    render(<RuleBuilder initialRules={initialRules} />)
    
    const deleteButton = screen.getByText('Delete')
    fireEvent.click(deleteButton)
    
    // Should show empty state again
    expect(screen.getByText('No rules defined. Click "Add Rule" to create your first rule.')).toBeInTheDocument()
  })
})
