# Webb Architecture Analysis

*Research conducted: December 23, 2025*
*Source: ~/example-corp/webb (Platform Operations CLI)*

## Overview

| Attribute | Value |
|-----------|-------|
| Language | Go 1.24.0 |
| Tool Count | 600+ |
| Modules | 42 |
| Total Client Code | 186,187 lines |
| Version | v6.30 |

---

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| CLI Framework | Cobra | Command hierarchy |
| MCP Framework | mark3labs/mcp-go v0.43.1 | Model Context Protocol |
| Deployment | AWS ECS (SSE) + Local (stdio) | Dual transport |
| Auth | GitHub organization membership | Collaborator checks |
| Observability | OpenTelemetry | Tracing, logging, metrics |

---

## Entry Points

### CLI Binary (`cmd/webb/main.go`)
- 30+ command groups organized hierarchically
- Categories: setup, login, k8s, slack, alerts, investigations, migrations, incidents
- Local CLI for interactive development and testing

### MCP Server (`cmd/webb-mcp/main.go`)
- Exposes 600+ tools to Claude via Model Context Protocol
- **Stdio mode:** Local Claude Code integration
- **SSE mode:** HTTP-based for AWS ECS with `/health`, `/metrics`, `/sse`, `/message` endpoints
- Async task event broadcasting via SSE

---

## Modular Tool System (Key Innovation)

### Architecture Pattern
```
ToolRegistry (singleton)
  ├── Module 1 (Kubernetes)
  │    ├── Tool A: webb_k8s_pods
  │    ├── Tool B: webb_k8s_deployments
  │    └── Tool C: webb_k8s_logs
  ├── Module 2 (AWS)
  ├── Module 3 (Slack)
  ├── ... 42 total modules
  └── Auto-registration via init()
```

### Module Structure
Each module in `internal/mcp/tools/<module>/module.go`:

```go
type Module struct{}

func (m *Module) Name() string { return "kubernetes" }
func (m *Module) Description() string { return "Kubernetes operations" }
func (m *Module) Tools() []tools.ToolDefinition { ... }

func init() {
    tools.GetRegistry().RegisterModule(&Module{})
}
```

### ToolDefinition Struct
```go
type ToolDefinition struct {
    Tool             mcp.Tool
    Handler          ToolHandlerFunc
    Category         string
    Subcategory      string
    Tags             []string
    UseCases         []string
    Complexity       ToolComplexity  // simple|moderate|complex
    ThinkingBudget   int             // For extended thinking (>0)
    OutputSchema     *OutputSchema   // For structured outputs
    IsWrite          bool            // For audit logging
    RequiresAuth     bool
    MinRole          string
}
```

### Modules by Category

**Core Domain Modules:**
| Module | Tools | Purpose |
|--------|-------|---------|
| kubernetes | 12 | pods, deployments, events, diagnostics |
| aws | 10 | ECS, RDS, CloudWatch, Lambda, ASG |
| tickets | 14 | Pylon, Shortcut, incident.io |
| slack | 17 | channels, search, threads, DMs, inbox |
| database | 10 | PostgreSQL, ClickHouse, RabbitMQ |

**Meta/Intelligence Modules:**
| Module | Tools | Purpose |
|--------|-------|---------|
| consolidated | 8 | Aggregated health checks (60-70% token savings) |
| investigation | 12 | RCA, pattern matching, symptom correlation |
| health | 6 | Cluster/queue/database health scoring |
| discovery | 3 | Tool discovery, progressive loading |

**New/Specialized Modules:**
| Module | Purpose |
|--------|---------|
| presentations | Google Slides, Nano Banana Pro |
| devtools | Feature planning, scaffolding |
| operations | ROI calculators, runbook management |
| research | Web research, knowledge graph |
| workers | Dev worker pool, task queue |

---

## Registry Capabilities

| Capability | Implementation |
|------------|----------------|
| Tool Lookup | O(1) map-based by name |
| Search | Relevance-ranked (name > tags > category > description) |
| Filtering | Category, subcategory, complexity, tag, write-capability |
| Pagination | Cursor-based for large tool sets |
| Stats | Real-time counts by category/complexity |
| Discovery | Lazy loading mode (~98% token savings) |

---

## Self-Improving Systems

### 1. PerpetualLearner (`perpetual_learner.go`)
Adaptive weight adjustment based on success rates:

```
NewWeight = OldWeight × (1 + LearningRate × (SuccessRate - 0.5))
```

**Configuration:**
| Parameter | Value |
|-----------|-------|
| Learning Rate | 0.2 (increased from 0.1) |
| Min Samples | 3 (reduced from 5) |
| Weight Bounds | 0.1 - 5.0 |

**Tracks:**
- Proposal success/failure outcomes
- PR merge times and review comments
- Source performance over time

### 2. KnowledgeGraphClient (`knowledgegraph.go`)
Transforms Obsidian vault into active intelligence:

**Graph Statistics (Post-Enhancement):**
| Metric | Value |
|--------|-------|
| Edges | 2,881 |
| Semantic Clusters | 14 |
| Avg Connections/Node | 21.18 |

**Edge Types & Weights:**
| Edge Type | Weight | Description |
|-----------|--------|-------------|
| `link` | 1.0 | Wiki-style `[[Page]]` links |
| `tag` | 0.5 | Shared tags |
| `entity` | 0.7 | Shared entity mentions |
| `temporal` | 0.3 | Same date documents |
| `resolution` | 1.0 | Incident → Fix |
| `escalation` | 0.9 | Escalation → Related incident |
| `dependency` | 0.8 | Document → Prerequisite |
| `supersedes` | 0.7 | New → Old document |

