# Agent Parity Templates

These templates back `scripts/hg-agent-parity-sync.sh`.

- `claude-settings.base.json` is the minimal repo baseline for Claude Code.
- `gemini-settings.base.json` is the minimal repo baseline for Gemini CLI.
- `hooks/` holds optional Claude hook intent fragments for repos that need richer lifecycle bridges.
- `extensions/` holds optional Gemini extension scaffolds for repos that need provider-native packaging beyond `.gemini/settings.json`.

The sync script owns the rendered repo files. Keep these templates generic and provider-safe.
