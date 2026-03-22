// /frontend/src/pages/HomePage.tsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { createSession } from '../api/sessionApi.ts';
import './HomePage.css';

function HomePage(): JSX.Element {
  const [loading, setLoading] = useState<boolean>(false);
  const [customSlug, setCustomSlug] = useState<string>('');
  const navigate = useNavigate();

  const handleCreateSession = async () => {
    setLoading(true);
    try {
      const data = await createSession(customSlug);
      // data.sessionId is returned
      navigate(`/votingSession/${data.sessionId}`);
    } catch (error: any) {
      alert(`Failed to create session: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="home-page-container">
      <h1>Question Voting App</h1>
      <div className="custom-slug-container">
        <input
          type="text"
          value={customSlug}
          onChange={(e) => setCustomSlug(e.target.value)}
          placeholder="Sessiontitle: Q&A Session"
          className="custom-slug-input"
        />
      </div>
      <button
        onClick={handleCreateSession}
        disabled={loading}
        className="start-session-button"
      >
        {loading ? 'Creating...' : '🚀 Start New Voting Session'}
      </button>
      <p className="home-page-info">
        Enter a custom name or start a session to generate a random one.
      </p>
    </div>
  );
}

export default HomePage;
