// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Volumetric lightning storm — branching bolts, screen flash, dark cloud backdrop

const int   BOLT_COUNT    = 3;
const float BOLT_INTERVAL = 2.2;   // seconds between bolts
const float BOLT_DURATION = 0.3;
const int   BOLT_JIT      = 6;     // jitter segments per bolt
const int   BRANCH_PER    = 4;
const float INTENSITY     = 0.55;

vec3 ls_pal(float t) {
    vec3 a = vec3(0.70, 0.85, 1.00);   // white-blue lightning
    vec3 b = vec3(0.90, 0.70, 1.00);   // purple-pink
    vec3 c = vec3(0.45, 0.80, 0.98);   // cyan
    float s = mod(t * 3.0, 3.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else              return mix(c, a, s - 2.0);
}

float ls_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Distance to a polyline defined by anchor points along a vertical path
// with per-segment horizontal jitter
float lightningBolt(vec2 p, float boltSeed, float t, float fadeIn) {
    float minDist = 1e9;
    vec2 prev = vec2((ls_hash(boltSeed * 3.1) - 0.5) * 1.4, 0.9);
    for (int i = 1; i <= BOLT_JIT; i++) {
        float fi = float(i);
        float ty = 0.9 - fi / float(BOLT_JIT) * 1.8;  // top to bottom
        float jit = (ls_hash(boltSeed * 7.0 + fi) - 0.5) * 0.35;
        // Jitter increases toward strike point
        float jitScale = 1.0 - (1.0 - fi / float(BOLT_JIT)) * 0.7;
        vec2 next = vec2(prev.x + jit * jitScale, ty);
        // Fade in by segment — bolt "descends" over BOLT_DURATION
        if (fi / float(BOLT_JIT) > fadeIn) break;
        // Segment distance
        vec2 ba = next - prev;
        vec2 pa = p - prev;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        float d = length(pa - ba * h);
        minDist = min(minDist, d);
        prev = next;
    }
    return minDist;
}

// A simple branch — offshoot segment from a random anchor
float branchBolt(vec2 p, float boltSeed, float branchIdx, float fadeIn) {
    vec2 prev = vec2((ls_hash(boltSeed * 3.1) - 0.5) * 1.4, 0.9);
    // Pick a random point along the main path as origin
    float anchorPhase = 0.25 + 0.6 * ls_hash(boltSeed * 11.0 + branchIdx);
    for (int i = 1; i <= BOLT_JIT; i++) {
        float fi = float(i);
        float ty = 0.9 - fi / float(BOLT_JIT) * 1.8;
        float jit = (ls_hash(boltSeed * 7.0 + fi) - 0.5) * 0.35;
        vec2 next = vec2(prev.x + jit * (1.0 - (1.0 - fi / float(BOLT_JIT)) * 0.7), ty);
        if (fi / float(BOLT_JIT) >= anchorPhase) {
            // Start branch here — perpendicular offshoot
            float angle = (ls_hash(boltSeed * 13.0 + branchIdx) - 0.5) * 1.8;
            float len = 0.08 + 0.12 * ls_hash(boltSeed * 17.0 + branchIdx);
            vec2 dir = vec2(cos(angle), sin(angle));
            vec2 end = next + dir * len * fadeIn;
            vec2 ba = end - next;
            vec2 pa = p - next;
            float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
            return length(pa - ba * h);
        }
        prev = next;
    }
    return 1e9;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Dark storm-cloud backdrop — turbulent low-frequency
    vec3 bg = vec3(0.01, 0.01, 0.04);

    vec3 col = bg;
    float totalFlash = 0.0;

    // Each bolt has its own schedule + seed
    for (int b = 0; b < BOLT_COUNT; b++) {
        float fb = float(b);
        float phase = mod(x_Time + fb * BOLT_INTERVAL * 0.33, BOLT_INTERVAL);
        float aliveFlag = step(0.0, phase) * step(phase, BOLT_DURATION);
        if (aliveFlag < 0.5) continue;

        float t01 = phase / BOLT_DURATION;              // [0,1] life
        float fadeIn = smoothstep(0.0, 0.3, t01);
        float fadeOut = 1.0 - smoothstep(0.5, 1.0, t01);
        float alpha = fadeIn * fadeOut;

        float boltSeed = floor(x_Time / BOLT_INTERVAL) * 19.3 + fb * 7.7;

        // Main bolt
        float d = lightningBolt(p, boltSeed, x_Time, fadeIn);
        float core = 1.0 - smoothstep(0.002, 0.008, d);
        float glow = exp(-d * 100.0) * 0.6;
        vec3 bc = ls_pal(fract(fb * 0.3 + x_Time * 0.05));
        col += (vec3(1.0) * core + bc * glow) * alpha;

        // Branches
        for (int br = 0; br < BRANCH_PER; br++) {
            float bd = branchBolt(p, boltSeed, float(br), fadeIn);
            float brCore = 1.0 - smoothstep(0.002, 0.006, bd);
            float brGlow = exp(-bd * 120.0) * 0.4;
            col += (vec3(1.0) * brCore * 0.7 + bc * brGlow) * alpha * 0.6;
        }

        // Screen flash — peaks at start of strike
        float flash = alpha * (1.0 - t01) * 0.1;
        totalFlash += flash;
    }

    // Add flash to whole screen
    col += vec3(0.5, 0.6, 0.9) * totalFlash;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
