# Troubleshooting

## Common Issues

### "telegram bridge is not configured"

Run `gt-telegram configure` first. See [Setup Guide](setup.md).

### "file has unsafe permissions"

The config file must have `0600` permissions:

```bash
chmod 600 $GT_TOWN/mayor/telegram.json
```

### "not in a Gas Town workspace"

Set the `GT_TOWN` environment variable:

```bash
export GT_TOWN=~/gt
```

### Bridge starts but no messages come through

**Check 1: Only one bridge instance**

Telegram only allows one `getUpdates` poller per bot token. If multiple
instances are running, they'll get 409 Conflict errors and messages will
be consumed unpredictably.

```bash
# Check for other instances
pgrep -fl "gt-telegram\|gt telegram"

# Kill extras
pkill -f "gt-telegram run"
```

**Check 2: Bot token works**

```bash
TOKEN=$(python3 -c "import json; print(json.load(open('$GT_TOWN/mayor/telegram.json'))['token'])")
curl -s "https://api.telegram.org/bot${TOKEN}/getMe" | python3 -m json.tool
```

**Check 3: `gt` and `bd` are on PATH**

The bridge shells out to `gt mail send`, `gt nudge`, and `bd list`. They
must be on PATH:

```bash
which gt bd
```

**Check 4: Dolt is running**

The bridge needs Dolt for mail delivery and bead queries:

```bash
gt daemon status
```

If Dolt is down, you'll see "circuit breaker is open" errors. Start the
daemon: `gt daemon start`.

### Duplicate messages on restart

On startup, the bridge seeds its forwarded set by querying all existing
messages. If Dolt isn't ready yet (cold start), the seed fails and old
messages may be re-forwarded once.

This is a one-time event on restart — subsequent messages won't duplicate.
To avoid it, wait for the daemon to fully start before running the bridge.

### "nudge failed (non-fatal)"

The Mayor session isn't running. The mail was still delivered — the Mayor
will read it when it starts. This is expected if you're testing without
agents running.

### Messages received but Mayor doesn't reply

The Mayor needs to be running in a tmux session (`hq-mayor`). Check:

```bash
gt daemon status
tmux -L gt-* list-sessions
```

The Mayor picks up Telegram messages via its existing `UserPromptSubmit`
hook (`gt mail check --inject`). No special configuration is needed on
the Mayor's side.

### macOS: Processes stuck in "UE" state

If `gt` processes get stuck in uninterruptible wait (state `UE`), they
can't be killed — even with `kill -9`. This happens when processes are
blocked on a dead Dolt socket.

**Fix:**

```bash
# Remove the binary to break the inode reference
rm ~/.local/bin/gt
# Copy a fresh build
cp /path/to/new/gt ~/.local/bin/gt
```

The UE processes are harmless (0 CPU) but won't go away until reboot.

### 409 Conflict errors

```
Conflict: terminated by other getUpdates request;
make sure that only one bot instance is running
```

Another process is polling the same bot token. Find and kill it:

```bash
pgrep -fl "gt-telegram\|gt telegram"
```

If the other instance is on a different machine, stop it there first.
Each bot token can only have one active `getUpdates` poller.

## Logs

The bridge logs to stderr. When running in the foreground, logs appear
in the terminal. For background operation, redirect to a file:

```bash
GT_TOWN=~/gt gt-telegram run 2>&1 | tee ~/gt-telegram.log
```

Key log messages:

| Message | Meaning |
|---------|---------|
| `reply forwarder: seeded N existing messages` | Startup dedup working |
| `telegram bridge: relay error` | Inbound message failed to relay (Dolt down?) |
| `reply forwarder: sent X to Telegram` | Outbound reply delivered |
| `telegram bridge: closed N delivered inbound bead(s)` | Cleanup working |
| `telegram bridge: run error (retrying in 5s)` | Connection lost, will retry |
| `telegram inbound: nudge failed (non-fatal)` | Mayor session not running |
