# Go Load Balancer

A production-inspired Layer 7 HTTP load balancer written in Go.

It supports round-robin request distribution, active health checks, retry logic, circuit breaker protection, graceful shutdown, middleware-based observability, and a metrics endpoint.

## Features

- Reverse proxy built on `httputil.ReverseProxy`
- Round-robin load balancing
- Active health checks
- Automatic retry on transport failures
- Circuit breaker
- Request metrics
- Structured logging middleware
- Panic recovery middleware
- Graceful shutdown
- YAML based configuration

---

## Architecture

```
                   Client
                      |
                      ▼
              ┌────────────────┐
              │ HTTP Server     │
              └────────────────┘
                      |
                      ▼
        ┌────────────────────────────┐
        │ Middleware Chain           │
        │                            │
        │  Recorder                  │
        │  Recovery                  │
        │  Logging                   │
        │  Metrics                   │
        └────────────────────────────┘
                      |
                      ▼
              Load Balancer
                      |
      ┌───────────────┼────────────────┐
      ▼               ▼                ▼
 Backend A        Backend B        Backend C
```

---

## Request Flow

```
Incoming Request
       │
       ▼
Select Backend
       │
       ▼
Healthy?
       │
       ├── No ─────► Try next backend
       │
       ▼
Circuit Open?
       │
       ├── Yes ───► Try next backend
       │
       ▼
Proxy Request
       │
       ├── Success
       │      │
       │      ▼
       │   Reset Circuit
       │
       └── Failure
              │
              ▼
      Increment Failure Count
              │
              ▼
          Retry (GET/HEAD/OPTIONS)
```

---

## Circuit Breaker

Each backend maintains its own circuit state.

```
Closed
   │
   │ failures >= threshold
   ▼
Open
   │
   │ timeout elapsed
   ▼
Half Open
   │
   ├── Success → Closed
   └── Failure → Open
```

Transport failures contribute towards opening the circuit.

---

## Metrics

The load balancer exposes

```
GET /metrics
```

Example response

```json
{
  "totalRequests": 1240,
  "successfulRequests": 1193,
  "failedRequests": 47,
  "retryCount": 15,
  "panicCount": 0
}
```

---

## Engineering Decisions

- Middleware pipeline keeps cross-cutting concerns (logging, metrics, recovery) separate from routing logic.
- Request-scoped context carries retry and response state without global variables.
- Circuit breaker and health checks are independent:
  - Health checks determine backend availability.
  - Circuit breaker protects against repeated transport failures.
- Atomic operations are used for metrics to avoid lock contention.
- `httputil.ReverseProxy` is used instead of implementing proxying manually.

## Configuration

```yaml
server:
  port: "8080"
  readTimeout: 10s
  writeTimeout: 30s

health:
  interval: 5s
  timeout: 2s

backends:
  - url: http://localhost:9001
  - url: http://localhost:9002
  - url: http://localhost:9003
```

---

## Running

```bash
go run ./cmd/lb
```

---

## Project Structure

```
cmd/
    lb/

config/

internal/

middlewares/

metrics/

request/
```

---

## Technologies

- Go
- net/http
- httputil.ReverseProxy
- sync/atomic
- sync.RWMutex
- context
- YAML configuration

---

## Current Limitations

- Retries are limited to idempotent HTTP methods (GET, HEAD, OPTIONS).
- Circuit breaker currently tracks transport-level failures.
- Response recorder does not implement optional `http.ResponseWriter` interfaces (`Flusher`, `Hijacker`), so streaming/WebSocket proxying is outside the current scope.
- Metrics are in-memory and reset on restart.

---

## Future Improvements

- Weighted round robin
- Least connections
- Passive health checks
- Prometheus metrics
- OpenTelemetry tracing
- Rate limiting
- Configuration reload
- Unit and integration tests
