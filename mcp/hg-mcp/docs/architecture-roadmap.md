# AFTRS-MCP Architecture Roadmap

> Research findings from analyzing ~/example-corp/webb - a mature MCP server with 1,254 tools

**Last Updated:** January 2026

## Executive Summary

Webb is an enterprise-grade MCP server for SRE/DevOps with sophisticated features that hg-mcp can adopt. Key learnings organized by implementation priority.

---

## Phase 1: Architecture Improvements (Foundation)

### 1.1 Enhanced Tool Registry (HIGH PRIORITY)
**Webb Pattern:** Rich metadata beyond basic tool definitions

```go
type ToolDefinition struct {
    Tool            mcp.Tool
    Handler         ToolHandlerFunc
    Category        string           // Primary grouping
    Subcategory     string           // Finer grouping
    Tags            []string         // Searchable keywords
    UseCases        []string         // Common use cases
    Complexity      ToolComplexity   // simple/moderate/complex
    IsWrite         bool             // State modification flag
    RequiresAuth    bool             // Auth requirement
    MinRole         string           // RBAC minimum role
    Deprecated      bool             // Deprecation tracking
    Successor       string           // Replacement tool
    OutputSchema    *OutputSchema    // Structured output schema
    ThinkingBudget  int              // Extended thinking tokens
}
```

**AFTRS Implementation:** ✅ COMPLETE
- [x] Add `Subcategory`, `UseCases`, `Complexity` to `ToolDefinition`
- [x] Add deprecation tracking fields (`Deprecated`, `Successor`)
- [ ] Add `OutputSchema` for Claude structured outputs (deferred)
- File: `internal/mcp/tools/registry.go`

### 1.2 MCP 2025 Annotations (HIGH PRIORITY)
**Webb Pattern:** Auto-apply annotations based on tool metadata

```go
func applyMCPAnnotations(td *ToolDefinition) {
    td.Tool.Annotations.Title = toolNameToTitle(td.Tool.Name)
    td.Tool.Annotations.ReadOnlyHint = !td.IsWrite
    td.Tool.Annotations.DestructiveHint = td.IsWrite
    td.Tool.Annotations.IdempotentHint = !td.IsWrite
    td.Tool.Annotations.OpenWorldHint = true
}
```

**AFTRS Implementation:** ✅ COMPLETE
- [x] Add `IsWrite` field to tool definitions
- [x] Auto-apply annotations in `RegisterWithServer()` via `applyMCPAnnotations()`
- File: `internal/mcp/tools/registry.go`

### 1.3 Request Context Pattern (MEDIUM PRIORITY)
**Webb Pattern:** Per-request isolation for multi-user

```go
type RequestContext struct {
    UserID      string
    DisplayName string
    Roles       []Role
    RequestID   string
    SessionID   string
    TokenSource string // "jwt", "api_key", "local"
}
```

**AFTRS Implementation:**
- [ ] Create `internal/mcp/context.go`
- [ ] Thread context through tool handlers
- [ ] Add request ID generation

### 1.4 Hook System (MEDIUM PRIORITY)
**Webb Pattern:** Pre/post execution hooks for audit, auth, tracking

```go
type HookManager struct {
    preHooks  []PreExecuteHook
    postHooks []PostExecuteHook
}
```

**AFTRS Implementation:**
- [ ] Create `internal/mcp/hooks.go`
- [ ] Add audit logging hook
- [ ] Add timing/metrics hook

---

## Phase 2: Progressive Tool Discovery (Token Efficiency)

### 2.1 Discovery Tools (HIGH PRIORITY) ✅ COMPLETE
**Webb Pattern:** 4-level progressive disclosure

| Tool | Tokens | Returns |
|------|--------|---------|
| `webb_tool_discover` (names) | ~500 | Tool names only |
| `webb_tool_discover` (signatures) | ~800 | Name + required params |
| `webb_tool_discover` (descriptions) | ~2000 | Full descriptions |
| `webb_tool_schema` | On-demand | Complete schemas |

**AFTRS Implementation:** ✅ COMPLETE (12 tools)
- [x] `aftrs_tool_discover` with `detail_level` parameter (names/signatures/descriptions/full)
- [x] `aftrs_tool_schema` for on-demand schema loading
- [x] `aftrs_tool_stats` for registry statistics
- [x] `aftrs_tool_search` for keyword search
- [x] `aftrs_tool_catalog` for category browsing
- [x] `aftrs_tools_related` for finding related tools
- [x] `aftrs_tools_workflow` for recommended sequences
- [x] `aftrs_tools_by_system` for system-based filtering
- [x] `aftrs_tools_recent` for session history
- [x] `aftrs_tools_favorites` for favorite management
- [x] `aftrs_tools_alias` for tool shortcuts
- File: `internal/mcp/tools/discovery/module.go`

### 2.2 Tool Catalog & Tree View ✅ COMPLETE

**Webb Pattern:** Hierarchical browsing

- [x] `aftrs_tool_catalog` - browse by category/subcategory
- [x] `aftrs_tools_related` - find related tools
- [ ] `aftrs_tool_tree` - visual hierarchy (deferred)

---

