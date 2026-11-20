#!/bin/bash
# tuios smart-split (POC)
#
# Port of apparatus's kitty smart-split for tuios. Instead of scanning
# the process tree (what the kitty version does via `kitty @ ls` + jq,
# and what we tried - with much more pain - against WezTerm's mux
# panes), this reads the new window's cwd directly: apparatus distrobox
# homes live at .../distrobox-homes/<name>, and tuios already reports
# a real `cwd` field per window via `list-windows --json`. If the new
# window's cwd falls under a distrobox home, re-enter that distrobox.
#
# NOTE: the tuios daemon runs hooks with a minimal PATH that does not
# include ~/.local/bin, so every call below uses the absolute path to
# the tuios binary - a bare `tuios` silently fails as "command not
# found" under the hook's environment (found via debug logging after
# the hook appeared to run but never actually entered the container).
TUIOS=/home/vincent/.local/bin/tuios

log() {
  printf '%s %s\n' "$(date +%H:%M:%S)" "$1" >> /tmp/smart-split-tuios-debug.log
}

id="${TUIOS_WINDOW_ID:-}"
log "--- hook fired, TUIOS_WINDOW_ID=$id PATH=$PATH which-tuios=$(command -v "$TUIOS" 2>&1) which-jq=$(command -v jq 2>&1)"
[ -z "$id" ] && { log "no id, exit"; exit 0; }

# The new window's cwd field is empty until its shell actually starts
# and reports it via OSC7 - that doesn't happen instantly, so poll for
# it instead of trusting a single fixed delay (found via debug
# logging: at a flat 0.3s sleep, cwd was still empty).
cwd=""
for _ in 1 2 3 4 5 6 7 8 9 10; do
  sleep 0.2
  raw=$("$TUIOS" list-windows --json 2>&1)
  cwd=$(printf '%s' "$raw" | jq -r --arg id "$id" '.windows[] | select(.window_id==$id) | .cwd // empty' 2>&1)
  [ -n "$cwd" ] && break
done
log "resolved cwd=[$cwd]"
[ -z "$cwd" ] && { log "no cwd after polling, exit"; exit 0; }

name=$(printf '%s' "$cwd" | sed -n 's#.*/distrobox-homes/\([^/]*\).*#\1#p')
log "resolved name=[$name]"
[ -z "$name" ] && { log "no name match, exit"; exit 0; }

# exec, not a plain command: replaces this pane's shell process with
# distrobox-enter instead of running it as a child of that shell.
# Without this, exiting the container just drops back to the host
# shell underneath instead of closing the pane - kitty's smart-split
# avoids the same problem the same way (`exec distrobox enter ...`).
out=$("$TUIOS" send-text -w "$id" "exec distrobox enter $name
" 2>&1)
log "send-text output: $out (exit=$?)"
