// /frontend/src/components/QuestionForm.tsx
import React, { useState } from 'react';
import { submitQuestion } from '../api/sessionApi.ts';
import './QuestionForm.css';

interface QuestionFormProps {
  sessionId: string;
  onQuestionSubmit: () => void;
}

function QuestionForm({ sessionId, onQuestionSubmit }: QuestionFormProps): JSX.Element {
  const [text, setText] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const QUESTION_MAX_LENGTH = 500;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!text.trim()) return;

    setLoading(true);
    try {
      await submitQuestion(sessionId, text.trim().slice(0, QUESTION_MAX_LENGTH));
      setText('');
      onQuestionSubmit(); // Refresh the list
    } catch (error: any) {
      alert(`Submission failed: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="question-form">
      <h3>Submit a Question</h3>
      <input
        type="text"
        value={text}
        onChange={(e) => setText(e.target.value)}
        placeholder="Type your question here..."
        disabled={loading}
        className="question-input"
        maxLength={QUESTION_MAX_LENGTH}
      />
      { text.length > 0 && <div className="character-count">{text.length}/{QUESTION_MAX_LENGTH}</div> }
      <button type="submit" disabled={loading} className="submit-button">
        {loading ? 'Submitting...' : 'Submit Question'}
      </button>
    </form>
  );
}

export default QuestionForm;
