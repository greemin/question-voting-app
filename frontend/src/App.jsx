// /frontend/src/App.jsx
import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import HomePage from './pages/HomePage';
import VotingSessionPage from './pages/VotingSessionPage';

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
