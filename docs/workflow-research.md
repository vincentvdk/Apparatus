# Workflow Research: Custom Silverblue/Bluefin Builds

Research on how other projects manage custom builds of Fedora Silverblue/Bluefin.

## Current Apparatus Workflows

| Feature | Status |
|---------|--------|
| Automated image builds | ✅ Daily schedule + push + PR |
| GHCR publishing | ✅ |
| Cosign signing (key-based) | ✅ |
| OCI metadata/labels | ✅ |
| Timestamp + SHA tagging | ✅ |
| ISO/disk image generation | ❌ Missing |
| Keyless signing (OIDC) | ❌ Missing |
| Justfile automation | ❌ Missing |
| ArtifactHub listing | ❌ Missing |
| Matrix builds (variants) | ❌ Missing |

---

## Reference Projects

| Project | Notable Features |
|---------|------------------|
| [ublue-os/image-template](https://github.com/ublue-os/image-template) | Official template with `build-disk.yml`, Justfile, S3 uploads, ArtifactHub metadata |
| [stephenreynolds/ublue-custom](https://github.com/stephenreynolds/ublue-custom) | `packages.json` manifest, Flatpak definitions, structured `system_files/` |
| [kleinbem/kleinbem-ublue](https://github.com/kleinbem/kleinbem-ublue) | Keyless signing via GitHub OIDC, disk_config/, Justfile with VM recipes |
| [ChrisLAS/bluenix](https://github.com/ChrisLAS/bluenix) | Community-focused, based on official template |

---

## Suggested Workflows

### 1. ISO/Disk Image Generation (High Value)

Add a `build-disk.yml` workflow using bootc-image-builder or Titanoboa:

```yaml
# .github/workflows/build-disk.yml
name: Build Disk Images
on:
  workflow_dispatch:
  workflow_run:
    workflows: ["build-apparatus-os"]
    types: [completed]

jobs:
  build-iso:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: osbuild/bootc-image-builder-action@v1
        with:
          image: ghcr.io/${{ github.repository_owner }}/apparatus-os:latest
          types: iso
          config-file: disk_config/iso.toml
      - uses: actions/upload-artifact@v4
        with:
          name: apparatus-iso
          path: "*.iso"
```

Requires a `disk_config/iso.toml` file for bootc-image-builder configuration.

**Resources:**
- [osbuild/bootc-image-builder-action](https://github.com/osbuild/bootc-image-builder-action) - Upstream ISO/disk builder
- [ublue-os/titanoboa](https://github.com/ublue-os/titanoboa) - Live ISO generation for bootc

---

### 2. Keyless Signing (OIDC) - Security Improvement

Migrate from stored private key to keyless signing (no secrets to manage):

```yaml
- name: Sign container image (keyless)
  if: github.event_name != 'pull_request'
  run: |
    cosign sign -y \
      ${{ steps.registry_case.outputs.lowercase }}/${{ steps.build_image.outputs.image }}@${{ steps.push.outputs.digest }}
  env:
    COSIGN_EXPERIMENTAL: 1
```

Uses GitHub OIDC and records signatures in Sigstore's transparency log.

---

### 3. Justfile for Local Development

```just
# Justfile
default:
  @just --list

# Build OS image locally
build-os:
  podman build -f os/Containerfile -t apparatus-os .

# Build distrobox locally
build-box:
  podman build -f distrobox/Containerfile -t apparatus-box .

# Run OS image in VM (requires qemu)
run-vm:
  podman run --rm -it --privileged \
    -v ./output:/output \
    quay.io/centos-bootc/bootc-image-builder:latest \
    --type qcow2 ghcr.io/$USER/apparatus-os:latest
  qemu-system-x86_64 -m 4G -enable-kvm ./output/qcow2/disk.qcow2

# Lint Containerfiles
lint:
  hadolint os/Containerfile distrobox/Containerfile
```

---

### 4. Matrix Builds for Variants

Build multiple image variants (e.g., with/without NVIDIA, different DEs):

```yaml
jobs:
  build:
    strategy:
      matrix:
        variant: [base, nvidia, hyprland]
    steps:
      - name: Build Image
        uses: redhat-actions/buildah-build@v2
        with:
          build-args: |
            VARIANT=${{ matrix.variant }}
```

---

### 5. ArtifactHub Metadata

Add `artifacthub-repo.yml` for discoverability:

```yaml
# artifacthub-repo.yml
repositoryID: <generate-uuid>
owners:
  - name: your-name
    email: your-email
```

---

### 6. Packages Manifest (Better Organization)

Use a `packages.json` to declaratively manage RPM packages:

```json
{
  "packages": {
    "install": ["neovim", "zsh", "tmux", "ripgrep"],
    "remove": ["firefox"]
  }
}
```

Then parse it in the Containerfile or build script.

---

### 7. Scheduled Upstream Sync Check

```yaml
name: Check Upstream Updates
on:
  schedule:
    - cron: '0 6 * * *'
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - name: Check for new base image
        run: |
          skopeo inspect docker://ghcr.io/ublue-os/silverblue-main:latest | \
            jq -r '.Digest' > /tmp/upstream
          # Compare with last known digest and notify if changed
```

---

## Priority Recommendations

| Priority | Workflow | Benefit |
|----------|----------|---------|
| High | ISO generation | Users can install directly without rebasing |
| High | Justfile | Local dev/test without pushing to CI |
| Medium | Keyless signing | No secrets management, better security |
| Medium | ArtifactHub | Discoverability in ublue ecosystem |
| Low | Matrix builds | Only if you need multiple variants |

---

## Sources

- [ublue-os/image-template](https://github.com/ublue-os/image-template) - Official Universal Blue template
- [osbuild/bootc-image-builder-action](https://github.com/osbuild/bootc-image-builder-action) - Upstream ISO/disk builder
- [ublue-os/titanoboa](https://github.com/ublue-os/titanoboa) - Live ISO generation for bootc
- [kleinbem/kleinbem-ublue](https://github.com/kleinbem/kleinbem-ublue) - Example with keyless signing
- [stephenreynolds/ublue-custom](https://github.com/stephenreynolds/ublue-custom) - Bluefin derivative example
- [Building your own custom Fedora Silverblue image](https://www.ypsidanger.com/building-your-own-fedora-silverblue-image/)
