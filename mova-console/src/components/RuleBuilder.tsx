'use client'

import React, { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { CodeHighlight } from '@/components/ui/code-highlight'

interface Condition {
  field: string
  operator: string
  value: any
  negate?: boolean
}

interface Action {
  type: string
  params: Record<string, any>
}

interface Rule {
  id: string
  name: string
  description: string
  priority: number
  enabled: boolean
  conditions: Condition[]
  actions: Action[]
  created_at?: string
  updated_at?: string
}

interface RuleSet {
  version: string
  name: string
  description: string
  rules: Rule[]
  metadata?: Record<string, string>
  created_at?: string
  updated_at?: string
}

interface RuleBuilderProps {
  initialRules?: Rule[]
  onSave?: (rules: Rule[]) => void
  onEvaluate?: (rules: Rule[], context: any) => Promise<any>
}

const OPERATORS = [
  { value: '==', label: 'Equals' },
  { value: '!=', label: 'Not Equals' },
  { value: '>', label: 'Greater Than' },
  { value: '>=', label: 'Greater Than or Equal' },
  { value: '<', label: 'Less Than' },
  { value: '<=', label: 'Less Than or Equal' },
  { value: 'contains', label: 'Contains' },
  { value: 'not_contains', label: 'Not Contains' },
  { value: 'regex', label: 'Regex Match' },
  { value: 'in', label: 'In List' },
  { value: 'not_in', label: 'Not In List' },
  { value: 'exists', label: 'Exists' },
  { value: 'not_exists', label: 'Not Exists' },
]

const ACTION_TYPES = [
  { value: 'set_var', label: 'Set Variable', params: ['variable', 'value'] },
  { value: 'retry', label: 'Retry', params: ['profile', 'max_attempts', 'delay'] },
  { value: 'http_call', label: 'HTTP Call', params: ['url', 'method', 'headers', 'body', 'timeout'] },
  { value: 'skip', label: 'Skip', params: ['reason'] },
  { value: 'log', label: 'Log', params: ['message', 'level'] },
  { value: 'route', label: 'Route', params: ['workflow', 'reason'] },
  { value: 'stop', label: 'Stop', params: ['reason'] },
  { value: 'transform', label: 'Transform', params: ['type', 'source', 'target'] },
]

export function RuleBuilder({ initialRules = [], onSave, onEvaluate }: RuleBuilderProps) {
  const [rules, setRules] = useState<Rule[]>(initialRules)
  const [selectedRule, setSelectedRule] = useState<Rule | null>(null)
  const [editMode, setEditMode] = useState<'list' | 'edit' | 'yaml'>('list')
  const [yamlContent, setYamlContent] = useState('')
  const [evaluationContext, setEvaluationContext] = useState('')
  const [evaluationResults, setEvaluationResults] = useState<any>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (editMode === 'yaml') {
      const ruleset: RuleSet = {
        version: '1.0.0',
        name: 'Custom RuleSet',
        description: 'Generated from Rule Builder',
        rules: rules,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }
      setYamlContent(JSON.stringify(ruleset, null, 2))
    }
  }, [rules, editMode])

  const createNewRule = (): Rule => ({
    id: `rule-${Date.now()}`,
    name: 'New Rule',
    description: 'A new rule',
    priority: 100,
    enabled: true,
    conditions: [{ field: '', operator: '==', value: '' }],
    actions: [{ type: 'log', params: { message: 'Rule matched', level: 'info' } }],
  })

  const addRule = () => {
    const newRule = createNewRule()
    setRules([...rules, newRule])
    setSelectedRule(newRule)
    setEditMode('edit')
  }

  const updateRule = (updatedRule: Rule) => {
    const updatedRules = rules.map(rule => 
      rule.id === updatedRule.id ? updatedRule : rule
    )
    setRules(updatedRules)
    setSelectedRule(updatedRule)
  }

  const deleteRule = (ruleId: string) => {
    const updatedRules = rules.filter(rule => rule.id !== ruleId)
    setRules(updatedRules)
    if (selectedRule?.id === ruleId) {
      setSelectedRule(null)
      setEditMode('list')
    }
  }

  const duplicateRule = (rule: Rule) => {
    const duplicatedRule: Rule = {
      ...rule,
      id: `rule-${Date.now()}`,
      name: `${rule.name} (Copy)`,
    }
    setRules([...rules, duplicatedRule])
  }

  const addCondition = () => {
    if (!selectedRule) return
    const updatedRule = {
      ...selectedRule,
      conditions: [...selectedRule.conditions, { field: '', operator: '==', value: '' }]
    }
    updateRule(updatedRule)
  }

  const updateCondition = (index: number, field: keyof Condition, value: any) => {
    if (!selectedRule) return
    const updatedConditions = [...selectedRule.conditions]
    updatedConditions[index] = { ...updatedConditions[index], [field]: value }
    updateRule({ ...selectedRule, conditions: updatedConditions })
  }

  const removeCondition = (index: number) => {
    if (!selectedRule) return
    const updatedConditions = selectedRule.conditions.filter((_, i) => i !== index)
    updateRule({ ...selectedRule, conditions: updatedConditions })
  }

  const addAction = () => {
    if (!selectedRule) return
    const updatedRule = {
      ...selectedRule,
      actions: [...selectedRule.actions, { type: 'log', params: { message: '', level: 'info' } }]
    }
    updateRule(updatedRule)
  }

  const updateAction = (index: number, field: keyof Action, value: any) => {
    if (!selectedRule) return
    const updatedActions = [...selectedRule.actions]
    updatedActions[index] = { ...updatedActions[index], [field]: value }
    updateRule({ ...selectedRule, actions: updatedActions })
  }

  const removeAction = (index: number) => {
    if (!selectedRule) return
    const updatedActions = selectedRule.actions.filter((_, i) => i !== index)
    updateRule({ ...selectedRule, actions: updatedActions })
  }

  const handleEvaluate = async () => {
    if (!onEvaluate || !evaluationContext) return
    
    setLoading(true)
    try {
      const context = JSON.parse(evaluationContext)
      const results = await onEvaluate(rules, context)
      setEvaluationResults(results)
    } catch (error) {
      console.error('Evaluation failed:', error)
      setEvaluationResults({ error: error instanceof Error ? error.message : 'Unknown error' })
    } finally {
      setLoading(false)
    }
  }

  const handleSave = () => {
    if (onSave) {
      onSave(rules)
    }
  }

  if (editMode === 'yaml') {
    return (
      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <h2 className="text-2xl font-bold">Rule YAML Editor</h2>
          <div className="space-x-2">
            <Button variant="outline" onClick={() => setEditMode('list')}>
              Back to List
            </Button>
            <Button onClick={handleSave}>Save Rules</Button>
          </div>
        </div>
        
        <Card>
          <CardHeader>
            <CardTitle>YAML Configuration</CardTitle>
            <CardDescription>
              Edit the complete ruleset configuration in YAML format
            </CardDescription>
          </CardHeader>
          <CardContent>
            <CodeHighlight language="yaml" code={yamlContent} />
          </CardContent>
        </Card>
      </div>
    )
  }

  if (editMode === 'edit' && selectedRule) {
    return (
      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <h2 className="text-2xl font-bold">Edit Rule: {selectedRule.name}</h2>
          <div className="space-x-2">
            <Button variant="outline" onClick={() => setEditMode('list')}>
              Back to List
            </Button>
            <Button onClick={handleSave}>Save Rule</Button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Rule Details */}
          <Card>
            <CardHeader>
              <CardTitle>Rule Details</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">Name</label>
                <input
                  type="text"
                  value={selectedRule.name}
                  onChange={(e) => updateRule({ ...selectedRule, name: e.target.value })}
                  className="w-full p-2 border rounded"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">Description</label>
                <Textarea
                  value={selectedRule.description}
                  onChange={(e) => updateRule({ ...selectedRule, description: e.target.value })}
                  rows={2}
                />
              </div>
              
              <div className="flex space-x-4">
                <div className="flex-1">
                  <label className="block text-sm font-medium mb-1">Priority</label>
                  <input
                    type="number"
                    value={selectedRule.priority}
                    onChange={(e) => updateRule({ ...selectedRule, priority: parseInt(e.target.value) })}
                    className="w-full p-2 border rounded"
                  />
                </div>
                
                <div className="flex items-center space-x-2 mt-6">
                  <input
                    type="checkbox"
                    id="enabled"
                    checked={selectedRule.enabled}
                    onChange={(e) => updateRule({ ...selectedRule, enabled: e.target.checked })}
                  />
                  <label htmlFor="enabled" className="text-sm font-medium">Enabled</label>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Conditions */}
          <Card>
            <CardHeader>
              <CardTitle>Conditions</CardTitle>
              <CardDescription>All conditions must be true for the rule to match</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {selectedRule.conditions.map((condition, index) => (
                  <div key={index} className="flex space-x-2 items-center p-3 border rounded">
                    <input
                      type="text"
                      placeholder="Field"
                      value={condition.field}
                      onChange={(e) => updateCondition(index, 'field', e.target.value)}
                      className="flex-1 p-1 border rounded text-sm"
                    />
                    
                    <select
                      value={condition.operator}
                      onChange={(e) => updateCondition(index, 'operator', e.target.value)}
                      className="p-1 border rounded text-sm"
                    >
                      {OPERATORS.map(op => (
                        <option key={op.value} value={op.value}>{op.label}</option>
                      ))}
                    </select>
                    
                    <input
                      type="text"
                      placeholder="Value"
                      value={typeof condition.value === 'string' ? condition.value : JSON.stringify(condition.value)}
                      onChange={(e) => {
                        try {
                          const parsed = JSON.parse(e.target.value)
                          updateCondition(index, 'value', parsed)
                        } catch {
                          updateCondition(index, 'value', e.target.value)
                        }
                      }}
                      className="flex-1 p-1 border rounded text-sm"
                    />
                    
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => removeCondition(index)}
                      className="text-red-600"
                    >
                      ×
                    </Button>
                  </div>
                ))}
                
                <Button variant="outline" onClick={addCondition} className="w-full">
                  Add Condition
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Actions */}
          <Card className="lg:col-span-2">
            <CardHeader>
              <CardTitle>Actions</CardTitle>
              <CardDescription>Actions to execute when rule conditions match</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {selectedRule.actions.map((action, index) => (
                  <div key={index} className="p-3 border rounded">
                    <div className="flex items-center space-x-2 mb-2">
                      <select
                        value={action.type}
                        onChange={(e) => {
                          const actionType = ACTION_TYPES.find(t => t.value === e.target.value)
                          const defaultParams: Record<string, any> = {}
                          if (actionType) {
                            actionType.params.forEach(param => {
                              defaultParams[param] = ''
                            })
                          }
                          updateAction(index, 'type', e.target.value)
                          updateAction(index, 'params', defaultParams)
                        }}
                        className="p-1 border rounded text-sm"
                      >
                        {ACTION_TYPES.map(type => (
                          <option key={type.value} value={type.value}>{type.label}</option>
                        ))}
                      </select>
                      
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => removeAction(index)}
                        className="text-red-600"
                      >
                        Remove
                      </Button>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-2">
                      {Object.entries(action.params).map(([key, value]) => (
                        <div key={key}>
                          <label className="block text-xs font-medium mb-1">{key}</label>
                          <input
                            type="text"
                            value={typeof value === 'string' ? value : JSON.stringify(value)}
                            onChange={(e) => {
                              const updatedParams = { ...action.params }
                              try {
                                updatedParams[key] = JSON.parse(e.target.value)
                              } catch {
                                updatedParams[key] = e.target.value
                              }
                              updateAction(index, 'params', updatedParams)
                            }}
                            className="w-full p-1 border rounded text-sm"
                          />
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
                
                <Button variant="outline" onClick={addAction} className="w-full">
                  Add Action
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">Rule Builder</h2>
        <div className="space-x-2">
          <Button variant="outline" onClick={() => setEditMode('yaml')}>
            YAML View
          </Button>
          <Button onClick={addRule}>Add Rule</Button>
          <Button onClick={handleSave}>Save All</Button>
        </div>
      </div>

      {/* Rules Table */}
      <Card>
        <CardHeader>
          <CardTitle>Rules ({rules.length})</CardTitle>
          <CardDescription>
            Manage your automation rules. Click on a rule to edit it.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Priority</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Conditions</TableHead>
                <TableHead>Actions</TableHead>
                <TableHead>Operations</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rules.map((rule) => (
                <TableRow key={rule.id} className="cursor-pointer hover:bg-gray-50">
                  <TableCell 
                    onClick={() => {
                      setSelectedRule(rule)
                      setEditMode('edit')
                    }}
                    className="font-medium"
                  >
                    {rule.name}
                  </TableCell>
                  <TableCell>{rule.priority}</TableCell>
                  <TableCell>
                    <span className={`px-2 py-1 rounded text-xs ${
                      rule.enabled 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-gray-100 text-gray-800'
                    }`}>
                      {rule.enabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </TableCell>
                  <TableCell>{rule.conditions.length}</TableCell>
                  <TableCell>{rule.actions.length}</TableCell>
                  <TableCell>
                    <div className="flex space-x-1">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation()
                          duplicateRule(rule)
                        }}
                      >
                        Copy
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation()
                          deleteRule(rule.id)
                        }}
                        className="text-red-600"
                      >
                        Delete
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          
          {rules.length === 0 && (
            <div className="text-center py-8 text-gray-500">
              No rules defined. Click "Add Rule" to create your first rule.
            </div>
          )}
        </CardContent>
      </Card>

      {/* Evaluation Section */}
      {onEvaluate && (
        <Card>
          <CardHeader>
            <CardTitle>Rule Evaluation</CardTitle>
            <CardDescription>
              Test your rules against a sample context without executing actions.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Evaluation Context (JSON)</label>
              <Textarea
                value={evaluationContext}
                onChange={(e) => setEvaluationContext(e.target.value)}
                placeholder='{"variables": {"status": "active"}, "request": {"user_id": "123"}, "response": {}, "metadata": {}}'
                rows={4}
              />
            </div>
            
            <Button 
              onClick={handleEvaluate} 
              disabled={loading || !evaluationContext}
            >
              {loading ? 'Evaluating...' : 'Evaluate Rules'}
            </Button>
            
            {evaluationResults && (
              <div className="mt-4">
                <h4 className="font-medium mb-2">Evaluation Results:</h4>
                <CodeHighlight 
                  language="json" 
                  code={JSON.stringify(evaluationResults, null, 2)} 
                />
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
