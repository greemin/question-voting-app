// /frontend/src/components/QuestionForm.tsx
import React, { useState } from 'react';
import { submitQuestion } from '../api/sessionApi.ts';
import { useTranslation } from '../hooks/useTranslation.ts';
import './QuestionForm.css';

interface QuestionFormProps {
  sessionId: string;
  onQuestionSubmit: () => void;
}

function QuestionForm({ sessionId, onQuestionSubmit }: QuestionFormProps): JSX.Element {
  const { t } = useTranslation();
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
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="question-form">
      <h3>{t.submitAQuestion}</h3>
      <input
        type="text"
        value={text}
        onChange={(e) => setText(e.target.value)}
        placeholder={t.typeYourQuestionHere}
        disabled={loading}
        className="question-input"
        maxLength={QUESTION_MAX_LENGTH}
      />
      <div className="form-footer">
        {text.length > 0 && <span className="character-count">{text.length}/{QUESTION_MAX_LENGTH}</span>}
        <button type="submit" disabled={loading} className="submit-button">
          {loading ? t.submitting : t.submitQuestion}
        </button>
      </div>
    </form>
  );
}

export default QuestionForm;
