// /frontend/src/components/QuestionItem.tsx
import React from 'react';
import { voteQuestion, deleteQuestion } from '../api/sessionApi.ts';
import { Question } from '../models/Question';

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
    if (!window.confirm("Are you sure you want to delete this question?")) return;
    
    try {
      await deleteQuestion(sessionId, question.id);
      // Note: We don't necessarily need to refresh the list manually here 
      // because our WebSocket `QUESTION_DELETED` event will instantly remove it!
    } catch (error: any) {
      alert(`Delete failed: ${error.message}`);
    }
  };

  return (
    <div style={{ border: '1px solid #ddd', padding: '15px', margin: '10px 0', borderRadius: '4px', backgroundColor: '#f9f9f9' }}>
      <p style={{ fontWeight: 'bold' }}>{question.text}</p>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '10px' }}>
        <span style={{ fontSize: '1.2em', color: '#007bff' }}>
          Votes: <strong data-testid="vote-count">{question.votes}</strong>
        </span>
        <div style={{ display: 'flex', gap: '10px' }}>
          <button onClick={handleVote} data-testid="vote-button" style={{ padding: '8px 15px', cursor: 'pointer', backgroundColor: '#28a745', color: 'white', border: 'none', borderRadius: '4px' }}>
            Vote Up
          </button>
          {isAdmin && (
            <button onClick={handleDelete} data-testid="delete-button" style={{ padding: '8px 15px', cursor: 'pointer', backgroundColor: '#dc3545', color: 'white', border: 'none', borderRadius: '4px' }}>
              Delete
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

export default QuestionItem;
