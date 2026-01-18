# Usage

## Quick Start

The TUI handles everything interactively — just run `muxchat`. Or use CLI commands:

```bash
muxchat init                      # Generate configs
muxchat bridge enable whatsapp    # Add a bridge
muxchat up                        # Start containers
muxchat open                      # Open Element
```

Sign into Element with your admin credentials. Start a chat with any bridge bot (e.g., `@whatsappbot:localhost`) — the bot will tell you how to authenticate.

**Finding your admin credentials:**
- Shown once during `muxchat init`
- Shown on each `muxchat up`
- Retrieve anytime with `muxchat config show --show-secrets`

> **Note:** If you change your password in Element, the stored password becomes stale. muxchat doesn't sync password changes from Synapse.

## Available Bridges

### Easy Login (QR code or credentials)

| Bridge | Description | Login |
|--------|-------------|-------|
| **WhatsApp** | Via linked device | QR code |
| **Signal** | Signal messenger | QR code |
| **Discord** | Discord servers & DMs | QR code |
| **Google Messages** | SMS/RCS | QR code |
| **Bluesky** | Bluesky social | Username/password |
| **IRC** | IRC networks | SASL auth |

### Cookie Extraction Required

These bridges require extracting cookies from your browser. Use the included helper scripts or [mautrix-manager](https://github.com/mautrix/manager).

| Bridge | Description | Helper |
|--------|-------------|--------|
| **Google Chat** | Google Workspace | `node helpers/googlechat.js` |
| **Google Voice** | SMS/calls via GV | `node helpers/gvoice.js` |
| **Slack** | Slack workspaces | `node helpers/slack.js` |
| **Meta** | Facebook/Instagram | Manual cURL |
| **Twitter** | Twitter/X DMs | Manual cookies |
| **LinkedIn** | LinkedIn messages | Manual cURL |

### Requires API Credentials

| Bridge | Description | Setup |
|--------|-------------|-------|
| **Telegram** | Telegram messenger | Register app at [my.telegram.org](https://my.telegram.org) |

### Not Supported

- **iMessage** — Requires macOS or jailbroken iOS, cannot run on Linux

## Helper Scripts

For bridges requiring cookie extraction:

```bash
cd helpers
npm install

node googlechat.js    # Google Chat (Workspace)
node gvoice.js        # Google Voice
node slack.js         # Slack
```

Or use [mautrix-manager](https://github.com/mautrix/manager) for automated cookie extraction.

## Connectivity Modes

**Local** (default): Access from your machine at `http://localhost:8080`. Good for testing.

**Private network** (VPN/Tailscale/Zerotier): Access from anywhere via your private network.
```bash
muxchat init --server-name 192.168.1.50
# or with Tailscale:
muxchat init --server-name your-machine.tailnet-name.ts.net
```

**Public HTTPS**: Expose to the internet with automatic SSL via [Caddy](https://caddyserver.com/).
```bash
muxchat init --https --domain chat.example.com --email you@example.com
```

For public HTTPS, you need:
- A domain pointing to your server (A record)
- Ports 80 and 443 open on your firewall/router

See [Caddy's automatic HTTPS docs](https://caddyserver.com/docs/automatic-https) for details.

> **Note:** Private network and Public HTTPS modes are not fully tested yet. Local mode is recommended for now. See [#2](https://github.com/tobocop2/muxchat/issues/2) for status.

## How It Works

muxchat runs a personal [Matrix](https://matrix.org) server (Synapse) with messaging bridges that connect to your accounts. You access everything through Element, a web-based Matrix client.

```
┌─────────────────────────────────────────────────┐
│                   Element Web                    │
│              (your unified inbox)                │
└─────────────────────────────────────────────────┘
                        │
┌─────────────────────────────────────────────────┐
│                    Synapse                       │
│               (Matrix server)                    │
└─────────────────────────────────────────────────┘
         │              │              │
    ┌────┴────┐    ┌────┴────┐    ┌────┴────┐
    │WhatsApp │    │ Discord │    │ Signal  │
    │ Bridge  │    │ Bridge  │    │ Bridge  │
    └────┬────┘    └────┬────┘    └────┬────┘
         │              │              │
    WhatsApp        Discord        Signal
```

All services run in Docker containers. muxchat manages the Docker Compose configuration automatically.

## Data Storage

Configuration and data follow XDG conventions:

- Config: `~/.config/muxchat/`
- Data: `~/.local/share/muxchat/`

Override with `XDG_CONFIG_HOME` and `XDG_DATA_HOME`.

## Troubleshooting

**Services won't start:**
```bash
muxchat health
muxchat logs
```

**Bridge not connecting:**
```bash
muxchat logs mautrix-whatsapp
```

**Start fresh:**
```bash
muxchat down && muxchat nuke --yes && muxchat init
```

**WhatsApp contacts not appearing:**
History sync can take a few minutes after login. Keep WhatsApp open on your phone. New messages appear immediately; historical conversations sync gradually.

**Bridge bot not responding to commands:**
If a bridge database was corrupted or reset, existing chat rooms with the bot become orphaned. Start a NEW direct message with the bot (e.g., `@whatsappbot:localhost`). The old room won't work; you need a fresh management room.

## CLI Reference

Run `muxchat --help` for all commands, or `muxchat <command> --help` for details.

### Core Commands

```
muxchat                   Launch the interactive TUI
muxchat init              Initialize configuration
muxchat up                Start all services
muxchat down              Stop all services
muxchat status            Show service status
muxchat open              Open Element Web in browser
```

### Bridge Management

```
muxchat bridge list              List available bridges
muxchat bridge enable <name>     Enable a bridge
muxchat bridge disable <name>    Disable a bridge
muxchat bridge login <name>      Show login instructions
```

### Logs & Monitoring

```
muxchat logs                     View all logs
muxchat logs <service>           View logs for specific service
muxchat logs -f                  Follow logs (live stream)
muxchat logs -n 50               Show last 50 lines
muxchat health                   Check service health
```

### Backup & Recovery

```
muxchat backup                   Create backup archive
muxchat backup -o backup.tar.gz  Specify output file
muxchat restore backup.tar.gz    Restore from backup
muxchat nuke                     Remove all data (with confirmation)
muxchat nuke -y                  Remove all data (skip confirmation)
```

### Configuration

```
muxchat config show              Show current configuration
muxchat init --server-name x     Set Matrix server name
muxchat init --https             Enable HTTPS mode
muxchat init --domain x.com      Set domain for HTTPS
muxchat init --email you@x.com   Set email for Let's Encrypt
muxchat init --no-element        Don't run Element Web
muxchat init --force             Overwrite existing config
```

### Other

```
muxchat setup-bots               Create DM rooms with bridge bots
muxchat tui                      Launch TUI (same as no args)
muxchat --version                Show version
```
