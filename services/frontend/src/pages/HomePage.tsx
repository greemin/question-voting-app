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
      navigate(`/${data.sessionId}`);
    } catch (error) {
      // The toast is already shown by the API layer.
      // We can just log the error for debugging.
      console.error("Failed to create session:", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="home-page-container">
      <h1>Question Voting App</h1>
      <p className="home-tagline">Real-time Q&amp;A for your event</p>
      <div className="home-card">
        <div className="custom-slug-container">
          <input
            type="text"
            value={customSlug}
            onChange={(e) => setCustomSlug(e.target.value)}
            placeholder="Session title (optional)"
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
          Enter a custom name or leave blank to generate a random one.
        </p>
      </div>
    </div>
  );
}

export default HomePage;
