// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
	httpclient "github.com/hairglasses-studio/mcpkit/client"
	"gopkg.in/yaml.v3"
)

// PluginType represents the type of plugin
type PluginType string

const (
	PluginTypeTDTox       PluginType = "touchdesigner_tox"
	PluginTypeTDComponent PluginType = "touchdesigner_component"
	PluginTypeResEffect   PluginType = "resolume_effect"
	PluginTypeResSource   PluginType = "resolume_source"
	PluginTypeGeneric     PluginType = "generic"
)

// PluginSource represents where a plugin comes from
type PluginSource string

const (
	PluginSourceGitHubOrg PluginSource = "github_org"
	PluginSourceManifest  PluginSource = "manifest"
	PluginSourceCommunity PluginSource = "community"
)

// Plugin represents a plugin in the registry
type Plugin struct {
	Name        string       `json:"name" yaml:"name"`
	Description string       `json:"description" yaml:"description"`
	Type        PluginType   `json:"type" yaml:"type"`
	Repo        string       `json:"repo" yaml:"repo"`
	Version     string       `json:"version,omitempty" yaml:"version,omitempty"`
	Author      string       `json:"author,omitempty" yaml:"author,omitempty"`
	Tags        []string     `json:"tags,omitempty" yaml:"tags,omitempty"`
	Source      PluginSource `json:"source" yaml:"-"`
	Installed   bool         `json:"installed" yaml:"-"`
	InstallPath string       `json:"install_path,omitempty" yaml:"-"`
	UpdateAvail bool         `json:"update_available,omitempty" yaml:"-"`
}

// PluginVersion represents a specific version of a plugin
type PluginVersion struct {
	Version     string    `json:"version"`
	Tag         string    `json:"tag"`
	ReleaseDate time.Time `json:"release_date"`
	Assets      []string  `json:"assets"`
	Prerelease  bool      `json:"prerelease"`
}

// PluginInstallResult represents the result of an installation
type PluginInstallResult struct {
	Success     bool   `json:"success"`
	Plugin      string `json:"plugin"`
	Version     string `json:"version"`
	InstallPath string `json:"install_path"`
	Error       string `json:"error,omitempty"`
}

// InstalledPlugin represents a locally installed plugin
type InstalledPlugin struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Type        PluginType `json:"type"`
	InstallPath string     `json:"install_path"`
	InstalledAt time.Time  `json:"installed_at"`
	Repo        string     `json:"repo"`
}

// PluginManifest represents the curated plugin manifest
type PluginManifest struct {
	Version int      `yaml:"version"`
	Plugins []Plugin `yaml:"plugins"`
	Sources struct {
		GitHubOrgs   []string `yaml:"github_orgs"`
		ManifestURLs []string `yaml:"manifest_urls"`
	} `yaml:"sources"`
}

// PluginTypeConfig defines configuration for a plugin type
type PluginTypeConfig struct {
	Name        string   `yaml:"name"`
	Extensions  []string `yaml:"extensions"`
	InstallPath string   `yaml:"install_path"`
	ValidateCmd string   `yaml:"validate_cmd,omitempty"`
	PostInstall string   `yaml:"post_install,omitempty"`
}

// PluginHealth represents plugin system health
type PluginHealth struct {
	Score            int      `json:"score"`
	Status           string   `json:"status"`
	PluginsTotal     int      `json:"plugins_total"`
	PluginsInstalled int      `json:"plugins_installed"`
	UpdatesAvailable int      `json:"updates_available"`
	Issues           []string `json:"issues,omitempty"`
	Recommendations  []string `json:"recommendations,omitempty"`
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int    `json:"size"`
	} `json:"assets"`
}

// PluginClient provides access to the plugin registry
type PluginClient struct {
	httpClient   *http.Client
	githubToken  string
	manifestPath string
	vaultPath    string
	typeConfigs  map[PluginType]PluginTypeConfig
}

// NewPluginClient creates a new plugin client
func NewPluginClient() (*PluginClient, error) {
	cfg := config.GetOrLoad()
	homeDir := cfg.Home
	if homeDir == "" {
		homeDir = "."
	}

	vaultPath := filepath.Join(cfg.AftrsVaultPath, "plugins")

	// Ensure vault directory exists
	os.MkdirAll(vaultPath, 0755)
	os.MkdirAll(filepath.Join(vaultPath, "cache"), 0755)

	client := &PluginClient{
		httpClient:   httpclient.Standard(),
		githubToken:  os.Getenv("GITHUB_TOKEN"),
		manifestPath: os.Getenv("PLUGIN_MANIFEST_PATH"),
		vaultPath:    vaultPath,
		typeConfigs:  defaultTypeConfigs(homeDir),
	}

	return client, nil
}

