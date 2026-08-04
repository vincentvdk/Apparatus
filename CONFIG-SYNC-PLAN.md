# Config Sync Implementation Plan

## Overview

Implement a unified configuration management system that allows the host OS to manage and sync configs to all distroboxes. This replaces the previous approach of managing configs independently in each distrobox.

## Goals

1. **Single source of truth**: All configs managed from host OS (`/usr/share/apparatus/`)
2. **Sync to distroboxes**: Push config updates from host to all distroboxes
3. **Skip capability**: Allow skipping specific distroboxes (bespoke configs)
4. **Migration**: Move from `butler.sh` to golang-based butler
5. **Initial sync**: Distrobox initialization includes config sync

---

## Current State Analysis

### Config Sources

1. **Host OS configs** (`/usr/share/apparatus/`):
   - Copied from `apparatus-dotfiles` during OS build
   - Contains: hyprland, waybar, mako, kitty, walker, uwsm, satty configs
   - Contains: themes (catppuccin-mocha, catppuccin-latte)

2. **apparatus-dotfiles repo** (`github.com/vincentvdk/apparatus-dotfiles`):
   - `home/dot_config/`: All user configs (hypr, kitty, waybar, mako, zsh, nvim, atuin, walker, uwsm, satty)
   - `themes/`: Theme configurations
   - `.chezmoiignore`: Conditional ignore rules based on environment

3. **Distrobox Containerfile** (`box/Containerfile`):
   - Currently only copies: zsh, atuin, nvim configs
   - Does NOT copy UI configs (hypr, kitty, waybar, etc.)
   - Runs `init.sh` on first login

4. **Distrobox init.sh** (`box/bin/init.sh`):
   - Runs on first login (or always for shared home)
   - Sets up zsh, antidote, p10k, nvm, atuin, git, neovim
   - Does NOT sync UI configs from host

5. **butler.sh**:
   - Has theme switching with symlinks (lines 418-479)
   - Only manages host OS configs

6. **golang butler** (`tools/cmd/butler/`):
   - Has theme switching in `os_mode.go` (lines 874-923)
   - Only manages host OS configs
   - No distrobox config sync capability

### Environment Detection

The `profile-custom.sh` sets up environment variables:
- `APPARATUS_OS_HOME=1`: Shared home with host (use container-specific config paths)
- `APPARATUS_OS_HOME=0`: Custom home directory (use standard paths)

---

## Implementation Plan

### Phase 1: Update Distrobox Initialization (SKIPPED - Per User Decision)

**Decision**: User does not want to sync configs when a new distrobox is initialized.
Config sync will be manual via golang butler TUI (`butler` → "Sync Configs") or CLI (`butler sync`).
The `init.sh` script does NOT perform config sync - it only sets up zsh, antidote, p10k, nvm, atuin, git, and neovim.

---

### Phase 2: Add Config Sync to Golang Butler (High Priority - DONE)

**Status**: Already implemented in `tools/cmd/butler/os_mode.go` and `tools/cmd/butler/main.go`

**Implemented**:
- Menu items: "Sync Configs" and "Skip List" 
- `syncDistroboxConfigs()` - initiates sync to all distroboxes
- `syncConfigsToAllDistroboxes()` - syncs to all non-skipped distroboxes
- `syncConfigToDistrobox()` - syncs to a single distrobox
- `getAllDistroboxes()` - gets list of all distroboxes with home directories
- `getDistroboxHomeDir()` - gets home directory for a specific distrobox
- `copyDir()` - recursive directory copy

---

### Phase 3: Skip List Management UI (High Priority - DONE)

**Status**: Already implemented in `tools/cmd/butler/os_mode.go`

**Implemented**:
- `showSkipList()` - displays list of distroboxes with skip checkbox indicators
- `handleSkipListToggle()` - toggles skip status for a distrobox
- `loadSkipList()` - loads skip list from `~/.config/apparatus/skip-distroboxes`
- `saveSkipList()` - saves skip list to file
- Skip list file format: one distrobox name per line

