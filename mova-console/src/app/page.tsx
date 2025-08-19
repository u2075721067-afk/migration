'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Textarea } from '@/components/ui/textarea';
import { Play, FileText } from 'lucide-react';
import Link from 'next/link';

const defaultEnvelope = `{
  "intent": "demo-workflow",
  "payload": {
    "message": "Hello MOVA"
  },
  "actions": [
    {
      "type": "set",
      "key": "greeting",
      "value": "Hello from MOVA Engine!"
    },
    {
      "type": "sleep",
      "duration": 1000
    }
  ]
}`;

export default function Home() {
  const [envelope, setEnvelope] = useState(defaultEnvelope);
  const [isRunning, setIsRunning] = useState(false);
  const [result, setResult] = useState<any>(null);

  const handleRun = async () => {
    setIsRunning(true);
    try {
      const envelopeObj = JSON.parse(envelope);
      const { api } = await import('@/lib/api');
      const result = await api.executeEnvelope(envelopeObj);
      setResult(result);
    } catch (error) {
      if (error instanceof SyntaxError) {
        setResult({ error: 'Invalid JSON format' });
      } else {
        setResult({ error: error instanceof Error ? error.message : 'Execution failed' });
      }
    } finally {
      setIsRunning(false);
    }
  };

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">MOVA Console</h1>
        <p className="text-muted-foreground">
          Web interface for MOVA Automation Engine - Execute envelopes and monitor workflows
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              Envelope Editor
            </CardTitle>
            <CardDescription>
              Enter your MOVA envelope JSON to execute
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Textarea
              value={envelope}
              onChange={(e) => setEnvelope(e.target.value)}
              className="min-h-[400px] font-mono text-sm"
              placeholder="Enter your MOVA envelope JSON..."
            />
            <div className="mt-4 flex gap-2">
              <Button 
                onClick={handleRun} 
                disabled={isRunning}
                className="flex items-center gap-2"
              >
                <Play className="h-4 w-4" />
                {isRunning ? 'Running...' : 'Run Envelope'}
              </Button>
              <Button variant="outline" asChild>
                <Link href="/runs">View Runs</Link>
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Result</CardTitle>
            <CardDescription>
              Execution result and output
            </CardDescription>
          </CardHeader>
          <CardContent>
            {result ? (
              <div className="bg-muted p-4 rounded-md text-sm overflow-auto max-h-[400px]">
                <pre className="language-json">
                  <code className="language-json">
                    {JSON.stringify(result, null, 2)}
                  </code>
                </pre>
              </div>
            ) : (
              <div className="text-muted-foreground text-center py-20">
                Click "Run Envelope" to see results here
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-lg">Quick Actions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Button variant="outline" className="w-full justify-start" asChild>
                <Link href="/runs">View All Runs</Link>
              </Button>
              <Button variant="outline" className="w-full justify-start" disabled>
                Load Template
              </Button>
              <Button variant="outline" className="w-full justify-start" disabled>
                Validate Schema
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-lg">Recent Runs</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-muted-foreground">
              No recent runs available
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-lg">System Status</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>API Server</span>
                <span className="text-green-600">●</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Executor</span>
                <span className="text-green-600">●</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Validator</span>
                <span className="text-green-600">●</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}