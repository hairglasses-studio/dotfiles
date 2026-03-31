---
paths:
  - "hyprland/**"
---

# Hyprland Config Rules (0.54+)

- Window rules MUST use block syntax with `name =` as the first field:
  ```
  windowrule {
      name = rule-name
      match:class = ^(pattern)$
      float = yes
  }
  ```
- Never use `windowrulev2` (deprecated in 0.54)
- Never use `windowrule = float, class:^(...)$` (old v1 syntax, broken)
- After any config change, verify with `hyprctl configerrors`
- Layer rules may need block syntax too — test before committing
- Environment variables use `env = KEY,VALUE` (no quotes around values)
- Keybinds: `bind = MOD, KEY, dispatcher, args` — use `binde` for repeatable, `bindl` for locked
