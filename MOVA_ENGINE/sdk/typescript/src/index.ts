export interface MOVAEnvelope {
  mova_version: string;
  intent: {
    name: string;
    version: string;
    description?: string;
    author?: string;
    tags?: string[];
    timeout?: number;
    retry?: {
      count: number;
      backoff_ms: number;
    };
    budget?: {
      tokens?: number;
      cost_usd?: number;
    };
  };
  payload: Record<string, any>;
  actions: Array<{
    type: string;
    name: string;
    description?: string;
    enabled?: boolean;
    timeout?: number;
    retry?: {
      count: number;
      backoff_ms: number;
    };
    config?: Record<string, any>;
  }>;
  variables?: Record<string, any>;
  secrets?: Record<string, string>;
}

export interface ExecutionResult {
  run_id: string;
  workflow_id: string;
  start_time: string;
  end_time?: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  variables: Record<string, any>;
  results: Record<string, any>;
  logs: Array<{
    timestamp: string;
    level: string;
    step: string;
    type: string;
    action?: string;
    message: string;
    params_redacted?: Record<string, any>;
    status: string;
    data?: Record<string, any>;
  }>;
}

export interface ValidationResult {
  valid: boolean;
  message: string;
  errors?: string[];
}

export interface MOVAClientOptions {
  baseURL?: string;
  timeout?: number;
  headers?: Record<string, string>;
}

export class MOVAClient {
  private baseURL: string;
  private timeout: number;
  private headers: Record<string, string>;

  constructor(options: MOVAClientOptions = {}) {
    this.baseURL = options.baseURL || 'http://localhost:8080';
    this.timeout = options.timeout || 30000;
    this.headers = {
      'Content-Type': 'application/json',
      ...options.headers,
    };
  }

  /**
   * Execute a MOVA workflow envelope
   * @param envelope - The MOVA envelope to execute
   * @param wait - Whether to wait for execution to complete (default: false)
   * @returns Promise<ExecutionResult>
   */
  async execute(envelope: MOVAEnvelope, wait: boolean = false): Promise<ExecutionResult> {
    const url = `${this.baseURL}/v1/execute${wait ? '?wait=true' : ''}`;
    
    const response = await this.fetch(url, {
      method: 'POST',
      body: JSON.stringify(envelope),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Execution failed: ${error.error || response.statusText}`);
    }

    return await response.json();
  }

  /**
   * Validate a MOVA envelope against the schema
   * @param envelope - The MOVA envelope to validate
   * @returns Promise<ValidationResult>
   */
  async validate(envelope: MOVAEnvelope): Promise<ValidationResult> {
    const url = `${this.baseURL}/v1/validate`;
    
    const response = await this.fetch(url, {
      method: 'POST',
      body: JSON.stringify(envelope),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Validation request failed: ${error.error || response.statusText}`);
    }

    return await response.json();
  }

  /**
   * Get the status and result of a workflow execution
   * @param runId - The run ID to retrieve
   * @returns Promise<ExecutionResult>
   */
  async getRun(runId: string): Promise<ExecutionResult> {
    const url = `${this.baseURL}/v1/runs/${runId}`;
    
    const response = await this.fetch(url, {
      method: 'GET',
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Failed to get run: ${error.error || response.statusText}`);
    }

    return await response.json();
  }

  /**
   * Get the logs for a workflow execution
   * @param runId - The run ID to retrieve logs for
   * @returns Promise<string[]> - Array of JSONL log entries
   */
  async getLogs(runId: string): Promise<string[]> {
    const url = `${this.baseURL}/v1/runs/${runId}/logs`;
    
    const response = await this.fetch(url, {
      method: 'GET',
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Failed to get logs: ${error.error || response.statusText}`);
    }

    const text = await response.text();
    return text.trim().split('\n').filter(line => line.length > 0);
  }

  /**
   * Get available schemas
   * @returns Promise<any>
   */
  async getSchemas(): Promise<any> {
    const url = `${this.baseURL}/v1/schemas`;
    
    const response = await this.fetch(url, {
      method: 'GET',
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Failed to get schemas: ${error.error || response.statusText}`);
    }

    return await response.json();
  }

  /**
   * Get a specific schema by name
   * @param name - Schema name (e.g., 'envelope', 'action')
   * @returns Promise<any>
   */
  async getSchema(name: string): Promise<any> {
    const url = `${this.baseURL}/v1/schemas/${name}`;
    
    const response = await this.fetch(url, {
      method: 'GET',
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Failed to get schema: ${error.error || response.statusText}`);
    }

    return await response.json();
  }

  /**
   * Get API introspection information
   * @returns Promise<any>
   */
  async introspect(): Promise<any> {
    const url = `${this.baseURL}/v1/introspect`;
    
    const response = await this.fetch(url, {
      method: 'GET',
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Failed to introspect: ${error.error || response.statusText}`);
    }

    return await response.json();
  }

  private async fetch(url: string, options: RequestInit): Promise<Response> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          ...this.headers,
          ...options.headers,
        },
        signal: controller.signal,
      });

      clearTimeout(timeoutId);
      return response;
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === 'AbortError') {
        throw new Error(`Request timeout after ${this.timeout}ms`);
      }
      throw error;
    }
  }
}

// Export default instance
export const mova = new MOVAClient();

// Export everything
export default MOVAClient;
