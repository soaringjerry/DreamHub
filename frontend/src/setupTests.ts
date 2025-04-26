import { vi } from 'vitest';

// Mock window.matchMedia for jsdom environment
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false, // Default to false or based on query if needed
    media: query,
    onchange: null,
    addListener: vi.fn(), // Deprecated but good to mock
    removeListener: vi.fn(), // Deprecated but good to mock
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// You can add other global test setup here if needed
// For example, importing jest-dom matchers:
// import '@testing-library/jest-dom';