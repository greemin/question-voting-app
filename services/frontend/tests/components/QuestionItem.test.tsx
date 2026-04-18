// /frontend/tests/components/QuestionItem.test.tsx
import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/preact';
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

describe('QuestionItem admin-only buttons', () => {
  const sessionId = 'test-session';
  const onVoteSuccess = vi.fn();
  const question: Question = {
    id: 'q1',
    session_id: sessionId,
    text: 'Is this a test question?',
    voters: [],
    votes: 5,
  };

  it('delete button is not visible to non-admin', () => {
    render(<QuestionItem sessionId={sessionId} question={question} isAdmin={false} onVoteSuccess={onVoteSuccess} />);
    expect(screen.queryByTestId('delete-button')).not.toBeInTheDocument();
  });

  it('delete button is visible to admin', () => {
    render(<QuestionItem sessionId={sessionId} question={question} isAdmin={true} onVoteSuccess={onVoteSuccess} />);
    expect(screen.getByTestId('delete-button')).toBeInTheDocument();
  });

  it('ban button is not visible to non-admin', () => {
    render(<QuestionItem sessionId={sessionId} question={question} isAdmin={false} onVoteSuccess={onVoteSuccess} />);
    expect(screen.queryByTestId('ban-button')).not.toBeInTheDocument();
  });

  it('ban button is visible to admin', () => {
    render(<QuestionItem sessionId={sessionId} question={question} isAdmin={true} onVoteSuccess={onVoteSuccess} />);
    expect(screen.getByTestId('ban-button')).toBeInTheDocument();
  });

  it('calls banSubmitter with sessionId and questionId on ban button click', async () => {
    const mockedBanSubmitter = vi.spyOn(sessionApi, 'banSubmitter').mockResolvedValue(null);

    render(<QuestionItem sessionId={sessionId} question={question} isAdmin={true} onVoteSuccess={onVoteSuccess} />);

    await act(async () => {
      fireEvent.click(screen.getByTestId('ban-button'));
    });

    expect(mockedBanSubmitter).toHaveBeenCalledWith(sessionId, question.id);
  });
});