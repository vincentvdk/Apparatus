#!/bin/bash
# tuios smart-split v2 (POC)
#
# Replaces the after-new-window hook approach (smart-split.sh): that
# version worked, but typing "exec distrobox enter <name>" into an
# already-running shell via send-text is visibly echoed by the pty
# before it runs, and the shell's own login banner prints first too -
# both unavoidable once a plain shell is already running there.
#
# This runs BEFORE the split happens instead: bound directly to the
# split keys via [[keybindings.command]], using
# $TUIOS_ACTIVE_PANE_CWD (the pane being split FROM) to decide the
# distrobox name, then spawning the new pane with
# `distrobox enter <name>` as its process from the start via
# `new-window` - no shell, nothing typed, nothing echoed.
#
# Trade-off: `new-window` has no direction parameter (unlike
# `split-window`), so a distrobox split's placement follows the
# session's own tiling scheme instead of the key you pressed. Plain
# (non-distrobox) splits still go through `split-window`, so they
# keep correct horizontal/vertical placement.
TUIOS=/home/vincent/.local/bin/tuios
direction="${1:-vertical}"
source "$(dirname "${BASH_SOURCE[0]}")/detect-distrobox.sh"

log() {
  printf '%s %s\n' "$(date +%H:%M:%S)" "$1" >> /tmp/smart-split-command-debug.log
}

cwd="${TUIOS_ACTIVE_PANE_CWD:-}"
session="${TUIOS_SESSION:-}"
pane_id="${TUIOS_ACTIVE_PANE_ID:-}"
log "fired direction=$direction TUIOS_ACTIVE_PANE_CWD=[$cwd] TUIOS_ACTIVE_PANE_ID=[$pane_id] TUIOS_SESSION=[$session] env_dump=$(env | grep '^TUIOS_')"

# Primary: CONTAINER_ID off the actual process tree - correct even for
# a distrobox sharing the host's home dir (no .../distrobox-homes/<name>
# path to match in that case). Falls back to the cwd pattern only if
# that lookup can't find anything (e.g. --cwd wasn't set on creation).
t0=$(date +%s.%N)
name=$(detect_container_id "$pane_id")
t1=$(date +%s.%N)
log "detect_container_id=[$name] took $(awk -v a="$t0" -v b="$t1" 'BEGIN{printf "%.3f", b-a}')s"
if [ -z "$name" ]; then
  name=$(printf '%s' "$cwd" | sed -n 's#.*/distrobox-homes/\([^/]*\).*#\1#p')
  log "fallback cwd-match resolved name=[$name]"
fi
log "final resolved name=[$name]"

# TUIOS_SESSION being set in the environment does NOT make the CLI
# target that session on its own - only the explicit -s flag does.
# Without this, new windows landed in whatever session happened to
# be "most recently active" instead of the one the key was actually
# pressed in (found by checking where a freshly created window
# showed up: not in the session the keypress came from).
sflag=()
[ -n "$session" ] && sflag=(-s "$session")

if [ -n "$name" ]; then
  # new-window's [name] positional is required before `--`, even
  # though it's documented as optional - without it, the first word
  # of the command gets consumed as the name instead (found the hard
  # way: "distrobox" became the window name and it tried to exec
  # "enter" as a program).
  # --cwd matters beyond cosmetics: a direct-exec'd process (no shell)
  # never emits the OSC7 escape sequence tuios otherwise tracks cwd
  # from, so without this the new pane's reported cwd stays stuck at
  # its creation-time default forever - breaking distrobox detection
  # for any further split made FROM this one (found by checking
  # TUIOS_ACTIVE_PANE_CWD on a second split: it showed the host home
  # dir despite the pane genuinely running inside the container).
  out=$("$TUIOS" new-window "$name" "${sflag[@]}" --cwd "$cwd" -- distrobox enter "$name" 2>&1)
  log "new-window output: $out"
else
  out=$("$TUIOS" split-window "$direction" "${sflag[@]}" 2>&1)
  log "split-window output: $out"
fi
