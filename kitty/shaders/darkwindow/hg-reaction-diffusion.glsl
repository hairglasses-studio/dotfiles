// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Pseudo reaction-diffusion — Turing-pattern spots via time-evolved FBM thresholding

const int   OCTAVES = 6;
const float INTENSITY = 0.5;

vec3 rd_pal(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(0.20, 0.95, 0.60);
    vec3 d = vec3(0.95, 0.30, 0.70);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float rd_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float rd_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(rd_hash(i), rd_hash(i + vec2(1,0)), u.x),
               mix(rd_hash(i + vec2(0,1)), rd_hash(i + vec2(1,1)), u.x), u.y);
}

float rd_fbm(vec2 p, float t) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        // Each octave offset in time at different rates → emergent pattern evolution
        v += a * rd_noise(p + vec2(t * (0.1 + float(i) * 0.02), t * (0.05 + float(i) * 0.015)));
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = x_PixelPos / x_WindowSize.y * 5.0;

    // Two competing "concentration" fields — one slow, one fast
    float a = rd_fbm(p, x_Time * 0.3);
    float b = rd_fbm(p * 2.0 + vec2(37.3), x_Time * 0.5);

    // Turing-like reaction: a-b crossings define pattern
    float diff = a - b * 0.8;

    // Binary-ish threshold with soft edge
    float pattern = smoothstep(0.35, 0.4, diff);

    // Multi-band pattern (rings around threshold gives spot+edge effect)
    float edge = 1.0 - smoothstep(0.0, 0.05, abs(diff - 0.375));

    // Color: base from diff value, edges bright
    vec3 baseCol = rd_pal(fract(diff * 1.5 + x_Time * 0.03));
    vec3 edgeCol = rd_pal(fract(diff * 1.5 + 0.4));

    vec3 col = baseCol * pattern * 0.6;
    col += edgeCol * edge * 1.4;

    // Dark regions (below threshold) get a subtle tint
    col += rd_pal(0.0) * (1.0 - pattern) * 0.12;

    // Pulse brightness subtly
    col *= 0.9 + 0.1 * sin(x_Time * 0.8);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility * 0.9);

    _wShaderOut = vec4(result, 1.0);
}
