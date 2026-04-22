// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Mystical neon glyphs — rune circles with rotating outer ring + sacred geometry + glow

const int   RUNES        = 12;
const float OUTER_R      = 0.35;
const float INTENSITY    = 0.55;

vec3 ng_pal(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98);
    vec3 b = vec3(0.90, 0.25, 0.70);
    vec3 c = vec3(0.10, 0.82, 0.92);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float ng_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// Mystical glyph — 5x5 binary pattern per rune seed
float runeGlyph(vec2 uvR, float seed) {
    vec2 pix = floor(uvR * 5.0);
    if (pix.x < 0.0 || pix.x > 4.0 || pix.y < 0.0 || pix.y > 4.0) return 0.0;
    float h = ng_hash(pix + vec2(seed, 0.0));
    // Symmetric glyph — mirror across vertical axis
    vec2 sym = vec2(4.0 - pix.x, pix.y);
    float hsym = ng_hash(sym + vec2(seed, 0.0));
    return step(0.5, max(h, hsym));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    float r = length(p);
    float a = atan(p.y, p.x);

    // Outer ring — main circle
    float outerRingD = abs(r - OUTER_R);
    float outerRing = exp(-outerRingD * outerRingD * 3000.0);
    col += ng_pal(fract(x_Time * 0.04)) * outerRing * 0.7;

    // Inner pentagon
    float pentA = mod(a + x_Time * 0.05, 6.28318 / 5.0) - 6.28318 / 10.0;
    float pentDist = OUTER_R * 0.5 - r * cos(pentA) / cos(0.0);
    float pentLine = exp(-pentDist * pentDist * 5000.0) * smoothstep(OUTER_R * 0.5, OUTER_R * 0.52, r) * smoothstep(OUTER_R * 0.55, OUTER_R * 0.53, r);
    col += ng_pal(0.3) * pentLine * 0.5;

    // Rune glyphs around the ring
    for (int i = 0; i < RUNES; i++) {
        float fi = float(i);
        float runeAngle = fi / float(RUNES) * 6.28 + x_Time * 0.15;
        vec2 runeCenter = vec2(cos(runeAngle), sin(runeAngle)) * OUTER_R;
        vec2 diff = p - runeCenter;
        // Rotate rune glyph to face inward
        float ca = cos(-runeAngle + 1.5708), sa = sin(-runeAngle + 1.5708);
        vec2 rotDiff = mat2(ca, -sa, sa, ca) * diff;
        // Glyph local UV (-0.03 to 0.03)
        vec2 glyphUV = (rotDiff + vec2(0.03)) / 0.06;
        if (glyphUV.x > 0.0 && glyphUV.x < 1.0 && glyphUV.y > 0.0 && glyphUV.y < 1.0) {
            float glyph = runeGlyph(glyphUV, fi);
            // Per-rune color
            vec3 runeCol = ng_pal(fract(fi * 0.085 + x_Time * 0.03));
            // Pulsing brightness
            float pulse = 0.6 + 0.4 * sin(x_Time * 2.0 + fi * 0.5);
            col += runeCol * glyph * pulse * 0.8;
        }
        // Rune halo dot
        float runeD = length(diff);
        col += ng_pal(fract(fi * 0.085)) * exp(-runeD * runeD * 800.0) * 0.2;
    }

    // Central symbol — triangle
    float triAng = mod(a + x_Time * 0.1, 6.28318 / 3.0) - 6.28318 / 6.0;
    float innerR = 0.1;
    float triDist = abs(r - innerR * cos(triAng) / cos(0.0));
    float triMask = exp(-triDist * triDist * 5000.0) * smoothstep(innerR * 0.6, innerR * 1.1, r) * smoothstep(innerR * 1.3, innerR * 1.1, r);
    col += ng_pal(0.8) * triMask * 0.7;

    // Pulsing center
    float centerGlow = exp(-r * r * 400.0) * (0.4 + 0.3 * sin(x_Time * 1.5));
    col += ng_pal(0.1) * centerGlow;

    // Emanating energy lines
    float energy = sin(r * 30.0 - x_Time * 3.0);
    energy = pow(max(0.0, energy), 4.0);
    float energyRadial = smoothstep(0.05, 0.3, r) * smoothstep(OUTER_R * 1.1, OUTER_R * 0.9, r);
    col += ng_pal(fract(a / 6.28)) * energy * energyRadial * 0.2;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
