# Developer Guide

Code architecture and extension guide for muxbee.

## Package Structure

```
muxbee/
├── cmd/                 # CLI commands (Cobra)
├── internal/
│   ├── bridges/         # Bridge registry (embedded YAML)
│   ├── config/          # Settings, XDG paths, load/save
│   ├── docker/          # Docker Compose wrapper
│   ├── generator/       # Template rendering
│   ├── matrix/          # Matrix client for bot setup
│   └── tui/             # Terminal UI (Bubble Tea)
└── main.go
```

Everything under `internal/` is private to the module. Only `cmd/` and `main.go` can import these packages.

## Package Dependencies

```
cmd/ ──────────────────────────────────────────────┐
  │                                                │
  ├── config     (settings, paths)                 │
  ├── docker     (compose wrapper)                 │
  ├── generator  (template rendering)              │
  ├── bridges    (bridge registry)                 │
  ├── matrix     (bot setup)                       │
  └── tui        (terminal UI)                     │
                                                   │
internal/tui ──────────────────────────────────────┤
  ├── config                                       │
  ├── docker                                       │
  ├── generator                                    │
  └── bridges                                      │
                                                   │
internal/generator ────────────────────────────────┤
  ├── config                                       │
  └── bridges                                      │
                                                   │
internal/docker ───────────────────────────────────┤
  └── config                                       │
                                                   │
internal/matrix ───────────────────────────────────┤
  ├── config                                       │
  └── bridges                                      │
                                                   │
internal/bridges ──────────────────────────────────┘
internal/config ───────────────────────────────────┘
  (no internal dependencies)
```

`config` and `bridges` are leaf packages with no internal dependencies.

## Design Patterns

### Model-View-Update (Elm Architecture)

The TUI uses Bubble Tea which implements the Elm architecture:

```go
type Model struct { ... }
func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string
```

Each screen (Dashboard, Bridges, Logs, Settings, Wizard) is its own sub-model composed into the main `Model`. The main model delegates to the active screen:

```go
switch m.screen {
case ScreenDashboard:
    m.dashboard, cmd = m.dashboard.Update(msg, m.config, m.compose)
case ScreenBridges:
    m.bridges, cmd = m.bridges.Update(msg, m.config)
// ...
}
```

### Message Passing

Async operations use typed messages:

```go
// Define a message type
type servicesUpdatedMsg struct {
    services []docker.ServiceStatus
}

// Return a command that produces the message
func (m *Model) fetchStatusCmd() tea.Cmd {
    return func() tea.Msg {
        services, _ := m.compose.Status()
        return servicesUpdatedMsg{services: services}
    }
}

// Handle the message in Update()
case servicesUpdatedMsg:
    m.services = msg.services
```

### Registry Pattern

Bridges are loaded from embedded YAML into a package-level registry:

```go
//go:embed bridges.yaml
var bridgesYAML []byte

var registry map[string]BridgeInfo

func init() {
    // parse YAML into registry
}

func Get(name string) *BridgeInfo
func List() []BridgeInfo
func Exists(name string) bool
```

### Factory Constructors

Every package uses `New*()` functions instead of direct struct initialization:

```go
compose := docker.New(cfg)
gen := generator.New()
client := matrix.NewClient(url)
```

Struct fields are private; interaction is through methods.

## Adding a New Bridge

### 1. Add to bridges.yaml

```yaml
mybridge:
  description: My messaging service
  port: 29330                              # Pick an unused port
  note: Requires something special         # Optional
  requires_api_credentials: false          # true if needs API keys
  login_instructions: |
    1. Chat with @mybridgebot:SERVER
    2. Send: login
    3. Follow the prompts

    Docs: https://docs.mau.fi/bridges/go/mybridge/
```

The `Name` field is set automatically from the YAML key.

### 2. Create the bridge config template

Create `internal/generator/templates/bridges/mybridge.yaml.tmpl`:

```yaml
homeserver:
  address: http://synapse:8008
  domain: {{.ServerName}}

appservice:
  address: http://mautrix-mybridge:{{.Port}}
  hostname: 0.0.0.0
  port: {{.Port}}
  database:
    type: sqlite3-fk-wal
    uri: /data/mybridge.db
  id: mybridge
  bot:
    username: {{.BotUsername}}
    displayname: My Bridge Bot
  as_token: {{.ASToken}}
  hs_token: {{.HSToken}}

bridge:
  username_template: mybridge_{{`{{.}}`}}
  displayname_template: {{`{{.Name}}`}}
  permissions:
    "*": relay
    "{{.ServerName}}": user
    "@{{.AdminUser}}:{{.ServerName}}": admin
  double_puppet:
    secrets:
      {{.ServerName}}: "{{.DoublePuppetSecret}}"

logging:
  min_level: info
  writers:
    - type: stdout
      format: pretty-colored
```

Template variables available:
- `{{.ServerName}}` - Matrix server name (e.g., "localhost")
- `{{.Port}}` - Bridge port from bridges.yaml
- `{{.AdminUser}}` - Admin username
- `{{.ASToken}}` - Appservice token (auto-generated)
- `{{.HSToken}}` - Homeserver token (auto-generated)
- `{{.BotUsername}}` - Bot username (bridgename + "bot")
- `{{.DoublePuppetSecret}}` - Double puppet secret

For Telegram specifically, `{{.TelegramAPIID}}` and `{{.TelegramAPIHash}}` are also available.

### 3. Add to docker-compose.yml

Add the service to `internal/docker/docker-compose.yml`:

```yaml
mautrix-mybridge:
  image: dock.mau.dev/mautrix/mybridge:latest
  profiles:
    - mybridge
  volumes:
    - ${DATA_DIR}/bridges/mybridge:/data
    - ${CONFIG_DIR}/bridges/mybridge/registration.yaml:/data/registration.yaml:ro
  depends_on:
    - synapse
  networks:
    - muxbee
  restart: unless-stopped
```

The profile name must match the bridge name in bridges.yaml.

### 4. Add welcome message (optional)

In `internal/matrix/setup.go`, add to `BotWelcomeMessages`:

```go
"mybridge": `Welcome to the My Bridge!

To connect:
1. Send: login
2. Follow the prompts

Commands:
• login - Start linking
• logout - Disconnect
• help - Show all commands

Docs: https://docs.mau.fi/bridges/go/mybridge/`,
```

### 5. Build and test

```bash
go build -o muxbee .
./muxbee bridge list              # Should show your bridge
./muxbee bridge enable mybridge   # Enable it
./muxbee up                       # Start everything
./muxbee logs mautrix-mybridge    # Check for errors
```

## Template System

Templates live in `internal/generator/templates/` and are embedded at compile time.

The generator creates files in two places:
- `~/.config/muxbee/` - Configuration files (Synapse reads these)
- `~/.local/share/muxbee/` - Data files (bridges write here)

Bridge configs go to the data directory because bridges expect a writable `/data` mount. Bridge registrations go to both (Synapse needs them in config, bridges need them in data).

## Token Persistence

Appservice tokens (AS/HS tokens) are generated once and stored in `settings.yaml`. This ensures tokens survive across `muxbee up` invocations. The generator checks for existing tokens before generating new ones:

```go
func (c *Config) GetOrCreateBridgeTokens(bridgeName string) (BridgeTokens, error) {
    if tokens, exists := c.BridgeTokens[bridgeName]; exists {
        return tokens, nil
    }
    // generate new tokens...
}
```

## Testing

```bash
go test ./...                      # Run all tests
go test ./internal/config/...      # Run specific package
go test -v ./internal/bridges/...  # Verbose output
```

Most tests use temporary directories to avoid touching real config files.
