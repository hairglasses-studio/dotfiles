// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Lava lamp — rising + falling metaball blobs in viscous fluid with warm uplight

const int   BLOBS    = 10;
const float THRESH   = 1.8;
const float INTENSITY = 0.5;

vec3 ll_pal(float t) {
    vec3 a = vec3(0.95, 0.25, 0.40);
    vec3 b = vec3(1.00, 0.55, 0.20);
    vec3 c = vec3(0.95, 0.85, 0.45);
    vec3 d = vec3(0.55, 0.30, 0.98);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float ll_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

vec2 blobPos(int i, float t) {
    float fi = float(i);
    float seed = fi * 7.31;
    // Slow rise-fall cycle
    float cycle = 6.0 + ll_hash(seed) * 4.0;
    float phase = fract((t + ll_hash(seed * 3.0) * cycle) / cycle);
    // Smooth up-down via sin
    float y = sin(phase * 6.28318) * 0.55;
    // Slight x drift
    float xBase = (ll_hash(seed * 5.1) - 0.5) * 0.4;
    float x = xBase + 0.03 * sin(t * 0.5 + fi);
    return vec2(x, y);
}

float blobSize(int i, float t) {
    float seed = float(i) * 7.31 + 100.0;
    return 0.07 + 0.04 * ll_hash(seed);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Vessel boundary — a vertical tube centered at 0, width ~0.25
    float tubeW = 0.3;
    if (abs(p.x) > tubeW) {
        _wShaderOut = terminal;
        return;
    }

    // Warm ambient gradient (uplit from base)
    vec3 bg = mix(vec3(0.45, 0.15, 0.05), vec3(0.10, 0.04, 0.08), (p.y + 0.5) / 1.0);
    vec3 col = bg;

    // Metaball field
    float field = 0.0;
    vec3 colorSum = vec3(0.0);
    float colorWeight = 0.0;
    for (int i = 0; i < BLOBS; i++) {
        vec2 bp = blobPos(i, x_Time);
        float r = blobSize(i, x_Time);
        float d = length(p - bp);
        float contrib = r * r / (d * d + 0.002);
        field += contrib;
        vec3 bc = ll_pal(fract(float(i) * 0.08 + x_Time * 0.04));
        colorSum += bc * contrib;
        colorWeight += contrib;
    }
    if (colorWeight > 0.001) colorSum /= colorWeight;

    // Inside isosurface
    if (field > THRESH) {
        col = colorSum * 0.9;
    }
    // Edge
    float edge = smoothstep(THRESH - 0.2, THRESH, field) * smoothstep(THRESH + 0.2, THRESH, field);
    col += colorSum * edge * 0.6;

    // Base uplight glow
    float baseGlow = exp(-pow(p.y + 0.5, 2.0) * 4.0) * 0.4;
    col += ll_pal(0.0) * baseGlow;

    // Top cap darken
    col *= smoothstep(0.8, 0.4, p.y);

    // Tube edges (slight rim)
    float edgeD = tubeW - abs(p.x);
    if (edgeD < 0.015) {
        col *= 0.7;
        col += vec3(0.3, 0.25, 0.2) * smoothstep(0.015, 0.003, edgeD) * 0.5;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
