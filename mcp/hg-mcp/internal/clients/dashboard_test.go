package clients

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestDashboardClientQuickStatusShowsDegradedHTTPSystem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not ready", http.StatusBadRequest)
	}))
	defer server.Close()
	host, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatalf("SplitHostPort() error = %v", err)
	}

	client := &DashboardClient{
		systems: map[string]*SystemConfig{
			"test-api": {
				Name:        "test-api",
				Category:    "network",
				Host:        host,
				Port:        mustAtoi(t, port),
				Protocol:    "http",
				Description: "HTTP health probe",
			},
		},
		httpClient: server.Client(),
	}

	status := client.GetQuickStatus(context.Background())
	if !strings.Contains(status, "1 degraded") {
		t.Fatalf("quick status should report one degraded system, got %q", status)
	}
	if !strings.Contains(status, "⚠test-api") {
		t.Fatalf("quick status should flag test-api as degraded, got %q", status)
	}
}

func TestCheckSystemDetailedReportsHTTPDegradation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not ready", http.StatusBadRequest)
	}))
	defer server.Close()
	host, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatalf("SplitHostPort() error = %v", err)
	}

	client := &DashboardClient{httpClient: server.Client()}
	status := client.checkSystemDetailed(context.Background(), &SystemConfig{
		Name:     "test-api",
		Category: "network",
		Host:     host,
		Port:     mustAtoi(t, port),
		Protocol: "http",
	})
	if status.Status != "degraded" {
		t.Fatalf("status = %q, want degraded", status.Status)
	}
	if status.Message != "HTTP 400" {
		t.Fatalf("message = %q, want HTTP 400", status.Message)
	}
}

func mustAtoi(t *testing.T, value string) int {
	t.Helper()
	port, err := strconv.Atoi(value)
	if err != nil {
		t.Fatalf("Atoi(%q) error = %v", value, err)
	}
	return port
}
