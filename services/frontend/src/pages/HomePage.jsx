// /frontend/src/pages/HomePage.jsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { createSession } from '../api/sessionApi';

function HomePage() {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleCreateSession = async () => {
    setLoading(true);
    try {
      const data = await createSession();
      // data.sessionId is returned, and adminId is confirmed via cookie
      navigate(`/votingSession/${data.sessionId}`);
    } catch (error) {
      alert(`Failed to create session: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ textAlign: 'center', padding: '50px' }}>
      <h1>Question Voting App</h1>
      <button 
        onClick={handleCreateSession} 
        disabled={loading}
        style={{ padding: '15px 30px', fontSize: '1.2em', cursor: 'pointer', backgroundColor: '#007bff', color: 'white', border: 'none', borderRadius: '5px' }}
      >
        {loading ? 'Creating...' : '🚀 Start New Voting Session'}
      </button>
      <p style={{ marginTop: '20px', color: '#666' }}>
        A session ID will be generated, and you will be the administrator.
      </p>
    </div>
  );
}

export default HomePage;
