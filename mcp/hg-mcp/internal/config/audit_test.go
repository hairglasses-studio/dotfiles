package config

import (
	"fmt"
	"os"
	"testing"
)

func TestIsCategoryConfigured_WithEnvSet(t *testing.T) {
	t.Setenv("NANOLEAF_HOST", "192.168.1.100")

	if !IsCategoryConfigured("nanoleaf") {
		t.Error("expected nanoleaf category to be configured when NANOLEAF_HOST is set")
	}
}

func TestIsCategoryConfigured_WithoutEnv(t *testing.T) {
	// Ensure the env var is unset
	t.Setenv("NANOLEAF_HOST", "")

	if IsCategoryConfigured("nanoleaf") {
		t.Error("expected nanoleaf category to be unconfigured when NANOLEAF_HOST is empty")
	}
}

func TestIsCategoryConfigured_UnknownCategory(t *testing.T) {
	// Unknown categories should return true (assumed configured)
	if !IsCategoryConfigured("totally_unknown_category_xyz") {
		t.Error("expected unknown category to be assumed configured")
	}
}

func TestIsCategoryConfigured_MultipleEnvVars(t *testing.T) {
	// Beatport has two env vars; setting one should suffice
	t.Setenv("BEATPORT_USERNAME", "")
	t.Setenv("BEATPORT_ACCESS_TOKEN", "some-token")

	if !IsCategoryConfigured("beatport") {
		t.Error("expected beatport to be configured when at least one env var is set")
	}
}

func TestIsServiceConfigured_Known(t *testing.T) {
	t.Setenv("HUE_BRIDGE_IP", "10.0.0.5")

	if !IsServiceConfigured("PhilipsHue") {
		t.Error("expected PhilipsHue to be configured when HUE_BRIDGE_IP is set")
	}
}

func TestIsServiceConfigured_KnownButMissing(t *testing.T) {
	t.Setenv("DISCORD_BOT_TOKEN", "")

	if IsServiceConfigured("Discord") {
		t.Error("expected Discord to be unconfigured when DISCORD_BOT_TOKEN is empty")
	}
}

func TestIsServiceConfigured_UnknownService(t *testing.T) {
	// Unknown services should return true (assumed configured)
	if !IsServiceConfigured("NonExistentService99") {
		t.Error("expected unknown service to be assumed configured")
	}
}

func TestCategoryToService_AllPointToValidServices(t *testing.T) {
	// Build a set of known service names
	serviceNames := make(map[string]bool, len(knownServices))
	for _, svc := range knownServices {
		serviceNames[svc.Name] = true
	}

	for category, serviceName := range CategoryToService {
		if !serviceNames[serviceName] {
			t.Errorf("CategoryToService[%q] = %q, which is not in knownServices",
				category, serviceName)
		}
	}
}

func TestAuditConfig_Basic(t *testing.T) {
	report := AuditConfig()
	// Should have some missing services (we're in test, no env vars set)
	if len(report.Missing) == 0 && len(report.Configured) == 0 {
		t.Error("expected at least some services in the report")
	}
	t.Logf("Configured: %d, Missing: %d", len(report.Configured), len(report.Missing))
}

func TestAuditConfig_ConfiguredService(t *testing.T) {
	t.Setenv("DISCORD_BOT_TOKEN", "test-token")

	report := AuditConfig()

	found := false
	for _, name := range report.Configured {
		if name == "Discord" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Discord to appear in Configured list when DISCORD_BOT_TOKEN is set")
	}

	// Discord should NOT appear in Missing
	for _, name := range report.Missing {
		if name == "Discord" {
			t.Error("Discord should not appear in Missing when DISCORD_BOT_TOKEN is set")
		}
	}
}

func TestAuditConfig_MissingService(t *testing.T) {
	t.Setenv("DISCORD_BOT_TOKEN", "")

	report := AuditConfig()

	found := false
	for _, name := range report.Missing {
		if name == "Discord" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Discord to appear in Missing list when DISCORD_BOT_TOKEN is empty")
	}
}

func TestAuditConfig_SkipsOptionalServices(t *testing.T) {
	report := AuditConfig()

	// Optional services (Core, Security, Observability, Tailscale) should not appear
	optionalNames := map[string]bool{
		"Core": true, "Security": true, "Observability": true, "Tailscale": true,
	}
	for _, name := range report.Configured {
		if optionalNames[name] {
			t.Errorf("optional service %q should not appear in Configured", name)
		}
	}
	for _, name := range report.Missing {
		if optionalNames[name] {
			t.Errorf("optional service %q should not appear in Missing", name)
		}
	}
}

func TestConnectivityCheckPass(t *testing.T) {
	// Temporarily add a service with a passing check
	orig := knownServices
	defer func() { knownServices = orig }()

	knownServices = []ServiceGroup{
		{
			Name:     "TestOK",
			EnvVars:  []string{},
			Optional: true,
			ConnectivityCheck: func() error {
				return nil
			},
		},
	}

	report := AuditConnectivity()
	if report.Skipped {
		t.Error("report should not be skipped")
	}
	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}
	if !report.Results[0].Reachable {
		t.Errorf("expected reachable, got error: %s", report.Results[0].Error)
	}
	if report.Results[0].Name != "TestOK" {
		t.Errorf("expected name TestOK, got %s", report.Results[0].Name)
	}
}

func TestConnectivityCheckFail(t *testing.T) {
	orig := knownServices
	defer func() { knownServices = orig }()

	knownServices = []ServiceGroup{
		{
			Name:     "TestFail",
			EnvVars:  []string{},
			Optional: true,
			ConnectivityCheck: func() error {
				return fmt.Errorf("connection refused")
			},
		},
	}

	report := AuditConnectivity()
	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}
	if report.Results[0].Reachable {
		t.Error("expected not reachable")
	}
	if report.Results[0].Error != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", report.Results[0].Error)
	}
}

func TestConnectivityCheckSkip(t *testing.T) {
	os.Setenv("SKIP_CONNECTIVITY_CHECKS", "1")
	defer os.Unsetenv("SKIP_CONNECTIVITY_CHECKS")

	orig := knownServices
	defer func() { knownServices = orig }()

	knownServices = []ServiceGroup{
		{
			Name:     "TestSkipped",
			EnvVars:  []string{},
			Optional: true,
			ConnectivityCheck: func() error {
				t.Fatal("check should not be called when skipped")
				return nil
			},
		},
	}

	report := AuditConnectivity()
	if !report.Skipped {
		t.Error("expected report to be skipped")
	}
	if len(report.Results) != 0 {
		t.Errorf("expected 0 results when skipped, got %d", len(report.Results))
	}
}

func TestConnectivityCheckUnconfigured(t *testing.T) {
	// Services without env vars set should be skipped
	orig := knownServices
	defer func() { knownServices = orig }()

	knownServices = []ServiceGroup{
		{
			Name:    "TestNotConfigured",
			EnvVars: []string{"NONEXISTENT_TEST_VAR_12345"},
			ConnectivityCheck: func() error {
				t.Fatal("check should not be called for unconfigured service")
				return nil
			},
		},
	}

	report := AuditConnectivity()
	if len(report.Results) != 0 {
		t.Errorf("expected 0 results for unconfigured service, got %d", len(report.Results))
	}
}
