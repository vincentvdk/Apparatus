# TUIOS — Terminal Multiplexer

Apparatus uses [tuios](https://tuios.dev) as its terminal multiplexer, running inside kitty on the host. A shared daemon holds all sessions, so host and every distrobox see and can attach to the same sessions. TUIOS combines with distroboxes so that each piece of work runs inside the right container without typing `distrobox enter` by hand.

## Concept model

TUIOS organizes work in four levels. In Apparatus they map to:

| tuios level | Apparatus meaning | Example |
|---|---|---|
| **Space** (groups sessions in the rail and `tuios ls`) | Customer | `autops` |
| **Session** | A piece of work, created together with its distrobox | `wharf` (space `autops`, runs in the `autops` distrobox) |
| **Workspace** (max 9 per session) | An aspect of that work | `edit`, `serve`, `agents`, `logs` |
| **Window** | A single pane | a shell, an editor, an agent |

```
daemon (shared by host + all distroboxes)
└── space "autops"
    └── session "wharf"            ← first window execs: distrobox enter autops
        ├── workspace 1 "edit"     → windows: nvim, shell
        ├── workspace 2 "serve"    → windows: dev server
        └── workspace 3 "agents"   → windows: coding agents
```

Key properties:

- **Sessions survive detach and terminal close** (daemon-backed), and can be attached from several terminals at once — including from inside a distrobox, since the daemon socket is shared via the home directory.
- **A session is not tied to a location.** Start tuios on the host; sessions are daemon-global.
- **Distrobox = environment, session = work.** A piece of work can use several distroboxes as several sessions under one customer space, expanding the old "one distrobox per project" rule.

## Starting tuios

```bash
# In a fresh terminal on the host (opened from the desktop, not from inside tuios):
tuios                 # attaches to startup.session ("main"), creating it if needed
tuios attach wharf    # attach a specific session
tuios new wharf       # create + attach a new session
tuios ls              # list sessions
```

Set a deterministic default in `~/.config/tuios/config.toml`:

```toml
[startup]
session = "main"      # plain `tuios` always lands here, not in the last-touched session
daemon = true
```

### "You are inside session …" error

Plain `tuios` is refused when the daemon thinks the current shell is already inside a tuios session:

```
ERROR
You are inside session "DISTRO-VTECH". Other sessions: HOST. To show one of them in this pane, run 'tuios attach NAME'. To start a new session, run 'tuios new NAME'.
```

The daemon detects "inside" from the process ancestry, the controlling tty, or pane env vars — **not** from hostname or `env` output. Two cases:

1. You really are in a pane of that session (e.g. a shell inside the box's tuios). Then use `tuios attach OTHER`, `tuios new NAME`, or switch sessions inside the TUI with `prefix+S` / `alt+shift+n` / `alt+shift+p` — never type plain `tuios` in a pane.
2. **False positive on the host** (known upstream issue, tracked in `ISSUES.md`): tuios compares raw pty device numbers, which collide between the host's `/dev/pts` and a container's private devpts. A host terminal on `pts/3` can be misclassified as inside a session whose container pane also reports `pts/3`. Workarounds: `tuios attach main` (named attach to another session is allowed), `tuios attach --force`, `TUIOS_ALLOW_NESTED=1 tuios`, or setting `startup.session` so plain `tuios` is a named attach.

## Distrobox integration

### Creating a work session in a distrobox

The `new-session` verb takes `command` (argv, no shell), `cwd` and `space`, so a session can be born inside its distrobox:

```bash
# Session "wharf" for customer autops, first window direct-exec'd into the autops box:
tuios new-session --name wharf --space autops \
  --cwd ~/distrobox-homes/autops \
  --command distrobox enter autops
```

No shell starts first, nothing is typed or echoed, and exiting the container closes the pane. The custom home (`~/distrobox-homes/<box>`) is pinned via `--cwd`, keeping the per-project home coupling.

In the TUI, the **project launcher** (`prefix+l`, hook `~/.config/tuios/hooks/project-launcher.sh`) does this interactively: pick a customer, it claims the next free workspace/session, names it, and spawns the first window inside their distrobox. If the work already exists (by name or by any pane's cwd), it just jumps there — no duplicates.

### Splits and new windows stay in the same distrobox

Smart-split hooks replace the plain split keys, ported from the kitty smart-split:

| Key | Action |
|---|---|
| `prefix+-` | Smart split horizontal |
| `prefix+\|` / `prefix+\` | Smart split vertical |
| `prefix+c` | Smart new window (re-enters the current workspace's distrobox) |

The source pane's distrobox is detected via `CONTAINER_ID` in the `podman exec` argv of the pane's process subtree (robust for shared-home boxes), falling back to the `~/distrobox-homes/<name>` cwd pattern. The new pane is created with `new-window <name> --cwd <cwd> -- distrobox enter <name>` — direct exec, no shell flash. Plain host splits still use `split-window` with correct direction.

Notes for hook authors: always pass `-s "$TUIOS_SESSION"` explicitly (the env var alone does not target the session), use the absolute path to the `tuios` binary (the daemon's hook PATH lacks `~/.local/bin`), and always pass `--cwd` — a direct-exec'd process never emits OSC 7, so without it the pane's reported cwd stays stale and breaks detection for further splits.

### Switching work

| Key | Action |
|---|---|
| `prefix+S` | Session switcher |
| `alt+shift+n` / `alt+shift+p` | Next / previous session |
| `alt+1..9` | Switch workspace |
| `alt+shift+1..9` | Move window to workspace |
| `prefix+d` | Detach (session keeps running in the daemon) |
| `prefix+w` / `prefix+W` | Workspace menu / workspace switcher |

## Agents

TUIOS tracks coding agents (Claude Code, Codex, Gemini CLI, …) in every pane: a state glyph (`working` ●, `needs_input` ▲, `done` ■, `errored` ×) on the title bar and in the rail, with everything blocked collected in the **Inbox** (`prefix+i`).

```bash
tuios start-agent claude --name reviewer --grants read,write
tuios start-agent --preset worker --prompt 'Run the tests and fix what fails.'
tuios list-agents                 # states of all agent panes
tuios wait-for agent-state -s wharf --until needs_input
tuios respond "yes"               # answer a blocked agent from the shell
```

Recommended config (synced to all boxes via butler):

```toml
[agents.permissions]
mode = "strict"                   # the daemon socket is reachable from every distrobox
grants = ["read", "write", "fan"] # no "respond": panes can't answer prompts for you

[agents.presets.worker]
agent = "claude"
grants = ["read", "write"]

[notifications.agent]
sound = true
quiet_hours = "22:00-07:00"

[notifications.agent.states]
needs_input = true
errored = true
done = true
```

Agent panes live in the work session's distrobox, so they inherit the project environment and butler-synced configs. For untrusted runs, start a session in a dedicated sandbox box (`--command distrobox enter sandbox`). Run `tuios integration install --all` **inside each box** (harness configs live in the box's own home) so agents self-report state via hooks; process-table autodetect works across boxes regardless.

## Configuration & sync

- Config: `~/.config/tuios/config.toml` (hot-reloaded on save; invalid files are ignored with an on-screen error), hooks in `~/.config/tuios/hooks/`.
- The `tuios` directory is part of butler's synced config list, so the same config and hooks reach every distrobox (`butler sync`). Host defaults are baked into the image at `/usr/share/apparatus/tuios/`.
- `tuios keybinds list` / `tuios keybinds doctor` inspect and fix bindings from the shell; most things are also editable live via the keybind manager (`prefix+k`) and settings (`prefix+,`).

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| `You are inside session "X"` on the host | See above: use `tuios attach NAME`, set `startup.session`, or see the pts-collision issue in `ISSUES.md` |
| New windows land in the wrong session | CLI calls need `-s <session>` explicitly; `TUIOS_SESSION` in the env is not enough |
| Hook runs but nothing happens | The daemon's hook PATH lacks `~/.local/bin` — call the absolute tuios path |
| Pane cwd stays at the host home after a distrobox split | Direct-exec'd panes don't emit OSC 7; always pass `--cwd` on `new-window` |
| Agent pane shows no state inside a distrobox | Run `tuios integration install <harness>` inside that box; check `tuios doctor agents` |
| Config change widening grants/hosts doesn't apply | It waits for you: run `tuios config apply` from outside tuios |
