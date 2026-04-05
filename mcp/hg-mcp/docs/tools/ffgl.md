# ffgl

> FFGL plugin development scaffolding and build tools for Resolume

**7 tools**

## Tools

- [`aftrs_ffgl_build`](#aftrs-ffgl-build)
- [`aftrs_ffgl_create_effect`](#aftrs-ffgl-create-effect)
- [`aftrs_ffgl_create_source`](#aftrs-ffgl-create-source)
- [`aftrs_ffgl_list`](#aftrs-ffgl-list)
- [`aftrs_ffgl_package`](#aftrs-ffgl-package)
- [`aftrs_ffgl_test`](#aftrs-ffgl-test)
- [`aftrs_ffgl_validate_shader`](#aftrs-ffgl-validate-shader)

---

## aftrs_ffgl_build

Build an FFGL plugin for the current platform.

**Complexity:** complex

**Tags:** `ffgl`, `build`, `compile`, `cmake`

**Use Cases:**
- Build plugin binary
- Compile for platform

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Plugin name to build |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_ffgl_create_effect

Scaffold a new FFGL effect plugin with shader template.

**Complexity:** moderate

**Tags:** `ffgl`, `effect`, `create`, `scaffold`

**Use Cases:**
- Create new effect plugin
- Start plugin development

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `author` | string |  | Plugin author name |
| `description` | string |  | Plugin description |
| `name` | string | Yes | Plugin name (no spaces) |

### Example

```json
{
  "author": "example",
  "description": "example",
  "name": "example"
}
```

---

## aftrs_ffgl_create_source

Scaffold a new FFGL source plugin (video generator).

**Complexity:** moderate

**Tags:** `ffgl`, `source`, `create`, `generator`

**Use Cases:**
- Create video source plugin
- Build generator

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `author` | string |  | Plugin author name |
| `description` | string |  | Plugin description |
| `name` | string | Yes | Plugin name (no spaces) |

### Example

```json
{
  "author": "example",
  "description": "example",
  "name": "example"
}
```

---

## aftrs_ffgl_list

List FFGL plugins in the development directory.

**Complexity:** simple

**Tags:** `ffgl`, `plugins`, `list`, `resolume`

**Use Cases:**
- View plugin projects
- Check development status

---

## aftrs_ffgl_package

Package an FFGL plugin for distribution.

**Complexity:** moderate

**Tags:** `ffgl`, `package`, `distribute`, `release`

**Use Cases:**
- Create distribution package
- Prepare for release

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Plugin name to package |
| `version` | string |  | Version string (default: 1.0.0) |

### Example

```json
{
  "name": "example",
  "version": "example"
}
```

---

## aftrs_ffgl_test

Run golden image tests on an FFGL plugin.

**Complexity:** moderate

**Tags:** `ffgl`, `test`, `golden`, `validation`

**Use Cases:**
- Test plugin output
- Validate rendering

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Plugin name to test |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_ffgl_validate_shader

Validate GLSL shader code for FFGL compatibility.

**Complexity:** simple

**Tags:** `ffgl`, `shader`, `glsl`, `validate`

**Use Cases:**
- Check shader syntax
- Validate GLSL code

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | Yes | Path to shader file |

### Example

```json
{
  "path": "example"
}
```

---

