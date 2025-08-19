import React, { useState, useEffect } from 'react';
import { Card } from './ui/card';
import { Button } from './ui/button';

interface DLQEntry {
  id: string;
  run_id: string;
  created_at: string;
  status: 'active' | 'retrying' | 'resolved' | 'archived';
  envelope: {
    intent: {
      name: string;
      description: string;
    };
  };
  error_details: {
    last_error: string;
    attempts: number;
    failure_reason: string;
  };
  metadata: {
    workflow_type: string;
    retry_count: number;
  };
}

interface DLQStats {
  total_entries: number;
  by_status: Record<string, number>;
  by_workflow_type: Record<string, number>;
}

export default function DLQManager() {
  const [entries, setEntries] = useState<DLQEntry[]>([]);
  const [stats, setStats] = useState<DLQStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedEntry, setSelectedEntry] = useState<DLQEntry | null>(null);
  const [filter, setFilter] = useState({
    status: '',
    workflow_type: '',
    limit: 50
  });

  useEffect(() => {
    fetchDLQData();
  }, [filter]);

  const fetchDLQData = async () => {
    try {
      setLoading(true);
      
      // Fetch entries
      const entriesParams = new URLSearchParams();
      if (filter.status) entriesParams.set('status', filter.status);
      if (filter.workflow_type) entriesParams.set('workflow_type', filter.workflow_type);
      if (filter.limit) entriesParams.set('limit', filter.limit.toString());
      
      const entriesResponse = await fetch(`/api/v1/dlq?${entriesParams}`);
      if (!entriesResponse.ok) {
        throw new Error(`Failed to fetch DLQ entries: ${entriesResponse.statusText}`);
      }
      const entriesData = await entriesResponse.json();
      setEntries(entriesData.entries || []);

      // Fetch stats
      const statsResponse = await fetch('/api/v1/dlq/stats');
      if (statsResponse.ok) {
        const statsData = await statsResponse.json();
        setStats(statsData);
      }

      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load DLQ data');
    } finally {
      setLoading(false);
    }
  };

  const handleRetry = async (dlqId: string, sandboxMode: boolean = true) => {
    try {
      const response = await fetch(`/api/v1/dlq/${dlqId}/retry`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ sandbox_mode: sandboxMode }),
      });

      if (!response.ok) {
        throw new Error(`Failed to retry DLQ entry: ${response.statusText}`);
      }

      const result = await response.json();
      alert(`Retry initiated successfully. New run ID: ${result.retry_run_id}`);
      
      // Refresh data
      await fetchDLQData();
    } catch (err) {
      alert(`Error: ${err instanceof Error ? err.message : 'Failed to retry'}`);
    }
  };

  const handleArchive = async (dlqId: string) => {
    if (!confirm('Are you sure you want to archive this DLQ entry?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/dlq/${dlqId}/archive`, {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error(`Failed to archive DLQ entry: ${response.statusText}`);
      }

      alert('DLQ entry archived successfully');
      
      // Refresh data
      await fetchDLQData();
    } catch (err) {
      alert(`Error: ${err instanceof Error ? err.message : 'Failed to archive'}`);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-red-100 text-red-800';
      case 'retrying': return 'bg-yellow-100 text-yellow-800';
      case 'resolved': return 'bg-green-100 text-green-800';
      case 'archived': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-600">Loading DLQ data...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4">
        <div className="text-red-800">Error: {error}</div>
        <Button 
          onClick={fetchDLQData} 
          className="mt-2"
          variant="outline"
          size="sm"
        >
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header and Stats */}
      <div className="flex justify-between items-start">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Dead Letter Queue</h2>
          <p className="text-gray-600">Manage failed workflow executions</p>
        </div>
        
        {stats && (
          <Card className="p-4 min-w-[200px]">
            <h3 className="font-semibold mb-2">Statistics</h3>
            <div className="space-y-1 text-sm">
              <div>Total: {stats.total_entries}</div>
              <div>Active: {stats.by_status.active || 0}</div>
              <div>Retrying: {stats.by_status.retrying || 0}</div>
              <div>Resolved: {stats.by_status.resolved || 0}</div>
              <div>Archived: {stats.by_status.archived || 0}</div>
            </div>
          </Card>
        )}
      </div>

      {/* Filters */}
      <Card className="p-4">
        <div className="flex gap-4 items-end">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Status
            </label>
            <select
              value={filter.status}
              onChange={(e) => setFilter({ ...filter, status: e.target.value })}
              className="border border-gray-300 rounded-md px-3 py-2 text-sm"
            >
              <option value="">All</option>
              <option value="active">Active</option>
              <option value="retrying">Retrying</option>
              <option value="resolved">Resolved</option>
              <option value="archived">Archived</option>
            </select>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Limit
            </label>
            <select
              value={filter.limit}
              onChange={(e) => setFilter({ ...filter, limit: parseInt(e.target.value) })}
              className="border border-gray-300 rounded-md px-3 py-2 text-sm"
            >
              <option value={10}>10</option>
              <option value={25}>25</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
            </select>
          </div>

          <Button onClick={fetchDLQData} variant="outline" size="sm">
            Refresh
          </Button>
        </div>
      </Card>

      {/* Entries List */}
      {entries.length === 0 ? (
        <Card className="p-8 text-center">
          <div className="text-gray-600">No DLQ entries found</div>
        </Card>
      ) : (
        <div className="space-y-4">
          {entries.map((entry) => (
            <Card key={entry.id} className="p-4">
              <div className="flex justify-between items-start">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <span className="font-mono text-sm text-gray-600">
                      {entry.id.substring(0, 8)}
                    </span>
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(entry.status)}`}>
                      {entry.status}
                    </span>
                    <span className="text-sm text-gray-500">
                      {entry.metadata.workflow_type}
                    </span>
                  </div>
                  
                  <h3 className="font-semibold text-gray-900 mb-1">
                    {entry.envelope.intent.name}
                  </h3>
                  
                  <p className="text-sm text-gray-600 mb-2">
                    {entry.envelope.intent.description}
                  </p>
                  
                  <div className="flex gap-4 text-xs text-gray-500">
                    <span>Run ID: {entry.run_id.substring(0, 12)}</span>
                    <span>Attempts: {entry.error_details.attempts}</span>
                    <span>Created: {formatDate(entry.created_at)}</span>
                  </div>
                  
                  {entry.error_details.last_error && (
                    <div className="mt-2 p-2 bg-red-50 border border-red-200 rounded text-sm text-red-700">
                      <strong>Last Error:</strong> {entry.error_details.last_error}
                    </div>
                  )}
                </div>
                
                <div className="flex gap-2 ml-4">
                  <Button
                    onClick={() => setSelectedEntry(entry)}
                    variant="outline"
                    size="sm"
                  >
                    View Details
                  </Button>
                  
                  {entry.status === 'active' && (
                    <Button
                      onClick={() => handleRetry(entry.id, true)}
                      size="sm"
                    >
                      Retry (Sandbox)
                    </Button>
                  )}
                  
                  {entry.status !== 'archived' && (
                    <Button
                      onClick={() => handleArchive(entry.id)}
                      variant="outline"
                      size="sm"
                    >
                      Archive
                    </Button>
                  )}
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Entry Details Modal */}
      {selectedEntry && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <Card className="max-w-4xl max-h-[90vh] overflow-auto p-6">
            <div className="flex justify-between items-start mb-4">
              <h3 className="text-xl font-bold">DLQ Entry Details</h3>
              <Button
                onClick={() => setSelectedEntry(null)}
                variant="outline"
                size="sm"
              >
                Close
              </Button>
            </div>
            
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <strong>ID:</strong> {selectedEntry.id}
                </div>
                <div>
                  <strong>Status:</strong> 
                  <span className={`ml-2 px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(selectedEntry.status)}`}>
                    {selectedEntry.status}
                  </span>
                </div>
                <div>
                  <strong>Run ID:</strong> {selectedEntry.run_id}
                </div>
                <div>
                  <strong>Workflow Type:</strong> {selectedEntry.metadata.workflow_type}
                </div>
                <div>
                  <strong>Created:</strong> {formatDate(selectedEntry.created_at)}
                </div>
                <div>
                  <strong>Retry Count:</strong> {selectedEntry.metadata.retry_count}
                </div>
              </div>

              <div>
                <strong>Workflow Intent:</strong>
                <div className="mt-1 p-3 bg-gray-50 rounded">
                  <div><strong>Name:</strong> {selectedEntry.envelope.intent.name}</div>
                  <div><strong>Description:</strong> {selectedEntry.envelope.intent.description}</div>
                </div>
              </div>

              <div>
                <strong>Error Details:</strong>
                <div className="mt-1 p-3 bg-red-50 border border-red-200 rounded">
                  <div><strong>Attempts:</strong> {selectedEntry.error_details.attempts}</div>
                  <div><strong>Failure Reason:</strong> {selectedEntry.error_details.failure_reason}</div>
                  <div><strong>Last Error:</strong> {selectedEntry.error_details.last_error}</div>
                </div>
              </div>

              <div className="flex gap-2 pt-4">
                {selectedEntry.status === 'active' && (
                  <>
                    <Button
                      onClick={() => {
                        handleRetry(selectedEntry.id, true);
                        setSelectedEntry(null);
                      }}
                    >
                      Retry in Sandbox
                    </Button>
                    <Button
                      onClick={() => {
                        handleRetry(selectedEntry.id, false);
                        setSelectedEntry(null);
                      }}
                      variant="outline"
                    >
                      Retry in Production
                    </Button>
                  </>
                )}
                
                {selectedEntry.status !== 'archived' && (
                  <Button
                    onClick={() => {
                      handleArchive(selectedEntry.id);
                      setSelectedEntry(null);
                    }}
                    variant="outline"
                  >
                    Archive
                  </Button>
                )}
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}
