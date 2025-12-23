# Repository Management

**Integrated repository analysis, health monitoring, and consolidation tools**

## Overview

The repository management system has been fully integrated into `aftrs_cli` as the `repo-mgmt` module. This provides comprehensive tools for analyzing, monitoring, and consolidating repositories across the aftrs-void organization.

## Quick Start

### Using AFTRS CLI

```bash
# Navigate to aftrs_cli directory
cd aftrs_cli

# Quick repository health check
./aftrs.sh repo-mgmt quick

# Full repository analysis
./aftrs.sh repo-mgmt analyze

# Repository health monitoring
./aftrs.sh repo-mgmt health

# Status overview
./aftrs.sh repo-mgmt status
```

### Available Commands

| Command | Description | Usage |
|---------|-------------|--------|
| `analyze` | Run comprehensive repository analysis | `./aftrs.sh repo-mgmt analyze` |
| `health` | Run repository health monitoring | `./aftrs.sh repo-mgmt health` |
| `consolidate` | Run repository consolidation | `./aftrs.sh repo-mgmt consolidate [--phase N]` |
| `status` | Show repository management status | `./aftrs.sh repo-mgmt status` |
| `commit` | Commit consolidation changes | `./aftrs.sh repo-mgmt commit` |
| `quick` | Quick status and health check | `./aftrs.sh repo-mgmt quick` |

## Integration Details

### Module Location
- **Primary Script:** `aftrs_cli/scripts/repository_management.py`
- **Integration Point:** `aftrs_cli/aftrs.sh` (repo-mgmt category)
- **Base Tools:** All original consolidation tools remain available

### Architecture
The repository management system uses a modular architecture:

```
aftrs_cli/
├── aftrs.sh                           # Main CLI dispatcher
├── scripts/
│   └── repository_management.py       # Repository management module
└── docs/
    └── repository-management.md       # This documentation
```

### Tool Integration
The system integrates the following tools from the root aftrs-void directory:

- `repo_analysis_script.py` - Repository analysis and categorization
- `repository_health_monitor.py` - Health monitoring with scoring
- `consolidation_script.py` - Phase 1 consolidation (documentation-only)
- `phase2_consolidation_script.py` - Phase 2 consolidation (backups/utilities)
- `repository_toolkit.py` - Unified toolkit interface
- `commit_consolidation_changes.sh` - Safe commit automation

## Usage Examples

### Daily Operations
```bash
# Quick health and status check
./aftrs.sh repo-mgmt quick

# Check current status
./aftrs.sh repo-mgmt status
```

### Analysis and Monitoring
```bash
# Run full repository analysis
./aftrs.sh repo-mgmt analyze

# Monitor repository health
./aftrs.sh repo-mgmt health
```

### Consolidation Operations
```bash
# Run Phase 1 consolidation (documentation-only repos)
./aftrs.sh repo-mgmt consolidate --phase 1

# Run Phase 2 consolidation (backup configs and utilities)
./aftrs.sh repo-mgmt consolidate --phase 2

# Commit consolidation changes
./aftrs.sh repo-mgmt commit
```

## Current Status

### Consolidation Results
As of September 2025, the repository consolidation has achieved:

- **Repository Reduction:** 35 → 25 active repositories (29% decrease)
- **Health Status:** 100% healthy repositories (25/25 perfect scores)
- **Content Preservation:** 100% with full attribution and audit trail

### Consolidated Repositories

**Phase 1 - Documentation & Plans:**
- `aftrs-terraform` → `aftrs_wiki/projects/aftrs-terraform.md`
- `secret_cli` → `aftrs_wiki/projects/secret_cli.md`
- `wizardstower-unraid` → `aftrs_wiki/projects/wizardstower-unraid.md`
- `tailscale-acl` → `aftrs_wiki/projects/tailscale-acl.md`
- `bench_cli` → `aftrs_wiki/projects/bench_cli.md`

**Phase 2 - Infrastructure Consolidation:**
- Config backups → `aftrs_wiki/infrastructure/backups/`
- Small utilities → `aftrs_wiki/projects/utilities/`

### Archive Locations
- **Phase 1 Archives:** `archived_repos_20250923_045634/`
- **Phase 2 Archives:** `archived_repos_phase2_20250923_050209/`

## Health Monitoring

### Health Criteria
Repositories are evaluated on:
- **Git Activity:** Recent commits and development activity
- **Repository Structure:** README files, proper organization
- **Content Quality:** Substantial vs. stub repositories
- **Maintenance Status:** Uncommitted changes, repository size

### Health Categories
- **🟢 Healthy (80-100 points):** Active, well-maintained repositories
- **🟡 Needs Attention (60-79 points):** May need updates or cleanup
- **🔴 Stale (0-59 points):** Inactive repositories, consolidation candidates

### Current Health Status
✅ **All 25 active repositories maintain healthy status (100% success rate)**

## Best Practices

### Regular Maintenance
- **Weekly:** Run `./aftrs.sh repo-mgmt quick` for health checks
- **Monthly:** Run `./aftrs.sh repo-mgmt analyze` for comprehensive analysis
- **As Needed:** Use `./aftrs.sh repo-mgmt health` for detailed monitoring

### Repository Organization
1. **Keep substantial projects separate** - Projects with active development deserve their own repositories
2. **Consolidate documentation** - Project plans work better in centralized wikis
3. **Monitor health regularly** - Address recommendations promptly
4. **Maintain audit trails** - Keep comprehensive logs of organizational changes

## Integration with Other Systems

### AFTRS CLI Integration
The repository management system is fully integrated into the AFTRS CLI ecosystem:

- **Performance Caching:** Uses AFTRS CLI's caching system
- **Logging:** Integrates with AFTRS CLI logging infrastructure
- **Error Handling:** Follows AFTRS CLI error handling patterns
- **Help System:** Integrated help and documentation

### Git Operations Integration
Works seamlessly with existing git operations:

- **Bulk Git Module:** `./aftrs.sh git status|commit|push`
- **Repository Management:** `./aftrs.sh repo-mgmt [commands]`
- **Repo Sync Module:** `./aftrs.sh repos clone|status|autopush`

## Support and Maintenance

### Getting Help
```bash
# Show repository management help
./aftrs.sh repo-mgmt

# Show general AFTRS CLI help
./aftrs.sh help
```

### Troubleshooting
1. **Tool execution fails:** Ensure Python 3.x is installed and accessible
2. **Git operations fail:** Check for uncommitted changes and git configuration
3. **Health issues:** Review specific repository recommendations from health reports

### Extending the System
The repository management system is modular and extensible:

- **Add new analysis rules:** Extend `repository_health_monitor.py`
- **Custom consolidation logic:** Modify consolidation scripts
- **New health criteria:** Update health scoring in the monitoring system
- **Additional workflows:** Add new commands to the CLI integration

---

*Last Updated: September 23, 2025*  
*Integration Status: Complete - Fully operational within aftrs_cli*  
*Health Status: 100% healthy repositories (25/25 active)*
