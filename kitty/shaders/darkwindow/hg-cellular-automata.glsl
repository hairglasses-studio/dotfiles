// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Cellular automata visualization — Game-of-Life-like pattern with time-evolved FBM proxy

const int   OCTAVES = 5;
const float CELL = 0.025;
const float INTENSITY = 0.55;

vec3 ca_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.25, 0.65);
    vec3 d = vec3(0.20, 0.95, 0.60);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float ca_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float ca_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(ca_hash(i), ca_hash(i + vec2(1,0)), u.x),
               mix(ca_hash(i + vec2(0,1)), ca_hash(i + vec2(1,1)), u.x), u.y);
}

float ca_fbm(vec2 p, float t) {
    float v = 0.0, a = 0.5;
    // Each octave time-shifts at different rates — pattern evolution
    for (int i = 0; i < OCTAVES; i++) {
        v += a * ca_noise(p + vec2(t * (0.3 + float(i) * 0.05), t * (0.2 + float(i) * 0.04)));
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Snap to cell grid
    vec2 cellId = floor(p / CELL);
    vec2 cellCenter = (cellId + 0.5) * CELL;

    // Current "aliveness" from time-evolved FBM at cell center
    float alive = ca_fbm(cellCenter * 4.0, x_Time * 0.5);
    alive = smoothstep(0.48, 0.52, alive);  // binary-ish

    // "Recent" state — sample slightly in past
    float alivePrev = ca_fbm(cellCenter * 4.0, x_Time * 0.5 - 0.15);
    alivePrev = smoothstep(0.48, 0.52, alivePrev);

    // Cell position (0-1) within cell
    vec2 inCell = fract(p / CELL);
    vec2 centered = inCell - 0.5;

    // Live cells: fill with color; Dying cells (alive→dead transition): red glow
    vec3 col = vec3(0.01, 0.02, 0.04);

    if (alive > 0.5) {
        // Cell body — rounded square
        float cellD = max(abs(centered.x), abs(centered.y)) - 0.35;
        float body = smoothstep(0.0, -0.05, cellD);
        float edge = smoothstep(0.02, 0.0, abs(cellD));

        vec3 cellCol = ca_pal(fract(cellId.x * 0.05 + cellId.y * 0.03 + x_Time * 0.04));
        col = mix(col, cellCol * 0.8, body);
        col += cellCol * edge * 0.4;

        // Just-born bright flash
        if (alivePrev < 0.5) {
            col += vec3(1.0) * body * 0.5;
        }
    } else if (alivePrev > 0.5) {
        // Dying cell — fading red glow
        float deathGlow = ca_fbm(cellCenter * 4.0, x_Time * 0.5 - 0.05);
        deathGlow = smoothstep(0.48, 0.52, deathGlow);
        if (deathGlow > 0.5) {
            float cellD = length(centered);
            col += vec3(0.95, 0.25, 0.3) * exp(-cellD * cellD * 60.0) * 0.6;
        }
    }

    // Grid lines (very subtle)
    vec2 gridDist = abs(centered);
    float gridLine = smoothstep(0.48, 0.5, max(gridDist.x, gridDist.y));
    col += vec3(0.02, 0.04, 0.05) * gridLine;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility * 0.9);

    _wShaderOut = vec4(result, 1.0);
}
