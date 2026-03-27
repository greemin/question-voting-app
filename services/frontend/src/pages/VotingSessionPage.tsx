// /frontend/src/pages/VotingSessionPage.tsx
import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getSessionData, endSession, checkAdminStatus, createSessionWebSocket } from '../api/sessionApi.ts';
import QuestionForm from '../components/QuestionForm.tsx';
import QuestionItem from '../components/QuestionItem.tsx';
import { Question } from '../models/Question';
import './VotingSessionPage.css';

function VotingSessionPage(): JSX.Element {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const [questions, setQuestions] = useState<Question[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [isAdmin, setIsAdmin] = useState<boolean>(false);
  const [sessionTitle, setSessionTitle] = useState<string>('');

  const fetchSession = useCallback(async () => {
    try {
      if (!sessionId) return;
      setLoading(true);
      const data = await getSessionData(sessionId); // This may contain an adminToken on creation

      // If an adminToken is returned, it means a new session was just created
      // and this user is the admin. Otherwise, we check the status separately.
      if (data.adminToken) {
        setIsAdmin(true);
      } else {
        const status = await checkAdminStatus(sessionId);
        setIsAdmin(status.isAdmin);
      }

      data.questions.sort((a: Question, b: Question) => b.votes - a.votes);
      setQuestions(data.questions);
      setSessionTitle(data.sessionTitle);
    } catch (error: any) {
      console.error('Failed to fetch session data:', error);
      alert(error.message);
      if (error.message.includes('not found')) {
        navigate('/');
      }
    } finally {
      setLoading(false);
    }
  }, [sessionId, navigate]);

  const handleEndSession = async () => {
    if (!window.confirm("Are you sure you want to end this voting session? This will delete all questions.")) return;
    
    try {
      if (!sessionId) return;
      await endSession(sessionId);
      alert('Session ended and data deleted successfully!');
      navigate('/');
    } catch (error: any) {
      alert(`Failed to end session: ${error.message}`);
    }
  };

  useEffect(() => {
    // Fetch initial session data
    fetchSession();

    if (!sessionId) return;

    const ws = createSessionWebSocket(sessionId);

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        switch (data.type) {
          case 'QUESTION_ADDED':
            setQuestions((prev) => [...prev, data.payload].sort((a, b) => b.votes - a.votes));
            break;

          case 'VOTE_UPDATED':
            setQuestions((prev) => {
              const updated = prev.map((q) => (q.id === data.payload.id ? data.payload : q));
              return updated.sort((a, b) => b.votes - a.votes);
            });
            break;

          case 'QUESTION_DELETED':
            setQuestions((prev) => prev.filter((q) => q.id !== data.payload.id));
            break;

          case 'SESSION_ENDED':
            alert('This session has been ended by the admin.');
            navigate('/');
            break;

          default:
            console.warn('Unknown WebSocket event:', data.type);
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    ws.onerror = (error) => console.error('WebSocket error:', error);

    return () => {
      // Clear handlers to avoid memory leaks or state updates on unmounted components
      ws.onmessage = null;
      ws.onerror = null;
      
      // If the socket is still connecting, waiting for it to open before closing 
      // prevents browser console errors ("connection interrupted") in React Strict Mode.
      if (ws.readyState === WebSocket.CONNECTING) {
        ws.onopen = () => ws.close();
      } else {
        ws.close();
      }
    };
  }, [sessionId, fetchSession, navigate]);


  if (loading) return <div className="loading-session">Loading session...</div>;
  const link = window.location.host + window.location.pathname;

  return (
    <div className="voting-session-container">
      <div className="title-container">
        <h1 className="session-title">
          {sessionTitle?.toUpperCase()}
        </h1>
        <a
            href={window.location.href}
            className="session-id"
            target="_blank"
            rel="noopener noreferrer"
          >
            {link}
          </a>
      </div>
      
      { isAdmin && 
          <button onClick={handleEndSession} className="end-session-button">
            End Session & Delete Data
          </button> 
      }

      <QuestionForm
        sessionId={sessionId!}
        onQuestionSubmit={() => {}} /* Websocket handles update */
      />
      
      <h2>Questions ({questions.length})</h2>
      {questions.length === 0 ? (
        <p>No questions submitted yet. Be the first!</p>
      ) : (
        <div className="questions-container">
          {questions.map((q) => (
            <QuestionItem 
              key={q.id} 
              sessionId={sessionId!} 
              question={q} 
              isAdmin={isAdmin}
              onVoteSuccess={() => {}} /* Websocket handles update */
            />
          ))}
        </div>
      )}
    </div>
  );
}

export default VotingSessionPage;
