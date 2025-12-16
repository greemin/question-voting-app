// /frontend/src/models/Question.ts
export interface Question {
  id: string;
  session_id: string;
  text: string;
  votes: number;
  voters: string[];
}
