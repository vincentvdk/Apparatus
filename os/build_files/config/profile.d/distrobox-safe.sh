# Safe distrobox wrapper - prevents accidental container creation
# Wraps 'distrobox enter' to only allow entering existing containers

distrobox() {
    if [ "$1" = "enter" ] && [ -n "$2" ]; then
        local container_name="$2"
        # Check if container exists
        if ! command distrobox list --no-color 2>/dev/null | tail -n +2 | awk '{print $3}' | grep -qx "$container_name"; then
            echo "Error: Container '$container_name' does not exist." >&2
            echo "" >&2
            echo "Available containers:" >&2
            command distrobox list 2>/dev/null
            echo "" >&2
            echo "To create a new container, use: distrobox create $container_name" >&2
            return 1
        fi
    fi
    command distrobox "$@"
}
