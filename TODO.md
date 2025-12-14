📝 Project Development Checklist (Question Voting App)
Phase 1: Testing & Refinement (Current Codebase)

    [ ] General project/files structure and setup

        [ ] Move session JSON files into own folder away from source code (e.g., create a /data directory outside of /backend).

        [ ] Review and refine the entire project file tree structure (current structure feels weird).
        
        [ ] move project to git platform and look for way for secure LLM access to current state of code

    [ ] Backend Unit Tests (Go):

        [ ] Test loadSessionData and saveSessionData for file I/O and locking integrity.

        [ ] Test CreateSessionHandler to ensure a file and Admin ID are correctly created.

        [ ] Test CheckAdminHandler to ensure the correct boolean is returned based on the userSessionId cookie.

        [ ] Test EndSessionHandler (DELETE) with both authorized (admin) and unauthorized (non-admin) userSessionId cookies.

        [ ] Test AddQuestionHandler with a malformed payload (e.g., missing question text).

        [ ] Test VoteHandler to verify that a user cannot vote more than once per question.

    [ ] Frontend Component Tests (RTL):

        [ ] Test HomePage component rendering and successful navigation on button click.

        [ ] Test QuestionForm component to ensure the input field works and calls the API correctly upon submission.

        [ ] Test QuestionItem component rendering of question text, votes, and user-friendly timestamp.

        [ ] Test QuestionItem vote button: check that it disables after a click and calls the API once.

        [ ] Test VotingSessionPage to ensure the Admin button correctly appears based on the checkAdminStatus API call.

Phase 2: Feature Implementation (New Functionality)

    [ ] Live Updates:

        [ ] Frontend: Update the polling mechanism to a more efficient solution like WebSockets (Go's golang.org/x/net/websocket or a third-party library) for real-time question and vote updates.

        [ ] Backend: Implement the WebSocket connection and broadcast vote/question updates to all connected clients.

    [ ] Voting Session Features:

        [ ] Backend: Add a feature to allow the Admin to delete an individual question.

        [ ] Frontend: Add a Delete Question button (visible only to Admin) to the QuestionItem component.

        [ ] Frontend: Implement confirmation modal/dialog before deleting a session or a question.

    [ ] UX/UI Improvements:

        [ ] Implement a simple "Copy Link" button for the session URL to make sharing easier.

        [ ] Add basic input validation on the frontend (e.g., max question length).

        [ ] Implement better error display for API failures (e.g., a toast notification instead of just console logging).
        
        [ ] general redesign
        
      	    [ ] components library like material ui? other inspiration from modern websites?
      	    
      	    [ ] make as much question cards visible with mobile first in mind
