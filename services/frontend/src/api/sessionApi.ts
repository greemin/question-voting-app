// /frontend/src/api/sessionApi.ts
import toast from 'react-hot-toast';
import { SessionData } from '../models/SessionData';
import { getT } from '../i18n/useTranslation.ts';

const API_BASE = '/api/session';

const handleResponse = async (response: Response) => {
  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || `API request failed with status ${response.status}`);
  }
  if (response.status === 204) return null;
  return response.json();
};

export const createSession = async (sessionId?: string): Promise<{ sessionId: string; sessionTitle: string; adminToken: string }> => {
  const request = async () => {
    const response = await fetch(API_BASE, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ sessionId }),
      credentials: 'include',
    });
    const data = await handleResponse(response);
    if (data.adminToken) {
      localStorage.setItem(`adminToken_${data.sessionId}`, data.adminToken);
    }
    if (data.sessionTitle) {
      localStorage.setItem(`sessionTitle_${data.sessionId}`, data.sessionTitle);
    }
    return data;
  };

  return toast.promise(request(), {
    loading: getT().creatingSession,
    success: getT().sessionCreated,
    error: (err) => err.message || getT().failedToCreateSession,
  });
};

export const getSessionData = async (sessionId: string): Promise<SessionData> => {
  try {
    const response = await fetch(`${API_BASE}/${encodeURIComponent(sessionId)}`);

    const data = await handleResponse(response) as SessionData;
    const adminToken = localStorage.getItem(`adminToken_${sessionId}`);
    if (!adminToken) {
      if (data.adminToken) {
        localStorage.setItem(`adminToken_${data.sessionId}`, data.adminToken);
      }
      if (data.sessionTitle) {
        localStorage.setItem(`sessionTitle_${data.sessionId}`, data.sessionTitle);
      }
    }

    return data;
  } catch (err: any) {
    toast.error(err.message || getT().failedToLoadSession);
    throw err;
  }
};

export const submitQuestion = async (sessionId: string, text: string): Promise<null> => {
  const request = async () => {
    const response = await fetch(`${API_BASE}/${encodeURIComponent(sessionId)}/questions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text }),
    });
    return handleResponse(response);
  };

  return toast.promise(request(), {
    loading: getT().submittingQuestion,
    success: getT().questionSubmitted,
    error: (err) => err.message || getT().failedToSubmitQuestion,
  });
};

export const voteQuestion = async (sessionId: string, questionId: string): Promise<null> => {
  const request = async () => {
    const response = await fetch(`${API_BASE}/${encodeURIComponent(sessionId)}/questions/${encodeURIComponent(questionId)}/vote`, {
      method: 'PUT',
    });
    return handleResponse(response);
  };

  return toast.promise(request(), {
    loading: getT().registeringVote,
    success: getT().voteRegistered,
    error: (err) => err.message || getT().failedToVote,
  });
};

export const deleteQuestion = async (sessionId: string, questionId: string): Promise<null> => {
  const request = async () => {
    const adminToken = localStorage.getItem(`adminToken_${sessionId}`);
    const headers: Record<string, string> = {};
    if (adminToken) {
        headers['Authorization'] = `Bearer ${adminToken}`;
    }
    const response = await fetch(`${API_BASE}/${encodeURIComponent(sessionId)}/questions/${encodeURIComponent(questionId)}`, {
      method: 'DELETE',
      headers,
    });
    return handleResponse(response);
  };

  return toast.promise(request(), {
    loading: getT().deletingQuestion,
    success: getT().questionDeleted,
    error: (err) => err.message || getT().failedToDeleteQuestion,
  });
};

export const endSession = async (sessionId: string): Promise<null> => {
  const request = async () => {
    const adminToken = localStorage.getItem(`adminToken_${sessionId}`);
    const headers: Record<string, string> = {};
    if (adminToken) {
        headers['Authorization'] = `Bearer ${adminToken}`;
    }
    const response = await fetch(`${API_BASE}/${encodeURIComponent(sessionId)}`, {
      method: 'DELETE',
      headers,
    });
    return handleResponse(response);
  };

  return toast.promise(request(), {
    loading: getT().endingSession,
    success: getT().sessionEnded,
    error: (err) => err.message || getT().failedToEndSession,
  });
};

/**
 * Checks if the current user has the admin token for the session.
 * @param {string} sessionId
 * @returns {Promise<boolean>}
 */
export const checkAdminStatus = async (sessionId: string): Promise<{ isAdmin: boolean }> => {
    try {
        const adminToken = localStorage.getItem(`adminToken_${sessionId}`);
        const headers: Record<string, string> = {};
        if (adminToken) {
            headers['Authorization'] = `Bearer ${adminToken}`;
        }
        const response = await fetch(`${API_BASE}/${encodeURIComponent(sessionId)}/check-admin`, { headers });
        return await handleResponse(response);
    } catch (err: any) {
        toast.error(err.message || getT().failedToCheckAdminStatus);
        throw err;
    }
};

/**
 * Creates a WebSocket connection for real-time session updates.
 * @param {string} sessionId
 * @returns {WebSocket}
 */
export const createSessionWebSocket = (sessionId: string): WebSocket => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}${API_BASE}/${encodeURIComponent(sessionId)}/ws`;
  return new WebSocket(wsUrl);
};
