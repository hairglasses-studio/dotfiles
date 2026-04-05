package security

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseRBACUsers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantNil  bool
		wantUser string
		wantRole Role
	}{
		{"empty", "", true, "", ""},
		{"whitespace", "  ", true, "", ""},
		{"single user", "alice:admin", false, "alice", RoleAdmin},
		{"operator", "bob:operator", false, "bob", RoleOperator},
		{"readonly", "viewer:readonly", false, "viewer", RoleReadonly},
		{"with spaces", " alice : admin , bob : operator ", false, "alice", RoleAdmin},
		{"invalid role skipped", "alice:superuser", true, "", ""},
		{"no colon skipped", "alice", true, "", ""},
		{"empty parts skipped", ":admin,bob:", true, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRBACUsers(tt.input)
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			roles, ok := result[tt.wantUser]
			if !ok {
				t.Fatalf("expected user %q in result", tt.wantUser)
			}
			found := false
			for _, r := range roles {
				if r == tt.wantRole {
					found = true
				}
			}
			if !found {
				t.Errorf("expected role %q for user %q, got %v", tt.wantRole, tt.wantUser, roles)
			}
		})
	}
}

func TestParseRBACUsers_MultipleUsers(t *testing.T) {
	result := parseRBACUsers("alice:admin,bob:operator,default:readonly")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != 3 {
		t.Errorf("expected 3 users, got %d", len(result))
	}
	if result["alice"][0] != RoleAdmin {
		t.Errorf("alice should be admin, got %v", result["alice"])
	}
	if result["bob"][0] != RoleOperator {
		t.Errorf("bob should be operator, got %v", result["bob"])
	}
	if result["default"][0] != RoleReadonly {
		t.Errorf("default should be readonly, got %v", result["default"])
	}
}

func TestNewRBAC_Defaults(t *testing.T) {
	// Without RBAC_USERS set, should use hardcoded defaults
	rbac := NewRBAC()
	roles := rbac.GetUserRoles("mitch")
	if len(roles) == 0 || roles[0] != RoleAdmin {
		t.Errorf("mitch should be admin by default, got %v", roles)
	}
	roles = rbac.GetUserRoles("unknown")
	if len(roles) == 0 || roles[0] != RoleReadonly {
		t.Errorf("unknown user should be readonly, got %v", roles)
	}
}

func TestRBAC_HasPermission(t *testing.T) {
	rbac := NewRBAC()
	if !rbac.HasPermission("mitch", PermAdmin) {
		t.Error("admin should have admin permission")
	}
	if !rbac.HasPermission("luke", PermWrite) {
		t.Error("operator should have write permission")
	}
	if rbac.HasPermission("luke", PermAdmin) {
		t.Error("operator should NOT have admin permission")
	}
	if rbac.HasPermission("unknown", PermWrite) {
		t.Error("readonly should NOT have write permission")
	}
}

func TestRBAC_CanAccessTool(t *testing.T) {
	rbac := NewRBAC()
	if !rbac.CanAccessTool("mitch", "aftrs_inventory_delete") {
		t.Error("admin should access delete tool")
	}
	if !rbac.CanAccessTool("luke", "aftrs_inventory_list") {
		t.Error("operator should access list tool")
	}
	if rbac.CanAccessTool("unknown", "aftrs_inventory_delete") {
		t.Error("readonly should NOT access delete tool")
	}
}

// ── Audit Logger Tests ──

func TestAuditLogger_LogAndRetrieve(t *testing.T) {
	logger := NewAuditLogger("", 100) // in-memory only, no file

	logger.LogToolCall("mitch", "aftrs_inventory_list", map[string]any{"category": "GPU"})
	logger.LogToolResult("mitch", "aftrs_inventory_list", 50*time.Millisecond, nil)

	events := logger.GetRecentEvents(10)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != AuditToolCall {
		t.Errorf("first event type = %q, want %q", events[0].Type, AuditToolCall)
	}
	if events[1].Type != AuditToolSuccess {
		t.Errorf("second event type = %q, want %q", events[1].Type, AuditToolSuccess)
	}
}

