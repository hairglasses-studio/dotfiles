# opnsense-llmagent

**Repository Path:** `/Users/mitch/Docs/aftrs-void/opnsense-llmagent`
**Category:** Llm Agents
**Size:** 0.13 MB
**Last Consolidated:** 2025-09-23 00:34:08

## Overview

# OPNsense LLM Agent Plugin (MCP Server)

An OPNsense plugin that runs an **MCP (Model Context Protocol)** server to expose
router tools to AI agents via the **OpenAI Agents Python SDK**. Agents can query
status and safely perform actions such as listing firewall rules, restarting
services via `configctl`, viewing logs, running diagnostics, and (optionally)
rebooting or updating the system.

> ⚠️ **Dangerous operations** (e.g., reboot, halt, updates, arbitrary shell exec)
> are **opt‑in** and must be enabled explicitly in the plugin settings or via
> environment variables. Defaults are **read‑only** plus safe service control.

## Features (current)

- MCP server using stdio transport (local process) with a tool registry
- Implemented tool categories:
  - **Firewall (read‑only)**: list rules via `pfctl`
  - **NAT (read‑only)**: list NAT/rdr rules and state table via `pfctl`
  - **Interfaces (read‑only)**: list interfaces, VLANs, and routes via `ifconfig`/`netstat`
  - **Services**: restart/stop/status for a safe whitelist using `configctl`
  - **System**: uptime, version info, tail logs (bounded), disk/memory stats
  - **Network diagnostics**: ping, traceroute
  - **Backups**: snapshot `/conf/config.xml` (+ pf rules dumps) into `/root/backups/opnsense`
  - **Plugin dev scaffolding**: create a skeleton for a new plugin in `/root/dev_plugins`
- Minimal OPNsense MVC scaffolding with Settings page (enable/disable server)
- `rc.d`+`configd` service wiring to launch the MCP server as a daemon
- Makefile for **dev install**, start/stop/status, and template reload

## Planned (roadmap excerpt)

- Full read/write firewall & NAT via OPNsense API/models
- Interface/VLAN management
- Backups & restore hooks
- SSE transport option (HTTP) for remote agents
- Fine‑grained RBAC/approval layers for destructive operations
- Test harness (tox/pytest) and GitHub Actions CI

## Quick start (DEV)

> Run on an OPNsense box (or a close FreeBSD 13/14 userland).

```sh
# 1) install deps (python, pip, agents sdk)
pkg install -y python3 py311-pip || true
pip3 install --upgrade pip
pip3 install -r scripts/llmagent/requirements.txt

# 2) install plugin scaffolding (copies files into /usr/local/opnsense/*)
make install

# 3) start service (daemonized MCP server)
make start
make status

# 4) connect an agent using MCP stdio
# e.g. in your agent code:
# from agents.mcp import MCPServerStdio
# subprocess to: /usr/local/opnsense/scripts/llmagent/agent_mcp.py
```

## Settings

In the OPNsense GUI (Services → **LLM Agent**), toggle **Enable server**.
You can also set environment flags in `/usr/local/opnsense/scripts/llmagent/env`:

```sh
LLMAGENT_ALLOW_DANGEROUS=0    # set to 1 to allow reboot/halt/update/exec
LLMAGENT_SERVICE_WHITELIST="unbound,dhcpd,sshd"  # comma list
LLMAGENT_LOG_TAIL_MAX=2000    # safety bound for log lines
```

## Security Notes

- The MCP server **does not bind a network port** by default. It runs as a local
  process launched by an orchestrator or via SSH. To expose a port intentionally,
  set `LLMAGENT_BIND="127.0.0.1:8822"` and use a firewall.
- Destructive tools require explicit enablement (`LLMAGENT_ALLOW_DANGEROUS=1`).

## Uninstall (DEV)

```sh
make stop || true
make uninstall
```

---

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) and [`docs/ROADMAP.md`](docs/ROADMAP.md) for details.


### Example tool calls

```python
# NAT rules
await client.call_tool("nat_list", {"verbose": True})

# Interfaces / VLANs / routes
await client.call_tool("interfaces_list", {})

# Backup config
await client.call_tool("config_backup", {"include_pf_dumps": True})
```


