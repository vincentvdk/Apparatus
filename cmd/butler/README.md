# Butler - Unified Configuration Tool

A single Go binary that provides configuration management for both the **Host OS** and **Distrobox containers** in Apparatus OS.

## Architecture

```
butler (single binary)
├── Auto-detect or APPARATUS_MODE=os|box
│
├── OS Mode (Host)
│   ├── Distrobox Manager (create/list/upgrade)
│   ├── System Config (terminal, monitors, audio, AI VRAM)
│   ├── Theme Switcher (Catppuccin)
│   ├── Font Switcher
│   └── Help / Keybindings
│
└── Box Mode (Container)
    ├── Tool Categories (cloud, k8s, languages)
    ├── Tool Version Manager (mise)
    ├── LSP Auto-Setup (neovim)
    └── Tool Search
```

## Environment Detection

### Auto-detect
```bash
butler
```

Checks for `hyprctl` (OS mode) or `mise` (box mode) on PATH.

### Explicit mode
```bash
butler os    # Force Host OS mode
butler box   # Force Distrobox mode
```

### Environment variable override
```bash
APPARATUS_MODE=os butler
APPARATUS_MODE=box butler
```

## Deployment

### Host OS (os/Containerfile.bootc)
```bash
cp os/build_files/apparatus/butler.sh /usr/local/bin/butler
```

### Distrobox Container (box/Containerfile)
```dockerfile
FROM docker.io/golang:1.22-alpine AS builder
WORKDIR /build
COPY cmd/butler/ .
RUN go build -o butler .
...
COPY --from=builder /build/butler /usr/local/bin/butler
```

## Migration from Legacy

| Old | New |
|-----|-----|
| `butler.sh` (host OS menu) | `butler` (auto-detects OS mode) |
| `butler-tools` (box TUI) | `butler` (auto-detects box mode) |
| `box/bin/tool-manager.sh` | Still used by box mode |

## Features

### OS Mode
- **Distrobox Management**: Create, list, enter, upgrade, stop, remove containers
- **Terminal Selection**: Set default terminal to kitty
- **Monitor Config**: View current monitors, launch hyprdynamicmonitors
- **Audio Config**: Launch pavucontrol
- **AI Workload**: Configure AMD VRAM allocation (pkexec)
- **Theme Switching**: Catppuccin Mocha/Latte (symlinks to all config files)
- **Font Switching**: Ioskeley Mono, JetBrains Mono, Hack Nerd Font
- **Help**: Keyboard shortcuts reference

### Box Mode
- **Tool Categories**: Cloud, Kubernetes, Languages
- **Version Management**: latest, current, available, setversion
- **LSP Auto-Setup**: Installs LSP server + copies Neovim config on `latest`
- **Tool Search**: Filter tools by name
- **Config**: `tools.yaml` defines categories and tools

## Configuration

### Box Mode Config
```yaml
# /usr/share/apparatus/tools.yaml
categories:
  languages:
    icon: "💻"
    tools:
      - name: golang
        description: Go programming language
        lsp:
          name: gopls
          install_cmd: go install golang.org/x/tools/gopls@latest
```

### Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| `APPARATUS_MODE` | auto-detect | Force `os` or `box` mode |
| `BUTLER_CONFIG_PATH` | `/usr/share/apparatus/tools.yaml` | Path to tools.yaml |
| `BUTLER_TOOL_MANAGER` | `/opt/bin/tool-manager.sh` | Tool manager script |
| `BUTLER_LSP_TEMPLATES` | `/usr/share/apparatus/lsp` | LSP config templates dir |

## Tech Stack
- **Language**: Go 1.22
- **TUI Framework**: Charm Bubbletea + Bubbles + Lipgloss
- **Config Parsing**: YAML (gopkg.in/yaml.v3)
- **External Tools**: mise, distrobox, hyprctl, pkexec
