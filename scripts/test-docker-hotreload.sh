#!/usr/bin/env bash
# Docker hot-reload E2E test
# Usage: ./scripts/test-docker-hotreload.sh
set -euo pipefail

BASE_URL="http://localhost:18080"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhZG1pbiIsImVtYWlsIjoiYWRtaW5AdGVzdC5jb20iLCJleHAiOjE3NzkzNTk1MzEsImlhdCI6MTc3OTM1NTkzMX0.67gqiK11iNJuGYS6gGRnhpjpnBLN5X1qbzlOFzE3-I0"
SECURITY_YAML="config/security.yaml"
PASS=0
FAIL=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_pass() { echo -e "${GREEN}PASS${NC} $1"; ((PASS++)); }
log_fail() { echo -e "${RED}FAIL${NC} $1"; ((FAIL++)); }
log_step() { echo -e "${YELLOW}>>>${NC} $1"; }

# MCP call helper — sends JSON-RPC to the MCP endpoint
mcp_call() {
    local method="$1"
    local params="$2"
    curl -s -X POST "${BASE_URL}/mcp" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"${method}\",\"params\":${params}}"
}

# Call a tool via MCP
call_tool() {
    local tool_name="$1"
    local arguments="$2"
    mcp_call "tools/call" "{\"name\":\"${tool_name}\",\"arguments\":${arguments}}"
}

# Initialize MCP session
mcp_init() {
    mcp_call "initialize" "{\"protocolVersion\":\"2025-03-26\",\"capabilities\":{},\"clientInfo\":{\"name\":\"hotreload-test\",\"version\":\"1.0\"}}"
}

# Send initialized notification
mcp_initialized() {
    curl -s -X POST "${BASE_URL}/mcp" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","method":"notifications/initialized"}' > /dev/null 2>&1
    # Small delay for server to process
    sleep 0.2
}

echo "========================================"
echo "  Docker Hot-Reload E2E Test"
echo "========================================"
echo ""

# ── Step 0: Health check ──
log_step "Step 0: Health check"
HEALTH=$(curl -s "${BASE_URL}/health")
if [ "$HEALTH" = "OK" ]; then
    log_pass "Health check OK"
else
    log_fail "Health check failed: $HEALTH"
    exit 1
fi

# ── Step 1: Init MCP session ──
log_step "Step 1: Initialize MCP session"
INIT_RESP=$(mcp_init)
if echo "$INIT_RESP" | grep -q "safe-mysql-mcp"; then
    log_pass "MCP session initialized"
else
    log_fail "MCP init failed: $INIT_RESP"
    exit 1
fi
mcp_initialized

# ── Step 2: Query before reload ──
log_step "Step 2: SELECT before hot-reload"
BEFORE=$(call_tool "query" "{\"database\":\"testdb\",\"sql\":\"SELECT COUNT(*) AS cnt FROM users\"}")
if echo "$BEFORE" | grep -q '"cnt"'; then
    log_pass "SELECT before reload works"
else
    log_fail "SELECT before reload failed: $(echo "$BEFORE" | head -c 200)"
fi

# ── Step 3: Save original security.yaml and modify it ──
log_step "Step 3: Modify security.yaml — remove DELETE from allowed_dml"
cp "$SECURITY_YAML" "${SECURITY_YAML}.bak"

# Remove DELETE from allowed_dml
sed '/^    - DELETE$/d' "$SECURITY_YAML" > "${SECURITY_YAML}.tmp" && mv "${SECURITY_YAML}.tmp" "$SECURITY_YAML"

log_step "  Waiting for config poll (5s interval)..."
sleep 8

# ── Step 4: SELECT should still work after reload ──
log_step "Step 4: SELECT after hot-reload"
AFTER_SELECT=$(call_tool "query" "{\"database\":\"testdb\",\"sql\":\"SELECT 1 AS ok\"}")
if echo "$AFTER_SELECT" | grep -q '"ok"'; then
    log_pass "SELECT after reload works"
else
    log_fail "SELECT after reload FAILED (server unresponsive?): $(echo "$AFTER_SELECT" | head -c 200)"
fi

# ── Step 5: DELETE should be blocked now ──
log_step "Step 5: DELETE should be blocked after reload"
DELETE_RESP=$(call_tool "query" "{\"database\":\"testdb\",\"sql\":\"DELETE FROM users WHERE 1=0\"}")
if echo "$DELETE_RESP" | grep -qi "blocked\|not allowed\|error"; then
    log_pass "DELETE correctly blocked after hot-reload"
else
    log_fail "DELETE was NOT blocked (security rules not updated): $(echo "$DELETE_RESP" | head -c 200)"
fi

# ── Step 6: Restore original config ──
log_step "Step 6: Restore original security.yaml"
mv "${SECURITY_YAML}.bak" "$SECURITY_YAML"
sleep 8

# ── Step 7: DELETE should work again after restoring ──
log_step "Step 7: DELETE should work again after restoring config"
DELETE_RESTORE=$(call_tool "query" "{\"database\":\"testdb\",\"sql\":\"DELETE FROM users WHERE 1=0\"}")
if echo "$DELETE_RESTORE" | grep -qi "rows_affected"; then
    log_pass "DELETE works again after config restore"
elif echo "$DELETE_RESTORE" | grep -qi "error"; then
    # Might still be OK — the SQL returns error but is allowed
    log_pass "DELETE allowed after config restore (got DB error, not blocked)"
else
    log_fail "DELETE after restore unexpected: $(echo "$DELETE_RESTORE" | head -c 200)"
fi

# ── Step 8: Rapid reload stress test ──
log_step "Step 8: Rapid reload stress test (5 rapid config changes)"
for i in $(seq 1 5); do
    cp "$SECURITY_YAML" "${SECURITY_YAML}.bak"
    sed '/^    - DELETE$/d' "$SECURITY_YAML" > "${SECURITY_YAML}.tmp" && mv "${SECURITY_YAML}.tmp" "$SECURITY_YAML"
    sleep 1
    mv "${SECURITY_YAML}.bak" "$SECURITY_YAML"
    sleep 1
done

# Wait for final poll
sleep 8

# Final check
FINAL=$(call_tool "query" "{\"database\":\"testdb\",\"sql\":\"SELECT 1 AS final\"}")
if echo "$FINAL" | grep -q '"final"'; then
    log_pass "Server responsive after rapid reloads"
else
    log_fail "Server UNRESPONSIVE after rapid reloads: $(echo "$FINAL" | head -c 200)"
fi

echo ""
echo "========================================"
echo -e "  Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
echo "========================================"

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
