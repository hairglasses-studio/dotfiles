# Shader Forge — integration pipeline

After writing `kitty/shaders/darkwindow/hg-<slug>.glsl`, run these in order.

## 1. Compile-check

```bash
cd /home/hg/hairglasses-studio/dotfiles
# The preamble defines DarkWindow uniforms so glslangValidator can compile.
cat > /tmp/darkwindow-preamble.glsl <<'PRE'
#version 460 core
uniform sampler2D _wTexture;
uniform vec2 _wResolution;
uniform float x_Time;
uniform vec2 x_CursorPos;
#define x_WindowSize _wResolution
#define x_PixelPos (gl_FragCoord.xy)
#define x_Texture(uv) texture(_wTexture, (uv))
out vec4 _wFragOutput;
PRE
cat > /tmp/darkwindow-footer.glsl <<'FOOT'
void main() {
    vec4 _wBase = x_Texture(x_PixelPos / x_WindowSize);
    windowShader(_wBase);
    _wFragOutput = _wBase;
}
FOOT

SHADER=kitty/shaders/darkwindow/hg-<slug>.glsl
tmp=$(mktemp /tmp/sc-XXXXXX.frag)
cat /tmp/darkwindow-preamble.glsl "$SHADER" /tmp/darkwindow-footer.glsl > "$tmp"
out=$(glslangValidator -S frag "$tmp" 2>&1)
if echo "$out" | grep -q ERROR; then
  echo "FAIL"; echo "$out" | grep ERROR | head -5
  exit 1
else
  echo "PASS: $(wc -c < $SHADER) bytes"
fi
rm -f "$tmp"
```

## 2. Register in playlists + sync registry + regen tiers

```bash
cd /home/hg/hairglasses-studio/dotfiles/kitty/shaders/playlists
SLUG=hg-<slug>.glsl

# Core rotations: always add to ambient + best-of + cyberpunk + high-intensity.
for pl in ambient best-of cyberpunk high-intensity; do
  { cat $pl.txt; echo "$SLUG"; } | sort -u > $pl.txt.new && mv $pl.txt.new $pl.txt
done

# Thematic playlists — pick 1-2 matching your shader's genre:
#   radial-plasma  — radial/plasma/tunnel/fractal
#   trippy         — glitch / kaleidoscopic / psychedelic / distortion
#   post-fx        — chromatic / scanlines / bloom / noise overlays
#   low-intensity  — soft ambient, slow motion
#   dev-console    — tech/coder aesthetic (hacker-terminal, spectrum, hologram)
#
# Example:
# { cat radial-plasma.txt; echo "$SLUG"; } | awk 'NF && $1 !~ /^#/' | sort -u > rp.body
# { head -4 radial-plasma.txt; cat rp.body; } > radial-plasma.txt.new && mv radial-plasma.txt.new radial-plasma.txt
# rm -f rp.body

cd /home/hg/hairglasses-studio/dotfiles

# Regenerate DarkWindow plugin registry from ambient.txt (authoritative).
awk '/^plugin:darkwindow/{exit} {print}' hyprland/darkwindow-shaders.conf > /tmp/reg-header.conf
{
  cat /tmp/reg-header.conf
  echo "plugin:darkwindow {"
  awk 'NF && $1 !~ /^#/' kitty/shaders/playlists/ambient.txt | sort -u | while IFS= read -r line; do
    name="${line%.glsl}"
    printf '    shader[%s] {\n        path = /home/hg/hairglasses-studio/dotfiles/kitty/shaders/darkwindow/%s\n    }\n' "$name" "$line"
  done
  echo "}"
} > hyprland/darkwindow-shaders.conf.new
mv hyprland/darkwindow-shaders.conf.new hyprland/darkwindow-shaders.conf

# Consistency gate — MUST be green before proceeding.
scripts/check-shader-consistency.sh || { echo "CONSISTENCY FAIL — abort"; exit 1; }

# Regenerate tier playlists (cheap / mid / heavy bins by file size).
kitty/shaders/bin/shader-tier.sh generate

# Reload Hyprland so DarkWindow sees the new shader.
hyprctl reload
```

## 3. Live-test

```bash
SLUG=hg-<slug>
# Save previous state, dispatch new shader, wait, restore, clear notifications.
PREV=$(cat $HOME/.local/state/kitty-shaders/current-shader 2>/dev/null || echo water)
hyprctl dispatch darkwindow:shadeactive "$SLUG" >/dev/null
sleep 10
hyprctl dispatch darkwindow:shadeactive "$PREV" >/dev/null
sleep 0.3
swaync-client --close-all >/dev/null 2>&1
echo "live-tested $SLUG for 10s, restored $PREV"
```

## 4. Report to user

One paragraph: shader name, genre, what makes it visually distinct, which tier bin it landed in (`grep -l "^$SLUG$" kitty/shaders/playlists/tier-*.txt`), cumulative count (`ls kitty/shaders/darkwindow/hg-*.glsl | wc -l`).

## Escape hatches

- **Compile fails with `Reserved word`** — rename the variable (common culprits: `sample`, `active`, `filter`, `input`, `output`, `noise1..4`, `texture1D..3D`). Save the offender to `~/.claude/projects/-home-hg-hairglasses-studio-dotfiles/memory/feedback_glsl_reserved.md` if new.
- **Consistency check fails** — usually means the registry didn't pick up the new entry. Re-run the `awk … > darkwindow-shaders.conf` block.
- **Dispatch returns non-ok** — most likely the shader isn't in the registry yet. Check: `grep "shader\[$SLUG]" hyprland/darkwindow-shaders.conf` (without the `.glsl` suffix).
- **Scope creep** — one invocation = one shader. If you have energy for more, that's the /loop runtime's job, not yours.
