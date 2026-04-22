// /frontend/src/pages/HomePage.tsx
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { createSession } from '../api/sessionApi.ts';
import { useTranslation } from '../i18n/useTranslation.ts';
import './HomePage.css';

function HomePage(): JSX.Element {
  const { t } = useTranslation();
  const [loading, setLoading] = useState<boolean>(false);
  const [customSlug, setCustomSlug] = useState<string>('');
  const navigate = useNavigate();

  const appName = (window as any).__APP_NAME__ ?? import.meta.env.VITE_APP_NAME ?? 'Question Voting App';
  
    useEffect(() => {
      document.title = appName;
    }, []);

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
      <h1>{t.appTitle}</h1>
      <p className="home-tagline">{t.tagline}</p>
      <div className="home-card">
        <div className="custom-slug-container">
          <input
            type="text"
            value={customSlug}
            onChange={(e) => setCustomSlug(e.target.value)}
            placeholder={t.sessionTitlePlaceholder}
            className="custom-slug-input"
          />
        </div>
        <button
          onClick={handleCreateSession}
          disabled={loading}
          className="start-session-button"
        >
          {loading ? t.creating : t.startNewVotingSession}
        </button>
        <p className="home-page-info">
          {t.homePageInfo}
        </p>
      </div>
    </div>
  );
}

export default HomePage;
