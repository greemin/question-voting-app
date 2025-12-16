// /frontend/src/App.tsx
import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import HomePage from './pages/HomePage.tsx';
import VotingSessionPage from './pages/VotingSessionPage.tsx';

function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      {/* Dynamic route for the voting session ID */}
      <Route path="/votingSession/:sessionId" element={<VotingSessionPage />} />
    </Routes>
  );
}

export default App;
