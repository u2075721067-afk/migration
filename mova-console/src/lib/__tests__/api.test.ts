import { api } from '../api'
import { MovaClient } from '@mova-engine/sdk-typescript'

// Mock the MOVA SDK
jest.mock('@mova-engine/sdk-typescript')

const mockMovaClient = {
  execute: jest.fn(),
  validate: jest.fn(),
}

const MockedMovaClient = MovaClient as jest.MockedClass<typeof MovaClient>

describe('API Client', () => {
  beforeEach(() => {
    MockedMovaClient.mockImplementation(() => mockMovaClient as any)
    mockMovaClient.execute.mockClear()
    mockMovaClient.validate.mockClear()
    jest.clearAllMocks()
  })

  describe('executeEnvelope', () => {
    it('calls MOVA client execute method', async () => {
      const envelope = {
        intent: 'test-workflow',
        payload: { test: true },
        actions: [{ type: 'set', key: 'test', value: 'value' }]
      }
      
      const expectedResult = { status: 'success', runId: 'run-123' }
      mockMovaClient.execute.mockResolvedValue(expectedResult)

      const result = await api.executeEnvelope(envelope)

      expect(mockMovaClient.execute).toHaveBeenCalledWith(envelope)
      expect(result).toEqual(expectedResult)
    })

    it('handles execution errors', async () => {
      const envelope = { intent: 'test', payload: {}, actions: [] }
      const error = new Error('Execution failed')
      
      mockMovaClient.execute.mockRejectedValue(error)
      
      await expect(api.executeEnvelope(envelope)).rejects.toThrow('Execution failed')
    })
  })

  describe('validateEnvelope', () => {
    it('calls MOVA client validate method', async () => {
      const envelope = {
        intent: 'test-workflow',
        payload: { test: true },
        actions: [{ type: 'set', key: 'test', value: 'value' }]
      }
      
      const expectedResult = { valid: true, errors: [] }
      mockMovaClient.validate.mockResolvedValue(expectedResult)

      const result = await api.validateEnvelope(envelope)

      expect(mockMovaClient.validate).toHaveBeenCalledWith(envelope)
      expect(result).toEqual(expectedResult)
    })

    it('handles validation errors', async () => {
      const envelope = { intent: 'invalid', payload: {}, actions: [] }
      const error = new Error('Validation failed')
      
      mockMovaClient.validate.mockRejectedValue(error)
      
      await expect(api.validateEnvelope(envelope)).rejects.toThrow('Validation failed')
    })
  })

  describe('placeholder methods', () => {
    it('getRuns returns empty array', async () => {
      const result = await api.getRuns()
      expect(result).toEqual([])
    })

    it('getRunDetails returns null', async () => {
      const result = await api.getRunDetails('run-123')
      expect(result).toBeNull()
    })

    it('getSystemStatus returns default status', async () => {
      const result = await api.getSystemStatus()
      
      expect(result).toEqual({
        api: 'healthy',
        executor: 'healthy',
        validator: 'healthy'
      })
    })
  })
})