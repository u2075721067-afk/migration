import { MOVAClient, MOVAEnvelope } from './index';

// Mock fetch
const mockFetch = global.fetch as jest.MockedFunction<typeof fetch>;

describe('MOVAClient', () => {
  let client: MOVAClient;
  
  const mockEnvelope: MOVAEnvelope = {
    mova_version: '3.1',
    intent: {
      name: 'test-workflow',
      version: '1.0.0',
      description: 'Test workflow',
    },
    payload: { test: 'data' },
    actions: [
      {
        type: 'set',
        name: 'test-action',
        config: {
          variable: 'test_var',
          value: 'test_value',
        },
      },
    ],
  };

  beforeEach(() => {
    client = new MOVAClient({ baseURL: 'http://localhost:8080' });
    jest.clearAllMocks();
  });

  describe('execute', () => {
    it('should execute workflow synchronously', async () => {
      const mockResult = {
        run_id: 'test-run-123',
        workflow_id: 'test-workflow',
        status: 'completed',
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-01-01T00:01:00Z',
        variables: { test_var: 'test_value' },
        results: {},
        logs: [],
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResult,
      } as Response);

      const result = await client.execute(mockEnvelope, true);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/v1/execute?wait=true',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(mockEnvelope),
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
        })
      );

      expect(result).toEqual(mockResult);
    });

    it('should execute workflow asynchronously', async () => {
      const mockResult = {
        run_id: 'test-run-123',
        status: 'accepted',
        message: 'Execution started asynchronously',
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResult,
      } as Response);

      const result = await client.execute(mockEnvelope, false);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/v1/execute',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(mockEnvelope),
        })
      );

      expect(result).toEqual(mockResult);
    });

    it('should handle execution errors', async () => {
      const mockError = {
        error: 'Validation failed',
        details: 'Invalid envelope structure',
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        json: async () => mockError,
      } as Response);

      await expect(client.execute(mockEnvelope)).rejects.toThrow(
        'Execution failed: Validation failed'
      );
    });
  });

  describe('validate', () => {
    it('should validate envelope successfully', async () => {
      const mockResult = {
        valid: true,
        message: 'Envelope is valid',
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResult,
      } as Response);

      const result = await client.validate(mockEnvelope);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/v1/validate',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(mockEnvelope),
        })
      );

      expect(result).toEqual(mockResult);
    });

    it('should handle validation errors', async () => {
      const mockError = {
        valid: false,
        message: 'Validation failed',
        errors: ['Missing required field: intent'],
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockError,
      } as Response);

      const result = await client.validate(mockEnvelope);

      expect(result).toEqual(mockError);
    });
  });

  describe('getRun', () => {
    it('should get run status successfully', async () => {
      const runId = 'test-run-123';
      const mockResult = {
        run_id: runId,
        workflow_id: 'test-workflow',
        status: 'completed',
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-01-01T00:01:00Z',
        variables: {},
        results: {},
        logs: [],
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResult,
      } as Response);

      const result = await client.getRun(runId);

      expect(mockFetch).toHaveBeenCalledWith(
        `http://localhost:8080/v1/runs/${runId}`,
        expect.objectContaining({
          method: 'GET',
        })
      );

      expect(result).toEqual(mockResult);
    });

    it('should handle run not found', async () => {
      const runId = 'non-existent-run';
      const mockError = {
        error: 'Run not found',
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: async () => mockError,
      } as Response);

      await expect(client.getRun(runId)).rejects.toThrow(
        'Failed to get run: Run not found'
      );
    });
  });

  describe('getLogs', () => {
    it('should get logs successfully', async () => {
      const runId = 'test-run-123';
      const mockLogs = [
        '{"timestamp":"2024-01-01T00:00:00Z","message":"Test log 1"}',
        '{"timestamp":"2024-01-01T00:00:01Z","message":"Test log 2"}',
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: async () => mockLogs.join('\n'),
      } as Response);

      const result = await client.getLogs(runId);

      expect(mockFetch).toHaveBeenCalledWith(
        `http://localhost:8080/v1/runs/${runId}/logs`,
        expect.objectContaining({
          method: 'GET',
        })
      );

      expect(result).toEqual(mockLogs);
    });

    it('should handle empty logs', async () => {
      const runId = 'test-run-123';

      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: async () => '',
      } as Response);

      const result = await client.getLogs(runId);

      expect(result).toEqual([]);
    });
  });

  describe('getSchemas', () => {
    it('should get schemas successfully', async () => {
      const mockSchemas = {
        schemas: [
          {
            name: 'envelope',
            version: '3.1',
            description: 'MOVA v3.1 envelope schema',
            url: '/v1/schemas/envelope',
          },
        ],
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockSchemas,
      } as Response);

      const result = await client.getSchemas();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/v1/schemas',
        expect.objectContaining({
          method: 'GET',
        })
      );

      expect(result).toEqual(mockSchemas);
    });
  });

  describe('getSchema', () => {
    it('should get specific schema successfully', async () => {
      const schemaName = 'envelope';
      const mockSchema = {
        $schema: 'http://json-schema.org/draft-07/schema#',
        type: 'object',
        properties: {
          mova_version: { type: 'string' },
          intent: { type: 'object' },
          payload: { type: 'object' },
          actions: { type: 'array' },
        },
        required: ['mova_version', 'intent', 'payload', 'actions'],
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockSchema,
      } as Response);

      const result = await client.getSchema(schemaName);

      expect(mockFetch).toHaveBeenCalledWith(
        `http://localhost:8080/v1/schemas/${schemaName}`,
        expect.objectContaining({
          method: 'GET',
        })
      );

      expect(result).toEqual(mockSchema);
    });
  });

  describe('introspect', () => {
    it('should get API information successfully', async () => {
      const mockInfo = {
        name: 'MOVA Automation Engine API',
        version: '1.0.0',
        mova_version: '3.1',
        supported_actions: ['set', 'http_fetch', 'parse_json', 'sleep'],
        endpoints: [],
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockInfo,
      } as Response);

      const result = await client.introspect();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/v1/introspect',
        expect.objectContaining({
          method: 'GET',
        })
      );

      expect(result).toEqual(mockInfo);
    });
  });

  describe('timeout handling', () => {
    it.skip('should handle request timeout', async () => {
      // TODO: Implement proper timeout testing with AbortController
      const client = new MOVAClient({ timeout: 100 });
      expect(client).toBeDefined();
    });
  });

  describe('custom headers', () => {
    it('should use custom headers', async () => {
      const client = new MOVAClient({
        headers: { 'X-Custom-Header': 'test-value' },
      });

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      } as Response);

      await client.execute(mockEnvelope);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
            'X-Custom-Header': 'test-value',
          }),
        })
      );
    });
  });
});
