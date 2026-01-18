# Contributing to muxchat

Thank you for your interest in contributing to muxchat!

## Development Setup

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Make (optional but recommended)

### Building from Source

```bash
git clone https://github.com/tobocop2/muxchat.git
cd muxchat
go build -o muxchat .
```

### Running Tests

```bash
go test ./...
```

### Project Structure

```
muxchat/
├── cmd/                    # CLI commands (Cobra)
├── internal/
│   ├── config/            # Configuration management
│   ├── bridges/           # Bridge registry
│   ├── docker/            # Docker Compose wrapper
│   ├── generator/         # Template generation
│   └── tui/               # Terminal UI (Bubble Tea)
├── templates/             # Embedded templates
└── scripts/               # Build and install scripts
```

## Guidelines

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and small
- Add comments for exported functions

### Commits

- Use clear, descriptive commit messages
- Reference issues when applicable
- Keep commits atomic (one logical change per commit)

### Pull Requests

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Run tests: `go test ./...`
5. Build and verify: `go build -o muxchat . && ./muxchat --version`
6. Submit a pull request

**PR Guidelines:**
- Keep PRs focused — one feature or fix per PR
- Include a clear description of what changed and why
- Update documentation if behavior changes
- Add tests for new functionality
- Ensure CI passes before requesting review

**PR Title Format:**
- `feat: add new feature` — new functionality
- `fix: resolve bug` — bug fixes
- `docs: update readme` — documentation only
- `refactor: restructure code` — code changes that don't add features or fix bugs
- `test: add tests` — test additions or fixes

### Testing

- Add tests for new functionality
- Ensure all existing tests pass
- Test on both Linux and macOS if possible

## Architecture Decisions

### Single Binary
All assets (templates, bridge definitions, docker-compose) are embedded using `go:embed` for single-binary distribution.

### Docker Compose via Shell
We shell out to `docker compose` rather than using the Docker SDK. This is simpler and has fewer dependencies.

### TUI with Bubble Tea
The interactive interface uses Bubble Tea for a clean, maintainable TUI architecture.

## Getting Help

- Open an issue for bugs or feature requests
- Check existing issues before creating new ones
- Be respectful and constructive in discussions

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
