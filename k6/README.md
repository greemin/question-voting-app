# Load Tests

k6 load tests targeting the production environment at `https://question-app.duckdns.org`.

## Scenarios

| Scenario | Description |
|---|---|
| `golden_path` | Full HTTP flow: create session, submit 3 questions, vote, end session |
| `ws_connections` | Concurrent WebSocket connections held open for 5s each |

## Running

**Default (250 VUs each):**
```bash
k6 run k6/load-test.js
```

**Save results:**
```bash
k6 run k6/load-test.js --out json=k6/results/$(date +%Y-%m-%d_%H-%M).json
```

**Custom VU counts** (useful for finding the WS bottleneck threshold):
```bash
k6 run --env HTTP_VUS=100 --env WS_VUS=100 k6/load-test.js
```

**WS bottleneck ladder** — run at increasing counts and compare `ws_connecting p95`:
```bash
for vus in 50 100 150 200 250; do
  k6 run --env WS_VUS=$vus --env HTTP_VUS=50 \
    --out json=k6/results/ws-ladder-${vus}vus.json \
    k6/load-test.js
done
```

## Thresholds

| Metric | Threshold |
|---|---|
| `http_req_duration` | p95 < 2000ms |
| `http_req_failed` | rate < 5% |
| `checks` | pass rate > 95% |

## Baseline Results (2 vCPU / 4 GB Hetzner VPS)

| VUs (HTTP + WS) | p95 HTTP | p95 WS connect | Errors |
|---|---|---|---|
| 10 + 50 | 33ms | 91ms | 0% |
| 250 + 250 | 1.24s | 10.11s | 0% |
