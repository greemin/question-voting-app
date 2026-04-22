// /frontend/src/pages/VotingSessionPage.tsx
import React, { useState, useEffect, useCallback } from 'react';
import toast from 'react-hot-toast';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { QRCodeSVG } from 'qrcode.react';
import { getSessionData, endSession, checkAdminStatus, createSessionWebSocket } from '../api/sessionApi.ts';
import QuestionForm from '../components/QuestionForm.tsx';
import QuestionItem from '../components/QuestionItem.tsx';
import { Question } from '../models/Question';
import { useTranslation } from '../i18n/useTranslation.ts';
import './VotingSessionPage.css';

function VotingSessionPage(): JSX.Element {
  const { t } = useTranslation();
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [questions, setQuestions] = useState<Question[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [isAdmin, setIsAdmin] = useState<boolean>(false);
  const [sessionTitle, setSessionTitle] = useState<string>('');
  const [showQR, setShowQR] = useState<boolean>(false);

  // If an adminToken is passed as a query param (e.g. via a shared admin link),
  // persist it to localStorage and strip it from the URL.
  useEffect(() => {
    const tokenFromUrl = searchParams.get('adminToken');
    if (tokenFromUrl && sessionId) {
      localStorage.setItem(`adminToken_${sessionId}`, tokenFromUrl);
      setSearchParams({}, { replace: true });
    }
  }, []);

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

      if (Array.isArray(data.questions) && data.questions.length > 0) {
        data.questions.sort((a: Question, b: Question) => b.votes - a.votes);
        setQuestions(data.questions);
      }

      setSessionTitle(data.sessionTitle);
    } finally {
      setLoading(false);
    }
  }, [sessionId, navigate]);

  const handleEndSession = async () => {
    if (!window.confirm(t.endSessionConfirm)) return;

    if (!sessionId) return;
    await endSession(sessionId);
    navigate('/');

  };

  const handleCopyLink = async () => {
    try {
      await navigator.clipboard.writeText(window.location.href);
      toast.success(t.linkCopied);
    } catch (err) {
      console.error('Failed to copy link:', err);
      toast.error(t.failedToCopyLink);
    }
  };

  const handleCopyAdminLink = async () => {
    if (!sessionId) return;
    const adminToken = localStorage.getItem(`adminToken_${sessionId}`);
    if (!adminToken) {
      toast.error(t.adminTokenNotFound);
      return;
    }
    try {
      const adminUrl = `${window.location.origin}${window.location.pathname}?adminToken=${encodeURIComponent(adminToken)}`;
      await navigator.clipboard.writeText(adminUrl);
      toast.success(t.adminLinkCopied);
    } catch (err) {
      console.error('Failed to copy admin link:', err);
      toast.error(t.failedToCopyAdminLink);
    }
  };

  useEffect(() => {
    // Fetch initial session data
    fetchSession();

    if (!sessionId) return;

    let ws: WebSocket;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let unmounted = false;

    const connect = () => {
      ws = createSessionWebSocket(sessionId);

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

            case 'IP_BANNED':
              setQuestions((prev) => prev.filter((q) => !data.payload.questionIds.includes(q.id)));
              break;

            case 'SESSION_ENDED':
              if(!isAdmin) {
                toast(t.sessionEndedByAdmin);
              }
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

      ws.onclose = () => {
        if (!unmounted) {
          reconnectTimer = setTimeout(connect, 3000);
        }
      };
    };

    connect();

    return () => {
      unmounted = true;
      if (reconnectTimer !== null) clearTimeout(reconnectTimer);
      ws.onmessage = null;
      ws.onerror = null;
      ws.onclose = null;

      // If the socket is still connecting, waiting for it to open before closing
      // prevents browser console errors ("connection interrupted") in React Strict Mode.
      if (ws.readyState === WebSocket.CONNECTING) {
        ws.onopen = () => ws.close();
      } else {
        ws.close();
      }
    };
  }, [sessionId, fetchSession, navigate]);

  if (loading) {
    return <div className="loading-session"><p>{t.loadingSession}</p></div>;
  }

  const link = window.location.host + window.location.pathname;

  return (
    <div className="voting-session-page">
      <header className="session-header">
        <h1 className="session-title">{sessionTitle?.toUpperCase()}</h1>
        <div className="header-actions">
          <a
            href={window.location.href}
            className="session-link"
            target="_blank"
            rel="noopener noreferrer"
          >
            {link}
          </a>
          <button onClick={handleCopyLink} className="copy-link-button">
            📋
          </button>
          <button onClick={() => setShowQR((v) => !v)} className="copy-link-button" title={t.showQrCode}>
            ▣
          </button>
          {isAdmin && <>
            <button onClick={handleCopyAdminLink} className="copy-link-button" title={t.copyAdminLink}>
              🔑
            </button>
            <button onClick={handleEndSession} className="end-session-button">
              {t.endSession}
            </button>
          </>}
        </div>
      </header>

      <div className={`qr-banner${showQR ? ' qr-banner--visible' : ''}`}>
        <QRCodeSVG value={`${window.location.origin}${window.location.pathname}`} size={200} />
        <p className="qr-banner-url">{window.location.host}{window.location.pathname}</p>
      </div>

      <main className="session-content">
        <QuestionForm
          sessionId={sessionId!}
          onQuestionSubmit={() => {}} /* Websocket handles update */
        />

        <div className="questions-heading">
          {t.questions}
          <span className="questions-count-badge">{questions.length}</span>
        </div>

        {questions.length === 0 ? (
          <p className="no-questions">{t.noQuestionsYet}</p>
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
      </main>
    </div>
  );
}

export default VotingSessionPage;
