#!/bin/sh
set -e

# Rexec Entrypoint Script
# This script handles the setup required for connecting to a remote Docker daemon.
# Supports: TCP (plain), TCP with TLS, and SSH connections.

echo "=============================================="
echo "🚀 Rexec Entrypoint Starting"
echo "=============================================="
echo "Date: $(date)"
echo "User: $(whoami) (UID: $(id -u))"
echo "HOME: $HOME"
echo "PWD: $(pwd)"
echo ""
echo "Environment variables:"
echo "  DOCKER_HOST=${DOCKER_HOST:-<not set>}"
echo "  DOCKER_TLS_VERIFY=${DOCKER_TLS_VERIFY:-<not set>}"
echo "  DOCKER_CERT_PATH=${DOCKER_CERT_PATH:-<not set>}"
echo "  DOCKER_CA_CERT length: ${#DOCKER_CA_CERT} chars"
echo "  DOCKER_CLIENT_CERT length: ${#DOCKER_CLIENT_CERT} chars"
echo "  DOCKER_CLIENT_KEY length: ${#DOCKER_CLIENT_KEY} chars"
echo ""

# ============================================================================
# Docker Connection Configuration
# ============================================================================
#
# Environment Variables:
#
# DOCKER_HOST - The Docker daemon endpoint
#   Examples:
#     - unix:///var/run/docker.sock     (local socket, default)
#     - tcp://docker-host:2375          (remote, no TLS)
#     - tcp://docker-host:2376          (remote, with TLS)
#     - ssh://user@docker-host          (remote via SSH)
#
# For TLS connections (tcp:// with port 2376):
#   DOCKER_TLS_VERIFY=1                 - Enable TLS verification
#   DOCKER_CERT_PATH=/path/to/certs     - Path to TLS certificates
#
#   Or provide certificates directly via environment:
#   DOCKER_CA_CERT     - CA certificate content (PEM)
#   DOCKER_CLIENT_CERT - Client certificate content (PEM)
#   DOCKER_CLIENT_KEY  - Client private key content (PEM)
#
# For SSH connections (ssh://):
#   SSH_PRIVATE_KEY    - SSH private key content
#
# ============================================================================

# Function to write PEM certificate content
# Handles the case where PaaS platforms replace newlines with spaces
write_pem() {
    local content="$1"
    local file="$2"
    local desc="$3"

    # Check if content is empty
    if [ -z "$content" ]; then
        echo "  ⚠ Warning: Empty content for $desc"
        return 1
    fi

    # Check if content already has proper newlines (more than 2 lines)
    local line_count
    line_count=$(printf '%s\n' "$content" | wc -l)

    if [ "$line_count" -gt 5 ]; then
        # Already has newlines, write directly
        printf '%s\n' "$content" > "$file"
    else
        # Content is on a single line with spaces instead of newlines
        # We need to carefully reconstruct the PEM format

        # Strategy: Use sed to:
        # 1. Add newline after "-----BEGIN XXXXX-----"
        # 2. Add newline before "-----END XXXXX-----"
        # 3. Split the base64 content (replace spaces with newlines, but not in markers)

        printf '%s' "$content" | sed -E '
            # Add newline after BEGIN marker
            s/(-----BEGIN [A-Z ]+-----) /\1\n/g
            # Add newline before END marker
            s/ (-----END [A-Z ]+-----)/\n\1/g
        ' | awk '
        {
            # For lines that are not BEGIN or END markers, split on spaces
            if (/^-----BEGIN/ || /^-----END/) {
                print
            } else {
                # Replace spaces with newlines for base64 content
                gsub(/ /, "\n")
                print
            }
        }' > "$file"
    fi

    # Verify the file has proper PEM structure
    if grep -q "^-----BEGIN" "$file" && grep -q "^-----END" "$file"; then
        local lines
        lines=$(wc -l < "$file")
        echo "  ✓ $desc configured ($lines lines)"
        return 0
    else
        echo "  ⚠ Warning: $desc may be malformed"
        echo "    First line: $(head -1 "$file")"
        echo "    Last line: $(tail -1 "$file")"
        return 1
    fi
}

