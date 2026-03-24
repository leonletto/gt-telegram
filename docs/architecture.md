# Architecture

## Overview

gt-telegram is a standalone bridge between Telegram and Gas Town's mail/nudge
system. It has **zero Go-level dependencies** on Gas Town internals — it shells
out to `gt` and `bd` CLI commands for all interactions with the Gas Town
workspace.

```
┌─────────────┐         ┌──────────────────────────────┐
│  Telegram   │         │        gt-telegram           │
│  Bot API    │◄───────►│                              │
│             │         │  Bot (long-poll + send)      │
└─────────────┘         │  InboundRelay                │
                        │  ReplyForwarder              │
                        │  OutboundNotifier            │
                        │  InboundCleanup              │
                        └──────┬───────────────────────┘
                               │ shells out to
                        ┌──────▼───────────────────────┐
                        │  gt mail send / gt nudge     │
                        │  bd list / bd close          │
                        │  (Gas Town CLI)              │
                        └──────────────────────────────┘
```

## Package Structure

```
internal/telegram/
  config.go      Config struct, validation, file I/O (0600 enforced)
  bot.go         Telegram Bot API: long-poll, send, AccessGate, RateLimiter
  bridge.go      Lifecycle orchestrator: retry loop, panic recovery, shutdown
  inbound.go     Telegram message → gt mail send + gt nudge
  reply.go       Overseer inbox poll → bot.SendMessage (with dedup seeding)
  outbound.go    .feed.jsonl tail → category filter → bot.SendMessage
  msgmap.go      Bidirectional Telegram msgID ↔ mail ThreadID (FIFO eviction)
  sender.go      Sender interface: CLISender (standalone), DirectSender (library)
```

## Components

### Bot (`bot.go`)

Wraps the `go-telegram-bot-api` library. Handles:

- **Long-polling** with configurable timeout (30s default)
- **HTTP client timeout** slightly longer than poll timeout to detect silently
  dropped TCP connections (common on cloud servers with NAT/firewall idle
  connection timeouts)
- **AccessGate**: rejects bots, then checks `allow_from` list (fail-closed —
  empty list blocks everyone)
- **RateLimiter**: per-user sliding window (default 30 msgs/min)

### Bridge (`bridge.go`)

Lifecycle orchestrator that manages all components:

- **Retry loop**: if `runOnce()` fails, waits 5s and retries
- **Panic recovery**: panics in any goroutine are caught and logged
- **Clean shutdown**: context cancellation, waits for all goroutines
- **Persistent state**: `MessageMap` and `CLIInboxReader` survive reconnects
  within the same process

Goroutines started per connection cycle:
1. `bot.Poll` — Telegram long-poll
2. `outbound.Run` — feed event notifications
3. `replyFwd.Run` — Mayor reply forwarding
4. `inboundCleanup` — close delivered inbound beads

### InboundRelay (`inbound.go`)

Converts Telegram messages to Gas Town mail:

1. Calls `gt mail send mayor/ -s Telegram -m <text>`
2. Calls `gt nudge hq-mayor --mode=queue` (non-fatal if fails)
3. Sets `BD_ACTOR=overseer` on subprocesses

### ReplyForwarder (`reply.go`)

Polls for Mayor replies to overseer and forwards to Telegram:

1. Runs `bd list --assignee overseer --label gt:message --include-infra --json`
2. Filters out `from:overseer` (our own outbound) and already-forwarded IDs
3. Forwards to Telegram first, then marks as forwarded + closes bead
4. **Seed on startup**: queries all existing messages and pre-populates the
   forwarded set to prevent duplicates on restart

### OutboundNotifier (`outbound.go`)

Tails `.feed.jsonl` for workspace events:

1. Seeks to end of file on startup (only new events)
2. Polls every 100ms for new lines
3. Filters by configured notification categories
4. Detects file rotation (inode change) and re-opens

### InboundCleanup (`bridge.go`)

Background goroutine that closes stale inbound beads:

- Every 60 seconds, queries open beads assigned to the target with
  `from:overseer` and `gt:message` labels
- Closes beads older than 30 seconds (gives Mayor time to read them first)
- Prevents inbound Telegram messages from polluting the issue queue

## Sender Interface

```go
type Sender interface {
    SendMail(ctx context.Context, to, subject, body string) error
    Nudge(ctx context.Context, session, message string) error
}
```

Two implementations:

- **CLISender** — shells out to `gt mail send` and `gt nudge`. Used by the
  standalone binary.
- **DirectSender** — calls injected Go functions directly. Available for
  library usage if someone wants to embed the bridge in another Go program.

## Security

| Concern | Mitigation |
|---------|-----------|
| Token storage | `0600` permissions enforced at load time. Token masked in logs/status |
| Inbound access | `allow_from` is fail-closed. Empty list blocks all users |
| Bot messages | Always rejected (`is_bot` check before allow-list) |
| Rate limiting | Per-user sliding window, configurable (default 30/min) |
| Outbound | Only sends to configured `chat_id` |

## Error Handling

- **Connection failures**: 5s backoff retry loop with panic recovery
- **Mail send failures**: logged, Telegram message channel continues
- **Duplicate prevention**: seed on startup + in-memory forwarded set +
  durable close via `bd close`
- **Feed file rotation**: detects inode changes, re-opens automatically
- **Dolt cold start**: seed failure is non-fatal; bridge starts without
  seeding and will forward any old unread messages (one-time, non-critical)
