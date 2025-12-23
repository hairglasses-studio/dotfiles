# Console-Hax Org: Consolidation Summary (2025-09-14)

## Inventory
See `console-hax-inventory.tsv` in this folder for a machine-readable list of all repos with archived/active status and key metadata.

## Status Overview
- Archived: legacy templates and experiments (ps2-homebrew-starter, -modern, -full, ps2_all, pcsx2_scaffold, visualizers, hg_aura, ps2_leaked_source_code_repos, etc.)
- Active core:
  - ps2-base-image: toolchain container
  - ps2-test-image: emulator/test automation container
  - aura: example/benchmark ELF and assets
  - ps2_cli: CLI for build/test/run orchestration
  - ps2-ndi, retro-session-orchestrator, mcp2-toolbox, ps2_sdks_internal_collection: auxiliary tools/data

## Actions Completed
- Migration/archival notices propagated to legacy/template repos
- PCSX2 automation centralized into ps2-test-image
- README and docs updated in core repos
- External resources integrated into docs

## Next Steps
- Merge auxiliary tools into ps2_cli where feasible
- Add end-to-end tests across aura and sample projects
- Expand AI agent workflows for build→run→assert loops


## Journal (2025-09-14)
- ps2-test-image: Added headless ELF smoke test (`testing/test-elf.sh`) producing a minimal JSON report for CI/agents.
- ps2-test-image: Updated README with aligned options, Quick Start invocation, and an AI agent workflow recipe.