## Write paths (proposals → API apply)

To avoid clobbering OPNsense-generated `pf.conf`, **writes** are handled in two stages:

1) **Propose** a change: it’s saved under `/root/llmagent/pending/*.json`  
2) **Apply** via API: generate a set of REST calls you can run with your OPNsense API key

```python
# Firewall create (proposal)
await client.call_tool("firewall_propose_rule", {
  "action":"pass", "interface":"lan", "proto":"tcp",
  "source":"10.0.1.0/24", "destination":"192.168.0.5/32", "dport":"22",
  "description":"Allow SSH from DMZ"
})
# NAT outbound (proposal)
await client.call_tool("nat_propose_rule", {
  "mode":"outbound", "interface":"wan", "proto":"any",
  "source":"10.0.1.0/24", "description":"DMZ outbound NAT"
})
# Review pending
await client.call_tool("firewall_list_proposals", {})
await client.call_tool("nat_list_proposals", {})
# Generate apply steps (returns curl plan)
await client.call_tool("firewall_apply_via_api", {
  "api_base": "https://opnsense.local", "key":"APIKEY", "secret":"APISECRET"
})
await client.call_tool("nat_apply_via_api", {
  "api_base": "https://opnsense.local", "key":"APIKEY", "secret":"APISECRET"
})
```

## VLAN write (ephemeral)

```python
await client.call_tool("vlan_create_ephemeral", {"parent":"igb0", "vlan_id": 30, "name":"dmz30"})
await client.call_tool("vlan_delete_ephemeral", {"name":"dmz30"})
```

## Restore a backup

```python
await client.call_tool("config_restore", {"backup_filename":"config-20250919-120001.xml", "reboot": False})
```

## Optional HTTP/SSE transport (experimental)

You can run a simple HTTP server for external agents:

```sh
# enable and install deps
pip3 install fastapi uvicorn sse-starlette
# set bind (default 127.0.0.1:8822)
sysrc llmagent_http_enable=YES
make start   # or: service llmagent_http start
curl http://127.0.0.1:8822/tools
```

This is not a strict MCP-over-SSE implementation yet, but is adequate for local
automation and can be evolved to a spec‑compliant transport.


## Direct writers (no proposal)

Where the official API supports it, you can write rules directly. Since API coverage varies by version, tools accept a **strategy**:

- `strategy="api"`: Attempts known endpoints (e.g., `/api/firewall/filter/addRule`) and applies.
- `strategy="xml"`: Carefully writes to `/conf/config.xml` (backup first), then reloads templates and filter.

### Examples

```python
await client.call_tool("firewall_rule_create_direct", {
  "strategy":"api",
  "action":"pass","interface":"lan","protocol":"tcp",
  "source":"10.0.1.0/24","destination":"192.168.0.5/32","dst_port":"22",
  "descr":"Allow SSH from DMZ",
  "api_base":"https://opnsense.local","key":"APIKEY","secret":"APISECRET"
})

await client.call_tool("nat_rule_create_direct", {
  "strategy":"xml","mode":"port_forward",
  "interface":"wan","protocol":"tcp",
  "source":"","destination":"192.168.0.5","dst_port":"22","descr":"SSH forward"
})

await client.call_tool("vlan_create_persistent_direct", {
  "strategy":"xml","parent":"igb0","vlan_id":30,"ifname":"OPT_DMZ30","descr":"DMZ 30"
})

await client.call_tool("config_templates_reload", {})
```

## Future Additions (Recommended)

This section consolidates the recommended next steps into concrete, shippable tracks. Items are grouped for planning; each has clear outcomes and acceptance criteria.

### 1) Protocol & Transport
- **Spec‑compliant MCP over SSE/WebSocket**: replace the simple FastAPI JSON with a standards‑compliant MCP transport (SSE first). Include connection auth, TLS, per‑tool rate limits, and concurrency gates.
- **Detach/reattach sessions**: support ephemeral sessions for long‑running jobs with resumable event streams.
- **Agent discovery**: optional mDNS/Tailscale advertising of the agent endpoint for your studio network.

