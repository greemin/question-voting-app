// /frontend/src/components/QuestionItem.tsx
import React from 'react';
import { voteQuestion, deleteQuestion, banSubmitter } from '../api/sessionApi.ts';
import { Question } from '../models/Question';
import { useTranslation } from '../i18n/useTranslation.ts';
import './QuestionItem.css';

interface QuestionItemProps {
  sessionId: string;
  question: Question;
  isAdmin: boolean;
  onVoteSuccess: () => void;
}

function QuestionItem({ sessionId, question, isAdmin, onVoteSuccess }: QuestionItemProps): JSX.Element {
  const { t } = useTranslation();
  const handleVote = async () => {
    await voteQuestion(sessionId, question.id);
    onVoteSuccess();
  };

  const handleDelete = async () => {
    await deleteQuestion(sessionId, question.id);
  };

  const handleBan = async () => {
    await banSubmitter(sessionId, question.id);
  };

  return (
    <div className="question-item-container">
      <div className="vote-pill">
        <strong data-testid="vote-count">{question.votes}</strong>
        <span>{t.votes}</span>
      </div>
      <div className="question-body">
        <p className="question-text">{question.text}</p>
        <div className="button-group">
          <button onClick={handleVote} data-testid="vote-button" className="vote-button">
            {t.voteUp}
          </button>
          {isAdmin && (
            <button onClick={handleDelete} data-testid="delete-button" className="delete-button">
              {t.delete}
            </button>
          )}
          {isAdmin && (
            <button onClick={handleBan} data-testid="ban-button" className="ban-button">
              {t.banSubmitter}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

export default QuestionItem;
