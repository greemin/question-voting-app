// /frontend/tests/pages/VotingSessionPage.test.tsx
import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { vi, Mock } from 'vitest';
import VotingSessionPage from '../../src/pages/VotingSessionPage';
import * as sessionApi from '../../src/api/sessionApi';
import { Question } from '../../src/models/Question';
import { SessionData } from '../../src/models/SessionData';

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useParams: () => ({
      sessionId: 'test-session',
    }),
    useNavigate: () => vi.fn(),
  };
});

vi.mock('../../src/api/sessionApi', () => ({
  createSession: vi.fn(),
  getSessionData: vi.fn(),
  submitQuestion: vi.fn(),
  voteQuestion: vi.fn(),
  deleteQuestion: vi.fn(),
  endSession: vi.fn(),
  checkAdminStatus: vi.fn(),
  createSessionWebSocket: vi.fn(() => ({
    onmessage: null,
    onerror: null,
    close: vi.fn(),
    readyState: 1, // WebSocket.OPEN
    onopen: null,
  })),
}));

const mockSessionData: SessionData = {
  sessionTitle: 'Test Session',
  sessionId: 'test-session',
  createdAt: new Date().toDateString(),
  isActive: true,
  questions: [  
    { id: 'q1', session_id: 'test-session', text: 'Question 1', votes: 3, voters: [] },
    { id: 'q2', session_id: 'test-session', text: 'Question 2', votes: 5, voters: [] },
  ] 
};

describe('VotingSessionPage', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('shows loading state initially, then fetches and displays questions', async () => {
    (sessionApi.getSessionData as Mock).mockResolvedValue(mockSessionData);
    (sessionApi.checkAdminStatus as Mock).mockResolvedValue({ isAdmin: true });

    render(
      <BrowserRouter>
        <VotingSessionPage />
      </BrowserRouter>
    );

    expect(screen.getByText('Loading session...')).toBeInTheDocument();

    expect(await screen.findByText('Question 1')).toBeInTheDocument();
    expect(screen.getByText('Question 2')).toBeInTheDocument();
  });
  
  it('displays admin panel and handles ending session', async () => {
    window.confirm = vi.fn(() => true);
    (sessionApi.getSessionData as Mock).mockResolvedValue(mockSessionData);
    (sessionApi.checkAdminStatus as Mock).mockResolvedValue({ isAdmin: true });
    (sessionApi.endSession as Mock).mockResolvedValue(null);

    render(
      <BrowserRouter>
        <VotingSessionPage />
      </BrowserRouter>
    );
    
    const endSessionButton = await screen.findByText('End Session & Delete Data');
    await act(async () => {
      fireEvent.click(endSessionButton);
    });

    expect(window.confirm).toHaveBeenCalled();
    expect(sessionApi.endSession).toHaveBeenCalledWith('test-session');
  });
});