'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { ArrowLeft, Eye, Calendar, Clock, CheckCircle, XCircle, Loader } from 'lucide-react';
import Link from 'next/link';

interface RunRecord {
  id: string;
  intent: string;
  status: 'running' | 'completed' | 'failed';
  created_at: string;
  duration?: number;
  actions_count: number;
}

const mockRuns: RunRecord[] = [
  {
    id: 'run-001',
    intent: 'demo-workflow',
    status: 'completed',
    created_at: '2024-01-15T10:30:00Z',
    duration: 2500,
    actions_count: 3
  },
  {
    id: 'run-002', 
    intent: 'data-processing',
    status: 'running',
    created_at: '2024-01-15T10:25:00Z',
    actions_count: 5
  },
  {
    id: 'run-003',
    intent: 'api-integration',
    status: 'failed',
    created_at: '2024-01-15T10:20:00Z',
    duration: 1200,
    actions_count: 2
  }
];

const StatusIcon = ({ status }: { status: RunRecord['status'] }) => {
  switch (status) {
    case 'completed':
      return <CheckCircle className="h-4 w-4 text-green-600" />;
    case 'failed':
      return <XCircle className="h-4 w-4 text-red-600" />;
    case 'running':
      return <Loader className="h-4 w-4 text-blue-600 animate-spin" />;
    default:
      return null;
  }
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

export default function RunsPage() {
  const [runs, setRuns] = useState<RunRecord[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // TODO: Replace with actual API call
    setTimeout(() => {
      setRuns(mockRuns);
      setLoading(false);
    }, 1000);
  }, []);

  if (loading) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="flex items-center justify-center h-64">
          <Loader className="h-8 w-8 animate-spin" />
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="mb-6">
        <div className="flex items-center gap-4 mb-4">
          <Button variant="outline" size="sm" asChild>
            <Link href="/" className="flex items-center gap-2">
              <ArrowLeft className="h-4 w-4" />
              Back to Console
            </Link>
          </Button>
        </div>
        <h1 className="text-3xl font-bold mb-2">Execution Runs</h1>
        <p className="text-muted-foreground">
          View and monitor all MOVA envelope executions
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Calendar className="h-5 w-5" />
            Recent Runs
          </CardTitle>
          <CardDescription>
            All envelope executions ordered by creation time
          </CardDescription>
        </CardHeader>
        <CardContent>
          {runs.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No runs found. Execute an envelope to see results here.
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Status</TableHead>
                  <TableHead>Run ID</TableHead>
                  <TableHead>Intent</TableHead>
                  <TableHead>Actions</TableHead>
                  <TableHead>Duration</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {runs.map((run) => (
                  <TableRow key={run.id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <StatusIcon status={run.status} />
                        <span className="capitalize">{run.status}</span>
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {run.id}
                    </TableCell>
                    <TableCell>{run.intent}</TableCell>
                    <TableCell>{run.actions_count}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        {formatDuration(run.duration)}
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(run.created_at)}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button variant="ghost" size="sm" asChild>
                        <Link 
                          href={`/runs/${run.id}`}
                          className="flex items-center gap-1"
                        >
                          <Eye className="h-3 w-3" />
                          View
                        </Link>
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
