// /frontend/src/pages/VotingSessionPage.tsx
import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getQuestions, endSession, checkAdminStatus, createSessionWebSocket } from '../api/sessionApi.ts';
import QuestionForm from '../components/QuestionForm.tsx';
import QuestionItem from '../components/QuestionItem.tsx';
import { Question } from '../models/Question';

function VotingSessionPage(): JSX.Element {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const [questions, setQuestions] = useState<Question[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [isAdmin, setIsAdmin] = useState<boolean>(false);

  const fetchQuestions = useCallback(async () => {
    try {
      if (!sessionId) return;
      const data = await getQuestions(sessionId);
      setQuestions(data);
      setLoading(false);
    } catch (error: any) {
      console.error('Failed to fetch questions:', error);
      alert(error.message);
      // If session is closed (404/403), redirect home or show message
      if (error.message.includes('not found')) {
        navigate('/');
      }
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

  // Update this useEffect hook to call the backend check
  useEffect(() => {
    async function verifyAdmin() {
      try {
        if (!sessionId) return;
        const status = await checkAdminStatus(sessionId);
        setIsAdmin(status.isAdmin);
      } catch (error) {
        console.error("Failed to verify admin status:", error);
      }
    }

    // Run the verification when the component mounts
    verifyAdmin();

    // Fetch initial questions
    fetchQuestions();

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
  }, [sessionId, fetchQuestions, navigate]);


  if (loading) return <div style={{ textAlign: 'center', padding: '50px' }}>Loading session...</div>;

  return (
    <div style={{ maxWidth: '800px', margin: '0 auto', padding: '20px' }}>
      <h1 style={{ borderBottom: '2px solid #333', paddingBottom: '10px' }}>Voting Session: <code style={{ color: '#dc3545' }}>{sessionId}</code></h1>
      
      {isAdmin && (
        <div style={{ border: '1px solid #dc3545', padding: '15px', marginBottom: '20px', backgroundColor: '#f8d7da', borderRadius: '5px' }}>
          <p style={{ fontWeight: 'bold', color: '#721c24' }}>Admin Panel</p>
          <button 
            onClick={handleEndSession}
            style={{ padding: '10px 20px', cursor: 'pointer', backgroundColor: '#dc3545', color: 'white', border: 'none', borderRadius: '4px' }}
          >
            End Session & Delete Data
          </button>
        </div>
      )}

      <QuestionForm sessionId={sessionId!} onQuestionSubmit={fetchQuestions} />

      <h2>Questions ({questions.length})</h2>
      {questions.length === 0 ? (
        <p>No questions submitted yet. Be the first!</p>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
          {questions.map((q) => (
            <QuestionItem 
              key={q.id} 
              sessionId={sessionId!} 
              question={q} 
              onVoteSuccess={fetchQuestions} // Re-fetch to update votes and sort
            />
          ))}
        </div>
      )}
    </div>
  );
}

export default VotingSessionPage;
