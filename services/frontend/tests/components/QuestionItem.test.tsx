// /frontend/tests/components/QuestionItem.test.tsx
import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { vi } from 'vitest';
import QuestionItem from '../../src/components/QuestionItem';
import * as sessionApi from '../../src/api/sessionApi';
import { Question } from '../../src/models/Question';

vi.mock('../../src/api/sessionApi');

describe('QuestionItem', () => {
  const sessionId = 'test-session';
  const onVoteSuccess = vi.fn();
  const question: Question = {
    id: 'q1',
    session_id: sessionId,
    text: 'Is this a test question?',
    voters: [],
    votes: 5,
  };

  it('renders the question and vote count', () => {
    render(<QuestionItem sessionId={sessionId} question={question} onVoteSuccess={onVoteSuccess} />);

    expect(screen.getByText('Is this a test question?')).toBeInTheDocument();
    expect(screen.getByTestId('vote-count')).toHaveTextContent('5');
  });

  it('calls voteQuestion and onVoteSuccess on vote button click', async () => {
    const mockedVoteQuestion = vi.spyOn(sessionApi, 'voteQuestion').mockResolvedValue(null);

    render(<QuestionItem sessionId={sessionId} question={question} onVoteSuccess={onVoteSuccess} />);

    await act(async () => {
      fireEvent.click(screen.getByTestId('vote-button'));
    });

    expect(mockedVoteQuestion).toHaveBeenCalledWith(sessionId, question.id);
    expect(onVoteSuccess).toHaveBeenCalledTimes(1);
  });
});