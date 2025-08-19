import React, { useState, useEffect } from 'react';
import { Card } from './ui/card';
import { Button } from './ui/button';
import { Textarea } from './ui/textarea';

interface RetryProfile {
  name: string;
  description: string;
  maxRetries: number;
  initialDelay: string;
  maxDelay: string;
  backoffMultiplier: number;
  jitter: number;
  timeout: string;
}

interface PolicyCondition {
  errorType?: string;
  httpStatus?: number;
  errorMessagePattern?: string;
  actionType?: string;
}

interface RetryPolicy {
  id?: string;
  name: string;
  description: string;
  retryProfile: string;
  conditions: PolicyCondition[];
  enabled: boolean;
}

const RetryPolicyEditor: React.FC = () => {
  const [policy, setPolicy] = useState<RetryPolicy>({
    name: '',
    description: '',
    retryProfile: 'balanced',
    conditions: [],
    enabled: true,
  });

  const [yamlPreview, setYamlPreview] = useState<string>('');
  const [profiles] = useState<RetryProfile[]>([
    {
      name: 'aggressive',
      description: 'Fast retry with minimal backoff for time-sensitive operations',
      maxRetries: 3,
      initialDelay: '100ms',
      maxDelay: '1s',
      backoffMultiplier: 1.5,
      jitter: 0.1,
      timeout: '5s',
    },
    {
      name: 'balanced',
      description: 'Balanced retry with exponential backoff for general use cases',
      maxRetries: 5,
      initialDelay: '500ms',
      maxDelay: '10s',
      backoffMultiplier: 2.0,
      jitter: 0.2,
      timeout: '30s',
    },
    {
      name: 'conservative',
      description: 'Conservative retry with long intervals for resource-intensive operations',
      maxRetries: 10,
      initialDelay: '2s',
      maxDelay: '60s',
      backoffMultiplier: 2.5,
      jitter: 0.3,
      timeout: '300s',
    },
  ]);

  useEffect(() => {
    updateYamlPreview();
  }, [policy]);

  const updateYamlPreview = () => {
    const yaml = `name: ${policy.name}
description: ${policy.description}
retryProfile: ${policy.retryProfile}
enabled: ${policy.enabled}
conditions:
${policy.conditions.map(cond => `  - errorType: ${cond.errorType || ''}
    httpStatus: ${cond.httpStatus || ''}
    errorMessagePattern: ${cond.errorMessagePattern || ''}
    actionType: ${cond.actionType || ''}`).join('\n')}`;
    
    setYamlPreview(yaml);
  };

  const addCondition = () => {
    setPolicy(prev => ({
      ...prev,
      conditions: [...prev.conditions, {}],
    }));
  };

  const updateCondition = (index: number, field: keyof PolicyCondition, value: string | number) => {
    setPolicy(prev => ({
      ...prev,
      conditions: prev.conditions.map((cond, i) => 
        i === index ? { ...cond, [field]: value } : cond
      ),
    }));
  };

  const removeCondition = (index: number) => {
    setPolicy(prev => ({
      ...prev,
      conditions: prev.conditions.filter((_, i) => i !== index),
    }));
  };

  const applyPolicy = async () => {
    try {
      const response = await fetch('/api/v1/policies', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(policy),
      });

      if (response.ok) {
        alert('Policy applied successfully!');
      } else {
        alert('Failed to apply policy');
      }
    } catch (error) {
      console.error('Error applying policy:', error);
      alert('Error applying policy');
    }
  };

  const selectedProfile = profiles.find(p => p.name === policy.retryProfile);

  return (
    <div className="space-y-6">
      <Card className="p-6">
        <h2 className="text-2xl font-bold mb-4">Retry Policy Editor</h2>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Policy Configuration */}
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">Policy Name</label>
              <input
                type="text"
                value={policy.name}
                onChange={(e) => setPolicy(prev => ({ ...prev, name: e.target.value }))}
                className="w-full p-2 border rounded-md"
                placeholder="Enter policy name"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Description</label>
              <Textarea
                value={policy.description}
                onChange={(e) => setPolicy(prev => ({ ...prev, description: e.target.value }))}
                placeholder="Describe the policy purpose"
                rows={3}
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Retry Profile</label>
              <select
                value={policy.retryProfile}
                onChange={(e) => setPolicy(prev => ({ ...prev, retryProfile: e.target.value }))}
                className="w-full p-2 border rounded-md"
              >
                {profiles.map(profile => (
                  <option key={profile.name} value={profile.name}>
                    {profile.name} - {profile.description}
                  </option>
                ))}
              </select>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="enabled"
                checked={policy.enabled}
                onChange={(e) => setPolicy(prev => ({ ...prev, enabled: e.target.checked }))}
                className="rounded"
              />
              <label htmlFor="enabled" className="text-sm font-medium">Enable Policy</label>
            </div>
          </div>

          {/* Profile Details */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold">Selected Profile: {policy.retryProfile}</h3>
            {selectedProfile && (
              <div className="bg-gray-50 p-4 rounded-md space-y-2">
                <div><strong>Max Retries:</strong> {selectedProfile.maxRetries}</div>
                <div><strong>Initial Delay:</strong> {selectedProfile.initialDelay}</div>
                <div><strong>Max Delay:</strong> {selectedProfile.maxDelay}</div>
                <div><strong>Backoff Multiplier:</strong> {selectedProfile.backoffMultiplier}</div>
                <div><strong>Jitter:</strong> {selectedProfile.jitter}</div>
                <div><strong>Timeout:</strong> {selectedProfile.timeout}</div>
              </div>
            )}
          </div>
        </div>
      </Card>

      {/* Conditions */}
      <Card className="p-6">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold">Conditions</h3>
          <Button onClick={addCondition} className="bg-blue-600 hover:bg-blue-700">
            Add Condition
          </Button>
        </div>

        {policy.conditions.length === 0 && (
          <p className="text-gray-500 text-center py-4">No conditions defined. Add conditions to make the policy more specific.</p>
        )}

        {policy.conditions.map((condition, index) => (
          <div key={index} className="border rounded-md p-4 mb-4">
            <div className="flex justify-between items-center mb-3">
              <h4 className="font-medium">Condition {index + 1}</h4>
              <Button 
                onClick={() => removeCondition(index)}
                className="bg-red-600 hover:bg-red-700 text-white px-2 py-1 text-sm"
              >
                Remove
              </Button>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Error Type</label>
                <input
                  type="text"
                  value={condition.errorType || ''}
                  onChange={(e) => updateCondition(index, 'errorType', e.target.value)}
                  className="w-full p-2 border rounded-md text-sm"
                  placeholder="e.g., timeout, network"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">HTTP Status</label>
                <input
                  type="number"
                  value={condition.httpStatus || ''}
                  onChange={(e) => updateCondition(index, 'httpStatus', parseInt(e.target.value) || 0)}
                  className="w-full p-2 border rounded-md text-sm"
                  placeholder="e.g., 408, 500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Error Message Pattern</label>
                <input
                  type="text"
                  value={condition.errorMessagePattern || ''}
                  onChange={(e) => updateCondition(index, 'errorMessagePattern', e.target.value)}
                  className="w-full p-2 border rounded-md text-sm"
                  placeholder="regex pattern"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Action Type</label>
                <input
                  type="text"
                  value={condition.actionType || ''}
                  onChange={(e) => updateCondition(index, 'actionType', e.target.value)}
                  className="w-full p-2 border rounded-md text-sm"
                  placeholder="e.g., http_fetch, parse_json"
                />
              </div>
            </div>
          </div>
        ))}
      </Card>

      {/* YAML Preview */}
      <Card className="p-6">
        <h3 className="text-lg font-semibold mb-4">YAML Preview</h3>
        <pre className="bg-gray-100 p-4 rounded-md overflow-x-auto text-sm">
          {yamlPreview}
        </pre>
      </Card>

      {/* Actions */}
      <div className="flex justify-end space-x-4">
        <Button 
          onClick={applyPolicy}
          className="bg-green-600 hover:bg-green-700"
          disabled={!policy.name || !policy.description}
        >
          Apply Policy
        </Button>
      </div>
    </div>
  );
};

export default RetryPolicyEditor;
