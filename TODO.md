# 📝 Project Roadmap: Question Voting App

This document outlines the development plan for the application. Phases are organized by priority, focusing on building a robust, scalable, and user-friendly platform.

---

### ✅ **Phase 0: Foundation & MVP (Completed)**

-   [x] **Project Scaffolding**: Initial repository and file structure setup for frontend and backend services.
-   [x] **Storage Abstraction**: Implemented a `Storer` interface in the backend to decouple storage logic.
-   [x] **Backend Unit Tests**: Comprehensive tests for API handlers and storage functionality.
-   [x] **TypeScript Migration**: Converted the frontend codebase to TypeScript for improved type safety.
-   [x] **Frontend Component Tests**: Built out a suite of tests for React components and pages using RTL.

---

### 🚀 **Phase 1: Core Architecture & Production Readiness (Highest Priority)**

*This phase focuses on upgrading the core architecture to be scalable, real-time, and ready for deployment.*

-   [x] **Containerize the Application (Docker)**
    -   Simplify development setup and standardize deployment. This is a prerequisite for easier database management.
    -   [x] **Action**: Create a `Dockerfile` for both the Go backend and the React frontend.
    -   [x] **Action**: Create a `docker-compose.yml` file to orchestrate the services (including a MongoDB service) for easy local development.
    -   [x] **Action**: Try out new docker setup
        -   [x] For Development: Run docker-compose up (override is loaded automatically).
        -   [x] For Production: Be explicit and run docker compose -f docker-compose.yml up (avoids the development override).
    -   [x] **Action**: Update README

-   [ ] **Database Migration to MongoDB**
    -   [x] Replace the current file-based storage with a MongoDB database to ensure scalability and reliability.
    -   [x] **Action**: Implement a new `MongoStorer` that satisfies the `Storer` interface.
    -   [x] **Action**: Use environment variables for the connection string and database configuration, provided by the Docker setup.
    -   **Action**: Check that mongodb is started in secure mode.

-   [ ] **Implement Real-Time Updates with WebSockets**
    -   Transition from HTTP polling to WebSockets for instant updates to questions and votes.
    -   **Action (Backend)**: Integrate a WebSocket library (e.g., `gorilla/websocket`) to broadcast updates to clients in a session.
    -   **Action (Frontend)**: Use the native `WebSocket` API to listen for and display real-time changes.

-   [ ] **Externalize Configuration**
    -   Remove hardcoded configuration values from the codebase.
    -   **Action**: Move all environment-specific values (ports, CORS origins, database URIs, cookie settings) to environment variables. Consider a library like `Viper` for structured configuration.


---

### 🛡️ **Phase 2: Security & Admin Features**

*This phase improves the application's security and expands admin capabilities.*

-   [ ] **Improve Admin Authorization**
    -   Replace the simple cookie-based admin check with a more secure method.
    -   **Action**: Implement a secret token-based system. When a session is created, return a unique admin token to the creator, who must then provide it in an `Authorization` header for protected actions.

-   [ ] **Allow Admin to Delete a Question**
    -   Give session admins more control over the content.
    -   **Action (Backend)**: Create a new `DELETE /api/session/{sessionId}/questions/{questionId}` endpoint, protected by the new admin authorization.
    -   **Action (Frontend)**: Add a "Delete" button to `QuestionItem.tsx` that is only visible to the admin.

---

### ✨ **Phase 3: UX/UI & Feature Enhancements**

*This phase focuses on improving the user experience and visual design.*

-   [ ] **UI Redesign & Component Library**
    -   Overhaul the visual design for a modern, mobile-first experience.
    -   **Action**: Adopt a React component library like **Material-UI (MUI)**, **Chakra UI**, or **Mantine** to standardize components and accelerate development.

-   [ ] **Implement Confirmation Modals**
    -   Prevent accidental destructive actions.
    -   **Action**: Use the chosen component library to add a confirmation dialog before an admin ends a session or deletes a question.

-   [ ] **Add "Copy Link" Button**
    -   Make it easier for users to share the session URL.
    -   **Action**: Add a "Copy to Clipboard" button on the `VotingSessionPage.tsx`.

-   [ ] **Improve Input Validation & Error Handling**
    -   Provide better feedback to the user.
    -   **Action (Frontend)**: Add basic input validation (e.g., max question length) to the `QuestionForm`.
    -   **Action (Frontend)**: Implement a toast notification system (e.g., `react-hot-toast`) to display API errors and other feedback.

---

### 🧪 **Phase 4: Testing & Deployment**

*This phase ensures the application is reliable and easy to deploy.*

-   [ ] **Add Integration Tests**
    -   Test the full application flow from end to end.
    -   **Action**: Write tests that cover user flows across both the frontend and backend (e.g., creating a session, submitting a question, and seeing it appear in real-time).

-   [ ] **Add Mongo Integration Tests**
    -   Write integration tests for the `MongoStorage` implementation.
    -   **Action**: Use Testcontainers to spin up an ephemeral MongoDB database during testing to ensure queries and connection logic work correctly.

-   [ ] **Set Up CI/CD Pipeline**
    -   Automate testing and deployment.
    -   **Action**: Create a GitHub Actions workflow that automatically runs all tests on push/pull request.
    -   **Action**: Extend the workflow to build and push Docker images, and eventually deploy to a hosting provider.
