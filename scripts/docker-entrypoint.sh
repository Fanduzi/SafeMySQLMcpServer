#!/bin/sh
# SafeMySQLMcpServer — Docker entrypoint with config validation
set -e

# Validate JWT_SECRET
if [ -z "$JWT_SECRET" ]; then
    echo "ERROR: JWT_SECRET is not set."
    echo "  Run 'make init' to configure, or set it in .env"
    exit 1
fi

if [ ${#JWT_SECRET} -lt 32 ]; then
    echo "ERROR: JWT_SECRET must be at least 32 characters (got ${#JWT_SECRET})."
    exit 1
fi

# Validate config file exists
if [ ! -f "$CONFIG_PATH" ]; then
    echo "ERROR: Config file not found at $CONFIG_PATH"
    echo "  Mount your config directory: -v ./config:/app/config:ro"
    exit 1
fi

# Create logs directory if writable
mkdir -p /app/logs 2>/dev/null || true

echo "Starting SafeMySQLMcpServer..."
echo "  Config: $CONFIG_PATH"
echo "  Poll interval: ${CONFIG_POLL_INTERVAL:-fsnotify only}"

exec /app/server -config "$CONFIG_PATH" ${CONFIG_POLL_INTERVAL:+-poll-interval "$CONFIG_POLL_INTERVAL"}
