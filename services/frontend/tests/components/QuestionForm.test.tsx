// /frontend/tests/components/QuestionForm.test.tsx
import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { vi } from 'vitest';
import QuestionForm from '../../src/components/QuestionForm';
import * as sessionApi from '../../src/api/sessionApi';

vi.mock('../../src/api/sessionApi');

describe('QuestionForm', () => {
  const sessionId = 'test-session';
  const onQuestionSubmit = vi.fn();

  it('renders the form and allows typing', () => {
    render(<QuestionForm sessionId={sessionId} onQuestionSubmit={onQuestionSubmit} />);
    
    expect(screen.getByPlaceholderText('Type your question here...')).toBeInTheDocument();
    
    const input = screen.getByPlaceholderText('Type your question here...');
    fireEvent.change(input, { target: { value: 'A new question' } });
    
    expect(input.value).toBe('A new question');
  });

  it('calls submitQuestion and onQuestionSubmit on form submission', async () => {
    const mockedSubmitQuestion = vi.spyOn(sessionApi, 'submitQuestion').mockResolvedValue(null);

    render(<QuestionForm sessionId={sessionId} onQuestionSubmit={onQuestionSubmit} />);
    
    const input = screen.getByPlaceholderText('Type your question here...');
    fireEvent.change(input, { target: { value: 'A new question' } });

    await act(async () => {
      fireEvent.click(screen.getByText('Submit Question'));
    });

    expect(mockedSubmitQuestion).toHaveBeenCalledWith(sessionId, 'A new question');
    expect(onQuestionSubmit).toHaveBeenCalledTimes(1);
  });
});