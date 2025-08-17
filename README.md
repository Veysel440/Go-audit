Audit Trail Service

A minimal, production-ready audit log ingestion and query API. Built with Go, MongoDB, and Docker. Supports API key auth, idempotency keys, TTL retention, rate limiting, and structured logging.

Contents

Overview

Features

Architecture

Folder Structure

Data Model

API

Configuration

Run Locally

Run with Docker Compose

Examples

Idempotency

Rate Limiting

Indexes and Retention

Logging

Testing

Deployment Notes

Make it yours

Roadmap

License

Overview

This service collects immutable audit events from other systems and exposes fast query endpoints. Typical use cases:

Security and compliance trails.

User activity history.

Change logs for business entities.

The write path is optimized for durability and idempotency. Reads support common filters and time ranges.

Features

HTTP JSON API (Go 1.25+, net/http).

MongoDB storage with TTL-based retention.

API key authentication via X-Api-Key.

Idempotency keys to de-duplicate writes.

Rate limiting per client IP.

Structured logs with log/slog.

Graceful shutdown and timeouts.

Dockerfile + docker-compose for quick boot.

Architecture

cmd/api: process entrypoint (HTTP server, graceful shutdown).

internal/httpx: routing, middlewares, auth, JSON helpers.

internal/service: validation and business logic.

internal/repo/mongo: MongoDB repository and indexes.

internal/core: domain types (Audit).

pkg/rate: simple in-memory token bucket.