---

### Phase 2: Add Config Sync to Golang Butler (High Priority)

**File**: `tools/cmd/butler/os_mode.go`

**New menu item** (add to `makeMainList()` around line 86-92):
```go
osItem{title: "Sync Distrobox Configs", description: "Sync configs from host to all distroboxes", action: "sync-configs"}
```

**New types**:
```go
// Distrobox info for sync operations
type distroboxInfo struct {
    Name     string
    HomeDir  string
    IsRunning bool
    Skip     bool
}
```

**New functions**:

1. `getDistroboxes() ([]distroboxInfo, error)`:
   - Parse `distrobox list --no-color` output
   - For each distrobox, get home directory via `podman inspect`
   - Check if running via `distrobox status`
   - Check skip list

2. `getDistroboxHomeDir(name string) (string, error)`:
   - Run `podman inspect <name> --format '{{json .Args}}'`
   - Parse JSON to find `--home` argument
   - Return home directory path

3. `loadSkipList() (map[string]bool, error)`:
   - Read `~/.config/apparatus/skip-distroboxes` (JSON or simple text)
   - Return map of distrobox names to skip

4. `saveSkipList(skipMap map[string]bool) error`:
   - Write skip list to file

5. `syncDistroboxConfigs() tea.Cmd`:
   - Get all distroboxes
   - Load skip list
   - For each non-skipped distrobox:
     - Copy configs from `/usr/share/apparatus/` to distrobox home
     - Apply theme symlinks
   - Return commandDoneMsg with results

**Config copy logic**:
```go
func copyConfigToDistrobox(distrobox distroboxInfo) error {
    src := "/usr/share/apparatus"
    dst := filepath.Join(distrobox.HomeDir, ".config")
    
    configs := []string{"hypr", "kitty", "waybar", "mako", "walker", "uwsm", "satty", "atuin", "nvim", "zsh"}
    
    for _, config := range configs {
        srcPath := filepath.Join(src, config)
        dstPath := filepath.Join(dst, config)
        
        if _, err := os.Stat(srcPath); err == nil {
            os.RemoveAll(dstPath)
            if err := copyDir(srcPath, dstPath); err != nil {
                return err
            }
        }
    }
    
    themesSrc := filepath.Join(src, "themes")
    themesDst := filepath.Join(dst, "apparatus", "themes")
    os.RemoveAll(themesDst)
    copyDir(themesSrc, themesDst)
    
    return nil
}
```

**Note on cross-container copying**:
For custom home directories, copy directly to host path.
For shared home, use `distrobox enter` or `podman cp`.

---

### Phase 3: Skip List Management UI (High Priority)

**File**: `tools/cmd/butler/os_mode.go`

**New menu item**:
```go
osItem{title: "Manage Skipped Distroboxes", description: "Select which distroboxes to skip during config sync", action: "manage-skip"}
```

**New popup mode** (add to constants around line 29-42):
```go
popupModeSkipList
```

**New functions**:
1. `showSkipListPopup()`: Show list of all distroboxes with checkbox for each
2. `handleSkipListAction(item osItem)`: Toggle skip status, save list

---

### Phase 4: CLI Command (Medium Priority - DONE)

**Status**: Already implemented in `tools/cmd/butler/main.go` (lines 384-402)

**Implemented**:
- `butler sync` - syncs configs to all distroboxes
- `butler sync <name>` - syncs configs to specific distrobox
- Uses existing `syncConfigsToAllDistroboxes()` and `syncConfigsToDistroboxCLI()` functions

---

### Phase 5: Deprecation Notice (Low Priority - SKIPPED)

**Decision**: User indicated deprecation notice is not needed.

---

### Phase 6: Update Containerfile (Medium Priority - PARTIALLY DONE)

**File**: `box/Containerfile`

**Status**: Already copying all configs (lines 65-78), but verify themes are included.

