#!/bin/bash
#
# Build new container image
echo "Building new container image.."
podman build --no-cache -t ghcr.io/vincentvdk/apparatus -f distrobox/Containerfile .

#echo "### Stopping toolbox-1.."
#distrobox stop toolbox-1
#echo "### Deleting toolbox-1.."
#distrobox rm -f toolbox-1
#echo "### Creating new toolbox.."
#distrobox assemble create --replace --file ./distrobox.ini