func defaultTypeConfigs(homeDir string) map[PluginType]PluginTypeConfig {
	tdPalette := os.Getenv("TD_PALETTE_PATH")
	if tdPalette == "" {
		tdPalette = filepath.Join(homeDir, "Documents", "Derivative", "Palette")
	}

	tdComponents := os.Getenv("TD_COMPONENTS_PATH")
	if tdComponents == "" {
		tdComponents = filepath.Join(homeDir, "Documents", "Derivative", "Components")
	}

	resEffects := os.Getenv("RESOLUME_EFFECTS_PATH")
	if resEffects == "" {
		resEffects = filepath.Join(homeDir, "Documents", "Resolume Arena 7", "Effects")
	}

	resSources := os.Getenv("RESOLUME_SOURCES_PATH")
	if resSources == "" {
		resSources = filepath.Join(homeDir, "Documents", "Resolume Arena 7", "Sources")
	}

	return map[PluginType]PluginTypeConfig{
		PluginTypeTDTox: {
			Name:        "TouchDesigner TOX",
			Extensions:  []string{".tox"},
			InstallPath: tdPalette,
		},
		PluginTypeTDComponent: {
			Name:        "TouchDesigner Component",
			Extensions:  []string{".tox", ".toe"},
			InstallPath: tdComponents,
		},
		PluginTypeResEffect: {
			Name:        "Resolume Effect",
			Extensions:  []string{".ffx", ".dll", ".vuo"},
			InstallPath: resEffects,
		},
		PluginTypeResSource: {
			Name:        "Resolume Source",
			Extensions:  []string{".ffx", ".dll", ".vuo"},
			InstallPath: resSources,
		},
		PluginTypeGeneric: {
			Name:        "Generic Plugin",
			Extensions:  []string{"*"},
			InstallPath: filepath.Join(homeDir, "aftrs-vault", "plugins", "generic"),
		},
	}
}

// ListPlugins returns all available plugins
func (c *PluginClient) ListPlugins(ctx context.Context, filter string) ([]Plugin, error) {
	plugins := []Plugin{}

	// Load from manifest
	manifest, err := c.LoadManifest(ctx)
	if err == nil {
		for _, p := range manifest.Plugins {
			p.Source = PluginSourceManifest
			if filter == "" || matchesFilter(p, filter) {
				plugins = append(plugins, p)
			}
		}
	}

	// Mark installed plugins
	installed, _ := c.ListInstalled(ctx)
	installedMap := make(map[string]InstalledPlugin)
	for _, ip := range installed {
		installedMap[ip.Name] = ip
	}

	for i := range plugins {
		if ip, ok := installedMap[plugins[i].Name]; ok {
			plugins[i].Installed = true
			plugins[i].InstallPath = ip.InstallPath
			plugins[i].Version = ip.Version
		}
	}

	return plugins, nil
}

func matchesFilter(p Plugin, filter string) bool {
	filter = strings.ToLower(filter)
	if strings.Contains(strings.ToLower(p.Name), filter) {
		return true
	}
	if strings.Contains(strings.ToLower(p.Description), filter) {
		return true
	}
	if strings.Contains(strings.ToLower(string(p.Type)), filter) {
		return true
	}
	for _, tag := range p.Tags {
		if strings.Contains(strings.ToLower(tag), filter) {
			return true
		}
	}
	return false
}

// SearchPlugins searches plugins by query
func (c *PluginClient) SearchPlugins(ctx context.Context, query string) ([]Plugin, error) {
	return c.ListPlugins(ctx, query)
}

