// /frontend/src/api/sessionApi.ts
import { Question } from '../models/Question';

const API_BASE = '/api/session';

const handleResponse = async (response: Response) => {
  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || `API request failed with status ${response.status}`);
  }
  if (response.status === 204) return null;
  return response.json();
};

export const createSession = async (sessionId?: string): Promise<{ sessionId: string; adminToken: string }> => {
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
  return data;
};

export const getQuestions = async (sessionId: string): Promise<Question[]> => {
  const response = await fetch(`${API_BASE}/${sessionId}/questions`);
  return handleResponse(response);
};

export const submitQuestion = async (sessionId: string, text: string): Promise<null> => {
  const response = await fetch(`${API_BASE}/${sessionId}/questions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text }),
  });
  return handleResponse(response);
};

export const voteQuestion = async (sessionId: string, questionId: string): Promise<null> => {
  const response = await fetch(`${API_BASE}/${sessionId}/questions/${questionId}/vote`, {
    method: 'PUT',
  });
  return handleResponse(response);
};

export const endSession = async (sessionId: string): Promise<null> => {
  const adminToken = localStorage.getItem(`adminToken_${sessionId}`);
  const headers: Record<string, string> = {};
  if (adminToken) {
      headers['Authorization'] = `Bearer ${adminToken}`;
  }
  const response = await fetch(`${API_BASE}/${sessionId}`, {
    method: 'DELETE',
    headers,
  });
  return handleResponse(response);
};

/**
 * Checks if the current user has the admin token for the session.
 * @param {string} sessionId
 * @returns {Promise<boolean>}
 */
export const checkAdminStatus = async (sessionId: string): Promise<{ isAdmin: boolean }> => {
    const adminToken = localStorage.getItem(`adminToken_${sessionId}`);
    const headers: Record<string, string> = {};
    if (adminToken) {
        headers['Authorization'] = `Bearer ${adminToken}`;
    }
    const response = await fetch(`${API_BASE}/${sessionId}/check-admin`, { headers });
    return await handleResponse(response);
};

/**
 * Creates a WebSocket connection for real-time session updates.
 * @param {string} sessionId
 * @returns {WebSocket}
 */
export const createSessionWebSocket = (sessionId: string): WebSocket => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}${API_BASE}/${sessionId}/ws`;
  return new WebSocket(wsUrl);
};
