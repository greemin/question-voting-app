// /frontend/src/pages/HomePage.tsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { createSession } from '../api/sessionApi.ts';

function HomePage(): JSX.Element {
  const [loading, setLoading] = useState<boolean>(false);
  const [customSlug, setCustomSlug] = useState<string>('');
  const navigate = useNavigate();

  const handleCreateSession = async () => {
    setLoading(true);
    try {
      const data = await createSession(customSlug);
      // data.sessionId is returned, and adminId is confirmed via cookie
      navigate(`/votingSession/${data.sessionId}`);
    } catch (error: any) {
      alert(`Failed to create session: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ textAlign: 'center', padding: '50px' }}>
      <h1>Question Voting App</h1>
      <div style={{ margin: '20px 0' }}>
        <input
          type="text"
          value={customSlug}
          onChange={(e) => setCustomSlug(e.target.value)}
          placeholder="Optional: custom-session-name"
          style={{ padding: '10px', width: '300px', fontSize: '1em' }}
        />
      </div>
      <button
        onClick={handleCreateSession}
        disabled={loading}
        style={{ padding: '15px 30px', fontSize: '1.2em', cursor: 'pointer', backgroundColor: '#007bff', color: 'white', border: 'none', borderRadius: '5px' }}
      >
        {loading ? 'Creating...' : '🚀 Start New Voting Session'}
      </button>
      <p style={{ marginTop: '20px', color: '#666' }}>
        Enter a custom name or start a session to generate a random one.
      </p>
    </div>
  );
}

export default HomePage;