func TestAuditLogger_ErrorEvents(t *testing.T) {
	logger := NewAuditLogger("", 100)

	logger.LogToolResult("mitch", "aftrs_inventory_delete", 10*time.Millisecond, fmt.Errorf("not found"))
	logger.LogToolResult("mitch", "aftrs_inventory_list", 5*time.Millisecond, nil)
	logger.LogAccessDenied("unknown", "aftrs_inventory_delete", "insufficient permissions")

	errors := logger.GetErrorEvents(10)
	if len(errors) != 2 {
		t.Fatalf("expected 2 error events, got %d", len(errors))
	}
}

func TestAuditLogger_FilterByUser(t *testing.T) {
	logger := NewAuditLogger("", 100)

	logger.LogToolCall("mitch", "tool_a", nil)
	logger.LogToolCall("luke", "tool_b", nil)
	logger.LogToolCall("mitch", "tool_c", nil)

	events := logger.GetEventsByUser("mitch", 10)
	if len(events) != 2 {
		t.Fatalf("expected 2 events for mitch, got %d", len(events))
	}
}

func TestAuditLogger_FilterByTool(t *testing.T) {
	logger := NewAuditLogger("", 100)

	logger.LogToolCall("mitch", "tool_a", nil)
	logger.LogToolCall("luke", "tool_a", nil)
	logger.LogToolCall("mitch", "tool_b", nil)

	events := logger.GetEventsByTool("tool_a", 10)
	if len(events) != 2 {
		t.Fatalf("expected 2 events for tool_a, got %d", len(events))
	}
}

func TestAuditLogger_Stats(t *testing.T) {
	logger := NewAuditLogger("", 100)

	logger.LogToolCall("mitch", "tool_a", nil)
	logger.LogToolResult("mitch", "tool_a", 100*time.Millisecond, nil)
	logger.LogToolResult("luke", "tool_b", 200*time.Millisecond, fmt.Errorf("fail"))

	stats := logger.GetStats()
	if stats.TotalEvents != 3 {
		t.Errorf("total events = %d, want 3", stats.TotalEvents)
	}
	if stats.ErrorCount != 1 {
		t.Errorf("error count = %d, want 1", stats.ErrorCount)
	}
}

func TestAuditLogger_CircularBuffer(t *testing.T) {
	logger := NewAuditLogger("", 3) // max 3 events

	logger.LogToolCall("a", "t1", nil)
	logger.LogToolCall("b", "t2", nil)
	logger.LogToolCall("c", "t3", nil)
	logger.LogToolCall("d", "t4", nil) // should evict first

	events := logger.GetRecentEvents(10)
	if len(events) != 3 {
		t.Fatalf("expected 3 events (circular), got %d", len(events))
	}
	if events[0].User != "b" {
		t.Errorf("oldest event user = %q, want %q (first should be evicted)", events[0].User, "b")
	}
}

func TestSanitizeParams(t *testing.T) {
	params := map[string]any{
		"name":     "test",
		"password": "secret123",
		"api_key":  "key-abc",
		"category": "GPU",
	}

	sanitized := sanitizeParams(params)
	if sanitized["name"] != "test" {
		t.Errorf("name should not be redacted")
	}
	if sanitized["password"] != "[REDACTED]" {
		t.Errorf("password should be redacted, got %v", sanitized["password"])
	}
	if sanitized["api_key"] != "[REDACTED]" {
		t.Errorf("api_key should be redacted, got %v", sanitized["api_key"])
	}
	if sanitized["category"] != "GPU" {
		t.Errorf("category should not be redacted")
	}
}

// ── Additional RBAC Tests ──

func TestRBAC_SetUserRoles(t *testing.T) {
	rbac := NewRBAC()

	rbac.SetUserRoles("newuser", []Role{RoleOperator})
	roles := rbac.GetUserRoles("newuser")
	if len(roles) != 1 || roles[0] != RoleOperator {
		t.Errorf("expected [operator], got %v", roles)
	}

	// Overwrite existing user
	rbac.SetUserRoles("newuser", []Role{RoleAdmin, RoleOperator})
	roles = rbac.GetUserRoles("newuser")
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
}

