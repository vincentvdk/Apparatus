# Versioning & Releases

Apparatus uses semantic versioning. The source of truth is a **git tag on `main`** — no separate version file to keep in sync.

## Scheme

- Releases are tagged `vX.Y.Z` (e.g. `v1.2.0`) on `main`.
- Container images get the same version as a tag, without the `v` prefix:
  - `ghcr.io/vincentvdk/apparatus-os:X.Y.Z`
  - `ghcr.io/vincentvdk/apparatus-box:X.Y.Z`
  - plus `latest` (follows the newest build on `main` or the newest release) and the short commit SHA.
- Non-release builds on `main` are tagged `dev-<sha>` + `yyyymmdd`; pull-request builds get `pr-<n>` and `<sha>` only (never `latest`).
- The version is embedded in the image:
  - OCI label `org.opencontainers.image.version`
  - `/usr/share/apparatus/image-info.json` → `image-tag`
  - `butler version`
- Local builds use `git describe --tags --always --dirty` (`just show-version`).

## Cutting a release

```bash
git tag v1.2.0
git push origin v1.2.0
```

Pushing the tag triggers the `build-os-bootc` and `build-box` workflows, which build and push the `X.Y.Z` images to GHCR.

## Using a specific version

```bash
# OS image (pin instead of :latest)
sudo bootc switch ostree-unverified-registry:ghcr.io/vincentvdk/apparatus-os:1.2.0

# Box image: pin in ~/.config/distrobox/distrobox.ini
# image = ghcr.io/vincentvdk/apparatus-box:1.2.0
```

## Day-to-day updates

Track `latest` and update atomically:

```bash
sudo bootc upgrade
systemctl reboot
```

Check what is staged with `bootc status`; roll back with `bootc rollback` if needed.

## Managing old and pinned deployments

`bootc status` shows the `booted`, `rollback`, `staged` and `otherDeployments` entries. Entries that were pinned on purpose show `pinned: true` (with the OSTree checksum) — use the YAML/JSON output to inspect them:

```bash
bootc status
bootc status --format yaml    # or --format json
```

To remove an old/pinned deployment, match its **checksum** to an index in `ostree admin status` (do not assume the ordering matches `bootc status`), then undeploy by index — no unpinning needed, and the booted deployment (index 0, marked `*`) cannot be removed:

```bash
ostree admin status                 # list deployments with indexes + checksums
sudo ostree admin undeploy 2        # INDEX, not the checksum ("invalid index" otherwise)
sudo ostree admin cleanup           # or drop all non-booted leftovers at once
```

If a deployment ever needs unpinning explicitly: `sudo ostree admin pin --unpin <INDEX>`.

Note there are two separate "pins":

1. **Deployment pin** (above) — keeps an old deployment around in the OSTree sysroot.
2. **Boot-menu default** — systemd-boot persists the selected default entry in the `LoaderEntryDefault` EFI variable, which overrides automatic selection. Clear or repoint it with:
   ```bash
   bootctl list                                   # see which entry is the default
   sudo bootctl set-default @current              # default to the currently booted entry
   # or remove the pin entirely:
   sudo rm /sys/firmware/efi/efivars/LoaderEntryDefault-4a67b082-0a4c-41cf-b6c7-440b29bb8c4f
   ```

Kernel arguments are inherited from the booted deployment on every `bootc upgrade`, so once a bad karg (e.g. `iommu=off`, see [Troubleshooting](troubleshooting.md#network-not-working)) has been removed from the booted entry's BLS config, it will not come back on new deployments.
