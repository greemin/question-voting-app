import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const BASE_URL = 'https://question-app.duckdns.org';
const WS_BASE = 'wss://question-app.duckdns.org';

const HTTP_VUS = parseInt(__ENV.HTTP_VUS) || 10;
const WS_VUS = parseInt(__ENV.WS_VUS) || 10;

// Tracks what fraction of requests were rate-limited (429) — informational, not a failure
const rateLimited = new Rate('rate_limited');

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
    // 429s excluded via responseCallback — only 5xx and network errors count as failures
    http_req_failed: ['rate<0.01'],
    // p95 across all requests including fast 429s from nginx — keeps overall latency low
    http_req_duration: ['p(95)<2000'],
  },
};

function randomId() {
  return Math.random().toString(36).substring(2, 10);
}

// Pre-create a session pool for the WS scenario so WS VUs don't each burn the session rate limit.
// Sessions created here are shared (round-robined) across all WS VUs.
export function setup() {
  const sessions = [];
  for (let i = 0; i < 6; i++) {
    const res = http.post(
      `${BASE_URL}/api/session`,
      JSON.stringify({ sessionId: `k6-ws-${randomId()}` }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    if (res.status === 201) sessions.push(JSON.parse(res.body));
    sleep(0.2);
  }
  return { sessions };
}

export function teardown(data) {
  for (const { sessionId, adminToken } of data.sessions) {
    http.del(`${BASE_URL}/api/session/${sessionId}`, null, {
      headers: { Authorization: `Bearer ${adminToken}` },
    });
    sleep(0.2);
  }
}

// Scenario 1: full golden path — create session, submit questions, vote, end.
// Rate limit zones hit: sessions (10r/m burst=5), questions (10r/m burst=5),
// votes (30r/m burst=10), api (60r/m burst=30).
// 429s are expected under load from a single IP and tracked via rate_limited metric.
export function goldenPath() {
  const res = http.post(
    `${BASE_URL}/api/session`,
    JSON.stringify({ sessionId: `k6-${randomId()}` }),
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'POST /api/session' },
      responseCallback: http.expectedStatuses(201, 429),
    }
  );
  rateLimited.add(res.status === 429);
  check(res, { 'session created': (r) => r.status === 201 });
  if (res.status !== 201) return;

  const { sessionId, adminToken } = JSON.parse(res.body);
  sleep(0.5);

  const questionIds = [];
  for (let i = 0; i < 3; i++) {
    const qRes = http.post(
      `${BASE_URL}/api/session/${sessionId}/questions`,
      JSON.stringify({ text: `Load test question ${i + 1} — VU ${__VU} iter ${__ITER}` }),
      {
        headers: { 'Content-Type': 'application/json' },
        tags: { name: 'POST /api/session/:id/questions' },
        responseCallback: http.expectedStatuses(201, 429),
      }
    );
    rateLimited.add(qRes.status === 429);
    check(qRes, { 'question submitted': (r) => r.status === 201 });
    if (qRes.status === 201) questionIds.push(JSON.parse(qRes.body).id);
    sleep(0.3);
  }

  for (const qId of questionIds) {
    const vRes = http.put(
      `${BASE_URL}/api/session/${sessionId}/questions/${qId}/vote`,
      null,
      {
        headers: { 'Content-Type': 'application/json' },
        tags: { name: 'PUT /api/session/:id/questions/:id/vote' },
        responseCallback: http.expectedStatuses(200, 403, 429),
      }
    );
    rateLimited.add(vRes.status === 429);
    check(vRes, { 'vote cast': (r) => r.status === 200 });
    sleep(0.2);
  }

  sleep(0.5);

  const endRes = http.del(`${BASE_URL}/api/session/${sessionId}`, null, {
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${adminToken}` },
    tags: { name: 'DELETE /api/session/:id' },
    responseCallback: http.expectedStatuses(204, 429),
  });
  rateLimited.add(endRes.status === 429);
  check(endRes, { 'session ended': (r) => r.status === 204 });

  sleep(1);
}

// Scenario 2: concurrent WebSocket connections — tests connection capacity.
// Uses pre-created sessions from setup() to avoid burning the session rate limit budget here.
// WS upgrades use the api zone (60r/m burst=30) — rate limiting visible at high WS VU counts.
export function wsConnections(data) {
  const { sessions } = data;
  if (!sessions || sessions.length === 0) return;

  const { sessionId } = sessions[__VU % sessions.length];
  const url = `${WS_BASE}/api/session/${sessionId}/ws`;

  const res = ws.connect(url, { tags: { name: 'GET /api/session/:id/ws' } }, (socket) => {
    socket.on('message', (msg) => {
      const parsed = JSON.parse(msg);
      check(parsed, {
        'valid ws event type': (m) =>
          ['QUESTION_ADDED', 'VOTE_UPDATED', 'QUESTION_DELETED', 'SESSION_ENDED'].includes(m.type),
      });
    });
    socket.on('error', (e) => console.error(`WS error [${sessionId}]:`, e));
    socket.setTimeout(() => socket.close(), 5000);
  });

  check(res, { 'ws handshake ok': (r) => r && r.status === 101 });
  sleep(1);
}
