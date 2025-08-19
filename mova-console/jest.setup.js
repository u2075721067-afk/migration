import '@testing-library/jest-dom'

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter() {
    return {
      push: jest.fn(),
      replace: jest.fn(),
      prefetch: jest.fn(),
    }
  },
  useSearchParams() {
    return new URLSearchParams()
  },
  usePathname() {
    return ''
  },
  useParams() {
    return {}
  },
}))

// Mock Prism.js
jest.mock('prismjs', () => ({
  highlightAll: jest.fn(),
}))

// Mock MOVA SDK
jest.mock('@mova-engine/sdk-typescript', () => ({
  MovaClient: jest.fn().mockImplementation(() => ({
    execute: jest.fn(),
    validate: jest.fn(),
  })),
}))

// API module will be mocked in individual test files

// Mock environment variables
process.env.NEXT_PUBLIC_MOVA_API_URL = 'http://localhost:8080'
