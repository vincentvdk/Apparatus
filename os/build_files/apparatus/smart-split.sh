#!/bin/bash
# Smart split for kitty terminal
# Detects if inside distrobox and enters the same container in new split

if [ -n "$CONTAINER_ID" ]; then
    exec distrobox enter "$CONTAINER_ID"
else
    exec "$SHELL"
fi