func TestRBAC_SetToolPermission(t *testing.T) {
	rbac := NewRBAC()

	// A _list tool normally requires PermRead
	perm := rbac.GetToolPermission("aftrs_inventory_list")
	if perm != PermRead {
		t.Errorf("expected read for _list tool, got %v", perm)
	}

	// Override it to require admin
	rbac.SetToolPermission("aftrs_inventory_list", PermAdmin)
	perm = rbac.GetToolPermission("aftrs_inventory_list")
	if perm != PermAdmin {
		t.Errorf("expected admin after override, got %v", perm)
	}

	// Operator should now be denied
	if rbac.CanAccessTool("luke", "aftrs_inventory_list") {
		t.Error("operator should NOT access tool overridden to admin")
	}
}

func TestRBAC_GetToolPermission_DefaultsToRead(t *testing.T) {
	rbac := NewRBAC()

	// A tool with no matching suffix should default to read
	perm := rbac.GetToolPermission("aftrs_something_unknown")
	if perm != PermRead {
		t.Errorf("expected read for unknown suffix, got %v", perm)
	}
}

func TestRBAC_GetToolPermission_AllSuffixes(t *testing.T) {
	rbac := NewRBAC()

	readTools := []string{"my_status", "my_list", "my_get", "my_health", "my_whoami", "my_discover"}
	for _, tool := range readTools {
		perm := rbac.GetToolPermission(tool)
		if perm != PermRead {
			t.Errorf("tool %q: expected read, got %v", tool, perm)
		}
	}

	writeTools := []string{"my_sync", "my_create", "my_update", "my_add", "my_restart", "my_trigger"}
	for _, tool := range writeTools {
		perm := rbac.GetToolPermission(tool)
		if perm != PermWrite {
			t.Errorf("tool %q: expected write, got %v", tool, perm)
		}
	}

	adminTools := []string{"my_delete", "my_reset", "my_rotate"}
	for _, tool := range adminTools {
		perm := rbac.GetToolPermission(tool)
		if perm != PermAdmin {
			t.Errorf("tool %q: expected admin, got %v", tool, perm)
		}
	}
}

func TestCheckAccess(t *testing.T) {
	ctx := context.Background()

	// Admin should pass
	err := CheckAccess(ctx, "mitch", "aftrs_inventory_delete")
	if err != nil {
		t.Errorf("admin should have access, got error: %v", err)
	}

	// Readonly user on delete tool should fail
	err = CheckAccess(ctx, "unknown_user_xyz", "aftrs_inventory_delete")
	if err == nil {
		t.Error("readonly user should be denied access to delete tool")
	}
	if err != nil && !containsStr(err.Error(), "access denied") {
		t.Errorf("error should contain 'access denied', got %q", err.Error())
	}
}

func TestGetAllRoles(t *testing.T) {
	roles := GetAllRoles()
	if len(roles) != 3 {
		t.Fatalf("expected 3 roles, got %d", len(roles))
	}

	roleNames := make(map[Role]bool)
	for _, r := range roles {
		roleNames[r.Name] = true
		if len(r.Permissions) == 0 {
			t.Errorf("role %q has no permissions", r.Name)
		}
	}

	for _, expected := range []Role{RoleAdmin, RoleOperator, RoleReadonly} {
		if !roleNames[expected] {
			t.Errorf("missing role %q in GetAllRoles()", expected)
		}
	}
}

func TestRBAC_GetUserAccess(t *testing.T) {
	rbac := NewRBAC()

	// Admin user
	info := rbac.GetUserAccess("mitch")
	if info.Username != "mitch" {
		t.Errorf("username = %q, want mitch", info.Username)
	}
	if len(info.Roles) == 0 {
		t.Fatal("expected at least one role for mitch")
	}
	if info.Roles[0] != RoleAdmin {
		t.Errorf("expected admin role, got %v", info.Roles[0])
	}
	// Admin should have all 4 permissions
	if len(info.Permissions) != 4 {
		t.Errorf("admin should have 4 permissions, got %d: %v", len(info.Permissions), info.Permissions)
	}

	// Unknown user (falls back to readonly)
	info = rbac.GetUserAccess("nobody")
	if info.Username != "nobody" {
		t.Errorf("username = %q, want nobody", info.Username)
	}
	if len(info.Permissions) != 1 {
		t.Errorf("readonly should have 1 permission, got %d: %v", len(info.Permissions), info.Permissions)
	}
}

