# Job Forge

![Job Forge](assets/logo.png)

A concurrent notification service in Go, starting with **email** and designed to scale to new channels (SMS, push, WhatsApp, and more) without changing the core processing pipeline.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)
![Status](https://img.shields.io/badge/status-roadmap-blue)
![License](https://img.shields.io/badge/license-MIT-green)

---

## Why this project

`job-forge` is a practical playground for mastering Go runtime and concurrency using a real-world use case:

- receive notification jobs through an API/event source
- enqueue and process jobs asynchronously
- execute jobs with multiple workers
- handle retries, timeouts, and graceful shutdown
- keep the architecture ready for multiple channels

Initial scope: **email channel only**.  
Design goal: **easy channel expansion** with minimal core changes.

---

## Core objectives

- Build an asynchronous email notification pipeline connected to a queue
- Apply Go concurrency fundamentals in progressive phases (`F1` to `F4`)
- Keep channel implementation pluggable via clear interfaces
- Provide observability and predictable runtime behavior

---

## High-level architecture

```text
Producer/API -> Queue -> Worker Pool -> Channel Router -> Email Provider
                              |
                              +-> Retry/Timeout Policy
                              |
                              +-> Results (fan-in) -> Metrics
```

### Main components

- **Queue Adapter**
  - Responsible for enqueue/dequeue operations
  - Starts with one queue implementation, designed for swappable adapters
- **Worker Pool**
  - Configurable number of workers consuming jobs concurrently
  - Handles lifecycle and shutdown
- **Channel Router**
  - Routes each job to a channel handler (`email` now, others later)
- **Channel Handler Interface**
  - `Handle(ctx, job) error` contract for each channel implementation
- **Retry/Timeout Engine**
  - Encapsulates resiliency logic with backoff and cancellation
- **Metrics Aggregator**
  - Tracks throughput, success/failure, retries, and latency

---

## Extensible channel model

Even with only email enabled, the domain is modeled for expansion:

- `email` (implemented first)
- `sms` (future)
- `push` (future)
- `whatsapp` (future)

The processing pipeline must stay channel-agnostic; only handlers vary by channel.

---

## Functional roadmap

This roadmap maps features directly to the concurrency fundamentals.

### F1 — Goroutines + channels + basic workers

**Goal:** create the first working asynchronous email flow.

**Status (implemented in this repo):** completed the basic end-to-end pipeline for `email`.

**What’s running in F1 now:**
- `cmd/notifier/main.go` boots the process, starts the RabbitMQ consumer, and feeds a Go `jobs` channel
- `internal/worker/pool.go` spawns `WORKER_COUNT` goroutines and runs jobs concurrently
- `internal/channels/router.go` routes by `job.Channel` (`email` now)
- `internal/channels/email/handler.go` decodes the payload and simulates delivery (with an optional failure flag)

**Failure semantics in F1 (early retry behavior):**
- On handler error: `Nack(requeue=true)` for the first failure
- On handler error again (RabbitMQ redelivers): `Nack(requeue=false)` so the message “expires” (or goes to DLQ if you configure one later)

**Output:** queue-fed email jobs processed concurrently by fixed workers.

---

### F2 — Configurable worker pool + WaitGroup + graceful shutdown

**Goal:** make runtime lifecycle production-like.

- Add configurable `WORKER_COUNT`
- Use `sync.WaitGroup` to control worker lifecycle
- Capture `SIGINT/SIGTERM` and start graceful shutdown flow
- Stop intake, drain in-flight jobs, and terminate cleanly

**Output:** reliable shutdown with no dropped in-flight jobs.

---

### F3 — Retry + timeout + select

**Goal:** add resiliency against transient provider failures.

- Implement retry strategy (exponential backoff + jitter)
- Add per-job timeout via `context.WithTimeout`
- Use `select` to coordinate success, timeout, and cancellation paths
- Track retry attempts and final failure classification

**Output:** robust delivery attempts with deterministic failure behavior.

---

### F4 — Fan-in / Fan-out + simple metrics

**Goal:** improve scalability pattern and visibility.

- Fan-out: distribute jobs across worker goroutines
- Fan-in: consolidate worker results into a single aggregator
- Record key metrics:
  - `processed_total`
  - `success_total`
  - `failed_total`
  - `retry_total`
  - `avg_latency_ms`
- Expose lightweight metrics endpoint

**Output:** measurable runtime behavior and ready base for observability tools.

---
## How to run F1 locally (RabbitMQ + worker)
1. Start RabbitMQ (docker):
   - `docker compose up -d`
2. Create `.env` from `.env.example` (adjust only if needed):
   - `cp .env.example .env`
3. Run the worker:
   - `go run ./cmd/notifier`
4. Publish a test message to the bound queue:
   - Exchange: `RABBITMQ_EXCHANGE` (`jobforge.exchange`)
   - Routing key: `RABBITMQ_ROUTING_KEY` (`email.send`)

Example job JSON:
```json
{
  "id": "job-1",
  "channel": "email",
  "payload": {
    "to": "user@example.com",
    "subject": "hello",
    "body": "test"
  },
  "metadata": { "source": "manual" },
  "attempt": 0
}
```

To test failure/“expira”:
```json
{
  "id": "job-2",
  "channel": "email",
  "payload": {
    "to": "user@example.com",
    "subject": "will fail",
    "body": "test",
    "should_fail": true
  },
  "metadata": { "source": "manual" },
  "attempt": 0
}
```

---
## Next steps (F2 -> F4)
1. F2: improve graceful shutdown by cancelling the AMQP consumer and ensuring in-flight jobs are handled deterministically
2. F3: add proper retry engine (attempt counter + backoff/jitter) and per-job timeout (`context.WithTimeout`)
3. F4: implement results fan-in + basic metrics counters/latency, and expose a lightweight metrics endpoint

---

## Feature backlog (after F4)

- Dead-letter queue (DLQ)
- Idempotency keys to avoid duplicate sends
- Priority queues (`high`, `normal`, `low`)
- Scheduled notifications (`send_at`)
- Multi-provider failover for email
- OpenTelemetry tracing + Prometheus integration
- Persistent job state storage

---

## Suggested project structure

```text
cmd/
  notifier/            # app bootstrap
internal/
  app/                 # wiring
  config/              # env/config loading
  domain/              # job models and interfaces
  queue/               # queue adapters
  worker/              # worker pool and execution
  channels/
    email/             # email handler/provider integration
  retry/               # retry policies
  metrics/             # counters/latency aggregation
```

---

## Contribution guide

Contributions are welcome. Please keep changes aligned with the roadmap phases.

1. Open an issue describing the feature or bug
2. Propose scope and affected roadmap phase (`F1`..`F4` or backlog)
3. Submit PR with:
   - clear behavior description
   - tests (when applicable)
   - notes on concurrency/runtime impact

---

## Development principles

- Prefer explicit `context.Context` propagation
- Keep queue/worker/channel concerns decoupled
- Avoid hidden goroutine leaks
- Graceful shutdown is mandatory for new runtime loops
- Add metrics for every critical execution path

---

## License

MIT
