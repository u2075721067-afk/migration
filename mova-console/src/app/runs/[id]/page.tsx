'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ArrowLeft, Calendar, Clock, CheckCircle, XCircle, Loader, FileText } from 'lucide-react';
import Link from 'next/link';
import { useParams } from 'next/navigation';

interface LogEntry {
  timestamp: string;
  level: 'info' | 'error' | 'debug';
  message: string;
  action?: string;
  context?: any;
}

interface RunDetails {
  id: string;
  intent: string;
  status: 'running' | 'completed' | 'failed';
  created_at: string;
  completed_at?: string;
  duration?: number;
  actions_count: number;
  envelope: any;
  result?: any;
  logs: LogEntry[];
}

const mockRunDetails: RunDetails = {
  id: 'run-001',
  intent: 'demo-workflow',
  status: 'completed',
  created_at: '2024-01-15T10:30:00Z',
  completed_at: '2024-01-15T10:30:02.500Z',
  duration: 2500,
  actions_count: 3,
  envelope: {
    intent: 'demo-workflow',
    payload: { message: 'Hello MOVA' },
    actions: [
      { type: 'set', key: 'greeting', value: 'Hello from MOVA Engine!' },
      { type: 'sleep', duration: 1000 }
    ]
  },
  result: {
    status: 'success',
    context: {
      greeting: 'Hello from MOVA Engine!',
      message: 'Hello MOVA'
    },
    execution_time: 2500
  },
  logs: [
    {
      timestamp: '2024-01-15T10:30:00.100Z',
      level: 'info',
      message: 'Starting envelope execution',
      action: 'start',
      context: { intent: 'demo-workflow' }
    },
    {
      timestamp: '2024-01-15T10:30:00.150Z',
      level: 'info', 
      message: 'Executing action: set',
      action: 'set',
      context: { key: 'greeting', value: 'Hello from MOVA Engine!' }
    },
    {
      timestamp: '2024-01-15T10:30:00.160Z',
      level: 'info',
      message: 'Action completed successfully',
      action: 'set'
    },
    {
      timestamp: '2024-01-15T10:30:01.160Z',
      level: 'info',
      message: 'Executing action: sleep',
      action: 'sleep',
      context: { duration: 1000 }
    },
    {
      timestamp: '2024-01-15T10:30:02.160Z',
      level: 'info',
      message: 'Action completed successfully',
      action: 'sleep'
    },
    {
      timestamp: '2024-01-15T10:30:02.500Z',
      level: 'info',
      message: 'Envelope execution completed',
      action: 'complete',
      context: { duration: 2500, status: 'success' }
    }
  ]
};

const StatusIcon = ({ status }: { status: RunDetails['status'] }) => {
  switch (status) {
    case 'completed':
      return <CheckCircle className="h-5 w-5 text-green-600" />;
    case 'failed':
      return <XCircle className="h-5 w-5 text-red-600" />;
    case 'running':
      return <Loader className="h-5 w-5 text-blue-600 animate-spin" />;
    default:
      return null;
  }
};

const LogLevelBadge = ({ level }: { level: LogEntry['level'] }) => {
  const colors = {
    info: 'bg-blue-100 text-blue-800',
    error: 'bg-red-100 text-red-800',
    debug: 'bg-gray-100 text-gray-800'
  };
  
  return (
    <span className={`px-2 py-1 rounded text-xs font-medium ${colors[level]}`}>
      {level.toUpperCase()}
    </span>
  );
};

const formatDuration = (ms?: number) => {
  if (!ms) return '-';
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
};

const formatDate = (dateString: string) => {
  const date = new Date(dateString);
  return date.toLocaleString();
};

