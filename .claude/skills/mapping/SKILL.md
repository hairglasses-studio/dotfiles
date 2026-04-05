---
description: "Manage controller/MIDI mapping profiles. $ARGUMENTS: (empty)=status, 'list'=profiles, 'learn'=capture inputs, 'create <name>'=new profile, 'edit <name>'=modify, 'test <rule>'=test resolution"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)**: Call `mcp__dotfiles__mapping_status` — show connected devices, active profiles, daemon
- **"list"**: Call `mcp__dotfiles__mapping_list_profiles` — all profiles with device info
- **"learn"**: Call `mcp__dotfiles__mapping_learn` — listen to controller for 10s, capture inputs, suggest mappings
- **"create <name>"**: Call `mcp__dotfiles__mapping_generate` with template selection
- **"edit <name>"**: Call `mcp__dotfiles__mapping_get_profile` then present for editing
- **"validate <name>"**: Call `mcp__dotfiles__mapping_validate` — syntax + schema check
- **"test <rule>"**: Call `mcp__dotfiles__mapping_resolve_test` — simulate rule resolution
- **"diff <a> <b>"**: Call `mcp__dotfiles__mapping_diff` — compare two profiles
- **"monitor"**: Call `mcp__dotfiles__mapping_monitor` — live event stream
- **"auto"**: Call `mcp__dotfiles__mapping_auto_setup` — detect→generate→restart