### 2) Security, RBAC, and Approvals
- **OPNsense ACL integration**: map plugin API to user roles; per‑tool allow/deny with scopes (e.g., `service:*`, `firewall:read`, `nat:write`).
- **Approval flows for dangerous ops**: one‑time tokens (UI‑issued), time‑boxed grants, and change windows (maintenance windows).
- **Secrets management**: store external API keys (if any) with rotation reminders; redact secrets in logs; KDF‑based at‑rest protection.

### 3) Reliability & HA
- **CARP/HA pairs**: agent becomes HA‑aware; only the master (CARP) executes writes; passive node stays read‑only.
- **Leader‑elected apply**: queued writes apply on master with back‑pressure; audit mirrored to backup.
- **Locking**: per‑domain locks (firewall, nat, interfaces) to prevent conflicting concurrent changes.

### 4) Observability & Audit
- **Structured JSON logs** with request‑id, tool name, args schema version, caller, and latency stats.
- **Audit journal**: append‑only signed log of tool calls, diffs, outcomes, and approvals (export to JSONL).
- **Metrics**: Prometheus exporter (requests, failures, latencies, queue depth), basic health endpoints, OpenTelemetry traces.
- **SIEM hooks**: syslog/Elastic output for change events and anomalies.

### 5) Web UI/UX
- **Tool catalog UI**: searchable list of tools, JSON Schema hints, example payloads, dry‑run preview.
- **Diff & impact analysis**: render config diffs (before/after), pf rules preview, NAT/route impact graphs.
- **Approval center**: pending requests queue with “allow once / deny / schedule” actions and expiry timers.
- **Job console**: tail logs, stream MCP events, show exit codes and artifacts (e.g., backup file paths).

### 6) Tool Coverage (Core Network & Security)
- **Firewall (full CRUD)** via model API with idempotent upserts, groups/aliases, ordering controls, anchors.
- **NAT (CRUD)**: outbound/manual, port forwards, one‑to‑one; conflict detection and auto‑fix suggestions.
- **Interfaces/VLANs (persistent)**: full lifecycle with validation against parent drivers and MTU constraints.
- **DHCP (v4/v6)**: scope options, static mappings, reservations, and lease insight tools.
- **DNS/Unbound**: overrides, views, split‑horizon helpers (align with your Tailscale split‑DNS).
- **VPNs (WireGuard/OpenVPN/IPsec)**: bring‑up/tear‑down, peer management, key rotation helpers.
- **IDS/IPS (Suricata) & Zenarmor**: toggle rulesets, update feeds, quick “quarantine subnet” tool.
- **ACME/Certificates**: cert issuance/renewal status; reload dependent services (HAProxy, GUI).

### 7) Safety, Testing & Simulation
- **Dry‑run engine**: evaluate proposed changes against current config; compute diffs and warnings (port overlaps, shadowed rules).
- **Reachability checks**: simulate rule/NAT effects for tuples (src,dst,proto,port); preflight diagnostics.
- **Golden tests**: corpus of sample configs; golden diffs; regression tests.
- **VM E2E**: QEMU/VMX CI environment to validate writes against a live OPNsense image.

### 8) Packaging & Release Engineering
- **Proper `os-llmagent` package**: plugin metadata, pkg build scripts, signed artifacts.
- **FreeBSD port** for dependencies and reproducible builds.
- **CI/CD**: GitHub Actions to lint (PHP/Python), test, build pkg, and attach artifacts to GitHub Releases.
- **SemVer + CHANGELOG** with migration notes and rollback guidance.

### 9) Integrations
- **GitOps mode**: push config diffs to a Git repo (aftrs‑shell org); PR‑based approvals with signed commits.
- **Chat integrations**: Slack/Discord webhooks for approvals, alerts, and summaries.
- **Home Assistant / Node‑RED**: trigger common ops, display health widgets, receive events.
- **Prometheus/Grafana dashboards**: prebuilt panels for approvals, failures, and device health.

### 10) Performance & Limits
- **Async tool runner** with bounded concurrency, per‑category queues, and rate limits.
- **Chunked log streaming** for large tail operations; cutoff guards; exponential backoff on failures.
- **Resource caps**: soft memory ceiling for the agent process; watchdog restart on leak detection.

