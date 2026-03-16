# :question: :ballot_box_with_check: Question Voting App

A real-time, lightweight application for crowdsourcing and ranking questions during a presentation, Q&A session, or meeting. Built using a Go backend for fast API handling and a React/Vite frontend for a modern, responsive user experience.

## 🛠️ Setup and Installation
Prerequisites

You must have **Docker** and **Docker Compose** installed on your system.

### 1. Running the Application (Local Development)

This project is fully containerized. To spin up the React frontend, Go backend, and MongoDB database, simply run:

    docker compose up --build

The following services will be available:
* **Frontend:** http://localhost:5174
* **Backend API:** http://localhost:8081
* **MongoDB:** mongodb://localhost:27017 (using the credentials `devroot` / `devpassword`)


### 2. Frontend Service Setup (React Client)

The frontend is a single-page application (SPA) built with React and bundled with Vite.

#### 1. Navigate to the frontend directory:
    cd services/frontend

#### 2.Install the Node dependencies:
    npm install

To stop the application, run `docker compose down`. If you want to wipe the local database completely, run `docker compose down -v`.
