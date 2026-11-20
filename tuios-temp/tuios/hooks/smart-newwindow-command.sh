#!/bin/bash
# tuios smart new-window (POC)
#
# Bound to prefix+c, replacing the built-in new_window action. The
# workspace NAME set by project-launcher.sh persists even after every
# pane in it closes (naming a workspace and having panes in it are
# independent - found this is exactly what's needed for "closed all
# windows in a customer's workspace, want the next one to still land
# in their box"). No per-command-key workspace env var exists, so the
# current workspace number comes from list-workspaces --json instead.
#
# If the current workspace is named after a REAL distrobox (checked
# against `distrobox list`, not a directory under $HOMES), opens the
# new window straight into it - same direct-exec, no-shell-flash
# approach as smart-split-command.sh. Otherwise falls back to a plain
# new window.
#
# Checking distrobox list rather than a directory under $HOMES matters
# for a distrobox sharing the host's home dir instead of a dedicated
# one - it would never have a $HOMES/<name> directory to check, even
# though it's a perfectly real distrobox.
TUIOS=/home/vincent/.local/bin/tuios
HOMES=/home/vincent/distrobox-homes
session="${TUIOS_SESSION:-}"
sflag=()
[ -n "$session" ] && sflag=(-s "$session")

log() {
  printf '%s %s\n' "$(date +%H:%M:%S)" "$1" >> /tmp/smart-newwindow-debug.log
}
log "fired TUIOS_SESSION=[$session] TUIOS_ACTIVE_PANE_ID=[${TUIOS_ACTIVE_PANE_ID:-}] TUIOS_ACTIVE_PANE_CWD=[${TUIOS_ACTIVE_PANE_CWD:-}]"

ws_json=$("$TUIOS" list-workspaces --json "${sflag[@]}" 2>&1)
current=$(printf '%s' "$ws_json" | jq -r '.current_workspace')
name=$(printf '%s' "$ws_json" | jq -r --arg w "$current" '.workspaces[] | select((.workspace|tostring)==$w) | .name')
log "current=[$current] name=[$name] ws_json=$ws_json"

is_real_distrobox=""
if [ -n "$name" ]; then
  if distrobox list --no-color 2>/dev/null | awk -F'|' 'NR>1{gsub(/^[ \t]+|[ \t]+$/,"",$2); print $2}' | grep -qx "$name"; then
    is_real_distrobox=1
  fi
fi
log "is_real_distrobox=[$is_real_distrobox]"

if [ -n "$is_real_distrobox" ]; then
  # Only set --cwd if this distrobox actually has a dedicated home dir -
  # a shared-home one has no $HOMES/<name> to start in, so let
  # new-window fall back to its own default instead of pointing it at
  # a path that doesn't exist.
  cwdflag=()
  [ -d "$HOMES/$name" ] && cwdflag=(--cwd "$HOMES/$name")
  out=$("$TUIOS" new-window "$name" "${sflag[@]}" "${cwdflag[@]}" -- distrobox enter "$name" 2>&1)
  log "new-window(distrobox) output: $out"
else
  out=$("$TUIOS" new-window "${sflag[@]}" 2>&1)
  log "new-window(plain) output: $out"
fi
