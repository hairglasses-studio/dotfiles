# security

> Security and access control tools for identity, audit logging, and RBAC

**6 tools**

## Tools

- [`security_access_check`](#security-access-check)
- [`security_audit_log`](#security-audit-log)
- [`security_audit_stats`](#security-audit-stats)
- [`security_roles`](#security-roles)
- [`security_user_access`](#security-user-access)
- [`security_whoami`](#security-whoami)

---

## security_access_check

Check if a user has access to a specific tool

**Complexity:** simple

**Tags:** `security`, `rbac`, `access`, `check`

**Use Cases:**
- Verify tool access
- Debug permission issues

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tool` | string | Yes | Tool name to check access for |
| `user` | string |  | Username to check (defaults to current user) |

### Example

```json
{
  "tool": "example",
  "user": "example"
}
```

---

## security_audit_log

View recent audit log entries for tool invocations

**Complexity:** simple

**Tags:** `security`, `audit`, `log`, `history`, `monitoring`

**Use Cases:**
- Review tool usage
- Debug issues
- Security audit

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `errors_only` | boolean |  | Show only error events |
| `limit` | number |  | Maximum number of entries to return (default: 50) |
| `tool` | string |  | Filter by tool name |
| `user` | string |  | Filter by username |

### Example

```json
{
  "errors_only": false,
  "limit": 0,
  "tool": "example",
  "user": "example"
}
```

---

## security_audit_stats

Get summary statistics for audit events

**Complexity:** simple

**Tags:** `security`, `audit`, `statistics`, `metrics`

**Use Cases:**
- View tool usage metrics
- Monitor system activity

---

## security_roles

List all available roles and their permissions

**Complexity:** simple

**Tags:** `security`, `rbac`, `roles`, `permissions`

**Use Cases:**
- View available roles
- Understand permissions

---

## security_user_access

Get access information for a user including roles and permissions

**Complexity:** simple

**Tags:** `security`, `rbac`, `user`, `access`, `permissions`

**Use Cases:**
- View user permissions
- Audit user access

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `user` | string |  | Username to get access info for (defaults to current user) |

### Example

```json
{
  "user": "example"
}
```

---

## security_whoami

Get current user identity including AWS, GitHub, and Kubernetes access info

**Complexity:** simple

**Tags:** `security`, `identity`, `whoami`, `user`, `access`

**Use Cases:**
- Check current user identity
- Verify credentials
- Debug access issues

---

