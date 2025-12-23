# Migration & Archival Journal (2025-09-14)

## Migration Steps
- Created aftrs-void/org-admin with sync scripts and legacy archives
- CLI consolidation optional (see aftrs_cli)
- Legacy org to be archived after migration

## Status
- All code and assets migrated
- Documentation updated
# org-admin Org Breakdown (2025-09-14)

## Overview
The org-admin org contains sync workflows, admin scripts, and migration checklists for managing GitHub orgs, repo transfers, and bulk operations.

## Inventory
See `org-admin-inventory.tsv` in this folder for a machine-readable list of repos with status/metadata.

## Key Projects
- sync: Scripts for bulk repo sync, archival, and migration across orgs; includes ghorg and GitHub CLI automation.
- admin scripts: Repo transfer, permission management, and audit logging for org maintenance.
- migration checklist: Step-by-step guide for transferring unique repos (e.g., aftrs-shell to aftrs-void) via GitHub web UI/API.

## Usage
- Use sync scripts for scheduled repo sync and archival.
- Run admin scripts for permission updates and repo transfers.
- Follow migration checklist for org-to-org repo moves requiring owner approval.

## Troubleshooting
- Sync failures: Check API tokens, rate limits, and repo permissions.
- Transfer issues: Verify org membership and repo settings.

## Next Steps
- Expand migration checklist with screenshots and API examples.
- Add audit logging for all bulk operations.
- Document best practices for org maintenance and repo hygiene.