export default function RunDetailsPage() {
  const params = useParams();
  const runId = params.id as string;
  const [runDetails, setRunDetails] = useState<RunDetails | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // TODO: Replace with actual API call using runId
    setTimeout(() => {
      setRunDetails({ ...mockRunDetails, id: runId });
      setLoading(false);
    }, 1000);
  }, [runId]);

  if (loading) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="flex items-center justify-center h-64">
          <Loader className="h-8 w-8 animate-spin" />
        </div>
      </div>
    );
  }

  if (!runDetails) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-8">
          <h2 className="text-xl font-semibold mb-2">Run not found</h2>
          <p className="text-muted-foreground mb-4">
            The run with ID "{runId}" could not be found.
          </p>
          <Button asChild>
            <Link href="/runs">Back to Runs</Link>
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="mb-6">
        <div className="flex items-center gap-4 mb-4">
          <Button variant="outline" size="sm" asChild>
            <Link href="/runs" className="flex items-center gap-2">
              <ArrowLeft className="h-4 w-4" />
              Back to Runs
            </Link>
          </Button>
        </div>
        <div className="flex items-center gap-3 mb-2">
          <StatusIcon status={runDetails.status} />
          <h1 className="text-3xl font-bold">Run {runDetails.id}</h1>
        </div>
        <p className="text-muted-foreground">
          Intent: {runDetails.intent} • {runDetails.actions_count} actions • {formatDuration(runDetails.duration)}
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              Envelope
            </CardTitle>
            <CardDescription>
              Original envelope that was executed
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="bg-muted p-4 rounded-md text-sm overflow-auto max-h-96">
              <pre className="language-json">
                <code className="language-json">
                  {JSON.stringify(runDetails.envelope, null, 2)}
                </code>
              </pre>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Result</CardTitle>
            <CardDescription>
              Execution result and final context
            </CardDescription>
          </CardHeader>
          <CardContent>
            {runDetails.result ? (
              <div className="bg-muted p-4 rounded-md text-sm overflow-auto max-h-96">
                <pre className="language-json">
                  <code className="language-json">
                    {JSON.stringify(runDetails.result, null, 2)}
                  </code>
                </pre>
              </div>
            ) : (
              <div className="text-muted-foreground text-center py-8">
                {runDetails.status === 'running' ? 'Execution in progress...' : 'No result available'}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Execution Logs</CardTitle>
          <CardDescription>
            Detailed JSONL logs from the execution
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 max-h-96 overflow-auto">
            {runDetails.logs.map((log, index) => (
              <div key={index} className="border rounded-lg p-3 bg-muted/50">
                <div className="flex items-center gap-2 mb-2">
                  <LogLevelBadge level={log.level} />
                  <span className="text-xs text-muted-foreground font-mono">
                    {new Date(log.timestamp).toLocaleTimeString()}
                  </span>
                  {log.action && (
                    <span className="text-xs bg-primary/10 text-primary px-2 py-1 rounded">
                      {log.action}
                    </span>
                  )}
                </div>
                <p className="text-sm mb-2">{log.message}</p>
                {log.context && (
                  <div className="text-xs bg-background p-2 rounded border overflow-auto">
                    <pre className="language-json">
                      <code className="language-json">
                        {JSON.stringify(log.context, null, 2)}
                      </code>
                    </pre>
                  </div>
                )}
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <div className="mt-6 grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-lg flex items-center gap-2">
              <Calendar className="h-4 w-4" />
              Timeline
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span>Started:</span>
                <span className="text-muted-foreground">
                  {formatDate(runDetails.created_at)}
                </span>
              </div>
              {runDetails.completed_at && (
                <div className="flex justify-between">
                  <span>Completed:</span>
                  <span className="text-muted-foreground">
                    {formatDate(runDetails.completed_at)}
                  </span>
                </div>
              )}
              <div className="flex justify-between">
                <span>Duration:</span>
                <span className="text-muted-foreground">
                  {formatDuration(runDetails.duration)}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-lg flex items-center gap-2">
              <Clock className="h-4 w-4" />
              Performance
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span>Actions:</span>
                <span className="text-muted-foreground">{runDetails.actions_count}</span>
              </div>
              <div className="flex justify-between">
                <span>Avg/Action:</span>
                <span className="text-muted-foreground">
                  {runDetails.duration ? formatDuration(runDetails.duration / runDetails.actions_count) : '-'}
                </span>
              </div>
              <div className="flex justify-between">
                <span>Status:</span>
                <span className={`capitalize ${
                  runDetails.status === 'completed' ? 'text-green-600' :
                  runDetails.status === 'failed' ? 'text-red-600' : 'text-blue-600'
                }`}>
                  {runDetails.status}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-lg">Actions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Button variant="outline" className="w-full justify-start" disabled>
                Re-run Envelope
              </Button>
              <Button variant="outline" className="w-full justify-start" disabled>
                Export Logs
              </Button>
              <Button variant="outline" className="w-full justify-start" disabled>
                Download Result
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
