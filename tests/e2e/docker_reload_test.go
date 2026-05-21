//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/auth"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func dockerToken(t *testing.T) string {
	t.Helper()
	secret := envOr("JWT_SECRET", "e2e-docker-hot-reload-test-secret-at-least-32-characters-long")
	v := auth.NewValidator(secret)
	tok, err := v.GenerateToken("docker-test-user", "docker@test.com", time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return tok
}

// TestDocker_HotReload_SecurityRuleChange tests hot-reload against a running
// Docker container. Requires the dev stack: make dev
//
// Prerequisites:
//   - Docker containers running (make dev)
//   - config/ mounted as volume
//   - CONFIG_POLL_INTERVAL set (5s in .env)
func TestDocker_HotReload_SecurityRuleChange(t *testing.T) {
	baseURL := envOr("DOCKER_APP_URL", "http://localhost:18080")
	securityPath := envOr("SECURITY_YAML_PATH", "../../config/security.yaml")
	pollInterval := 8 * time.Second // CONFIG_POLL_INTERVAL=5s + buffer

	// Step 0: health check
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Skipf("Docker container not running at %s: %v", baseURL, err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Skipf("Docker health check failed: status %d", resp.StatusCode)
	}
	t.Log("Docker container healthy")

	// Step 1: generate token and connect MCP client
	token := dockerToken(t)

	client := mcpsdk.NewClient(&mcpsdk.Implementation{
		Name:    "docker-hotreload-test", Version: "1.0.0",
	}, nil)

	httpClient := &http.Client{
		Transport: &authTransport{base: http.DefaultTransport, token: token},
	}

	session, err := client.Connect(context.Background(), &mcpsdk.StreamableClientTransport{
		Endpoint:             baseURL + "/mcp",
		HTTPClient:           httpClient,
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("MCP connect: %v", err)
	}
	defer session.Close()

	dbName := "testdb"

	// Step 2: SELECT before reload — must work
	mustCallQuery(t, session, dbName, "SELECT 1 AS before_reload")
	t.Log("SELECT before reload: OK")

	// Step 3: DELETE before reload — must work (DELETE is in allowed_dml)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	delBefore, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name: "query",
		Arguments: map[string]any{
			"database": dbName,
			"sql":      "DELETE FROM products WHERE id = -999",
		},
	})
	if err != nil {
		t.Fatalf("DELETE before reload: %v", err)
	}
	if delBefore.IsError {
		t.Fatalf("DELETE before reload should succeed but got error: %v", delBefore.Content)
	}
	t.Log("DELETE before reload: allowed (as expected)")

	// Step 4: modify security.yaml — remove DELETE from allowed_dml
	original, err := os.ReadFile(securityPath)
	if err != nil {
		t.Fatalf("read security.yaml: %v", err)
	}

	modified := []byte{}
	for _, line := range splitLines(string(original)) {
		if line != "    - DELETE" {
			modified = append(modified, []byte(line+"\n")...)
		}
	}

	if err := os.WriteFile(securityPath, modified, 0644); err != nil {
		t.Fatalf("write modified security.yaml: %v", err)
	}
	t.Log("Modified security.yaml: removed DELETE from allowed_dml")

	// Ensure restore on exit
	defer func() {
		if err := os.WriteFile(securityPath, original, 0644); err != nil {
			t.Logf("WARNING: failed to restore security.yaml: %v", err)
		}
	}()

	// Step 5: wait for poll interval to pick up change
	t.Logf("Waiting %v for config poll...", pollInterval)
	time.Sleep(pollInterval)

	// Step 6: SELECT after reload — must still work
	mustCallQuery(t, session, dbName, "SELECT 2 AS after_reload")
	t.Log("SELECT after reload: OK")

	// Step 7: DELETE after reload — must be BLOCKED
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	delAfter, err := session.CallTool(ctx2, &mcpsdk.CallToolParams{
		Name: "query",
		Arguments: map[string]any{
			"database": dbName,
			"sql":      "DELETE FROM products WHERE id = -999",
		},
	})
	if err != nil {
		t.Fatalf("DELETE after reload call: %v", err)
	}

	if !delAfter.IsError {
		t.Fatal("DELETE should be BLOCKED after hot-reload but it succeeded — security rules NOT updated")
	}
	t.Logf("DELETE after reload: correctly blocked (%v)", delAfter.Content)
}

// TestDocker_HotReload_RapidReload rapidly changes config and verifies
// the server stays responsive throughout.
func TestDocker_HotReload_RapidReload(t *testing.T) {
	baseURL := envOr("DOCKER_APP_URL", "http://localhost:18080")
	securityPath := envOr("SECURITY_YAML_PATH", "../../config/security.yaml")

	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Skipf("Docker container not running at %s: %v", baseURL, err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Skipf("Docker health check failed: status %d", resp.StatusCode)
	}

	token := dockerToken(t)
	client := mcpsdk.NewClient(&mcpsdk.Implementation{
		Name: "docker-rapid-test", Version: "1.0.0",
	}, nil)
	httpClient := &http.Client{
		Transport: &authTransport{base: http.DefaultTransport, token: token},
	}
	session, err := client.Connect(context.Background(), &mcpsdk.StreamableClientTransport{
		Endpoint:             baseURL + "/mcp",
		HTTPClient:           httpClient,
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("MCP connect: %v", err)
	}
	defer session.Close()

	dbName := "testdb"
	mustCallQuery(t, session, dbName, "SELECT 1 AS warmup")

	// Save original
	original, err := os.ReadFile(securityPath)
	if err != nil {
		t.Fatalf("read security.yaml: %v", err)
	}
	defer os.WriteFile(securityPath, original, 0644)

	// Rapidly toggle DELETE 5 times
	for i := 0; i < 5; i++ {
		// Remove DELETE
		modified := []byte{}
		for _, line := range splitLines(string(original)) {
			if line != "    - DELETE" {
				modified = append(modified, []byte(line+"\n")...)
			}
		}
		os.WriteFile(securityPath, modified, 0644)
		time.Sleep(2 * time.Second)

		// Restore
		os.WriteFile(securityPath, original, 0644)
		time.Sleep(2 * time.Second)
	}

	// Wait for final poll
	time.Sleep(8 * time.Second)

	// Server must still be responsive
	mustCallQuery(t, session, dbName, fmt.Sprintf("SELECT %d AS after_rapid_reload", 42))
	t.Log("Server responsive after rapid reloads: OK")
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
