# ISSUES — to file upstream later

## 1. Nested-attach guard false-positives on host ptys that collide with container ptys (dev_t not unique across devpts instances)

**Repo:** `Gaurav-Gosain/tuios`
**Severity:** medium (blocks plain `tuios` on the host in common setups with distrobox/podman containers sharing a tuios daemon)

### Summary

Running plain `tuios` (unnamed attach) from a normal host terminal is refused with:

```
ERROR
You are inside session "DISTRO-VTECH". Other sessions: HOST. To show one of them in this pane, run 'tuios attach NAME'. To start a new session, run 'tuios new NAME'.
```

even though the shell is provably not inside any tuios pane: clean process ancestry
(`bash -> kitty -> Hyprland -> systemd`), no `TUIOS_*` variables in `/proc/$$/environ`,
and the host is not a container.

### Root cause

`Daemon.paneSession()` (`internal/session/nested_attach.go`) builds a map of
`tty_nr -> session` from each pane shell's `/proc/<pid>/stat` field 7, then places any
attaching client whose own controlling terminal has the same `tty_nr`
(`paneOriginTTY`).

For Unix98 ptys, `tty_nr` is just `major*256 + minor` (e.g. pts/3 -> 34819). That value
is **not unique across devpts instances**: a container (distrobox/podman) with its own
devpts mount has its own pts/0, pts/1, ... with the same dev_t numbering as the host.
A tuios daemon whose sessions include panes running inside a container therefore maps
container pts numbers onto host pts numbers.

### Repro

1. On a host, start a tuios daemon (session `HOST`).
2. Inside a distrobox container sharing the same home (and thus the same tuios socket),
   run `tuios new DISTRO-VTECH` and create a few windows. The pane shells of that
   session are container processes whose ptys live in the container's devpts.
3. On the host, open a terminal that happens to get pts index *n* (e.g. pts/3), where
   a DISTRO-VTECH pane shell also reports pts *n*:

   ```
   # host shell:
   $ tty
   /dev/pts/3
   $ cat /proc/$$/stat | sed 's/.*) //' | awk '{print $5}'   # tty_nr
   34819

   # pane shell of DISTRO-VTECH (visible from host /proc):
   $ cat /proc/<pane-shell-pid>/stat | sed 's/.*) //' | awk '{print $5}'
   34819        # same value: container devpts pts/3 collides with host devpts pts/3
   ```

4. On the host terminal, run plain `tuios`:

   ```
   ERROR
   You are inside session "DISTRO-VTECH". Other sessions: HOST. ...
   ```

Observed pane shells of the container session: tty_nr 34819..34825 (pts/3..pts/9).

### Expected behavior

A client on a host pty should never be classified as "inside" a session just because a
container pane happens to have the same pts minor number in a different devpts
instance. Plain `tuios` on the host should attach to the default/last session.

### Actual behavior

The client is placed in the container session via `paneOriginTTY`, and unnamed attaches
from a placed client are always refused (`daemon_handlers.go`: "an unnamed attach from
a pane is usually the pane's own session").

### Workarounds (confirmed against the code)

- `tuios attach HOST` — named attach to a different session passes the `shownInside`
  chain check.
- `tuios new NAME` — creating a session is unaffected.
- `tuios attach --force` or `TUIOS_ALLOW_NESTED=1 tuios` — bypasses the check.
- Setting `startup.session = "..."` in config.toml makes plain `tuios` a *named* attach,
  which is allowed.

### Suggested fix

In `paneSession()` (and anywhere `tty_nr` is used to identify a terminal), do not compare
raw `tty_nr` alone. Distinguish devpts instances, e.g. by comparing the mount ID of the
pty: `statx(..., STATX_MNT_ID)` on `/proc/<pid>/fd/0`, or the `mnt_id` from
`/proc/<pid>/fdinfo/0`, or by resolving the fd symlink and comparing the devpts mount
point. Container devpts and host devpts must not be treated as the same terminal.

### Environment

- tuios: (fill in `tuios version`)
- Host: Fedora (Apparatus OS), Hyprland, kitty terminal
- Containers: rootless podman via distrobox, sharing the host home (shared tuios daemon
  socket via `~/.local/share/tuios` / `TUIOS_SOCKET`)
- Both sessions (`HOST`, `DISTRO-VTECH`) visible from host and container — shared daemon
