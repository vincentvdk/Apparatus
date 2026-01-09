# Running Podman Inside Distrobox

## Problem

Running podman inside a distrobox container fails with user namespace errors:

```
ERRO[0000] running `/usr/bin/newuidmap 4082174 0 1000 1 1 10000 55537`: newuidmap: write to uid_map failed: Operation not permitted
ERRO[0000] invalid internal status, try resetting the pause process with "podman system migrate": cannot set up namespace using "/usr/bin/newuidmap": should have setuid or have filecaps setuid: exit status 1
```

This happens because rootless podman requires nested user namespaces, which are restricted inside containers.

## Solution

Instead of running podman inside the container, use `distrobox-host-exec` to forward podman commands to the host.

Create a symlink that makes distrobox-host-exec masquerade as podman:

```bash
sudo ln -s /usr/bin/distrobox-host-exec /usr/local/bin/podman
```

The `distrobox-host-exec` binary detects it was called as "podman" (based on the symlink name) and forwards the command to the host's podman.

## Additional Tools

The same approach works for other container tools:

```bash
sudo ln -s /usr/bin/distrobox-host-exec /usr/local/bin/docker
sudo ln -s /usr/bin/distrobox-host-exec /usr/local/bin/docker-compose
```

## Verification

```bash
# Check the symlink exists
ls -la /usr/local/bin/podman
# Should show: /usr/local/bin/podman -> /usr/bin/distrobox-host-exec

# Test it works
podman --version
podman ps
```
