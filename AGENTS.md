# Apparatus Project - Agent Guidelines

## Documentation

### Documentation Location
All end-user documentation must be placed in the `docs/` folder as markdown files.

### Documentation Rendering
- Documentation is rendered using **Zensical**
- All markdown files in `docs/` are automatically rendered
- Use clear, user-focused language
- Include practical examples

### Documentation Update Policy
- **ALL** new features, changes, or bug fixes must have corresponding documentation updates
- Documentation should be updated in the same commit/PR as the feature change
- If a feature is added or modified, update the relevant docs file(s)
- If a new concept or workflow is introduced, create a new docs file

### Documentation Structure
```
docs/
├── index.md              # Main landing page / overview
├── installation.md       # How to install Apparatus OS
├── getting-started.md    # First steps after installation
├── usage.md              # General usage guide
├── butler.md             # Butler tool documentation
├── distrobox.md          # Distrobox usage in Apparatus
├── configuration.md      # Configuration overview
├── themes.md             # Theme management
├── keybindings.md        # Keyboard shortcuts (existing)
└── troubleshooting.md     # Common issues and solutions
```

## Config Sync Feature

### Overview
Apparatus OS implements a unified configuration management system that allows the host OS to manage and sync configs to all distroboxes. This ensures consistency across all development environments.

### Implementation Details
- **Source of truth**: `/usr/share/apparatus/` on the host contains all system configs
- **Sync mechanism**: `butler sync` command copies configs to all distroboxes
- **Skip list**: `~/.config/apparatus/skip-distroboxes` allows excluding specific distroboxes
- **Initial sync**: Distroboxes automatically sync on first login via `init.sh`
- **Config locations**: `~/.config/{hypr,kitty,waybar,mako,satty,atuin,nvim,zsh,uwsm,walker}/`

### Files Modified
- `tools/cmd/butler/os_mode.go` - Sync functions, skip list management, TUI menu items
- `tools/cmd/butler/main.go` - CLI sync command, helper functions
- `box/bin/init.sh` - Initial config sync on distrobox startup
- `box/Containerfile` - Copy all configs during container build

### Usage
```bash
# Sync all distroboxes
butler sync

# Sync specific distrobox
butler sync mybox

# Manage skip list via TUI
butler  # Navigate to "Skip List"
```

## Code Changes

### Before Implementing
1. Check the existing codebase for similar patterns
2. Update the relevant documentation in `docs/`
3. Add entry to CHANGELOG if significant

### After Implementing
1. Verify documentation is complete
2. Test the feature works as documented
3. Ensure no breaking changes without migration guide

## Testing

- Manual testing is expected for all changes
- Document test cases for complex features
- Update docs with any user-facing test procedures

## Style Guidelines

### Documentation
- Use consistent markdown formatting
- Prefer tables for command references and keybindings
- Include code examples where helpful
- Link to external resources when appropriate
- Keep language clear and concise

### Code
- Follow existing code style in each file
- Use descriptive variable and function names
- Add comments for non-obvious logic
- Keep changes minimal and focused

## Project Structure

```
apparatus/
├── tools/                      # Go applications
│   └── cmd/
│       └── butler/            # Main butler application
│           ├── main.go
│           ├── os_mode.go
│           ├── box_mode.go
│           └── go.mod
├── box/                       # Distrobox container configuration
│   ├── Containerfile
│   └── bin/
├── os/                        # Host OS build files
│   └── build_files/
├── docs/                      # User documentation (Zensical)
│   ├── index.md
│   ├── installation.md
│   └── ...
└── AGENTS.md                  # This file
```
