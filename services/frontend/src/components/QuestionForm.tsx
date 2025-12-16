// /frontend/src/components/QuestionForm.tsx
import React, { useState } from 'react';
import { submitQuestion } from '../api/sessionApi.ts';

interface QuestionFormProps {
  sessionId: string;
  onQuestionSubmit: () => void;
}

function QuestionForm({ sessionId, onQuestionSubmit }: QuestionFormProps): JSX.Element {
  const [text, setText] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!text.trim()) return;

    setLoading(true);
    try {
      await submitQuestion(sessionId, text);
      setText('');
      onQuestionSubmit(); // Refresh the list
    } catch (error: any) {
      alert(`Submission failed: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ margin: '20px 0', padding: '20px', border: '1px solid #ccc', borderRadius: '8px' }}>
      <h3>Submit a Question</h3>
      <input
        type="text"
        value={text}
        onChange={(e) => setText(e.target.value)}
        placeholder="Type your question here..."
        disabled={loading}
        style={{ width: '100%', padding: '10px', marginBottom: '10px', border: '1px solid #ddd', borderRadius: '4px' }}
      />
      <button type="submit" disabled={loading} style={{ padding: '10px 20px', cursor: 'pointer' }}>
        {loading ? 'Submitting...' : 'Submit Question'}
      </button>
    </form>
  );
}

export default QuestionForm;