func TestRBAC_GetUserRoles_NoDefaultFallback(t *testing.T) {
	// Create RBAC with no "default" entry
	rbac := &RBAC{
		userRoles:     map[string][]Role{"alice": {RoleAdmin}},
		toolOverrides: make(map[string]Permission),
	}

	// Unknown user with no "default" key should get RoleReadonly
	roles := rbac.GetUserRoles("bob")
	if len(roles) != 1 || roles[0] != RoleReadonly {
		t.Errorf("expected [readonly] fallback, got %v", roles)
	}
}

// ── Additional Audit Logger Tests ──

func TestAuditLogger_WriteEventToFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test-audit.log")

	logger := NewAuditLogger(logFile, 100)
	defer logger.Close()

	logger.LogToolCall("mitch", "tool_a", map[string]any{"key": "value"})

	// Give the background writer time to process
	time.Sleep(100 * time.Millisecond)

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if len(data) == 0 {
		t.Error("log file should not be empty")
	}
	content := string(data)
	if !containsStr(content, "tool_a") {
		t.Errorf("log file should contain tool_a, got: %s", content)
	}
	if !containsStr(content, "mitch") {
		t.Errorf("log file should contain user mitch, got: %s", content)
	}
}

func TestAuditLogger_Close(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test-close.log")

	logger := NewAuditLogger(logFile, 100)
	logger.LogToolCall("user", "tool", nil)
	time.Sleep(50 * time.Millisecond)
	logger.Close()
	// Closing should not panic; second close would panic on closed channel,
	// so we only close once.
}

func TestAuditLogger_LogWithExplicitID(t *testing.T) {
	logger := NewAuditLogger("", 100)

	logger.Log(AuditEvent{
		ID:   "custom-id-123",
		Type: AuditToolCall,
		User: "mitch",
		Tool: "tool_x",
	})

	events := logger.GetRecentEvents(1)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != "custom-id-123" {
		t.Errorf("expected custom ID, got %q", events[0].ID)
	}
}

func TestAuditLogger_GetRecentEvents_ZeroLimit(t *testing.T) {
	logger := NewAuditLogger("", 100)

	logger.LogToolCall("a", "t1", nil)
	logger.LogToolCall("b", "t2", nil)

	// limit <= 0 should return all events
	events := logger.GetRecentEvents(0)
	if len(events) != 2 {
		t.Errorf("expected 2 events with limit=0, got %d", len(events))
	}

	events = logger.GetRecentEvents(-1)
	if len(events) != 2 {
		t.Errorf("expected 2 events with limit=-1, got %d", len(events))
	}
}

func TestAuditLogger_Stats_AccessDenied(t *testing.T) {
	logger := NewAuditLogger("", 100)

	logger.LogAccessDenied("baduser", "secret_tool", "no perms")
	logger.LogAccessDenied("baduser", "other_tool", "no perms")

	stats := logger.GetStats()
	if stats.AccessDenied != 2 {
		t.Errorf("access denied = %d, want 2", stats.AccessDenied)
	}
	if stats.TotalEvents != 2 {
		t.Errorf("total events = %d, want 2", stats.TotalEvents)
	}
}

func TestAuditLogger_Stats_NoLatency(t *testing.T) {
	logger := NewAuditLogger("", 100)

	// Events with no duration should result in zero average latency
	logger.LogToolCall("a", "t1", nil)
	stats := logger.GetStats()
	if stats.AverageLatency != 0 {
		t.Errorf("average latency should be 0 with no duration events, got %v", stats.AverageLatency)
	}
}

func TestLogToolInvocation(t *testing.T) {
	// Save and restore global logger
	orig := GlobalAuditLogger
	defer func() { GlobalAuditLogger = orig }()

	logger := NewAuditLogger("", 100)
	GlobalAuditLogger = logger

	LogToolInvocation(context.Background(), "mitch", "test_tool", map[string]any{"x": 1})

	events := logger.GetRecentEvents(10)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != AuditToolCall {
		t.Errorf("expected tool_call type, got %v", events[0].Type)
	}
	if events[0].User != "mitch" {
		t.Errorf("expected user mitch, got %q", events[0].User)
	}
}

