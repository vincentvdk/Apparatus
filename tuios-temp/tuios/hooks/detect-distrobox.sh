# Shared helper, sourced (not run directly) by smart-split-command.sh and
# smart-newwindow-command.sh.
#
# detect_container_id <pane_id>: prints the distrobox name a pane is
# actually inside, by reading it off `podman exec`'s own argv
# (--env=CONTAINER_ID=<name>, which distrobox always passes) - not by
# guessing from the pane's cwd. The cwd approach (.../distrobox-
# homes/<name>) only works for distroboxes with a dedicated home dir;
# one sharing the host's home never matches that pattern, even though
# CONTAINER_ID is passed the same way either way.
#
# Why not walk the process tree down into the container to read
# CONTAINER_ID as an env var there (the first version of this did
# that): conmon and everything inside the container get reparented to
# systemd during container startup - they detach from whatever invoked
# them on purpose, so they are never descendants of the pane's shell
# in the process tree. Found by tracing a real pane's tree by hand:
# `shell -> distrobox -> podman` is a normal parent/child chain, but
# `conmon` (and the actual container shell under it) show up with
# PPID 1 (reparented), nowhere near the pane's own subtree. podman's
# own argv still has the container name though, and podman *is* a
# normal child - 1 hop down if the pane's root process is `distrobox`
# itself (new-window --direct-exec), 2 hops down if it's a plain shell
# the name was typed into by hand (shell -> distrobox -> podman).
#
# Only ever looks at the pane's own small subtree (its pid, children,
# grandchildren) - not a system-wide process dump. The first version
# did dump the whole table and took ~0.6s on this machine (hundreds of
# processes); this one is a handful of `ps`/proc reads.
detect_container_id() {
  local pane_id="$1"
  local shell_pgid
  shell_pgid=$("$TUIOS" get-window "$pane_id" --json 2>/dev/null | jq -r '.shell_pgid // empty')
  [ -z "$shell_pgid" ] && return 1

  local children grandchildren
  children=$(ps -o pid= --ppid "$shell_pgid" 2>/dev/null)
  grandchildren=""
  local c
  for c in $children; do
    grandchildren+=" $(ps -o pid= --ppid "$c" 2>/dev/null)"
  done

  local pid cid
  for pid in "$shell_pgid" $children $grandchildren; do
    cid=$(tr '\0' ' ' < "/proc/$pid/cmdline" 2>/dev/null | grep -oP -- '--env=CONTAINER_ID=\S+' | head -1)
    cid="${cid#--env=CONTAINER_ID=}"
    if [ -n "$cid" ]; then
      printf '%s' "$cid"
      return 0
    fi
  done
  return 1
}
