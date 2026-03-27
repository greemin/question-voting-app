import { Question } from './Question';

export interface SessionData {
  sessionId: string;
  sessionTitle: string;
  isActive: boolean;
  createdAt: string;
  questions: Question[];
  adminToken?: string;
}
