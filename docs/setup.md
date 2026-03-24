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

## Step 2: Get Your Chat ID

1. Open Telegram and send `/start` to your new bot
2. Send any message (e.g., "hello")
3. In a terminal, run:

```bash
curl -s "https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates" | python3 -m json.tool
```

4. Find `result[0].message.chat.id` — that's your **chat ID**
5. Also note `result[0].message.from.id` — that's your **user ID** (usually the same as chat ID for private chats)

## Step 3: Install gt-telegram

```bash
# Option A: go install
go install github.com/leonletto/gt-telegram@latest

# Option B: build from source
git clone https://github.com/leonletto/gt-telegram
cd gt-telegram
go build -o gt-telegram .
cp gt-telegram ~/.local/bin/  # or wherever you keep binaries
```

## Step 4: Configure the Bridge

Set `GT_TOWN` to your Gas Town root directory:

```bash
export GT_TOWN=~/gt  # add to .bashrc/.zshrc for persistence
```

Run the configure command:

```bash
gt-telegram configure \
    --token "<YOUR_BOT_TOKEN>" \
    --chat-id <YOUR_CHAT_ID> \
    --allow-from <YOUR_USER_ID>
```

This creates `$GT_TOWN/mayor/telegram.json` with `0600` permissions.

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

## Step 5: Run the Bridge

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

## Step 6: Test

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