**Key Operations:**
- `BuildGraph()` - Extract links, tags, entities
- `Enrich()` - Add missing edges via auto-linking
- `QueryTraversal()` - Find related content
- `PatternDetection()` - Identify RCA patterns

### 3. MultiLLMAnalyzer (`multi_llm.go`)
Parallel analysis across providers with consensus detection:

**Providers:**
| Provider | Model | Use Case |
|----------|-------|----------|
| Claude | claude-sonnet-4 | Native extended thinking |
| OpenAI | o1/o1-mini | Reasoning tasks |
| Gemini | gemini-2.0-flash-thinking | Extended analysis |

**Features:**
- Parallel execution via goroutines
- Consensus detection (strong agreement)
- Divergence analysis (disagreement)
- Combined summary with multi-perspective insights
- Graceful degradation if providers unavailable

---

## Caching & Optimization

### Result Caching (aligned with Claude's 1-hour cache)
| Category | TTL | Purpose |
|----------|-----|---------|
| tool_catalog | 1 hour | Tool definitions |
| k8s_context | 1 hour | kubectl contexts |
| customer | 30 min | Customer configs |
| cost | 30 min | AWS/GCP costs |
| knowledge | 15 min | Knowledge graph queries |
| default | 5 min | Most API calls |
| queues | 1 min | RabbitMQ (freshness needed) |
| alerts | 30 sec | Grafana (critical freshness) |

### Consolidation Tools (60-70% token savings)
| Old Pattern | New Tool | Savings |
|-------------|----------|---------|
| k8s_pods + deployments + events + alerts + queues | `webb_cluster_health_full` | ~65% |
| pylon_list + incidentio_list + shortcut_search | `webb_ticket_summary` | ~60% |
| rabbitmq_queues + job_queue_wait + dlq + stuck | `webb_queue_health_full` | ~70% |

### Discovery-Only Mode (~98% token savings)
```bash
WEBB_LAZY_TOOLS=true  # Enable discovery mode
```
- Register only 3 discovery tools initially
- Models use `webb_tool_discover` to browse
- Load full schemas on-demand with `webb_tool_schema`

---

## Handler Pattern

### Wrapper Function (`registry.go` lines 643-735)
All handlers wrapped with:

1. **Parameter Validation** - Check required fields
2. **OTel Instrumentation** - Span creation, attributes
3. **Panic Recovery** - Capture panics, log traces
4. **Timeout Enforcement** - 30s default
5. **Execution Metrics** - Duration tracking
6. **Structured Logging** - Category, tool, duration

```go
func (r *ToolRegistry) wrapHandler(toolName string, handler ToolHandlerFunc) ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // Validation, OTel, panic recovery, timeout, metrics, logging
        return handler(ctx, req)
    }
}
```

---

## Client Patterns

### Constructor Pattern
```go
client, err := clients.NewK8sClient(clusterContext)
client, err := clients.NewAWSClient(ctx, region)
client, err := clients.NewSlackClient()
```

### Formatter Pattern
```go
output := client.FormatPodList(pods, clusterContext, "full")
return tools.TextResult(output), nil
```

### LLMProvider Interface
```go
type LLMProvider interface {
    Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error)
    GetModel() string
    GetProvider() ProviderType
    SupportsThinking() bool
    SupportsStructuredOutput() bool
}
```

---

## Key Files to Port

| File | Lines | Purpose |
|------|-------|---------|
| `internal/mcp/tools/registry.go` | ~800 | Core registry pattern |
| `cmd/webb-mcp/main.go` | ~200 | MCP server entry |
| `internal/clients/knowledgegraph.go` | ~600 | Knowledge graph |
| `internal/clients/perpetual_learner.go` | ~300 | Learning system |
| `internal/mcp/tools/consolidated/module.go` | ~400 | Consolidation pattern |
| `internal/mcp/tools/discovery/module.go` | ~200 | Tool discovery |

---

## Architectural Innovations

### 1. Progressive Tool Loading
- Expose only discovery tools (3 tools, minimal tokens)
- Relevance-ranked search
- Load schemas on-demand
- 98% token savings

### 2. Consolidation + Scoring
```go
type HealthScore struct {
    Score          int      // 0-100
    Status         string   // healthy|degraded|critical
    IssueCount     int
    Issues         []Issue
    Recommendations []string
}
```

### 3. Knowledge Graph + RCA Linking
- `resolution` edges link problem → solution
- `similarity` edges connect investigations
- Pattern extraction for future issues

### 4. Multi-LLM Consensus
- Parallel execution across providers
- Catches individual hallucinations
- Combined insights from multiple perspectives

### 5. Adaptive Source Weighting
- Success rate drives weight adjustment
- Quick adaptation (min 3 samples)
- Bounded weights (0.1-5.0)

---

## Deployment Patterns

### Local Development (Stdio)
```bash
./webb-mcp  # Stdio mode, direct Claude Code integration
```

### Remote Access (SSE)
```bash
MCP_MODE=sse PORT=8080 ./webb-mcp
# Endpoints: /sse, /message, /health, /metrics
```

### Docker Build (for ECS)
```bash
docker buildx build --platform linux/amd64 -t webb-mcp:latest -f Dockerfile . --load
```

---

## Patterns for hg-mcp

1. **Module-per-domain** with init() auto-registration
2. **Consolidation tools** for token efficiency
3. **Knowledge graph** connecting vault documents
4. **Pattern learning** from successful operations
5. **Handler wrapper** for cross-cutting concerns
6. **Discovery-first** lazy loading
7. **Health scoring** (0-100) for all aggregations
