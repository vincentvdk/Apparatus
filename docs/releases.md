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