### 11) Reboot/Upgrade Safety
- **Maintenance windows**: schedule reboots/updates; ensure snapshot/backup preconditions are met.
- **Preflight checks**: verify persistence, free disk, package health; pause on critical alarms.
- **Rollback**: quick restore to last good config with auto‑apply and reboot option.

### 12) LLM Safety & Policy
- **Tool allowlist per role** with explicit dangerous‑op gating.
- **Prompt‑injection hardening**: strip/validate inputs, canonicalize hosts/ports, length limits.
- **Content redaction**: PII/IPs in logs can be hashed; secrets never echoed.
- **Risk scoring**: require higher approvals for unusual patterns (e.g., “allow any to WAN”).

### 13) Plugin Dev Acceleration
- **`plugin init` generator**: full MVC skeleton + configd + rc.d with tests.
- **`plugin publish`**: tag, changelog generator, pkg build, GitHub Release, checksum manifest.
- **Code style**: PHP PSR‑12, Python black/ruff, pre‑commit hooks.

### 14) Stable Schemas & API
- **Versioned tool schemas** using JSON Schema; reject unknown fields; sensible defaults.
- **Compatibility matrix** by OPNsense release; tool behavior gates based on detected version.
- **Typed results**: machine‑friendly outputs for downstream agents/dashboards.

### 15) Documentation & Playbooks
- **Threat model** & hardening guide (firewalling the agent, key handling, ACLs).
- **Runbooks**: common tasks (e.g., “open 443 to host X for 2h”), scheduled expirations, audit.
- **Examples**: complete MCP client snippets (Python/JS), plus CI usage.
- **Diagrams**: architecture + data flow; approval states; HA/CARP behavior.


## Documentation Files (4)

### README.md

*Large file (13080 bytes) - see original at `/Users/mitch/Docs/aftrs-void/opnsense-llmagent/README.md`*

### docs/ARCHITECTURE.md