**Current**:
```dockerfile
ARG DOTFILES_REPO="https://github.com/vincentvdk/apparatus-dotfiles.git"
RUN git clone --depth=1 "$DOTFILES_REPO" /tmp/dotfiles && \
    mkdir -p /usr/share/apparatus && \
    cd /tmp/dotfiles/home/dot_config && \
    for dir in zsh atuin nvim hypr kitty waybar mako walker uwsm satty; do \
        [ -d "$dir" ] && cp -r "$dir" /usr/share/apparatus/; \
    done && \
    cd /usr/share/apparatus && \
    find . -name 'dot_*' | while read f; do \
        d=$(dirname "$f"); b=$(basename "$f"); \
        mv "$f" "$d/.${b#dot_}"; \
    done && \
    cp -r /tmp/dotfiles/themes /usr/share/apparatus/themes && \
    rm -rf /tmp/dotfiles
```

---

## File Changes Summary

| File | Changes | Priority | Status |
|------|---------|----------|--------|
| `box/bin/init.sh` | No config sync changes | High | **SKIPPED** |
| `tools/cmd/butler/os_mode.go` | Add sync menu item, sync functions, skip list management | High | **DONE** |
| `tools/cmd/butler/main.go` | Add `sync` CLI command | Medium | **DONE** |
| `os/build_files/apparatus/butler.sh` | Add deprecation notice | Low | **SKIPPED** |
| `box/Containerfile` | Copy all configs, not just CLI tools | Medium | **DONE** |

## Remaining Work

### Config Sync Feature
- [x] Phase 2: Golang butler sync implementation
- [x] Phase 3: Skip list management UI
- [x] Phase 4: CLI command
- [x] Phase 5: Deprecation notice in butler.sh (**SKIPPED**)
- [x] Phase 6: Containerfile (already copies all configs)

### Other Tasks
- [x] Fix build error: paths in build.sh are correct (files exist at expected locations)
- [ ] Satty configuration review and improvements
- [x] Verify Containerfile copies all needed configs correctly (confirmed: copies all configs + themes)

---

## Config File Locations

### Source (Host OS)
```
/usr/share/apparatus/
├── hypr/
├── kitty/
├── waybar/
├── mako/
├── walker/
├── uwsm/
├── satty/
├── atuin/
├── nvim/
├── zsh/
└── themes/
    ├── catppuccin-mocha/
    └── catppuccin-latte/
```

### Destination (Distrobox)

For **shared home** (`APPARATUS_OS_HOME=1`):
```
$APPARATUS_CONFIG_HOME/
├── apparatus-box/
└── apparatus/
    └── themes/
```

For **custom home** (`APPARATUS_OS_HOME=0`):
```
$HOME/.config/
├── hypr/
├── kitty/
├── waybar/
└── apparatus/
    └── themes/
```

---

## User Workflow

### Initial Setup
1. User creates distrobox with custom home
2. On first login, `init.sh` runs and sets up zsh, antidote, p10k, nvm, atuin, git
3. User must manually run `butler sync` to sync configs from host

### Config Update
1. User updates configs in `apparatus-dotfiles` repo
2. User rebuilds OS image (updates `/usr/share/apparatus/`)
3. User runs: `butler sync` (CLI) or uses TUI menu

### Skip Management
1. User manages skip list via TUI or CLI

---

## Technical Considerations

### Cross-Container Copying
- Custom home: Copy directly to host filesystem
- Shared home: Use `distrobox enter` or `podman cp`

### File Permissions
- Ensure correct ownership for copied files

### Performance
- Use `rsync` for efficient transfers
- Parallelize where possible
- Show progress in TUI

### Error Handling
- Continue syncing other distroboxes if one fails
- Report successes/failures
- Log errors for debugging

---

## Decisions

1. **Sync everything**: All configs from apparatus-dotfiles
2. **No process reloads**: Just copy files, services auto-pickup
3. **Create directories**: Parent dirs created automatically
4. **Copy files, not symlinks**: Symlinks won't work across container filesystems
