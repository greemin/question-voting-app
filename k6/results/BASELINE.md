# Load Test Baseline Results

Infrastructure: Hetzner VPS — 2 vCPU / 4 GB RAM  
Target: https://question-app.duckdns.org  
Date: 2026-04-17

## Run 1 — Smoke (10 HTTP + 50 WS VUs)

| Metric | avg | p90 | p95 | max |
|---|---|---|---|---|
| http_req_duration | 26.18ms | 30.79ms | 33.1ms | 70.22ms |
| ws_connecting | 81.23ms | 88.32ms | 91.26ms | 373.14ms |
| ws_session_duration | 5.08s | 5.08s | 5.09s | 5.37s |

| | |
|---|---|
| Total requests | 3444 |
| Errors | 0% |
| Checks passed | 100% |
| Iterations | 1050 |

## Run 2 — Load (250 HTTP + 250 WS VUs)

| Metric | avg | p90 | p95 | max |
|---|---|---|---|---|
| http_req_duration | 469.87ms | 793.13ms | 1.24s | 13.86s |
| ws_connecting | 1.52s | 3.82s | 10.11s | 13.95s |
| ws_session_duration | 6.52s | 8.82s | 15.11s | 18.95s |

| | |
|---|---|
| Total requests | 57966 |
| Errors | 0% |
| Checks passed | 100% |
| Iterations | 11379 |

## Notes

- HTTP performance stays within threshold (p95 < 2s) at 500 VUs
- WS connection time degrades significantly at 250 concurrent WS VUs (p95 = 10s)
- Root cause: MongoDB session lookup on WS upgrade queues under CPU pressure
- No failures at any load level — ceiling is latency, not errors
