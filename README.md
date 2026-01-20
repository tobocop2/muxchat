# muxchat (name still pending)

[![CI](https://github.com/tobocop2/muxchat/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/tobocop2/muxchat/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/tobocop2/muxchat/graph/badge.svg)](https://codecov.io/gh/tobocop2/muxchat)

Run one command. Toggle your bridges. Wait. Your chats appear.

![muxchat demo](assets/demo.gif)

```bash
muxchat
```

That's it. The TUI walks you through setup, starts services, and you're done. Toggle WhatsApp, Signal, Discord — whatever you want — and message the bot to link your account.

> **Note:** The UI briefly blocks during startup and when toggling bridges while Docker operations complete. This is [expected behavior](https://github.com/tobocop2/muxchat/issues/3) for now.

## What It Does

muxchat is a single binary that sets up a self-hosted Matrix server with messaging bridges. All your chats from different platforms in one place. No config files to edit, no secrets to manage — just run it.

- **Synapse** (Matrix homeserver)
- **Element Web** (bundled chat interface — disable via TUI or `muxchat init --no-element`)
- **Messaging bridges** (WhatsApp, Signal, Discord, Telegram, etc.)

Like [Bitlbee](https://www.bitlbee.org/), you interact with bridge bots to link accounts (e.g., message `@whatsappbot` and follow the prompts). Unlike Bitlbee, messages sync in real-time, you don't miss messages when offline, and modern features like reactions, threads, and encryption work.

### Linking Your Accounts

Message a bridge bot in Element to connect your accounts. Each bot walks you through authentication.

<img src="assets/element.gif" alt="Bridge bots in Element" width="1000">

### Sane Defaults

muxchat configures Synapse and bridges so things just work:

- **Auto-accept invites** — New chat rooms from bridges appear automatically, no manual accept needed
- **High rate limits** — Sync thousands of messages without getting throttled
- **Personal filtering spaces** — WhatsApp chats grouped together, Discord together, etc.
- **Full history sync** — Get your old messages, not just new ones
- **Double puppeting** — Messages you send from your phone show up as "you" in Element
- **Bot auto-setup** — Bridge bots are ready to message after running `muxchat setup-bots`

### Why not just a docker-compose.yml?

A static docker-compose.yml can't:
- **Generate secrets** — Each install needs unique passwords, appservice tokens, and signing keys. muxchat generates these on first run.
- **Register bridges dynamically** — When you enable a bridge, Synapse needs its registration.yaml added and a restart.
- **Adapt to your setup** — Domain, ports, which bridges — these require regenerating config files that reference each other.

muxchat handles all of this. Everything is generated from your choices and can be regenerated anytime.

## Background

Matrix is difficult to set up. Synapse alone has hundreds of configuration options. Add bridges and you're dealing with: appservice registration files with cryptographic tokens, database configuration for each bridge, Docker networking, rate limit tuning, and double-puppeting setup. Each bridge has its own config format. Getting it all working together is a real project.

[Bitlbee](https://www.bitlbee.org/) with libpurple was a great solution for years — an orchestrator for chat plugins, all accessible via IRC. muxchat is similar in spirit: an orchestrator for [mautrix](https://github.com/mautrix) bridges, all accessible via Matrix. Bitlbee's limitations: bridges poll (delayed messages), you miss messages when disconnected, no encryption, no reactions/threads/edits, and formatting gets mangled. Matrix handles all of this.

[Beeper](https://beeper.com) also solves this problem with a polished app and cloud-hosted bridges. muxchat is for tinkerers who want full control — no app installs, no cloud dependencies, runs on your hardware. There's some manual setup (messaging bridge bots), but it's simple for QR code bridges like WhatsApp, Discord, and Google Messages. You can also point Beeper or any Matrix client at the Synapse server muxchat sets up (untested).

## Supported Bridges

| Bridge | Login | Notes |
|--------|-------|-------|
| **WhatsApp** | QR code | Via linked device |
| **Signal** | QR code | No history sync (Signal limitation) |
| **Discord** | QR code or token | May violate ToS |
| **Telegram** | Phone + code | Requires API credentials from my.telegram.org |
| **Google Messages** | QR code | SMS/RCS |
| **Google Chat** | Cookies | Workspace accounts |
| **Google Voice** | Cookies | SMS/calls |
| **Slack** | Token + cookie | May violate ToS |
| **Bluesky** | Username/password | |
| **Meta** | Cookies | Facebook/Instagram |
| **Twitter** | Cookies | |
| **LinkedIn** | Cookies | |
| **IRC** | SASL | |

**Not supported:** iMessage (requires macOS or jailbroken iOS)

See [USAGE.md](USAGE.md) for login details and helper scripts for cookie extraction.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/tobocop2/muxchat/main/scripts/install.sh | sh
```

Downloads the correct binary for your platform to the current directory. Move it to your PATH.

Or build from source:
```bash
git clone https://github.com/tobocop2/muxchat.git
cd muxchat && go build -o muxchat .
```

See all releases: https://github.com/tobocop2/muxchat/releases

## Requirements

- Docker (with Docker Compose)
- 2GB RAM (4GB recommended)
- 10GB disk space

## Usage

Run `muxchat` for the TUI, or `muxchat --help` for CLI commands.

See [USAGE.md](USAGE.md) for detailed documentation on bridges, connectivity modes, troubleshooting, and more.

## Credits

muxchat uses the [mautrix bridges](https://github.com/mautrix) by [Tulir Asokan](https://github.com/tulir).

## License

MIT
