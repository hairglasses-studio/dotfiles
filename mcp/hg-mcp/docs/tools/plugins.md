# plugins

> GitHub-based plugin registry for TouchDesigner and Resolume plugins

**12 tools**

## Tools

- [`aftrs_plugin_check_updates`](#aftrs-plugin-check-updates)
- [`aftrs_plugin_health`](#aftrs-plugin-health)
- [`aftrs_plugin_info`](#aftrs-plugin-info)
- [`aftrs_plugin_install`](#aftrs-plugin-install)
- [`aftrs_plugin_installed`](#aftrs-plugin-installed)
- [`aftrs_plugin_list`](#aftrs-plugin-list)
- [`aftrs_plugin_refresh`](#aftrs-plugin-refresh)
- [`aftrs_plugin_search`](#aftrs-plugin-search)
- [`aftrs_plugin_sources`](#aftrs-plugin-sources)
- [`aftrs_plugin_uninstall`](#aftrs-plugin-uninstall)
- [`aftrs_plugin_update`](#aftrs-plugin-update)
- [`aftrs_plugin_versions`](#aftrs-plugin-versions)

---

## aftrs_plugin_check_updates

Check for available plugin updates.

**Complexity:** simple

**Tags:** `plugins`, `updates`, `check`

**Use Cases:**
- Check for updates
- See available upgrades

---

## aftrs_plugin_health

Check plugin system health.

**Complexity:** simple

**Tags:** `plugins`, `health`, `status`

**Use Cases:**
- Check plugin system health
- Diagnose issues

---

## aftrs_plugin_info

Get detailed information about a plugin.

**Complexity:** simple

**Tags:** `plugins`, `info`, `details`

**Use Cases:**
- View plugin details
- Check plugin compatibility

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Plugin name |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_plugin_install

Install a plugin from the registry.

**Complexity:** moderate

**Tags:** `plugins`, `install`, `download`

**Use Cases:**
- Install a plugin
- Add new functionality

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Plugin name to install |
| `version` | string |  | Version to install (default: latest) |

### Example

```json
{
  "name": "example",
  "version": "example"
}
```

---

## aftrs_plugin_installed

List all installed plugins.

**Complexity:** simple

**Tags:** `plugins`, `installed`, `local`

**Use Cases:**
- View installed plugins
- Check what's installed

---

## aftrs_plugin_list

List all available plugins from the registry.

**Complexity:** simple

**Tags:** `plugins`, `list`, `registry`, `touchdesigner`, `resolume`

**Use Cases:**
- Browse available plugins
- Find TD or Resolume plugins

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `filter` | string |  | Filter by name, type, or tag |
| `type` | string |  | Filter by plugin type: touchdesigner_tox, touchdesigner_component, resolume_effect, resolume_source |

### Example

```json
{
  "filter": "example",
  "type": "example"
}
```

---

## aftrs_plugin_refresh

Refresh plugin cache from sources.

**Complexity:** simple

**Tags:** `plugins`, `refresh`, `cache`

**Use Cases:**
- Refresh plugin list
- Update cache

---

## aftrs_plugin_search

Search for plugins by keyword.

**Complexity:** simple

**Tags:** `plugins`, `search`, `find`

**Use Cases:**
- Find plugins by keyword
- Search for specific functionality

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `query` | string | Yes | Search query (matches name, description, tags) |

### Example

```json
{
  "query": "example"
}
```

---

## aftrs_plugin_sources

List configured plugin sources.

**Complexity:** simple

**Tags:** `plugins`, `sources`, `config`

**Use Cases:**
- View plugin sources
- Check configured repositories

---

## aftrs_plugin_uninstall

Uninstall an installed plugin.

**Complexity:** moderate

**Tags:** `plugins`, `uninstall`, `remove`

**Use Cases:**
- Remove a plugin
- Clean up unused plugins

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Plugin name to uninstall |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_plugin_update

Update a plugin to the latest version.

**Complexity:** moderate

**Tags:** `plugins`, `update`, `upgrade`

**Use Cases:**
- Update plugin to latest
- Get new features

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Plugin name to update |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_plugin_versions

List available versions for a plugin.

**Complexity:** simple

**Tags:** `plugins`, `versions`, `releases`

**Use Cases:**
- Check available versions
- View release history

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Plugin name |

### Example

```json
{
  "name": "example"
}
```

---

