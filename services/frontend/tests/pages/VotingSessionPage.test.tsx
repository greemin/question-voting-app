// /frontend/tests/pages/VotingSessionPage.test.tsx
import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { vi } from 'vitest';
import VotingSessionPage from '../../src/pages/VotingSessionPage';
import * as sessionApi from '../../src/api/sessionApi';
import { Question } from '../../src/models/Question';

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

vi.mock('../../src/api/sessionApi');

const mockQuestions: Question[] = [
  { id: 'q1', session_id: 'test-session', text: 'Question 1', vote_count: 3, author: 'test', created_at: '' },
  { id: 'q2', session_id: 'test-session', text: 'Question 2', vote_count: 5, author: 'test', created_at: '' },
];

describe('VotingSessionPage', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('shows loading state initially, then fetches and displays questions', async () => {
    vi.spyOn(sessionApi, 'getQuestions').mockResolvedValue(mockQuestions);
    vi.spyOn(sessionApi, 'checkAdminStatus').mockResolvedValue({ isAdmin: true });

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
    vi.spyOn(sessionApi, 'getQuestions').mockResolvedValue(mockQuestions);
    vi.spyOn(sessionApi, 'checkAdminStatus').mockResolvedValue({ isAdmin: true });
    vi.spyOn(sessionApi, 'endSession').mockResolvedValue(null);

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

  /* Disabled because we are gonna switch to sockets soon
  it('polls for new questions', async () => {
    vi.useFakeTimers();
    const getQuestionsMock = vi.spyOn(sessionApi, 'getQuestions').mockResolvedValue(mockQuestions);
    vi.spyOn(sessionApi, 'checkAdminStatus').mockResolvedValue({ isAdmin: true });

    render(
      <BrowserRouter>
        <VotingSessionPage />
      </BrowserRouter>
    );
    
    // Wait for initial render and fetch
    await screen.findByText('Question 1');
    expect(getQuestionsMock).toHaveBeenCalledTimes(1);

    // Advance timers to trigger polling
    await act(async () => {
      vi.advanceTimersByTime(3000);
    });
    
    expect(getQuestionsMock).toHaveBeenCalledTimes(2);
    
    vi.useRealTimers();
  });*/
});