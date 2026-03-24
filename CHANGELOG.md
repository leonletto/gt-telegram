# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.1] - 2026-03-24

### Changed

- Post-pairing output now shows both foreground and tmux daemon options

## [0.3.0] - 2026-03-24

### Added

- **Interactive pairing flow** — `gt-telegram configure --token <TOKEN>` now
  auto-pairs by connecting to Telegram and capturing your user ID and chat ID
  from the first message you send. No manual ID lookup needed.
- **`pair` subcommand** — pair your Telegram account with an already-configured
  bridge (`gt-telegram pair`)
- **Pairing security model** — short window (60s default, 300s max), explicit
  `[y/n]` consent, probe message never relayed to Gas Town, fail-closed on
  timeout or decline

### Changed

- `configure` now auto-pairs when `--allow-from` is not provided. Use
  `--skip-pair` to save config without pairing, or `--allow-from <ID>` to
  skip pairing entirely.
- Split `Config.Validate()` into `ValidateToken()` (pre-pairing) and
  `Validate()` (full validation for running the bridge)

## [0.2.0] - 2026-03-23

### Added

- Release infrastructure with signed macOS binaries
- GoReleaser config for linux/darwin x amd64/arm64
- GitHub Actions CI and Release workflows
- Apple codesigning and notarization for macOS binaries
- Homebrew cask via `leonletto/homebrew-tap`
- Bash installer with checksum verification (`scripts/install.sh`)
- Makefile with build, install, test, fmt, vet, ci targets
- Version injection via ldflags with `version` subcommand
- README badges (License, Go Report Card, CI, Release, Go Version)

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

[0.3.1]: https://github.com/leonletto/gt-telegram/releases/tag/v0.3.1
[0.3.0]: https://github.com/leonletto/gt-telegram/releases/tag/v0.3.0
[0.2.0]: https://github.com/leonletto/gt-telegram/releases/tag/v0.2.0
[0.1.0]: https://github.com/leonletto/gt-telegram/releases/tag/v0.1.0
