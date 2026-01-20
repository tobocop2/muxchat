# muxbee Architecture

This document describes muxbee's architecture, design decisions, and the sane defaults it provides out of the box.

## Overview

muxbee is a thin orchestration layer around proven open-source components:

```
┌─────────────────────────────────────────────────────────────────┐
│                         muxbee CLI                              │
│              (Go binary with embedded templates)                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                    Docker Compose
                              │
┌─────────────────────────────────────────────────────────────────┐
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │ Element  │  │  Synapse │  │ Postgres │  │  Caddy   │        │
│  │  (Web)   │  │ (Matrix) │  │   (DB)   │  │ (HTTPS)  │        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │
│                      │                                          │
│       ┌──────────────┼──────────────┐                          │
│       │              │              │                          │
│  ┌────┴────┐   ┌────┴────┐   ┌────┴────┐                      │
│  │WhatsApp │   │Telegram │   │ Signal  │   ...more bridges    │
│  │ Bridge  │   │ Bridge  │   │ Bridge  │                      │
│  └─────────┘   └─────────┘   └─────────┘                      │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### Synapse (Matrix Homeserver)
- **What**: The Matrix protocol server that stores messages and manages rooms
- **Image**: `matrixdotorg/synapse:latest`
- **Why Matrix**: Open federation protocol with excellent bridge ecosystem

### Element Web
- **What**: Polished web-based Matrix client
- **Image**: `vectorim/element-web:latest`
- **Pre-configured**: Points directly at local Synapse, no manual homeserver URL entry needed

### PostgreSQL
- **What**: Database backend for Synapse
- **Image**: `postgres:17`
- **Why not SQLite**: Better performance and reliability for production use

### Caddy (optional)
- **What**: Reverse proxy with automatic HTTPS
- **Image**: `caddy:latest`
- **When used**: Only with `--https` mode for public deployments

### Mautrix Bridges
- **What**: Protocol bridges connecting Synapse to external messaging platforms
- **Images**: `dock.mau.dev/mautrix/*:latest`
- **Supported**: WhatsApp, Telegram, Signal, Google Messages, Discord, Slack, and more

## Design Decisions

### Single Binary Distribution
muxbee is distributed as a single Go binary with embedded assets:
- Docker Compose configuration
- Config templates (Synapse, Element, bridges)
- Bridge registry (ports, login instructions, categories)

This means users download one file. No cloning repos, no managing config files manually.

### Docker Compose via Shell (not SDK)
We shell out to `docker compose` rather than using the Docker SDK because:
- Simpler implementation
- Users can inspect and modify the generated compose file
- Fewer dependencies in the binary
- Easier debugging ("just run docker compose yourself")

### XDG Directory Layout
Configuration and data follow XDG Base Directory conventions:
```
~/.config/muxbee/          # Configuration (settings, generated configs)
├── settings.yaml           # User settings and credentials
├── docker-compose.yml      # Generated compose file
├── synapse/
│   └── homeserver.yaml     # Synapse configuration
├── element/
│   └── config.json         # Element configuration
└── bridges/
    └── <bridge>/
        └── registration.yaml

~/.local/share/muxbee/     # Persistent data (databases, media)
├── synapse/                # Synapse data, media, signing keys
├── postgres/               # PostgreSQL data
└── bridges/
    └── <bridge>/           # Bridge databases and state
```

### Bridge Categories
Bridges are organized by authentication complexity:

**QR/Phone Bridges** (simplest setup):
- WhatsApp, Signal, Google Messages, Discord, Bluesky
- Scan a QR code or enter a phone number

**Cookie Extraction Bridges** (requires browser DevTools):
- Google Chat, Google Voice, Meta (Facebook/Instagram), Twitter, LinkedIn, Slack
- Extract authentication cookies from a logged-in browser session
- Helper scripts in `scripts/cookie-helpers/` directory simplify extraction

**API Credential Bridges**:
- Telegram: Requires API ID/hash from my.telegram.org (free)
- IRC: Standard IRC server connection details

Each bridge's "note" field explains any special requirements. Bridges without notes work with standard QR/phone authentication.

## Sane Defaults

muxbee configures services with production-ready defaults so things "just work."

### Auto-Accept Room Invites
**Problem**: When a bridge syncs, it creates rooms and invites the user. Without auto-accept, users must manually accept hundreds of invites.

**Solution**: Synapse's `auto_accept_invites` is enabled:
```yaml
auto_accept_invites:
  enabled: true
  only_for_direct_messages: false    # Accept group chats too
  only_from_local_users: false       # Accept from bridge bots
```

### High Rate Limits
**Problem**: Bridge syncs create many rooms simultaneously, triggering Synapse's rate limits (429 errors).

**Solution**: Rate limits are relaxed for local operations:
```yaml
rc_joins:
  local:
    per_second: 100
    burst_count: 100
rc_joins_per_room:
  per_second: 100
  burst_count: 100
rc_invites:
  per_room:
    per_second: 100
    burst_count: 100
  per_user:
    per_second: 100
    burst_count: 100
```

### Bridge Permissions
Each bridge is configured with sensible permission levels:
```yaml
permissions:
  "*": relay                              # Anyone can relay messages
  "localhost": user                       # Local users get full access
  "@admin:localhost": admin               # Admin user gets admin rights
```

### Double Puppeting
**Problem**: Without double puppeting, messages you send from your phone appear in Element as coming from a "ghost user" (e.g., `@whatsapp_123456:localhost`) rather than your real Matrix account.

**Solution**: muxbee configures appservice-based double puppeting. A dedicated `doublepuppet` appservice is registered with Synapse, and bridges are configured with its token:
```yaml
double_puppet:
  secrets:
    localhost: "as_token:YOUR_DOUBLEPUPPET_TOKEN"
```

This allows bridges to send messages on behalf of your real Matrix user, so your messages appear consistently whether sent from Element or your phone.

### Full Chat Sync
**Problem**: By default, many bridges only sync recent chats (e.g., last 10). Users expect all their conversations to appear.

**Solution**: Bridges are configured to sync all chats:
```yaml
# Go bridges (WhatsApp, Signal, etc.)
network:
  initial_chat_sync_count: -1    # -1 = unlimited

# Python bridges (Telegram, Google Chat)
bridge:
  initial_chat_sync: -1
  sync_create_limit: -1
```

### Message Delivery
Bridges are configured for reliable message delivery:
```yaml
matrix:
  delivery_receipts: true        # Confirm message delivery
  message_error_notices: true    # Notify on send failures
```

### Element Configuration
- **Dark theme** by default
- **No guest access** (this is a personal server)
- **No 3PID login** (no email/phone login, just username)
- **Pre-configured homeserver URL** (no manual entry needed)

### Latest Images
All container images use recent stable versions:
- Avoids missing features (like `auto_accept_invites` added in Synapse 1.109)
- Avoids security vulnerabilities in old versions
- Bridges use `:latest` to stay compatible with protocol changes

### Generated Secrets
All secrets are auto-generated with cryptographic randomness:
- PostgreSQL password (32 chars)
- Admin user password (16 chars)
- Synapse registration secret (64 chars)
- Bridge appservice tokens (64 chars each)

Tokens are persisted in `settings.yaml` and reused across restarts.

## Network Architecture

All services communicate on a private Docker network (`muxbee`):

```
External Access:
  - Port 8080: Element Web
  - Port 8008: Matrix Client API (for mobile apps)
  - Ports 80/443: Caddy (HTTPS mode only)

Internal (Docker network only):
  - synapse:8008 - Synapse API
  - postgres:5432 - PostgreSQL
  - mautrix-*:29XXX - Bridge appservice ports
```

Bridges connect to Synapse via the internal Docker network. They're not exposed externally.

## File Generation Flow

When `muxbee up` runs:

1. **Load settings** from `~/.config/muxbee/settings.yaml`
2. **Generate configs** from embedded templates:
   - Synapse homeserver.yaml (with DB creds, bridge registrations)
   - Element config.json (with homeserver URL)
   - Bridge configs (with tokens, permissions)
   - Bridge registrations (for Synapse to recognize them)
3. **Write docker-compose.yml** to config directory
4. **Run `docker compose up`** with appropriate profiles

Templates use Go's `text/template` with data from settings.

## Bridge Registration

Each bridge needs two files:
1. **config.yaml** (in data dir): Bridge's own configuration
2. **registration.yaml** (in config dir): Tells Synapse about the bridge

The registration contains:
- Appservice ID and URL
- AS token (bridge→Synapse auth)
- HS token (Synapse→bridge auth)
- User/alias namespace regex (which Matrix IDs the bridge controls)

Both files must have matching tokens. muxbee generates tokens once and persists them in settings.yaml.

## Security Considerations

### Local Mode
- Services bind to localhost only (except Element on 8080)
- No TLS (acceptable for local development)
- Secrets stored in plaintext in settings.yaml (user's home directory)

### HTTPS Mode
- Caddy handles TLS termination with automatic Let's Encrypt
- All external traffic encrypted
- Internal Docker network remains unencrypted (acceptable, isolated network)

### Bridge Security
- Bridges run as unprivileged containers
- Each bridge has its own isolated data directory
- Bridge tokens are unique per-bridge

## Extending muxbee

### Adding a New Bridge

See [DEVELOPERS.md](DEVELOPERS.md) for a detailed walkthrough including template variables, docker-compose setup, and testing.

Quick version:
1. Add entry to `internal/bridges/bridges.yaml`
2. Create template in `internal/generator/templates/bridges/<name>.yaml.tmpl`
3. Add service to `internal/docker/docker-compose.yml`
4. Rebuild: `go build -o muxbee .`

### Custom Synapse Configuration
Edit `internal/generator/templates/synapse/homeserver.yaml.tmpl` and rebuild. Or for one-off changes, edit the generated file in `~/.config/muxbee/synapse/` (will be overwritten on next `muxbee up`).

### Using External Database
Modify `internal/generator/templates/synapse/homeserver.yaml.tmpl` to point to your PostgreSQL instance and remove the postgres service from `internal/docker/docker-compose.yml`.
