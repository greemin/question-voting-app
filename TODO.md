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

### ✅ 🚀 **Phase 1: Core Architecture & Production Readiness (Highest Priority)**

*This phase focuses on upgrading the core architecture to be scalable, real-time, and ready for deployment.*

-   [x] **Containerize the Application (Docker)**
    -   Simplify development setup and standardize deployment. This is a prerequisite for easier database management.
    -   [x] **Action**: Create a `Dockerfile` for both the Go backend and the React frontend.
    -   [x] **Action**: Create a `docker-compose.yml` file to orchestrate the services (including a MongoDB service) for easy local development.
    -   [x] **Action**: Try out new docker setup
        -   [x] For Development: Run docker-compose up (override is loaded automatically).
        -   [x] For Production: Be explicit and run docker compose -f docker-compose.yml up (avoids the development override).
    -   [x] **Action**: Update README

-   [x] **Database Migration to MongoDB**
    -   [x] Replace the current file-based storage with a MongoDB database to ensure scalability and reliability.
    -   [x] **Action**: Implement a new `MongoStorer` that satisfies the `Storer` interface.
    -   [x] **Action**: Use environment variables for the connection string and database configuration, provided by the Docker setup.
    -   [x] **Action**: Check that mongodb is started in secure mode.

-   [x] **Routes/Handlers & Human-Readable URLs**
    -   Users should be able to name their sessions freely with human-readable slugs instead of UUIDs.
    -   [x] **Action (Backend - Database)**: Add `CreatedAt` to the Session model/interface. Configure MongoDB on startup to create a unique index on `sessionId` and a TTL index on `createdAt` (e.g., 24-48 hours) to automatically purge old sessions.
    -   [x] **Action (Backend - API)**: Change router/handlers to accept custom `sessionId` strings instead of strict UUIDs. Securely validate the input (URL-safe characters only).
    -   [x] **Action (Backend - Collision Handling)**: If a `sessionId` collision occurs during insertion (MongoDB duplicate key error E11000), a random 4-character suffix is automatically appended, and the final unique slug is returned.
    -   [x] **Action (Frontend)**: Update the "Create Session" API call and UI to optionally capture and send a user-defined slug. Redirect the user to the actually created session URL returned by the backend.
    -   [x] **Action (Backend - Tests)**: Update tests to changes and tests slugify logic and slugcollision behavior.
    -   [x] **Action (Backend)**: Readd IsDuplicateKeyError check for mongo collision and find out why the check is failing and fix it.
    -   [x] **Action (Backend)**: Use unicode character classes in slug regexp
    check in order to support non-English languages.

-   [x] **Propagate Contexts to Database Layer**
    -   [x] **Action**: Update the `Storer` interface and `MongoStorage` implementation to accept a `context.Context` from HTTP handlers instead of hardcoding `context.Background()`. This ensures database queries are automatically cancelled if an HTTP request times out or is aborted by the user.

-   [x] **Externalize Configuration**
    -   Remove hardcoded configuration values from the codebase.
    -   [x] **Action**: Move all environment-specific values (ports, CORS origins, database URIs, cookie settings) to environment variables. Consider a library like `Viper` for structured configuration.

-   [x] **Implement Real-Time Updates with WebSockets**
    -   Transition from HTTP polling to WebSockets for instant updates to questions and votes.
    -   [x] **Action (Backend)**: Integrate a WebSocket library (e.g., `gorilla/websocket`) to broadcast updates to clients in a session.
    -   [x] **Action (Frontend)**: Use the native `WebSocket` API to listen for and display real-time changes.


---

### ✅ 🛡️ **Phase 2: Security & Admin Features**

*This phase improves the application's security and expands admin capabilities.*

-   [x] **Improve Admin Authorization**
    -   Replace the simple cookie-based admin check with a more secure method.
    -   [x] **Action**: Implement a secret token-based system. When a session is created, return a unique admin token to the creator, who must then provide it in an `Authorization` header for protected actions.

-   [x] **Allow Admin to Delete a Question**
    -   Give session admins more control over the content.
    -   [x] **Action (Backend)**: Create a new `DELETE /api/session/{sessionId}/questions/{questionId}` endpoint, protected by the new admin authorization.
    -   [x] **Action (Frontend)**: Add a "Delete" button to `QuestionItem.tsx` that is only visible to the admin.

---

### ✅ ✨ **Phase 3: UX/UI & Feature Enhancements**

*This phase focuses on improving the user experience and visual design.*

-   [x] **UI Redesign**
    -   Overhaul the visual design for a modern, mobile-first experience without adding heavy dependencies.
    -   [x] **Action (Frontend)**: Implement basic grid layout using media queries for breakpoints.
    -   [x] **Action (Frontend)**: Settle on default theme for the app.