// GetPlugin returns a specific plugin by name
func (c *PluginClient) GetPlugin(ctx context.Context, name string) (*Plugin, error) {
	plugins, err := c.ListPlugins(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, p := range plugins {
		if p.Name == name {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("plugin not found: %s", name)
}

// GetPluginVersions returns available versions for a plugin
func (c *PluginClient) GetPluginVersions(ctx context.Context, name string) ([]PluginVersion, error) {
	plugin, err := c.GetPlugin(ctx, name)
	if err != nil {
		return nil, err
	}

	releases, err := c.getGitHubReleases(ctx, plugin.Repo)
	if err != nil {
		return nil, err
	}

	versions := []PluginVersion{}
	for _, r := range releases {
		if r.Draft {
			continue
		}
		assets := []string{}
		for _, a := range r.Assets {
			assets = append(assets, a.Name)
		}
		versions = append(versions, PluginVersion{
			Version:     r.TagName,
			Tag:         r.TagName,
			ReleaseDate: r.PublishedAt,
			Assets:      assets,
			Prerelease:  r.Prerelease,
		})
	}

	return versions, nil
}

func (c *PluginClient) getGitHubReleases(ctx context.Context, repo string) ([]GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.githubToken != "" {
		req.Header.Set("Authorization", "token "+c.githubToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// InstallPlugin installs a plugin
func (c *PluginClient) InstallPlugin(ctx context.Context, name, version string) (*PluginInstallResult, error) {
	result := &PluginInstallResult{Plugin: name, Version: version}

	plugin, err := c.GetPlugin(ctx, name)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Get releases
	releases, err := c.getGitHubReleases(ctx, plugin.Repo)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	if len(releases) == 0 {
		result.Error = "no releases found"
		return result, fmt.Errorf("no releases found for %s", name)
	}

	// Find the requested version or latest
	var targetRelease *GitHubRelease
	for i, r := range releases {
		if version == "" || version == "latest" {
			if !r.Prerelease && !r.Draft {
				targetRelease = &releases[i]
				break
			}
		} else if r.TagName == version {
			targetRelease = &releases[i]
			break
		}
	}

	if targetRelease == nil {
		result.Error = "version not found"
		return result, fmt.Errorf("version %s not found", version)
	}

	result.Version = targetRelease.TagName

	// Find appropriate asset
	typeConfig := c.typeConfigs[plugin.Type]
	var assetURL, assetName string

	for _, asset := range targetRelease.Assets {
		for _, ext := range typeConfig.Extensions {
			if ext == "*" || strings.HasSuffix(strings.ToLower(asset.Name), ext) {
				assetURL = asset.BrowserDownloadURL
				assetName = asset.Name
				break
			}
		}
		if assetURL != "" {
			break
		}
	}

	if assetURL == "" {
		result.Error = "no compatible asset found"
		return result, fmt.Errorf("no compatible asset found for plugin type %s", plugin.Type)
	}

	// Download asset
	cachePath := filepath.Join(c.vaultPath, "cache", assetName)
	if err := c.downloadFile(ctx, assetURL, cachePath); err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Copy to install path
	installDir := typeConfig.InstallPath
	os.MkdirAll(installDir, 0755)
	installPath := filepath.Join(installDir, assetName)

	if err := c.copyFile(cachePath, installPath); err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.InstallPath = installPath
	result.Success = true

	// Record installation
	installed := InstalledPlugin{
		Name:        name,
		Version:     result.Version,
		Type:        plugin.Type,
		InstallPath: installPath,
		InstalledAt: time.Now(),
		Repo:        plugin.Repo,
	}
	c.recordInstallation(installed)

	return result, nil
}

func (c *PluginClient) downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	if c.githubToken != "" {
		req.Header.Set("Authorization", "token "+c.githubToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (c *PluginClient) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (c *PluginClient) recordInstallation(plugin InstalledPlugin) error {
	installed, _ := c.ListInstalled(context.Background())

	// Update or add
	found := false
	for i := range installed {
		if installed[i].Name == plugin.Name {
			installed[i] = plugin
			found = true
			break
		}
	}
	if !found {
		installed = append(installed, plugin)
	}

	return c.saveInstalled(installed)
}

func (c *PluginClient) saveInstalled(plugins []InstalledPlugin) error {
	path := filepath.Join(c.vaultPath, "installed.json")
	data, err := json.MarshalIndent(plugins, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// UninstallPlugin removes an installed plugin
func (c *PluginClient) UninstallPlugin(ctx context.Context, name string) error {
	installed, err := c.ListInstalled(ctx)
	if err != nil {
		return err
	}

	var target *InstalledPlugin
	for i := range installed {
		if installed[i].Name == name {
			target = &installed[i]
			break
		}
	}

	if target == nil {
		return fmt.Errorf("plugin not installed: %s", name)
	}

	// Remove file
	if err := os.Remove(target.InstallPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Update installed list
	var remaining []InstalledPlugin
	for _, ip := range installed {
		if ip.Name != name {
			remaining = append(remaining, ip)
		}
	}

	return c.saveInstalled(remaining)
}

// ListInstalled returns installed plugins
func (c *PluginClient) ListInstalled(ctx context.Context) ([]InstalledPlugin, error) {
	path := filepath.Join(c.vaultPath, "installed.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []InstalledPlugin{}, nil
		}
		return nil, err
	}

	var installed []InstalledPlugin
	if err := json.Unmarshal(data, &installed); err != nil {
		return nil, err
	}

	return installed, nil
}

// CheckUpdates checks for available updates
func (c *PluginClient) CheckUpdates(ctx context.Context) ([]Plugin, error) {
	installed, err := c.ListInstalled(ctx)
	if err != nil {
		return nil, err
	}

	updates := []Plugin{}

	for _, ip := range installed {
		plugin, err := c.GetPlugin(ctx, ip.Name)
		if err != nil {
			continue
		}

		versions, err := c.GetPluginVersions(ctx, ip.Name)
		if err != nil || len(versions) == 0 {
			continue
		}

		latest := versions[0].Version
		if latest != ip.Version {
			plugin.Version = ip.Version
			plugin.UpdateAvail = true
			updates = append(updates, *plugin)
		}
	}

	return updates, nil
}

// UpdatePlugin updates a plugin to the latest version
func (c *PluginClient) UpdatePlugin(ctx context.Context, name string) (*PluginInstallResult, error) {
	// First uninstall
	c.UninstallPlugin(ctx, name)

	// Then install latest
	return c.InstallPlugin(ctx, name, "latest")
}

// LoadManifest loads the plugin manifest
func (c *PluginClient) LoadManifest(ctx context.Context) (*PluginManifest, error) {
	// Try local manifest first
	manifestPath := c.manifestPath
	if manifestPath == "" {
		// Look in configs directory
		cwd, _ := os.Getwd()
		manifestPath = filepath.Join(cwd, "configs", "plugin-manifest.yaml")
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		// Try remote manifest
		remoteURL := os.Getenv("PLUGIN_MANIFEST_URL")
		if remoteURL != "" {
			return c.loadRemoteManifest(ctx, remoteURL)
		}
		// Return empty manifest
		return &PluginManifest{Version: 1}, nil
	}

	var manifest PluginManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (c *PluginClient) loadRemoteManifest(ctx context.Context, url string) (*PluginManifest, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to load manifest: %s", resp.Status)
	}

	var manifest PluginManifest
	if err := yaml.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// GetSources returns configured plugin sources
func (c *PluginClient) GetSources(ctx context.Context) ([]string, error) {
	manifest, err := c.LoadManifest(ctx)
	if err != nil {
		return nil, err
	}

	sources := []string{}
	for _, org := range manifest.Sources.GitHubOrgs {
		sources = append(sources, fmt.Sprintf("github:%s", org))
	}
	for _, url := range manifest.Sources.ManifestURLs {
		sources = append(sources, fmt.Sprintf("manifest:%s", url))
	}

	return sources, nil
}

// GetHealth returns plugin system health
func (c *PluginClient) GetHealth(ctx context.Context) (*PluginHealth, error) {
	health := &PluginHealth{
		Score:  100,
		Status: "healthy",
	}

	// Get all plugins
	plugins, err := c.ListPlugins(ctx, "")
	if err == nil {
		health.PluginsTotal = len(plugins)
	} else {
		health.Score -= 20
		health.Issues = append(health.Issues, "Failed to load plugin list")
	}

	// Get installed
	installed, err := c.ListInstalled(ctx)
	if err == nil {
		health.PluginsInstalled = len(installed)
	}

	// Check updates
	updates, err := c.CheckUpdates(ctx)
	if err == nil {
		health.UpdatesAvailable = len(updates)
		if len(updates) > 0 {
			health.Recommendations = append(health.Recommendations,
				fmt.Sprintf("%d plugin(s) have updates available", len(updates)))
		}
	}

	// Check GitHub token
	if c.githubToken == "" {
		health.Score -= 10
		health.Recommendations = append(health.Recommendations,
			"Set GITHUB_TOKEN for better API rate limits")
	}

	// Set status
	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	return health, nil
}

// GetTypeConfigs returns plugin type configurations
func (c *PluginClient) GetTypeConfigs() map[PluginType]PluginTypeConfig {
	return c.typeConfigs
}