## Phase 3: Gateway/Consolidated Tools (Mega-Tools)

### 3.1 Gateway Pattern (HIGH PRIORITY) ✅ COMPLETE

**Webb Pattern:** Single entry point for domain operations

```go
// Instead of 15 separate AWS tools:
Tool: mcp.Tool{
    Name: "webb_aws",
    InputSchema: mcp.ToolInputSchema{
        Properties: map[string]interface{}{
            "domain": {"enum": ["ecs", "rds", "ec2", "elb", "s3", "lambda"]},
            "action": {"description": "Operation within domain"},
        },
    },
}
```

**Webb has 20 gateway tools:**
- `webb_aws` - AWS operations
- `webb_slack` - Slack operations
- `webb_k8s` - Kubernetes operations
- `webb_devops` - DevOps operations
- `webb_tickets` - Unified tickets
- etc.

**AFTRS Gateways:** ✅ COMPLETE (5 tools)
- [x] `aftrs_dj` - Unified DJ operations (Serato, Rekordbox, Traktor)
- [x] `aftrs_av` - AV control (Resolume, TouchDesigner, OBS)
- [x] `aftrs_lighting` - Lighting (grandMA3, DMX, WLED)
- [x] `aftrs_audio` - Audio (Ableton, Dante, MIDI)
- [x] `aftrs_streaming` - Streaming (Twitch, YouTube, NDI)
- File: `internal/mcp/tools/gateway/module.go`

### 3.2 Consolidated Health Tools ✅ COMPLETE

**Webb Pattern:** Multi-source aggregation

| Webb Tool | Sources Combined |
|-----------|------------------|
| `webb_cluster_health_full` | K8s + queues + DB + alerts |
| `webb_ticket_summary` | Pylon + Shortcut + Incident.io |
| `webb_standup_briefing` | All sources for morning |

**AFTRS Consolidated Tools:** ✅ COMPLETE (8 tools)

- [x] `aftrs_studio_health_full` - All systems status (TD + Resolume + DMX + NDI + UNRAID)
- [x] `aftrs_show_preflight` - Pre-show checklist
- [x] `aftrs_morning_check` - Daily status
- [x] `aftrs_stream_dashboard` - Streaming overview
- [x] `aftrs_performance_overview` - Visual systems status
- [x] `aftrs_investigate_show` - Show investigation
- [x] `aftrs_pre_stream_check` - Pre-stream checklist
- [x] `aftrs_equipment_audit` - Equipment inventory
- File: `internal/mcp/tools/consolidated/module.go`

---

## Phase 4: Workflow/Chain System ✅ COMPLETE

### 4.1 Chain Executor (HIGH VALUE) ✅ COMPLETE

**Webb Pattern:** Multi-step workflows with gates

```go
type Chain struct {
    ID          string
    Name        string
    Description string
    Steps       []ChainStep
    Parameters  []ChainParameter
    Triggers    []ChainTrigger  // cron, event, manual
}

type ChainStep struct {
    Type     string  // "tool", "chain", "parallel", "branch", "gate"
    Tool     string  // Tool to execute
    Inputs   map[string]interface{}
    OnError  string  // "stop", "continue", "retry"
}
```

**AFTRS Chain Examples:** ✅ IMPLEMENTED

- [x] `show_startup` - Power on -> Check systems -> Load project -> Test outputs
- [x] `stream_start` - OBS scene -> Go live -> Notify Discord
- [x] `backup_daily` - Stop services -> Backup -> Verify -> Restart

**Implementation:** ✅ COMPLETE (8 MCP tools)

- [x] `internal/chains/types.go` - Chain, ChainStep, ChainParameter types
- [x] `internal/chains/executor.go` - ChainExecutor with gate support
- [x] `aftrs_chain_list` - List available chains
- [x] `aftrs_chain_get` - Get chain details
- [x] `aftrs_chain_execute` - Execute a chain
- [x] `aftrs_chain_status` - Check execution status
- [x] `aftrs_chain_approve` - Approve/reject gates
- [x] `aftrs_chain_cancel` - Cancel execution
- [x] `aftrs_chain_pending` - List pending gates
- [x] `aftrs_chain_history` - View execution history
- File: `internal/mcp/tools/chains/module.go`

### 4.2 Event-Driven Triggers
**Webb Pattern:** Chains triggered by events

```go
type ChainTrigger struct {
    Type     string  // "cron", "event", "manual"
    Schedule string  // "0 9 * * *" for cron
    Event    string  // "show.started", "stream.ended"
}
```

---

## Phase 5: Knowledge Graph & Vault

### 5.1 Knowledge Graph (MEDIUM PRIORITY)
**Webb Pattern:** Entity relationships for context

```go
type KnowledgeGraph struct {
    nodes map[string]*Node  // incident, runbook, customer, etc.
    edges map[string][]Edge // relationships
}

type Edge struct {
    Source   string
    Target   string
    Type     string  // "link", "tag", "similarity", "temporal"
    Weight   float64
}
```

**AFTRS Graph Entities:**
- Shows, Venues, Setlists, Tracks, Equipment, Presets