-   [x] **Add "Copy Link" Button**
    -   Make it easier for users to share the session URL.
    -   [x] **Action**: Add a "Copy to Clipboard" button on the `VotingSessionPage.tsx`.
    -   [x] **Action**: Add admin links to session.
    -   [x] **Action**: Add QR code links to session.

-   [x] **Improve Input Validation & Error Handling**
    -   Provide better feedback to the user.
    -   [x] **Action (Frontend)**: Add basic input validation (e.g., max question length) to the `QuestionForm`.
    -   [x] **Action (Frontend)**: Implement a toast notification system (e.g., `react-hot-toast`) to display API errors and other feedback.

---

### ✅ 🧪 **Phase 4: Testing & CI/CD Pipeline**

*This phase ensures the application is reliable and easy to deploy.*

-   [x] **Add Integration Tests**
    -   Test the full application flow from end to end.
    -   [x] **Action**: Write tests that cover user flows across both the frontend and backend (e.g., creating a session, submitting a question, and seeing it appear in real-time).

-   [x] **Add Mongo Integration Tests**
    -   Write integration tests for the `MongoStorage` implementation.
    -  [x] **Action**: Use Testcontainers to spin up an ephemeral MongoDB database during testing to ensure queries and connection logic work correctly.

-   [x] **Set Up CI/CD Pipeline**
    -   Automate testing and deployment.
    -   [x] **Action**: Create a GitHub Actions workflow that automatically runs all tests on push/pull request.
    -   [x] **Action**: Create a GitHub Actions workflow that builds and pushes backend and frontend Docker images to GHCR.

---

### ✅ 🚀 **Phase 5: Deployment**

*First real deployment to a hosted environment.*

-   [x] **Choose a Hosting Provider**
    -   [x] **Action**: Provision a Hetzner VPS, clone the repo, and configure the `.env` file with production secrets.
    -   [x] **Action**: Verify MongoDB hosting strategy — managed Atlas free tier or self-hosted on the same VPS.

-   [x] **Provision Production Environment**
    -   [x] **Action**: Set up production environment variables (MongoDB URI, CORS origin, cookie settings, admin secrets).
    -   [x] **Action**: Ensure MongoDB runs in auth-enabled mode with a dedicated user for the app.
    -   [x] **Action**: Configure TLS — HTTPS for the frontend and WSS for WebSocket connections.

-   [x] **Deploy**
    -   [x] **Action**: Deploy using `docker-compose.prod.yml` — pull latest images and restart containers.
    -   [x] **Action**: Smoke test the golden path after deploy: create session, submit question, vote, end session.
    -   [x] **Action**: Verify WebSocket connections work end-to-end in the hosted environment.

-   [x] **Observability**
    -   [x] **Action**: Set up basic logging and error visibility (provider logs or a lightweight tool like Grafana/Loki).
    -   [x] **Action**: Set up an uptime monitor (e.g. UptimeRobot) to alert on downtime.

-   [x] **Load Testing (post-deploy)**
    -   Only meaningful against the real hosted infrastructure — see Long-Term Goals.
    -   [x] **Action**: Point k6 at the production URL and run the session/WebSocket load scenarios.
    -   [x] **Action**: Record baseline metrics (concurrent sessions, connections, response times) to inform future performance decisions.

-   [x] **Cloudflare (post load tests):** Load test results (0 errors at 500 VUs, Hetzner DDoS protection in place) show no immediate need. Revisit if the app scales beyond a demo instance.

---

### ✅ 🔍 **Phase 6: Post-Deployment Reevaluation**

*Structured review after first production deployment — verifying what was built, closing gaps identified under real load, and making infrastructure decisions while the codebase is still fresh.*

-   [x] **Load Testing:** k6 scripts written and executed against the Hetzner VPS. Baseline results documented in `k6/results/BASELINE.md`. Key finding: HTTP stays within thresholds at 500 VUs; WS connect time degrades at 250 concurrent WS VUs (p95 = 10s) due to synchronous MongoDB lookup on upgrade — tracked separately under WebSocket hardening.

