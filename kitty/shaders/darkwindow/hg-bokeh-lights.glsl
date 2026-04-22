// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Bokeh lights — defocused circular light spots with hexagonal aperture hints

const int   LIGHTS = 40;
const float INTENSITY = 0.55;

vec3 bk_pal(float t) {
    vec3 a = vec3(0.95, 0.30, 0.70);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(1.00, 0.75, 0.35);
    vec3 d = vec3(0.55, 0.30, 0.98);
    vec3 e = vec3(0.20, 0.95, 0.60);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

float bk_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  bk_hash2(float n) { return vec2(bk_hash(n), bk_hash(n * 1.37 + 11.0)); }

// Hexagonal bokeh shape
float sdHexBokeh(vec2 p, float r) {
    p = abs(p);
    float d = max(p.x * 0.866 + p.y * 0.5, p.y);
    return d - r;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.02, 0.02, 0.05);

    for (int i = 0; i < LIGHTS; i++) {
        float fi = float(i);
        float seed = fi * 3.71;
        // Position — slight drift
        vec2 basePos = bk_hash2(seed) * 2.0 - 1.0;
        basePos.x *= x_WindowSize.x / x_WindowSize.y;
        basePos += 0.05 * vec2(sin(x_Time * 0.2 + fi), cos(x_Time * 0.15 + fi * 1.3));

        // Bokeh size varies (focus distance)
        float size = 0.04 + bk_hash(seed * 3.7) * 0.06;
        // Angle (for hex aperture)
        float angle = seed;
        float ca = cos(angle), sa = sin(angle);
        vec2 rp = mat2(ca, -sa, sa, ca) * (p - basePos);

        // Hexagonal edge SDF
        float hexD = sdHexBokeh(rp, size);
        // Fill with dim interior
        float inside = smoothstep(0.0, -size * 0.8, hexD);
        // Bright edge
        float edge = exp(-hexD * hexD * 2000.0) * 0.5;

        vec3 bokehCol = bk_pal(fract(seed * 0.05 + x_Time * 0.03));

        // Brightness flickers per light
        float flicker = 0.7 + 0.3 * sin(x_Time * (1.0 + bk_hash(seed * 5.3)) + fi);
        float brightness = flicker * (0.4 + 0.6 * bk_hash(seed * 7.3));

        col += bokehCol * inside * brightness * 0.35;
        col += bokehCol * edge * brightness * 0.7;

        // Extra inner hotspot
        float hotspot = exp(-dot(rp, rp) / (size * 0.3 * size * 0.3) * 2.0);
        col += bokehCol * hotspot * brightness * 0.25;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.8, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
