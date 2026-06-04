# Apparatus OS

An opinionated, immutable Fedora-based desktop environment engineered for SREs, DevOps, and Systems Engineers. 

Designed for stability, reproducibility, and minimal maintenance, Apparatus uses `bootc` and `ostree` to provide a hardened, predictable workstation that stays out of your way.

## 🛠 Architecture & Philosophy

Apparatus is built on the principle of **minimal drift**. By leveraging Fedora's immutable architecture, it provides a consistent environment that doesn't rot over time.

- **Immutable Core**: Powered by `bootc`, ensuring that the base system is cryptographically verifiable and resistant to configuration drift.
- **Zero-Dependency Build**: The build engine is engineered to be entirely self-contained, using simple shell primitives to parse versions from `apparatus.env` without requiring `yq` or `python` in the build stage.
- **Single Source of Truth**: All critical component versions (OS, Hyprland, Tooling) are managed via a single `.env` file, facilitating reproducible image builds.
- **Container-First Workflow**: Development tools, runtimes, and dependencies do not touch the host system; they live in isolated, versioned **Distrobox** containers.

## 🚀 Deployment & Rebase

To deploy Apparatus on a fresh Fedora Silverblue installation, follow the rebase workflow.

### Step 1: Initial Rebase (Unverified)
For testing and initial deployment, use the unverified registry path:

```shell
# Define the image path
IMAGE_PATH=ghcr.io/vincentvdk/apparatus-os

# Rebase using bootc
sudo bootc rebase ostree-unverified-registry:$IMAGE_PATH:latest

# Reboot to apply changes
systemctl reboot
```

### Step 2: Production Rebase (Signed)
Once you have validated the image, switch to the trusted, signed production image:

```shell
# Define the image path
IMAGE_PATH=ghcr.io/vincentvdk/apparatus-os

# Rebase using bootc
sudo bootc rebase ostree-image-signed:docker://$IMAGE_PATH:latest

# Reboot to complete the transition
systemctl reboot
```

## ⚙️ Operational Workflow

### Workspace Initialization
After the first boot, use the `butler` utility to initialize user-level configurations (Flathub, user-specific symlinks, and profile defaults):

```bash
# Initialize user-level environment
butler init
```

### Environment Isolation (Distrobox)
All development work occurs within isolated containers. This ensures that the host OS remains clean, lightweight, and immutable.

```shell
# Create a dedicated environment for a specific stack (e.g., Go/Kubernetes)
distrobox create -i ghcr.io/vinintvdk/apparatus-box:latest -n k8s-dev

# Enter the environment to run workloads
distrobox enter k8s-dev
```

## 📦 Build Engineering

The build system is designed for high reliability in CI/CD pipelines.

- **Containerfile.bootc**: Uses a multi-stage build approach. A `delivery` stage prepares the build context (scripts, `.env` files) before being mounted via `--mount=type=bind` into the main `bootc` build stage.
- **Dependency-Free Parsing**: The `justfile` orchestrator extracts versions from `apparatus.env` using standard `grep`/`cut`, ensuring the build can run in any minimal container environment.
- **Automated Validation**: Every build concludes with `bootc container lint` and `ostree container commit` to ensure image integrity.

## 🔗 Links
- [Fedora Silverblue/Bootc Documentation](https://fedoraproject.org/atomic-desktops/silverblue/)
- [Hyprland Wiki](https://wiki.hypr.land/)
- [Universal Blue](https://universal-blue.org/)
