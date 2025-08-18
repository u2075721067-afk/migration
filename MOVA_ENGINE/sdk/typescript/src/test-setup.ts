// Test setup file for Jest
// Add global test utilities and mocks here

// Mock fetch globally for tests
global.fetch = jest.fn();

// Reset mocks before each test
beforeEach(() => {
  jest.resetAllMocks();
});
