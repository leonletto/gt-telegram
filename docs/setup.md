# Setup Guide

Complete walkthrough for setting up the Telegram bridge with Gas Town.

## Prerequisites

- A running [Gas Town](https://github.com/steveyegge/gastown) instance
- `gt` and `bd` commands on your PATH
- A Telegram account

## Step 1: Create a Telegram Bot

1. Open Telegram and search for [@BotFather](https://t.me/BotFather)
2. Send `/newbot`
3. Choose a display name (e.g., "My GT Bot")
4. Choose a username (must end in `bot`, e.g., `my_gt_bot`)
5. BotFather replies with your **bot token** — save this securely

The token looks like: `123456789:ABCDefGHIjklMNOpqrsTUVwxyz`

## Step 2: Install gt-telegram

```bash
# Option A: one-line installer (recommended — signed binary with checksum verification)
curl -fsSL https://raw.githubusercontent.com/leonletto/gt-telegram/main/scripts/install.sh | sh

# Option B: Homebrew
brew install leonletto/tap/gt-telegram

# Option C: go install
go install github.com/leonletto/gt-telegram@latest

# Option D: build from source
git clone https://github.com/leonletto/gt-telegram
cd gt-telegram
make install
```

## Step 3: Configure and Pair

Set `GT_TOWN` to your Gas Town root directory:

```bash
export GT_TOWN=~/gt  # add to .bashrc/.zshrc for persistence
```

Run configure with your bot token:

```bash
gt-telegram configure --token "<YOUR_BOT_TOKEN>"
```

This saves the token and starts an interactive pairing flow:

```
Telegram bridge configured (/home/user/gt/mayor/telegram.json).

Pairing — send any message to your bot from Telegram (timeout: 60s)...

Message from: Your Name (ID: 123456789)
  Allow this user? [y/n]: y

Paired! chat_id=123456789, allow_from=[123456789]
  Bridge is ready — run 'gt-telegram run' to start.
```

The pairing flow automatically captures your chat ID and user ID from the
first message you send to the bot. No manual ID lookup needed.

### Alternative: Skip Pairing

If you already know your IDs (e.g., from another bot), you can skip pairing:

```bash
gt-telegram configure \
    --token "<YOUR_BOT_TOKEN>" \
    --chat-id <YOUR_CHAT_ID> \
    --allow-from <YOUR_USER_ID>
```

Or save the token now and pair later:

```bash
gt-telegram configure --token "<YOUR_BOT_TOKEN>" --skip-pair
gt-telegram pair  # run this when you're ready
```

### Optional: Configure Notifications

By default, only `escalations` are forwarded. To add more:

```bash
gt-telegram configure --notify escalations,stuck_agents,merge_failures
```

### Verify Configuration

```bash
gt-telegram status
```

Should show your config with the token masked.

## Step 4: Run the Bridge

```bash
gt-telegram run
```

The bridge runs in the foreground. Press Ctrl-C to stop.

### Running in the Background

Using tmux:

```bash
tmux new-session -d -s telegram 'GT_TOWN=~/gt gt-telegram run'
```

Using nohup:

```bash
GT_TOWN=~/gt nohup gt-telegram run > ~/gt-telegram.log 2>&1 &
```

Using systemd (Linux):

```ini
# ~/.config/systemd/user/gt-telegram.service
[Unit]
Description=Gas Town Telegram Bridge
After=network.target

[Service]
Environment=GT_TOWN=%h/gt
ExecStart=%h/.local/bin/gt-telegram run
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
```

```bash
systemctl --user daemon-reload
systemctl --user enable --now gt-telegram
```

## Step 5: Test

1. Send a message to your bot on Telegram
2. The bridge relays it as mail to the Mayor
3. The Mayor picks it up on its next turn and replies
4. The reply appears on Telegram

If the Mayor isn't running, you'll see a "nudge failed" log (non-fatal) —
the mail is still delivered and will be read when the Mayor starts.

## Updating Configuration

You can update individual fields without re-entering everything:

```bash
gt-telegram configure --notify escalations,stuck_agents
gt-telegram configure --rate-limit 60
gt-telegram configure --token "<NEW_TOKEN>" --yes  # skip confirmation
```

## Uninstalling

1. Stop the bridge (Ctrl-C or kill the process)
2. Remove the config: `rm $GT_TOWN/mayor/telegram.json`
3. Remove the binary: `rm ~/.local/bin/gt-telegram`
4. Optionally, delete the bot via @BotFather: `/deletebot`
