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
    try {
      await voteQuestion(sessionId, question.id);
      onVoteSuccess();
    } catch (error: any) {
      alert(`Vote failed: ${error.message}`);
    }
  };

  const handleDelete = async () => {    
    try {
      await deleteQuestion(sessionId, question.id);
      // Note: We don't necessarily need to refresh the list manually here 
      // because our WebSocket `QUESTION_DELETED` event will instantly remove it!
    } catch (error: any) {
      alert(`Delete failed: ${error.message}`);
    }
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
