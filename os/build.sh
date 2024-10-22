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
unzip -d /tmp/hack-font /tmp/Hack-v3.003-ttf.zip
cp -r /tmp/hack-font /usr/share/fonts/
fc-cache -f -v


# -- Configure distrobox


# -- Bootstrap

# -- Desktop Environment
