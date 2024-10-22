#!/bin/bash

set -ouex pipefail

RELEASE="$(rpm -E %fedora)"

### Install packages
rpm-ostree install distrobox docker libvirt-daemon-kvm qemu-kvm virt-manager

#### Enabling Systemd Unit File
systemctl enable podman.socket
systemctl enable docker


#### System Configuration
# -- Fonts
curl -OL --output-dir /tmp https://github.com/source-foundry/Hack/releases/download/v3.003/Hack-v3.003-ttf.zip
unzip -d /tmp /tmp/Hack-v3.003-ttf.zip
cp /tmp/ttf/* /usr/share/fonts/
ls -lah /tmp/ttf
# -- Configure distrobox


# -- Bootstrap

# -- Desktop Environment
