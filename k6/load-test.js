import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';

const BASE_URL = 'https://question-app.duckdns.org';
const WS_BASE = 'wss://question-app.duckdns.org';

const HTTP_VUS = parseInt(__ENV.HTTP_VUS) || 250;
const WS_VUS = parseInt(__ENV.WS_VUS) || 250;

export const options = {
  scenarios: {
    golden_path: {
      executor: 'ramping-vus',
      exec: 'goldenPath',
      startVUs: 0,
      stages: [
        { duration: '1m', target: HTTP_VUS },
        { duration: '2m', target: HTTP_VUS },
        { duration: '30s', target: 0 },
      ],
    },
    ws_connections: {
      executor: 'ramping-vus',
      exec: 'wsConnections',
      startVUs: 0,
      stages: [
        { duration: '1m', target: WS_VUS },
        { duration: '2m', target: WS_VUS },
        { duration: '30s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.05'],
    checks: ['rate>0.95'],
  },
};

function randomId() {
  return Math.random().toString(36).substring(2, 10);
}

function createSession() {
  const res = http.post(
    `${BASE_URL}/api/session`,
    JSON.stringify({ sessionId: `k6-${randomId()}` }),
    { headers: { 'Content-Type': 'application/json' }, tags: { name: 'POST /api/session' } }
  );
  check(res, { 'session created': (r) => r.status === 201 });
  if (res.status !== 201) return null;
  return JSON.parse(res.body);

}

function endSession(sessionId, adminToken) {
  http.del(`${BASE_URL}/api/session/${sessionId}`, null, {
    headers: { Authorization: `Bearer ${adminToken}` },
    tags: { name: 'DELETE /api/session/:id' },
  });
}

// Scenario 1: full golden path — create session, submit questions, vote, end
export function goldenPath() {
  const session = createSession();
  if (!session) return;

  const { sessionId, adminToken } = session;

  sleep(0.5);

  const questionIds = [];
  for (let i = 0; i < 3; i++) {
    const res = http.post(
      `${BASE_URL}/api/session/${sessionId}/questions`,
      JSON.stringify({ text: `Load test question ${i + 1} — VU ${__VU} iter ${__ITER}` }),
      { headers: { 'Content-Type': 'application/json' }, tags: { name: 'POST /api/session/:id/questions' } }
    );
    check(res, { 'question submitted': (r) => r.status === 201 });
    if (res.status === 201) questionIds.push(JSON.parse(res.body).id);
    sleep(0.3);
  }

  for (const qId of questionIds) {
    const res = http.put(
      `${BASE_URL}/api/session/${sessionId}/questions/${qId}/vote`,
      null,
      { headers: { 'Content-Type': 'application/json' }, tags: { name: 'PUT /api/session/:id/questions/:id/vote' } }
    );
    check(res, { 'vote cast': (r) => r.status === 200 });
    sleep(0.2);
  }

  sleep(0.5);

  const endRes = http.del(`${BASE_URL}/api/session/${sessionId}`, null, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${adminToken}`,
    },
    tags: { name: 'DELETE /api/session/:id' },
  });
  check(endRes, { 'session ended': (r) => r.status === 204 });

  sleep(1);
}

// Scenario 2: concurrent WebSocket connections — tests connection capacity
export function wsConnections() {
  const session = createSession();
  if (!session) return;

  const { sessionId, adminToken } = session;
  const url = `${WS_BASE}/api/session/${sessionId}/ws`;

  const res = ws.connect(url, { tags: { name: 'GET /api/session/:id/ws' } }, (socket) => {
    socket.on('message', (data) => {
      const msg = JSON.parse(data);
      check(msg, {
        'valid ws event type': (m) =>
          ['QUESTION_ADDED', 'VOTE_UPDATED', 'QUESTION_DELETED', 'SESSION_ENDED'].includes(
            m.type
          ),
      });
    });

    socket.on('error', (e) => console.error(`WS error [${sessionId}]:`, e));

    socket.setTimeout(() => socket.close(), 5000);
  });

  check(res, { 'ws handshake ok': (r) => r && r.status === 101 });

  endSession(sessionId, adminToken);
}
