import { MovaClient } from '@mova-engine/sdk-typescript';

// Create MOVA client instance
const movaClient = new MovaClient({
  baseUrl: process.env.NEXT_PUBLIC_MOVA_API_URL || 'http://localhost:8080',
  timeout: 30000,
});

export { movaClient };

// Types for the API responses
export interface RunRecord {
  id: string;
  intent: string;
  status: 'running' | 'completed' | 'failed';
  created_at: string;
  completed_at?: string;
  duration?: number;
  actions_count: number;
}

export interface RunDetails extends RunRecord {
  envelope: any;
  result?: any;
  logs: LogEntry[];
}

export interface LogEntry {
  timestamp: string;
  level: 'info' | 'error' | 'debug';
  message: string;
  action?: string;
  context?: any;
}

// API functions using the MOVA SDK
export const api = {
  // Execute an envelope
  async executeEnvelope(envelope: any) {
    try {
      const result = await movaClient.execute(envelope);
      return result;
    } catch (error) {
      console.error('Failed to execute envelope:', error);
      throw error;
    }
  },

  // Validate an envelope
  async validateEnvelope(envelope: any) {
    try {
      const result = await movaClient.validate(envelope);
      return result;
    } catch (error) {
      console.error('Failed to validate envelope:', error);
      throw error;
    }
  },

  // Get list of runs
  async getRuns(): Promise<RunRecord[]> {
    try {
      // TODO: Implement when SDK supports runs listing
      // For now, return mock data
      return [];
    } catch (error) {
      console.error('Failed to fetch runs:', error);
      throw error;
    }
  },

  // Get run details by ID
  async getRunDetails(id: string): Promise<RunDetails | null> {
    try {
      // TODO: Implement when SDK supports run details
      // For now, return null
      return null;
    } catch (error) {
      console.error('Failed to fetch run details:', error);
      throw error;
    }
  },

  // Get system status
  async getSystemStatus() {
    try {
      // TODO: Implement when SDK supports system status
      return {
        api: 'healthy',
        executor: 'healthy', 
        validator: 'healthy'
      };
    } catch (error) {
      console.error('Failed to fetch system status:', error);
      throw error;
    }
  }
};
