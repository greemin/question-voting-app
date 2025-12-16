// /frontend/src/api/sessionApi.ts
import { Question } from '../models/Question';

const API_BASE = '/api/session';

// Helper function to get the current userSessionId from the document cookies
const getAdminIdFromCookie = (): string | null => {
  const cookieMatch = document.cookie.match(new RegExp('(^| )userSessionId=([^;]+)'));
  return cookieMatch ? cookieMatch[2] : null;
};

const handleResponse = async (response: Response) => {
  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || `API request failed with status ${response.status}`);
  }
  if (response.status === 204) return null;
  return response.json();
};

export const createSession = async (): Promise<{ sessionId: string; adminId: string | null }> => {
  const response = await fetch(API_BASE, {
    method: 'POST',
  });
  const data = await handleResponse(response);
  // Admin ID is now guaranteed to be in the cookie
  data.adminId = getAdminIdFromCookie();
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
  const response = await fetch(`${API_BASE}/${sessionId}`, {
    method: 'DELETE',
  });
  return handleResponse(response);
};

/**
 * Checks if the current user session is the admin for the session.
 * This relies on the Go backend reading the HttpOnly cookie.
 * @param {string} sessionId
 * @returns {Promise<boolean>}
 */
export const checkAdminStatus = async (sessionId: string): Promise<{ isAdmin: boolean }> => {
    const response = await fetch(`${API_BASE}/${sessionId}/check-admin`);
    return await handleResponse(response);
};
