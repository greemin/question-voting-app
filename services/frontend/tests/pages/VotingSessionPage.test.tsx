// /frontend/tests/pages/VotingSessionPage.test.tsx
import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/preact';
import { BrowserRouter, MemoryRouter } from 'react-router-dom';
import { vi, Mock } from 'vitest';
import VotingSessionPage from '../../src/pages/VotingSessionPage';
import * as sessionApi from '../../src/api/sessionApi';
import { SessionData } from '../../src/models/SessionData';

const mockNavigate = vi.fn();
const mockUseSearchParams = vi.fn(() => [new URLSearchParams(), vi.fn()] as const);

vi.mock('qrcode.react', () => ({
  QRCodeSVG: () => null,
}));

vi.mock('react-router-dom', () => ({
  BrowserRouter: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  MemoryRouter: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useParams: () => ({ sessionId: 'test-session' }),
  useNavigate: () => mockNavigate,
  useSearchParams: () => mockUseSearchParams(),
}));

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
    localStorage.clear();
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
    
    const endSessionButton = await screen.findByText('End Session');
    await act(async () => {
      fireEvent.click(endSessionButton);
    });

    expect(window.confirm).toHaveBeenCalled();
    expect(sessionApi.endSession).toHaveBeenCalledWith('test-session');
  });
});

describe('QR code banner', () => {
  afterEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('is hidden by default', async () => {
    (sessionApi.getSessionData as Mock).mockResolvedValue(mockSessionData);
    (sessionApi.checkAdminStatus as Mock).mockResolvedValue({ isAdmin: false });

    const { container } = render(
      <BrowserRouter>
        <VotingSessionPage />
      </BrowserRouter>
    );

    await screen.findByText('Question 1');

    expect(container.querySelector('.qr-banner')).not.toHaveClass('qr-banner--visible');
  });

  it('shows when QR button is clicked and hides on second click', async () => {
    (sessionApi.getSessionData as Mock).mockResolvedValue(mockSessionData);
    (sessionApi.checkAdminStatus as Mock).mockResolvedValue({ isAdmin: false });

    const { container } = render(
      <BrowserRouter>
        <VotingSessionPage />
      </BrowserRouter>
    );

    await screen.findByText('Question 1');
    const qrButton = screen.getByTitle('Show QR code');
    const banner = container.querySelector('.qr-banner');

    await act(async () => { fireEvent.click(qrButton); });
    expect(banner).toHaveClass('qr-banner--visible');

    await act(async () => { fireEvent.click(qrButton); });
    expect(banner).not.toHaveClass('qr-banner--visible');
  });
});

describe('admin link', () => {
  beforeEach(() => {
    Object.assign(navigator, {
      clipboard: { writeText: vi.fn().mockResolvedValue(undefined) },
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('copy admin link button is not visible to non-admins', async () => {
    (sessionApi.getSessionData as Mock).mockResolvedValue(mockSessionData);
    (sessionApi.checkAdminStatus as Mock).mockResolvedValue({ isAdmin: false });

    render(
      <BrowserRouter>
        <VotingSessionPage />
      </BrowserRouter>
    );

    await screen.findByText('Question 1');
    expect(screen.queryByTitle('Copy admin link')).not.toBeInTheDocument();
  });

  it('copy admin link button is visible to admins', async () => {
    (sessionApi.getSessionData as Mock).mockResolvedValue(mockSessionData);
    (sessionApi.checkAdminStatus as Mock).mockResolvedValue({ isAdmin: true });

    render(
      <BrowserRouter>
        <VotingSessionPage />
      </BrowserRouter>
    );

    await screen.findByText('End Session');
    expect(screen.getByTitle('Copy admin link')).toBeInTheDocument();
  });

  it('copies URL with admin token to clipboard when clicked', async () => {
    (sessionApi.getSessionData as Mock).mockResolvedValue(mockSessionData);
    (sessionApi.checkAdminStatus as Mock).mockResolvedValue({ isAdmin: true });
    localStorage.setItem('adminToken_test-session', 'my-secret-token');

    render(
      <BrowserRouter>
        <VotingSessionPage />
      </BrowserRouter>
    );

    await screen.findByText('End Session');
    await act(async () => { fireEvent.click(screen.getByTitle('Copy admin link')); });

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining('adminToken=my-secret-token')
    );
  });
});

describe('admin token from URL', () => {
  afterEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    mockUseSearchParams.mockReturnValue([new URLSearchParams(), vi.fn()]);
  });

  it('stores token from URL query param in localStorage', async () => {
    mockUseSearchParams.mockReturnValue([new URLSearchParams('adminToken=url-token'), vi.fn()]);
    (sessionApi.getSessionData as Mock).mockResolvedValue(mockSessionData);
    (sessionApi.checkAdminStatus as Mock).mockResolvedValue({ isAdmin: false });

    render(
      <MemoryRouter initialEntries={['/test-session?adminToken=url-token']}>
        <VotingSessionPage />
      </MemoryRouter>
    );

    await screen.findByText('Question 1');
    expect(localStorage.getItem('adminToken_test-session')).toBe('url-token');
  });
});