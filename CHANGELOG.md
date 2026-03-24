# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-03-23

### Added

- Standalone Telegram bridge extracted from Gas Town core
- Inbound relay: Telegram messages to `gt mail send` + `gt nudge`
- Outbound replies: Poll overseer inbox, forward to Telegram with dedup
- Outbound notifications: Tail `.feed.jsonl`, filter by category
- CLI with `configure`, `status`, and `run` subcommands (stdlib, no cobra)
- Bidirectional message-thread mapping with FIFO eviction
- AccessGate and RateLimiter for Telegram input
- Retry loop with panic recovery
- Inbound bead cleanup goroutine
- Setup guide, architecture doc, and troubleshooting guide

[0.1.0]: https://github.com/leonletto/gt-telegram/releases/tag/v0.1.0
