# Usage

## Quick Start

The TUI handles everything interactively — just run `muxbee`. Or use CLI commands:

```bash
muxbee init                      # Generate configs
muxbee bridge enable whatsapp    # Add a bridge
muxbee up                        # Start containers
muxbee open                      # Open Element
```

Sign into Element with your admin credentials. Start a chat with any bridge bot (e.g., `@whatsappbot:localhost`) — the bot will tell you how to authenticate.

<img src="assets/element.gif" alt="Element with bridge bots" width="1000">

**Finding your admin credentials:**
- Shown during `muxbee init` and `muxbee up`
- **Missed it?** Run `muxbee config show --show-secrets` and look for the `admin` section with `username` and `password`

> **Note:** If you change your password in Element, the stored password becomes stale. muxbee doesn't sync password changes from Synapse.

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
muxbee init --server-name 192.168.1.50
# or with Tailscale:
muxbee init --server-name your-machine.tailnet-name.ts.net
```

**Public HTTPS**: Expose to the internet with automatic SSL via [Caddy](https://caddyserver.com/).
```bash
muxbee init --https --domain chat.example.com --email you@example.com
```

For public HTTPS, you need:
- A domain pointing to your server (A record)
- Ports 80 and 443 open on your firewall/router

See [Caddy's automatic HTTPS docs](https://caddyserver.com/docs/automatic-https) for details.

> **Note:** Private network and Public HTTPS modes are not fully tested yet. Local mode is recommended for now. See [#2](https://github.com/tobocop2/muxbee/issues/2) for status.

## How It Works

muxbee runs a personal [Matrix](https://matrix.org) server (Synapse) with messaging bridges that connect to your accounts. You access everything through Element, a web-based Matrix client.

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

All services run in Docker containers. muxbee manages the Docker Compose configuration automatically.

## Data Storage

Configuration and data follow XDG conventions:

- Config: `~/.config/muxbee/`
- Data: `~/.local/share/muxbee/`

Override with `XDG_CONFIG_HOME` and `XDG_DATA_HOME`.

## Updating Services

Bridges use `:latest` Docker images. Pull updates when things break or you want new versions:

```bash
muxbee update    # CLI
# or press 'u' in TUI dashboard
```

This pulls latest images and restarts all services. Run `muxbee status` to see current versions.

## Troubleshooting

**Services won't start:**
```bash
muxbee health
muxbee logs
```

**Bridge not connecting:**
```bash
muxbee logs mautrix-whatsapp
```

**Start fresh:**
```bash
muxbee down && muxbee nuke --yes && muxbee init
```

**WhatsApp contacts not appearing:**
History sync can take a few minutes after login. Keep WhatsApp open on your phone. New messages appear immediately; historical conversations sync gradually.

**Bridge bot not responding to commands:**
If a bridge database was corrupted or reset, existing chat rooms with the bot become orphaned. Start a NEW direct message with the bot (e.g., `@whatsappbot:localhost`). The old room won't work; you need a fresh management room.

## CLI Reference

Run `muxbee --help` for all commands, or `muxbee <command> --help` for details.

### Core Commands

```
muxbee                   Launch the interactive TUI
muxbee init              Initialize configuration
muxbee up                Start all services
muxbee down              Stop all services
muxbee status            Show service status with versions
muxbee update            Pull latest images and restart
muxbee open              Open Element Web in browser
```

### Bridge Management

```
muxbee bridge list              List available bridges
muxbee bridge enable <name>     Enable a bridge
muxbee bridge disable <name>    Disable a bridge
muxbee bridge login <name>      Show login instructions
```

### Logs & Monitoring

```
muxbee logs                     View all logs
muxbee logs <service>           View logs for specific service
muxbee logs -f                  Follow logs (live stream)
muxbee logs -n 50               Show last 50 lines
muxbee health                   Check service health
```

### Backup & Recovery

```
muxbee backup                   Create backup archive
muxbee backup -o backup.tar.gz  Specify output file
muxbee restore backup.tar.gz    Restore from backup
muxbee nuke                     Remove all data (with confirmation)
muxbee nuke -y                  Remove all data (skip confirmation)
```

### Configuration

```
muxbee config show              Show current configuration
muxbee init --server-name x     Set Matrix server name
muxbee init --https             Enable HTTPS mode
muxbee init --domain x.com      Set domain for HTTPS
muxbee init --email you@x.com   Set email for Let's Encrypt
muxbee init --no-element        Don't run Element Web
muxbee init --force             Overwrite existing config
```

### Other

```
muxbee setup-bots               Create DM rooms with bridge bots
muxbee tui                      Launch TUI (same as no args)
muxbee --version                Show version
```
