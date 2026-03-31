// /frontend/src/components/QuestionItem.tsx
import React from 'react';
import { voteQuestion, deleteQuestion } from '../api/sessionApi.ts';
import { Question } from '../models/Question';
import './QuestionItem.css';

interface QuestionItemProps {
  sessionId: string;
  question: Question;
  isAdmin: boolean;
  onVoteSuccess: () => void;
}

function QuestionItem({ sessionId, question, isAdmin, onVoteSuccess }: QuestionItemProps): JSX.Element {
  const handleVote = async () => {
    await voteQuestion(sessionId, question.id);
    onVoteSuccess();
  };

  const handleDelete = async () => {    
    await deleteQuestion(sessionId, question.id);
  };

  return (
    <div className="question-item-container">
      <p className="question-text">{question.text}</p>
      <div className="question-details">
        <span className="vote-count">
          Votes: <strong data-testid="vote-count">{question.votes}</strong>
        </span>
        <div className="button-group">
          <button onClick={handleVote} data-testid="vote-button" className="vote-button">
            Vote Up
          </button>
          {isAdmin && (
            <button onClick={handleDelete} data-testid="delete-button" className="delete-button">
              Delete
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

export default QuestionItem;
