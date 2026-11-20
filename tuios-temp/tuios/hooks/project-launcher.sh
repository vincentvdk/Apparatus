#!/bin/bash
# tuios project launcher (POC)
#
# Customer -> distrobox (1:1) -> tuios workspace. Picking a customer:
# - if their workspace already exists, just jumps to it. Checked two
#   ways: a workspace explicitly named after the customer (set by this
#   script), OR - since panes can also get created outside the
#   launcher, e.g. via smart-split - any existing window whose cwd is
#   already under that customer's distrobox home. Found the hard way:
#   name-only matching missed a customer's workspace that had been
#   populated manually, and created a duplicate.
# - otherwise claims the next free (unnamed, empty) workspace, names
#   it, and spawns a pane there that enters their distrobox directly -
#   no manual "distrobox enter" step. Just a plain shell, no agent
#   auto-started.
#
# No fzf on this host, so this is a plain numbered menu instead of a
# fuzzy picker - swap in fzf later if it gets installed.
TUIOS=/home/vincent/.local/bin/tuios
HOMES=/home/vincent/distrobox-homes
session="${TUIOS_SESSION:-poc}"

mapfile -t customers < <(ls "$HOMES")
if [ "${#customers[@]}" -eq 0 ]; then
  echo "No customers found under $HOMES"
  read -r -p "Press Enter to close..."
  exit 1
fi

echo "Pick a customer:"
select customer in "${customers[@]}"; do
  [ -n "$customer" ] && break
  echo "Invalid choice, try again."
done
[ -z "$customer" ] && exit 0

ws_json=$("$TUIOS" list-workspaces --json -s "$session" 2>&1)
win_json=$("$TUIOS" list-windows --json -s "$session" 2>&1)

existing=$(printf '%s' "$ws_json" | jq -r --arg c "$customer" '.workspaces[] | select(.name==$c) | .workspace' | head -1)
if [ -z "$existing" ]; then
  # Match on the trailing path, not a fixed root: /home and /var/home
  # are the same directory via symlink but different strings, and
  # tuios reports the /var/home form - comparing against a hardcoded
  # $HOMES prefix silently never matched (found by dumping the real
  # window list: cwd was /var/home/vincent/... while $HOMES assumed
  # /home/vincent/...).
  existing=$(printf '%s' "$win_json" | jq -r --arg c "/distrobox-homes/$customer" \
    '.windows[] | select(.cwd != null) | select((.cwd | endswith($c)) or (.cwd | contains($c + "/"))) | .workspace' | head -1)
fi

if [ -n "$existing" ]; then
  "$TUIOS" select-workspace "$existing" -s "$session"
  exit 0
fi

next=$(printf '%s' "$ws_json" | jq -r '.workspaces[] | select(.name=="" and .window_count==0) | .workspace' | head -1)
if [ -z "$next" ]; then
  echo "No free workspace (all 9 are named/occupied)."
  read -r -p "Press Enter to close..."
  exit 1
fi

"$TUIOS" set-workspace-name "$next" "$customer" -s "$session"
"$TUIOS" new-window "$customer" -s "$session" --workspace "$next" --cwd "$HOMES/$customer" \
  -- distrobox enter "$customer"
"$TUIOS" select-workspace "$next" -s "$session"
