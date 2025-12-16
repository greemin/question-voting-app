// /frontend/tests/setup.ts
import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock ResizeObserver
const ResizeObserverMock = vi.fn(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

vi.stubGlobal('ResizeObserver', ResizeObserverMock);

// Mock window.alert
vi.spyOn(window, 'alert').mockImplementation(() => {});