-   [ ] **Nginx rate limit smoke test (post-deploy):** After deploying, manually verify each rate limit zone returns 429 (not 503) when exceeded.

    -   [ ] **Questions (10r/m, burst=5):** `for i in $(seq 1 15); do curl -s -o /dev/null -w "%{http_code}\n" -X POST https://question-app.duckdns.org/api/session/<session_id>/questions -H 'Content-Type: application/json' -b 'userSessionId=<cookie>' -d '{"text":"test"}'; done` — expect first 6 (1 + burst) to return 201, remainder 429.

    -   [ ] **Sessions (10r/m, burst=5):** `for i in $(seq 1 15); do curl -s -o /dev/null -w "%{http_code}\n" -X POST https://question-app.duckdns.org/api/session -H 'Content-Type: application/json' -d '{}'; done` — expect first 6 (1 + burst) to return 200 or 201, remainder 429.

    -   [ ] **Votes (30r/m, burst=10):** First create a question and grab its ID, then: `for i in $(seq 1 40); do curl -s -o /dev/null -w "%{http_code}\n" -X POST https://question-app.duckdns.org/api/session/<session_id>/questions/<qid>/vote -b 'userSessionId=<cookie>'; done` — expect first 11 (1 + burst) to return 200 or 409 (already voted), remainder 429.

-   [ ] **E2E Tests in CI:** The Playwright e2e job is currently disabled (`if: false` in `ci.yml`) due to flaky startup timing — the docker stack (especially MongoDB) takes longer to initialise on cold CI runners than the wait timeout allows. Fix options: tune timeouts further, add per-service healthcheck polling, or use a pre-built image cache to speed up the stack startup.

-   [ ] **SQLite storage backend:** Add `SQLiteStorer` implementing the existing `Storer` interface as an alternative to MongoDB. Store questions and voters as a JSON column on the sessions table (maps naturally to the current whole-doc-replace update pattern). Replace MongoDB's TTL index with a periodic cleanup goroutine. Wire up via a `DB_DRIVER` env var in `main.go`. Use `modernc.org/sqlite` (pure Go, no CGo). Selectable via `DB_DRIVER` env var alongside MongoDB — not a replacement, just an additional option. When using SQLite the MongoDB container can simply be omitted from the stack.

---

### 🔮 **Backlog, Tech-Debt and Long-Term Goals**

*Low-urgency items — revisit if the app sees meaningful outside interest or scale.*

-   [x] **i18n:** Add localization via lazy loaded json files based on browser language.
-   [x] **Shorten Session URLs:** Change session path from "BASE_URL/votingSession/SESSION_SLUG" to "BASE_URL/SESSION_SLUG"
-   [x] **Insecure Direct Object Reference (IDOR):** Resolved — router migrated to Go 1.22 native path parameters (`{session_id}`, `{question_id}`). All handlers use `r.PathValue()`, no manual path splitting remains.

-   [ ] **IP-based vote deduplication:** Votes are currently deduplicated by `userSessionId` cookie. A script that fetches a fresh cookie per request (one `GET /api/session/{id}` then one `PUT .../vote`) can still cast unlimited votes — nginx rate limiting (30/min) slows this but doesn't stop it. Fix: store the submitter IP on each vote entry and reject if the same IP already voted on that question, mirroring the existing `BannedIPs` / `SubmitterIP` mechanism on questions. Tradeoff: one vote per question per shared IP (office NAT, university networks).

-   [ ] **Missing Input Validation (XSS):** Session ID character validity is handled by `slugify()`. Remaining gap: question text in `SubmitQuestionHandler` is stored as-is with no backend sanitization. React prevents XSS on the frontend, but defence in depth would require sanitizing or escaping on the backend as well.

-   [ ] **WebSocket hardening (two related items):** At 500 VUs, the session fetch on WS upgrade is the main bottleneck — load test showed p95 WS connect time of 10s on a 2 vCPU box. Two fixes, ~75 lines combined, no new dependencies: (1) **Session cache** — skip the MongoDB lookup on WS connect by caching recently verified session IDs in an in-memory map with a short TTL; invalidate on session end. (2) **Per-IP connection cap** — track open connections per IP in the Hub and reject upgrades above a threshold (e.g. 5) before goroutines are allocated, closing the goroutine exhaustion avenue for flooding attacks.

-   [ ] **CI Auto-Deploy to Hetzner:** Extend the CI workflow to automatically deploy to the VPS after a successful image build (e.g. via SSH + `docker compose pull && up -d`).

-   [ ] **MongoDB Connection Pool Tuning:** Review and tune the Go MongoDB driver's connection pool size (`maxPoolSize`) for high-concurrency workloads. Default pool may be undersized for 500+ concurrent requests, contributing to query queuing under load.

-   [ ] **Voter tracking:** The current implementation stores an array of `voterID`s for each question. This could become inefficient for questions with many votes. A different data structure might be better, or a separate collection/table to track votes.

-   [ ] **Separation of Concerns:** The handlers are directly interacting with the storage layer. In a larger application, it would be better to have a service layer in between to handle business logic.

-   [ ] **Cloudflare / Custom Domain (production scale):** For a larger production instance, evaluate moving DNS to Cloudflare for DDoS protection beyond Hetzner's baseline, caching, and hiding the server IP. Requires a custom domain.

