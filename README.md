# 🚀 Question Voting App

A real-time, lightweight application for crowdsourcing and ranking questions during a presentation, Q&A session, or meeting. Built using a Go backend for fast API handling and a React/Vite frontend for a modern, responsive user experience.

## 🛠️ Setup and Installation
Prerequisites

You must have the following installed on your system:

- Go: Version 1.22 or higher

- Node.js & npm: Version 18 or higher


### 1. Backend Service Setup (Go API)

The backend handles session management, voting, and storage.

#### 1. Navigate to the backend directory:

    cd services/backend

#### 2. Download the required Go modules:

    go mod tidy

Start the Go server:

    go run ./cmd/server/main.go

The API should start on http://localhost:8080.

### 2. Frontend Service Setup (React Client)

The frontend is a single-page application (SPA) built with React and bundled with Vite.

#### 1. Navigate to the frontend directory:
    cd services/frontend

#### 2.Install the Node dependencies:
    npm install

#### 3. Start the development server:
    npm run dev

The application will typically be available at http://localhost:5173 (or similar).

### 3. Running the App 🏁

Open your browser to the address provided by the frontend server (e.g., http://localhost:5173).