```
# Architecture

## Overview

This plugin follows OPNsense’s MVC layout and uses **configd** for privileged
operations. The **MCP server** (Python) is launched as a daemon via an `rc.d`
script and exposes router operations as **tools** to LLM agents adhering to the
Model Context Protocol.

```
opnsense-llmagent/
├─ src/opnsense/mvc/app/
│  ├─ controllers/OPNsense/LLMAgent/
│  │  ├─ IndexController.php
│  │  └─ Api/{SettingsController,ServiceController}.php
│  ├─ models/OPNsense/LLMAgent/{LLMAgent.php,LLMAgent.xml,Menu/Menu.xml}
│  └─ views/OPNsense/LLMAgent/index.volt
├─ service/conf/actions.d/actions_llmagent.conf
├─ service/templates/rc.d/rc.llmagent
├─ scripts/llmagent/{agent_mcp.py,lib/*.py,requirements.txt,env}
└─ Makefile
```

### Why Python + OpenAI Agents SDK?

- First‑class support for **MCP** transports and tool definitions
- Batteries‑included ergonomics for registering tools and arguments
- Easier to evolve and test compared to FreeBSD‑specific C/Go shims

### Why configd?

- Canonical, auditable way to perform privileged actions from OPNsense UI/API
- Uniform mapping of “service actions” to rc scripts (`start/restart/status`)
- Centralized permission point (future: wire to GUI ACL)

## Data Flow

1. Admin enables plugin (checkbox) → `ServiceController` starts `rc.llmagent` via
   `configctl llmagent start` (configd reads `actions_llmagent.conf`).
2. Daemon (`rc.llmagent`) launches `agent_mcp.py` with environment from `env`.
3. An external agent connects via **stdio** transport (local process / SSH) and
   calls tools like `service_restart`, `firewall_list_rules`, `log_tail`…
4. Each tool uses a safe wrapper to call `configctl`, `pfctl`, etc.

## Extending Tools

Add functions in `scripts/llmagent/lib/tools_*.py` and register them in
`agent_mcp.py`. Prefer **read‑only** tools first, then add controlled writers.

```

### docs/ROADMAP.md

```
# Roadmap

## v0.1.0 (this zip)
- OPNsense MVC scaffolding (UI enable toggle)
- rc.d + configd actions for service management
- MCP server (stdio) with initial tools:
  - firewall_list_rules (read‑only)
  - service_restart / service_status / service_stop (whitelist)
  - system_info / system_uptime / log_tail (bounded)
  - net_ping / net_traceroute (bounded)
  - plugin_scaffold (dev helper under /root/dev_plugins)
- Makefile targets for install/start/stop/status/uninstall

## v0.2.0 (delivered)
- NAT and interface read‑only inventory (tools: `nat_list`, `interfaces_list`)
- Safe config backup tool (tool: `config_backup`) – writes to /root/backups/opnsense
- (Planned) SSE transport option (disabled by default)
- (Planned) Approvals UX for dangerous ops

## v0.3.0
- Write APIs for firewall/NAT via OPNsense models
- VLAN management (create/delete)
- Service discovery (enumerate configd‑managed services)
- Tests and CI

## v0.4.0 and beyond
- Web UI dashboard for tool runs + logs
- RBAC integration with OPNsense ACLs
- Multi‑agent sessions and audit trails
- In‑place plugin generator (full skeleton with tests)

```

### scripts/llmagent/requirements.txt

```
openai-agents>=0.1.0
psutil>=5.9.8
pydantic>=2.7.0

requests>=2.32.2

```

## Configuration Files (2)

- `service/conf/actions.d/actions_llmagent.conf`
- `service/conf/actions.d/actions_llmagent_http.conf`

## OPNsense XML Configurations (3)

- `src/opnsense/mvc/app/models/OPNsense/LLMAgent/LLMAgent.xml`
- `src/opnsense/mvc/app/models/OPNsense/LLMAgent/Menu/Menu.xml`
- `src/opnsense/mvc/app/controllers/OPNsense/LLMAgent/forms/general.xml`

## Scripts (23)

- `scripts/llmagent/agent_http.py`
- `scripts/llmagent/agent_mcp.py`
- `scripts/llmagent/lib/tools_firewall.py`
- `scripts/llmagent/lib/tools_restore.py`
- `scripts/llmagent/lib/opnsense_api.py`
- `scripts/llmagent/lib/pending_store.py`
- `scripts/llmagent/lib/tools_services.py`
- `scripts/llmagent/lib/tools_writers_direct.py`
- `scripts/llmagent/lib/tools_network.py`
- `scripts/llmagent/lib/tools_firewall_config.py`
- `scripts/llmagent/lib/tools_backup.py`
- `scripts/llmagent/lib/tools_system.py`
- `scripts/llmagent/lib/tools_nat_config.py`
- `scripts/llmagent/lib/utils.py`
- `scripts/llmagent/lib/xml_writers.py`
- `scripts/llmagent/lib/tools_vlan.py`
- `scripts/llmagent/lib/tools_nat.py`
- `scripts/llmagent/lib/tools_interfaces.py`
- `scripts/llmagent/lib/tools_plugin_dev.py`
- `src/opnsense/mvc/app/models/OPNsense/LLMAgent/LLMAgent.php`
- `src/opnsense/mvc/app/controllers/OPNsense/LLMAgent/IndexController.php`
- `src/opnsense/mvc/app/controllers/OPNsense/LLMAgent/Api/ServiceController.php`
- `src/opnsense/mvc/app/controllers/OPNsense/LLMAgent/Api/SettingsController.php`

## Integration Notes

This repository is part of the OPNsense ecosystem and may integrate with:

- **AFTRS CLI:** Network management and automation
- **UNRAID Monolith:** Server-side network coordination
- **Tailscale:** VPN and remote access
- **Other OPNsense repositories:** See [../](../) for related components

## Original Repository

**Location:** `/Users/mitch/Docs/aftrs-void/opnsense-llmagent`
**Git Repository:** `/Users/mitch/Docs/aftrs-void/opnsense-llmagent`

For the latest version and to make changes, work directly with the original repository.

---
*Consolidated on 2025-09-23 00:34:08 by OPNsense Documentation Consolidation Script*
