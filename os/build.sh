#!/bin/bash

set -ouex pipefail

RELEASE="$(rpm -E %fedora)"

### Install packages
rpm-ostree install dnf-plugins-core
rpm-ostree install distrobox docker libvirt-daemon-kvm qemu-kvm virt-manager hack-fonts

#### Enabling Systemd Unit File
systemctl enable podman.socket
systemctl enable docker


#### System Configuration

# -- Configure distrobox


# -- Bootstrap

# -- Desktop Environment
