// /frontend/src/App.tsx
import React from 'react';
import { Toaster } from 'react-hot-toast';
import { Routes, Route } from 'react-router-dom';
import HomePage from './pages/HomePage.tsx';
import VotingSessionPage from './pages/VotingSessionPage.tsx';

function App() {
  return (
    <>
      <Toaster />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/:sessionId" element={<VotingSessionPage />} />
      </Routes>
    </>
  );
}

export default App;
