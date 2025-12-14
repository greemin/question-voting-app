// /frontend/src/pages/VotingSessionPage.jsx
import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getQuestions, endSession, checkAdminStatus } from '../api/sessionApi';
import QuestionForm from '../components/QuestionForm';
import QuestionItem from '../components/QuestionItem';

// Helper to get the admin ID from the cookie for UI check
const getCookieAdminID = () => {
  const cookieMatch = document.cookie.match(new RegExp('(^| )userSessionId=([^;]+)'));
  return cookieMatch ? cookieMatch[2] : null;
};

function VotingSessionPage() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const [questions, setQuestions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [isAdmin, setIsAdmin] = useState(false);
  const [adminId, setAdminId] = useState(null);

  const fetchQuestions = useCallback(async () => {
    try {
      const data = await getQuestions(sessionId);
      setQuestions(data);
      setLoading(false);

      // Simple check to determine if the current user is the admin
      if (data && data.length > 0) {
        // Assume the adminID check will be done later when we try to end session,
        // but for initial admin UI, we need the stored admin ID.
        // For simplicity here, we assume the admin ID will be fetched (or we rely on the cookie).
        // A real app would have a dedicated endpoint to check admin status.
        // We'll rely solely on the cookie matching the one stored on session creation.
        const currentCookieId = getCookieAdminID();
        // Since the Go backend doesn't expose the stored adminId, we'll check on DELETE.
        // For UI purposes, we'll optimistically set the adminId from the cookie.
        setAdminId(currentCookieId);
      }
    } catch (error) {
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
      await endSession(sessionId);
      alert('Session ended and data deleted successfully!');
      navigate('/');
    } catch (error) {
      alert(`Failed to end session: ${error.message}`);
    }
  };

  // Update this useEffect hook to call the backend check
  useEffect(() => {
    async function verifyAdmin() {
      try {
        const status = await checkAdminStatus(sessionId);
        setIsAdmin(status);
      } catch (error) {
        console.error("Failed to verify admin status:", error);
      }
    }

    // Run the verification when the component mounts
    verifyAdmin();

    // The fetchQuestions polling can be kept separate
    fetchQuestions(); 

    // Polling setup: Fetch questions every 3 seconds (and verify admin status less frequently if desired, but we'll stick to just fetching questions)
    const intervalId = setInterval(fetchQuestions, 3000);
    return () => clearInterval(intervalId);
  }, [sessionId, fetchQuestions]);


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

      <QuestionForm sessionId={sessionId} onQuestionSubmit={fetchQuestions} />

      <h2>Questions ({questions.length})</h2>
      {questions.length === 0 ? (
        <p>No questions submitted yet. Be the first!</p>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
          {questions.map((q) => (
            <QuestionItem 
              key={q.id} 
              sessionId={sessionId} 
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