**Implementation:**
- [ ] Create `internal/clients/knowledgegraph.go`
- [ ] Add `aftrs_graph_link`, `aftrs_graph_search`
- [ ] Link to Obsidian vault

### 5.2 Session Memory (MEDIUM PRIORITY)
**Webb Pattern:** 4-component memory system

1. **User memories** - #remember team knowledge
2. **Session insights** - Auto-generated learnings
3. **Symptom-resolution mappings**
4. **Pitfalls and best practices**

**Implementation:**
- [ ] Create `internal/clients/session_memory.go`
- [ ] Add `aftrs_memory_remember`, `aftrs_memory_search`

---

## Phase 6: Advanced Features

### 6.1 Self-Healing/Remediation
**Webb Pattern:** Automated response to issues

```go
type RemediationPlaybook struct {
    ID          string
    Name        string
    Trigger     string  // "high_cpu", "stream_dropped"
    RiskScore   int     // 0-100
    Steps       []RemediationStep
    Cooldown    time.Duration
}
```

**AFTRS Playbooks:**
- [ ] `restart_crashed_service` - Auto-restart OBS/Resolume
- [ ] `reconnect_ndi_source` - Re-establish NDI
- [ ] `failover_stream` - Switch to backup

### 6.2 MCP Federation
**Webb Pattern:** Connect to remote MCP servers

```go
type FederatedTool struct {
    ServerURL   string
    ToolName    string
    AuthType    string  // "oauth", "apikey"
}
```

**AFTRS Use Cases:**
- [ ] Federate TouchDesigner MCP tools
- [ ] Connect to home-assistant MCP
- [ ] Integrate with external services

### 6.3 Research/Swarm System
**Webb Pattern:** Autonomous research workers

- 15+ worker types running in parallel
- Pattern discovery from usage
- Automatic improvement suggestions

**AFTRS Application:**
- [ ] Track troubleshooting patterns
- [ ] Learn from show issues
- [ ] Suggest automation improvements

---

## Phase 7: Observability Enhancements

### 7.1 OpenTelemetry (Already Have - Enhance)
**Webb Pattern:** Comprehensive instrumentation

- [ ] Add per-tool latency histograms
- [ ] Track tool success/failure rates
- [ ] Add token usage metrics

### 7.2 Feedback Loop
**Webb Pattern:** Tool routing optimization

```go
type FeedbackMetrics struct {
    ToolName     string
    SuccessRate  float64
    AvgLatency   time.Duration
    BoostScore   float64  // Routing preference
}
```

---

## Implementation Priority Matrix

| Phase | Feature | Effort | Value | Priority | Status |
|-------|---------|--------|-------|----------|--------|
| 1.1 | Enhanced Registry | Low | High | P0 | ✅ COMPLETE |
| 1.2 | MCP Annotations | Low | High | P0 | ✅ COMPLETE |
| 2.1 | Progressive Discovery | Medium | High | P0 | ✅ COMPLETE |
| 3.1 | Gateway Tools | Medium | High | P1 | ✅ COMPLETE |
| 4.1 | Chain System | High | High | P1 | ✅ COMPLETE |
| 3.2 | Consolidated Tools | Low | Medium | P1 | ✅ COMPLETE |
| 5.1 | Knowledge Graph | High | Medium | P2 | ✅ COMPLETE |
| 6.1 | Self-Healing | High | Medium | P2 | ✅ COMPLETE |
| 5.2 | Session Memory | Medium | Medium | P2 | ✅ COMPLETE |
| 6.2 | MCP Federation | Medium | Low | P3 | ✅ COMPLETE |

---

## Tool Count Comparison

| Metric | Webb | AFTRS Current | AFTRS Target |
|--------|------|---------------|--------------|
| Total Tools | 1,254 | 719 | 1,000+ |
| Categories | 57 | 72 | 80+ |
| Gateways | 20 | 0 | 10+ |
| Discovery Tools | 31 | 9 | 15+ |

---

## Key Files to Reference

**Webb Architecture:**
- `~/example-corp/webb/internal/mcp/tools/registry.go` (2047 lines)
- `~/example-corp/webb/internal/mcp/context.go`
- `~/example-corp/webb/internal/mcp/hooks.go`
- `~/example-corp/webb/internal/chains/executor.go`
- `~/example-corp/webb/internal/clients/knowledgegraph.go` (130KB)
- `~/example-corp/webb/internal/clients/self_healing.go` (98KB)

**AFTRS Files to Modify:**
- `internal/mcp/tools/registry.go` - Enhance ToolDefinition
- `internal/mcp/tools/discovery/module.go` - Progressive discovery
- `internal/mcp/tools/consolidated/module.go` - Gateway tools
- New: `internal/chains/` - Workflow system
- New: `internal/clients/knowledgegraph.go`

---

## Next Steps

1. **Immediate (P0):** Enhanced registry + MCP annotations + progressive discovery
2. **Short-term (P1):** Gateway tools for DJ/AV/Lighting + chain system
3. **Medium-term (P2):** Knowledge graph + session memory
4. **Long-term (P3):** MCP federation + research swarm

---

## Contributing

See [CLAUDE.md](../CLAUDE.md) for development guidelines.
