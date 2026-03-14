---
type: scout
date: 2026-03-15
subject: GoClaw Server Codebase Analysis
---

# Scout Report — GoClaw Server Analysis

## Summary

GoClaw is a multi-agent AI gateway in Go 1.25 with PostgreSQL, WebSocket RPC, and HTTP REST API. Serves as the server that this CLI will manage.

## Key Numbers

- **44** HTTP handler files (REST API endpoints)
- **28** WebSocket RPC method files
- **19** database migrations (PostgreSQL + pgvector)
- **30+** dashboard page routes (React SPA)
- **13+** LLM providers supported
- **7** messaging channels (Telegram, Discord, Slack, Zalo OA, Zalo Personal, Feishu, WhatsApp)
- **16** database stores (agents, sessions, skills, cron, MCP, teams, etc.)

## Auth Model

| Path | Mechanism | Role |
|------|-----------|------|
| Token | `Authorization: Bearer {GOCLAW_GATEWAY_TOKEN}` | admin |
| Device pairing (reconnect) | `sender_id` from previous pairing | operator |
| Device pairing (new) | No token → pairing code → admin approves | pending→operator |

## Core Entities

Agents, Sessions, Skills, Cron Jobs, MCP Servers, Custom Tools, Built-in Tools, LLM Providers, Teams (members, tasks, workspace), Channels (instances, contacts, writers), Memory (documents + embeddings), Knowledge Graph, Traces (spans), Paired Devices, Config Secrets, Audit Logs, CLI Credentials

## API Transport

- **REST:** `/v1/*` endpoints, standard CRUD, Bearer token auth
- **WebSocket:** v3 protocol, `req`/`res`/`event` frames, `connect` handshake first, streaming via events

## Dashboard Features Covered

Every feature in the web dashboard has corresponding REST or WS API endpoints. Full parity achievable via CLI.