# Setup TLS certificates if provided via environment variables
if [ -n "$DOCKER_CA_CERT" ] || [ -n "$DOCKER_CLIENT_CERT" ] || [ -n "$DOCKER_CLIENT_KEY" ]; then
    echo "📜 Configuring Docker TLS certificates..."

    # Create cert directory
    CERT_DIR="${DOCKER_CERT_PATH:-$HOME/.docker}"
    mkdir -p "$CERT_DIR"
    chmod 700 "$CERT_DIR"

    # Write CA certificate
    if [ -n "$DOCKER_CA_CERT" ]; then
        write_pem "$DOCKER_CA_CERT" "$CERT_DIR/ca.pem" "CA certificate"
        chmod 644 "$CERT_DIR/ca.pem"
    fi

    # Write client certificate
    if [ -n "$DOCKER_CLIENT_CERT" ]; then
        write_pem "$DOCKER_CLIENT_CERT" "$CERT_DIR/cert.pem" "Client certificate"
        chmod 644 "$CERT_DIR/cert.pem"
    fi

    # Write client key
    if [ -n "$DOCKER_CLIENT_KEY" ]; then
        write_pem "$DOCKER_CLIENT_KEY" "$CERT_DIR/key.pem" "Client key"
        chmod 600 "$CERT_DIR/key.pem"
    fi

    # Set cert path if not already set
    if [ -z "$DOCKER_CERT_PATH" ]; then
        export DOCKER_CERT_PATH="$CERT_DIR"
    fi

    # Enable TLS verification by default when certs are provided
    if [ -z "$DOCKER_TLS_VERIFY" ]; then
        export DOCKER_TLS_VERIFY=1
    fi

    # Debug: show first and last line of CA cert to verify format
    if [ -f "$CERT_DIR/ca.pem" ]; then
        echo "  📄 CA cert preview:"
        echo "      First line: $(head -1 "$CERT_DIR/ca.pem")"
        echo "      Last line:  $(tail -1 "$CERT_DIR/ca.pem")"
    fi

    echo "✅ Docker TLS configuration complete"
    echo "   DOCKER_CERT_PATH=$DOCKER_CERT_PATH"
fi

# Setup SSH for remote Docker connection via SSH
if [ -n "$SSH_PRIVATE_KEY" ]; then
    echo "🔑 Configuring SSH for remote Docker connection..."

    # Ensure .ssh directory exists
    mkdir -p "$HOME/.ssh"
    chmod 700 "$HOME/.ssh"

    # Write the private key, handling escaped newlines
    write_pem "$SSH_PRIVATE_KEY" "$HOME/.ssh/id_rsa" "SSH private key"
    chmod 600 "$HOME/.ssh/id_rsa"

    # Configure SSH to disable strict host key checking
    cat > "$HOME/.ssh/config" <<EOF
Host *
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
    LogLevel ERROR
    ServerAliveInterval 60
    ServerAliveCountMax 3
EOF
    chmod 600 "$HOME/.ssh/config"

    echo "✅ SSH configuration complete"
fi

# Log Docker connection info
if [ -n "$DOCKER_HOST" ]; then
    echo "🐳 Docker Host: $DOCKER_HOST"
    if [ -n "$DOCKER_TLS_VERIFY" ] && [ "$DOCKER_TLS_VERIFY" = "1" ]; then
        echo "🔒 TLS verification enabled"
        echo "   Cert path: ${DOCKER_CERT_PATH:-$HOME/.docker}"
    fi
else
    echo "🐳 Docker Host: unix:///var/run/docker.sock (default)"
fi

echo ""
echo "Starting application..."
echo "=============================================="

# Execute the main command (usually "rexec")
exec "$@"