func TestLogToolCompletion(t *testing.T) {
	orig := GlobalAuditLogger
	defer func() { GlobalAuditLogger = orig }()

	logger := NewAuditLogger("", 100)
	GlobalAuditLogger = logger

	LogToolCompletion(context.Background(), "luke", "test_tool", 42*time.Millisecond, nil)

	events := logger.GetRecentEvents(10)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != AuditToolSuccess {
		t.Errorf("expected tool_success type, got %v", events[0].Type)
	}

	// Also test with error
	LogToolCompletion(context.Background(), "luke", "fail_tool", 10*time.Millisecond, fmt.Errorf("boom"))
	events = logger.GetRecentEvents(10)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[1].Type != AuditToolError {
		t.Errorf("expected tool_error type, got %v", events[1].Type)
	}
}

func TestSanitizeParams_Nil(t *testing.T) {
	result := sanitizeParams(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
}

func TestSanitizeParams_AllSensitiveKeys(t *testing.T) {
	sensitiveKeys := []string{
		"password", "secret", "token", "key", "credential",
		"auth", "bearer", "api_key", "private", "oauth",
	}

	for _, k := range sensitiveKeys {
		params := map[string]any{k: "sensitive_value"}
		sanitized := sanitizeParams(params)
		if sanitized[k] != "[REDACTED]" {
			t.Errorf("key %q should be redacted, got %v", k, sanitized[k])
		}
	}
}

func TestSanitizeParams_CaseInsensitiveContains(t *testing.T) {
	// Keys containing sensitive substrings should also be redacted
	params := map[string]any{
		"my_Password_field": "secret",
		"AUTH_header":       "bearer xyz",
		"user_token_value":  "tok123",
		"safe_field":        "ok",
	}
	sanitized := sanitizeParams(params)

	if sanitized["my_Password_field"] != "[REDACTED]" {
		t.Errorf("my_Password_field should be redacted, got %v", sanitized["my_Password_field"])
	}
	if sanitized["AUTH_header"] != "[REDACTED]" {
		t.Errorf("AUTH_header should be redacted, got %v", sanitized["AUTH_header"])
	}
	if sanitized["user_token_value"] != "[REDACTED]" {
		t.Errorf("user_token_value should be redacted, got %v", sanitized["user_token_value"])
	}
	if sanitized["safe_field"] != "ok" {
		t.Errorf("safe_field should not be redacted")
	}
}

func TestContains_EdgeCases(t *testing.T) {
	// Empty substring
	if !contains("anything", "") {
		t.Error("empty substring should match")
	}

	// Substring longer than string
	if contains("ab", "abcdef") {
		t.Error("longer substring should not match")
	}

	// Exact match
	if !contains("password", "password") {
		t.Error("exact match should return true")
	}

	// Case insensitive
	if !contains("MyPassword", "password") {
		t.Error("case insensitive match should work")
	}
	if !contains("password", "PASSWORD") {
		t.Error("case insensitive match should work (upper pattern)")
	}
}

func TestAuditLogger_QueueFull(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "queue-full.log")

	// Create logger but don't start background writer by using a trick:
	// We create with a file, then block the writer by filling the queue.
	logger := &AuditLogger{
		events:     make([]AuditEvent, 0, 100),
		maxEvents:  100,
		logFile:    logFile,
		writeQueue: make(chan AuditEvent, 1), // tiny buffer
		stopCh:     make(chan struct{}),
	}
	// Don't start backgroundWriter - the queue will fill up

	// Fill queue
	logger.Log(AuditEvent{Type: AuditToolCall, User: "a", Tool: "t1"})
	// This one should be dropped (queue full) but not panic
	logger.Log(AuditEvent{Type: AuditToolCall, User: "b", Tool: "t2"})

	// In-memory buffer should still have both
	events := logger.GetRecentEvents(10)
	if len(events) != 2 {
		t.Errorf("expected 2 in-memory events, got %d", len(events))
	}
}

// helper to avoid importing strings just for Contains
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
