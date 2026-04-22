// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Fireflies meadow — 120 pulsing glowing insects drifting over moonlit grass

const int   FIREFLIES = 120;
const float INTENSITY = 0.5;

vec3 ff_pal(float t) {
    vec3 warm = vec3(0.95, 0.85, 0.40);
    vec3 mint = vec3(0.25, 0.95, 0.60);
    vec3 cyan = vec3(0.20, 0.80, 0.95);
    vec3 gold = vec3(1.00, 0.70, 0.30);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(warm, mint, s);
    else if (s < 2.0) return mix(mint, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, gold, s - 2.0);
    else              return mix(gold, warm, s - 3.0);
}

float ff_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  ff_hash2(float n) { return vec2(ff_hash(n), ff_hash(n * 1.37 + 11.0)); }

// Firefly position — lazy drift
vec2 flyPos(int i, float t) {
    float fi = float(i);
    float seed = fi * 3.71;
    // Base wander anchor
    vec2 anchor = ff_hash2(seed) * 1.6 - 0.8;
    anchor.x *= x_WindowSize.x / x_WindowSize.y;
    // Slow drift
    float sa = t * (0.15 + ff_hash(seed * 2.0) * 0.1) + seed;
    anchor.x += 0.05 * cos(sa) + 0.025 * cos(sa * 1.7);
    anchor.y += 0.04 * sin(sa * 1.3) + 0.02 * sin(sa * 2.3);
    return anchor;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Moonlit sky/meadow gradient
    vec3 bg = mix(vec3(0.03, 0.08, 0.15), vec3(0.01, 0.02, 0.05), 1.0 - uv.y);
    vec3 col = bg;

    // Grass silhouette at bottom — dark, with occasional lit blade from firefly
    if (uv.y < 0.3) {
        // Procedural grass blades
        float grassDense = sin(uv.x * 100.0 + ff_hash(floor(uv.x * 50.0))) * 0.5 + 0.5;
        float grassHeight = 0.05 + 0.05 * ff_hash(floor(uv.x * 80.0));
        if (uv.y < grassHeight) {
            col = vec3(0.02, 0.03, 0.02);
        }
    }

    // Moon glow at top corner
    vec2 moonPos = vec2(0.3, 0.4);
    float moonD = length(p - moonPos);
    col += vec3(0.95, 0.95, 0.85) * exp(-moonD * moonD * 200.0) * 0.6;
    col += vec3(0.5, 0.55, 0.45) * exp(-moonD * moonD * 8.0) * 0.07;

    // Fireflies
    for (int i = 0; i < FIREFLIES; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        vec2 fp = flyPos(i, x_Time);

        float d = length(p - fp);

        // Independent blink pattern — random on/off with fade
        float blinkF = 0.3 + ff_hash(seed * 5.1) * 0.6;
        float blinkPhase = fract(x_Time * blinkF + ff_hash(seed * 7.3));
        // Bright peak, slow fade — piecewise
        float brightness;
        if (blinkPhase < 0.08) {
            brightness = blinkPhase / 0.08;  // rise
        } else if (blinkPhase < 0.35) {
            brightness = 1.0 - (blinkPhase - 0.08) / 0.27;  // fall
        } else {
            brightness = 0.0;
        }

        // Hot white center + colored halo
        float coreSize = 0.003;
        float core = exp(-d * d / (coreSize * coreSize) * 2.0);
        float halo = exp(-d * d * 1800.0) * 0.3;
        float farHalo = exp(-d * d * 150.0) * 0.07;

        vec3 fc = ff_pal(fract(seed * 0.05 + x_Time * 0.03));
        col += fc * (core * 1.2 + halo) * brightness;
        col += fc * farHalo * brightness * 0.5;
        col += vec3(1.0) * core * brightness * 0.8;
    }

    // Very subtle drifting mist
    float mist = 0.05 * sin(p.x * 8.0 + x_Time * 0.1) * smoothstep(0.3, 0.0, uv.y);
    col += vec3(0.3, 0.4, 0.5) * max(0.0, mist) * 0.3;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
