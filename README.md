# gt-telegram

Telegram bridge for [Gas Town](https://github.com/steveyegge/gastown) overseer
communication. Chat with the Mayor agent and receive workspace notifications
from any Telegram client.

[![License](https://img.shields.io/github/license/leonletto/gt-telegram)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/leonletto/gt-telegram)](https://goreportcard.com/report/github.com/leonletto/gt-telegram)
[![CI](https://github.com/leonletto/gt-telegram/actions/workflows/ci.yml/badge.svg)](https://github.com/leonletto/gt-telegram/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/leonletto/gt-telegram)](https://github.com/leonletto/gt-telegram/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/leonletto/gt-telegram)](go.mod)

[![Demo](https://img.youtube.com/vi/1WmGka_8fE8/maxresdefault.jpg)](https://youtu.be/1WmGka_8fE8)

## Install

```bash
# One-line installer (recommended — downloads signed binary with checksum verification)
curl -fsSL https://raw.githubusercontent.com/leonletto/gt-telegram/main/scripts/install.sh | sh
```

Or with Homebrew:

```bash
brew install leonletto/tap/gt-telegram
```

Or with Go:

```bash
go install github.com/leonletto/gt-telegram@latest
```

Or build from source:

```bash
git clone https://github.com/leonletto/gt-telegram
cd gt-telegram
make install
```

## Quick Start

```bash
# 1. Create a bot via @BotFather on Telegram, copy the token
# 2. Send /start to your bot, then get your chat ID:
#    curl https://api.telegram.org/bot<TOKEN>/getUpdates
#    → result[0].message.chat.id

# 3. Configure (from your Gas Town root, or set GT_TOWN)
export GT_TOWN=~/gt
gt-telegram configure \
    --token "123456789:AAH..." \
    --chat-id <CHAT_ID> \
    --allow-from <USER_ID>

# 4. Run the bridge
gt-telegram run
```

## Commands

| Command | Description |
|---------|-------------|
| `gt-telegram configure` | Set token, chat ID, allow list, notifications |
| `gt-telegram status` | Show bridge configuration (token masked) |
| `gt-telegram status --json` | Machine-readable status |
| `gt-telegram run` | Run bridge in foreground (Ctrl-C to stop) |

## How It Works

```
You (Telegram)                    Gas Town
    │                                │
    │  "Hi mayor, status?"           │
    ├───────────────────────────────→ │
    │   Bot API long-poll             │
    │   → AccessGate (allow_from)     │
    │   → RateLimiter (30/min)        │
    │   → gt mail send mayor/         │
    │   → gt nudge hq-mayor           │
    │                                 │
    │                          Mayor processes
    │                          mail, replies to
    │                          overseer inbox
    │                                 │
    │  "@mayor/: All systems green"   │
    │ ←──────────────────────────────┤
    │   ReplyForwarder polls          │
    │   overseer inbox every 3s       │
    │   → bot.SendMessage()           │
```

### Inbound (Telegram → Mayor)

1. Bot long-polls Telegram with 30s timeout
2. Access gate rejects bots, checks `allow_from` list (fail-closed)
3. Rate limiter enforces per-user sliding window (default 30 msgs/min)
4. Message sent as mail: `from: overseer`, `to: mayor/`, `subject: Telegram`
5. Nudge queued to `hq-mayor` so Mayor picks it up on its next turn

### Outbound (Mayor → Telegram)

1. Reply forwarder polls overseer inbox every 3s via `bd list`
2. Forwards Mayor's replies to Telegram with thread mapping
3. Seeds forwarded set on startup to prevent duplicates on restart

### Event Notifications

Tails `.feed.jsonl` and forwards matching events:

| Category | Events |
|----------|--------|
| `stuck_agents` | `mass_death`, `session_death` |
| `escalations` | `escalation_sent` |
| `merge_failures` | `merge_failed` |

## Configuration

Config lives at `<GT_TOWN>/mayor/telegram.json` with `0600` permissions.

```json
{
  "token": "123456789:AAH...",
  "chat_id": 123456789,
  "allow_from": [123456789],
  "target": "mayor/",
  "enabled": true,
  "notify": ["escalations"],
  "rate_limit": 30
}
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GT_TOWN` | Gas Town root directory (default: cwd) |
| `GT_TOWN_ROOT` | Alias for GT_TOWN |

## Configuration Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `token` | string | required | BotFather bot token |
| `chat_id` | int64 | required | Telegram chat for outbound messages |
| `allow_from` | []int64 | `[]` | Allowed user IDs (fail-closed) |
| `target` | string | `"mayor/"` | Mail recipient for inbound messages |
| `enabled` | bool | `true` | Enable/disable the bridge |
| `notify` | []string | `["escalations"]` | Notification categories |
| `rate_limit` | int | `30` | Max inbound messages per user per minute |

## Documentation

- [Setup Guide](docs/setup.md) — step-by-step walkthrough from bot creation to running
- [Architecture](docs/architecture.md) — component design, package structure, security model
- [Troubleshooting](docs/troubleshooting.md) — common issues and solutions

## Requirements

- A running [Gas Town](https://github.com/steveyegge/gastown) instance
- `gt` and `bd` commands on PATH
- A Telegram bot token (from [@BotFather](https://t.me/BotFather))

## License

[MIT](LICENSE)
