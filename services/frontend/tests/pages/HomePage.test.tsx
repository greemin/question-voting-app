// /frontend/tests/pages/HomePage.test.tsx
import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/preact';
import { BrowserRouter } from 'react-router-dom';
import { vi } from 'vitest';
import HomePage from '../../src/pages/HomePage';
import * as sessionApi from '../../src/api/sessionApi';

vi.mock('react-router-dom', () => ({
  BrowserRouter: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useNavigate: () => vi.fn(),
}));

// Mock the sessionApi
vi.mock('../../src/api/sessionApi');

describe('HomePage', () => {
  it('renders the main heading and button', () => {
    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );
    expect(screen.getByText('Question Voting App')).toBeInTheDocument();
    expect(screen.getByText('🚀 Start New Voting Session')).toBeInTheDocument();
  });

  it('calls createSession and navigates on button click', async () => {
    let resolveCreateSession;
    const createSessionPromise = new Promise(resolve => {
      resolveCreateSession = resolve;
    });
    const mockedCreateSession = vi.spyOn(sessionApi, 'createSession').mockReturnValue(createSessionPromise);
    
    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    act(() => {
      fireEvent.click(screen.getByText('🚀 Start New Voting Session'));
    });

    expect(screen.getByText('Creating...')).toBeInTheDocument();
    expect(mockedCreateSession).toHaveBeenCalledTimes(1);

    await act(async () => {
      resolveCreateSession({ sessionId: '123' });
      await createSessionPromise;
    });

    // After the promise resolves, the loading should be false again
    expect(screen.getByText('🚀 Start New Voting Session')).toBeInTheDocument();
  });
});
