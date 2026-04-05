# learning

> Pattern learning and troubleshooting tools that improve over time

**8 tools**

## Tools

- [`aftrs_equipment_history`](#aftrs-equipment-history)
- [`aftrs_fix_suggest`](#aftrs-fix-suggest)
- [`aftrs_pattern_learn`](#aftrs-pattern-learn)
- [`aftrs_pattern_match`](#aftrs-pattern-match)
- [`aftrs_rca_generate`](#aftrs-rca-generate)
- [`aftrs_symptom_correlate`](#aftrs-symptom-correlate)
- [`aftrs_troubleshoot`](#aftrs-troubleshoot)
- [`aftrs_venue_patterns`](#aftrs-venue-patterns)

---

## aftrs_equipment_history

Get issue history and reliability score for specific equipment.

**Complexity:** simple

**Tags:** `learning`, `equipment`, `history`, `reliability`

**Use Cases:**
- Check equipment reliability
- Review past issues

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `equipment` | string | Yes | Equipment name |

### Example

```json
{
  "equipment": "example"
}
```

---

## aftrs_fix_suggest

Suggest fixes based on current symptoms and learned patterns.

**Complexity:** simple

**Tags:** `learning`, `fix`, `suggest`, `resolution`

**Use Cases:**
- Get quick fix suggestions
- Troubleshoot issues

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `symptoms` | string | Yes | Comma-separated list of symptoms |

### Example

```json
{
  "symptoms": "example"
}
```

---

## aftrs_pattern_learn

Learn a pattern from a resolved issue. Captures symptoms, root cause, and resolution for future matching.

**Complexity:** moderate

**Tags:** `learning`, `pattern`, `resolution`, `capture`

**Use Cases:**
- Document resolved issues
- Build pattern library

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `equipment` | string |  | Comma-separated list of affected equipment |
| `resolution` | string | Yes | How it was resolved |
| `root_cause` | string | Yes | Root cause of the issue |
| `symptoms` | string | Yes | Comma-separated list of symptoms |
| `venue` | string |  | Venue where issue occurred |

### Example

```json
{
  "equipment": "example",
  "resolution": "example",
  "root_cause": "example",
  "symptoms": "example",
  "venue": "example"
}
```

---

## aftrs_pattern_match

Match current symptoms to learned patterns. Returns best matches with confidence scores.

**Complexity:** simple

**Tags:** `learning`, `pattern`, `match`, `diagnosis`

**Use Cases:**
- Find similar past issues
- Get fix suggestions

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `equipment` | string |  | Equipment involved (optional) |
| `symptoms` | string | Yes | Comma-separated list of current symptoms |
| `venue` | string |  | Current venue (optional) |

### Example

```json
{
  "equipment": "example",
  "symptoms": "example",
  "venue": "example"
}
```

---

## aftrs_rca_generate

Generate a root cause analysis based on symptoms and learned patterns.

**Complexity:** moderate

**Tags:** `learning`, `rca`, `analysis`, `root cause`

**Use Cases:**
- Diagnose complex issues
- Generate post-mortem

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `issue` | string | Yes | Description of the issue |
| `symptoms` | string | Yes | Comma-separated list of symptoms |

### Example

```json
{
  "issue": "example",
  "symptoms": "example"
}
```

---

## aftrs_symptom_correlate

Find symptoms that commonly occur together with the given symptom.

**Complexity:** simple

**Tags:** `learning`, `symptom`, `correlate`, `analysis`

**Use Cases:**
- Find related symptoms
- Predict cascading issues

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `symptom` | string | Yes | The symptom to find correlations for |

### Example

```json
{
  "symptom": "example"
}
```

---

## aftrs_troubleshoot

Get a guided troubleshooting sequence based on the issue and learned patterns.

**Complexity:** moderate

**Tags:** `learning`, `troubleshoot`, `guide`, `steps`

**Use Cases:**
- Step-by-step troubleshooting
- Guided diagnosis

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `issue` | string | Yes | Description of the issue |

### Example

```json
{
  "issue": "example"
}
```

---

## aftrs_venue_patterns

Get venue-specific patterns, common issues, and recommendations.

**Complexity:** simple

**Tags:** `learning`, `venue`, `patterns`, `history`

**Use Cases:**
- Prepare for venue
- Review venue quirks

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `venue` | string | Yes | Venue name |

### Example

```json
{
  "venue": "example"
}
```

---

