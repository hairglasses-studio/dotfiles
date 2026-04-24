// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Lightning Field — homage to Walter de Maria's installation: grid of thin stainless-steel rods across a desert plain in perspective, receding to horizon; periodic jagged lightning bolts strike rod tips from a stormy sky.

const int   GRID_COLS = 10;
const int   GRID_ROWS = 7;
const int   BOLTS = 3;
const float HORIZON_Y = 0.22;
const float INTENSITY = 0.55;

vec3 lf_pal(float t) {
    vec3 vio1   = vec3(0.22, 0.12, 0.35);
    vec3 mag    = vec3(0.75, 0.20, 0.55);
    vec3 amber  = vec3(1.00, 0.65, 0.30);
    vec3 cyan   = vec3(0.30, 0.85, 1.00);
    vec3 white  = vec3(1.00, 0.98, 0.90);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(vio1, mag, s);
    else if (s < 2.0) return mix(mag, amber, s - 1.0);
    else if (s < 3.0) return mix(amber, cyan, s - 2.0);
    else              return mix(cyan, white, s - 3.0);
}

float lf_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float lf_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float lf_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(lf_hash2(i), lf_hash2(i + vec2(1, 0)), u.x),
               mix(lf_hash2(i + vec2(0, 1)), lf_hash2(i + vec2(1, 1)), u.x), u.y);
}

// Map grid cell (c, r) to world position, then to screen position with perspective
vec2 rodBase(int c, int r) {
    // World grid: column c ∈ [0, COLS-1], row r ∈ [0, ROWS-1]
    // Row determines depth: 0 = far, ROWS-1 = near
    float depth = 1.0 - float(r) / float(GRID_ROWS - 1);  // 1 = far, 0 = near
    // Perspective Y: far rows near horizon, near rows toward bottom
    float y = mix(-0.55, HORIZON_Y - 0.02, depth * depth);
    // Perspective X: columns spread outward based on depth
    float colFrac = (float(c) / float(GRID_COLS - 1) - 0.5);
    float xSpread = mix(1.2, 0.22, depth);  // near rows spread wide, far rows narrow
    float x = colFrac * xSpread;
    return vec2(x, y);
}

// Rod tip height above base (perspective-scaled)
float rodHeight(int r) {
    float depth = 1.0 - float(r) / float(GRID_ROWS - 1);
    return mix(0.15, 0.04, depth);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.008, 0.020);

    // === Sky: stormy gradient + distant cloud layers ===
    if (p.y > HORIZON_Y) {
        float skyT = (p.y - HORIZON_Y) / (0.9 - HORIZON_Y);
        vec3 skyTop = vec3(0.04, 0.03, 0.12);
        vec3 skyHor = vec3(0.35, 0.15, 0.25);
        col = mix(skyHor, skyTop, skyT);
        // Clouds
        float cloud = lf_noise(p * vec2(3.0, 8.0) + vec2(x_Time * 0.05, 0.0));
        cloud = smoothstep(0.45, 0.65, cloud);
        col = mix(col, vec3(0.15, 0.10, 0.18), cloud * 0.4);
    } else {
        // Desert ground
        float depth = (HORIZON_Y - p.y) / 0.8;
        depth = clamp(depth, 0.0, 1.0);
        vec3 groundFar = vec3(0.25, 0.15, 0.15);
        vec3 groundNear = vec3(0.08, 0.05, 0.07);
        col = mix(groundFar, groundNear, depth);
        // Desert texture
        float grainy = lf_noise(p * 20.0) * 0.15;
        col += vec3(0.20, 0.14, 0.10) * grainy * (1.0 - depth);
    }

    // === Grid rods ===
    // For each pixel, find the closest rod
    for (int c = 0; c < GRID_COLS; c++) {
        for (int r = 0; r < GRID_ROWS; r++) {
            vec2 base = rodBase(c, r);
            float h = rodHeight(r);
            // Rod occupies a thin vertical segment from (base.x, base.y) to (base.x, base.y + h)
            if (p.x > base.x - 0.005 && p.x < base.x + 0.005
                && p.y > base.y && p.y < base.y + h) {
                // Rod body (thin dark metal with rim shine)
                float xCenter = abs(p.x - base.x);
                float bodyMask = smoothstep(0.003, 0.0, xCenter);
                // Depth-based brightness (far rods dimmer)
                float depth = 1.0 - float(r) / float(GRID_ROWS - 1);
                float rodBright = mix(0.6, 0.2, depth);
                col = mix(col, vec3(0.85, 0.85, 0.95), bodyMask * rodBright);
            }
            // Rod tip glow
            vec2 tip = vec2(base.x, base.y + h);
            float td = length(p - tip);
            float depth = 1.0 - float(r) / float(GRID_ROWS - 1);
            float tipSize = mix(0.012, 0.004, depth);
            float tipGlow = exp(-td * td / (tipSize * tipSize) * 2.0);
            col += vec3(0.90, 0.95, 1.00) * tipGlow * (1.0 - depth * 0.5) * 0.55;
        }
    }

    // === Lightning bolts — strikes hitting random rod tips at random times ===
    for (int b = 0; b < BOLTS; b++) {
        float fb = float(b);
        float seed = fb * 11.1;
        // Cycle: ~3s per strike
        float strikeCycle = 3.0 + lf_hash(seed) * 2.0;
        float t0 = fract((x_Time + lf_hash(seed * 3.1) * strikeCycle) / strikeCycle);
        // Flash envelope: brief
        float flashEnv = 0.0;
        if (t0 < 0.04) flashEnv = 1.0 - t0 / 0.04;
        else if (t0 < 0.08) flashEnv = (t0 - 0.04) / 0.04 * 0.4;

        if (flashEnv > 0.01) {
            // Target rod
            int targetC = int(lf_hash(seed * 5.1) * float(GRID_COLS));
            int targetR = int(lf_hash(seed * 7.3) * float(GRID_ROWS));
            targetC = targetC % GRID_COLS;
            targetR = targetR % GRID_ROWS;
            vec2 rtBase = rodBase(targetC, targetR);
            float rtH = rodHeight(targetR);
            vec2 tip = vec2(rtBase.x, rtBase.y + rtH);
            // Bolt starts at top of sky and zigzags to tip
            vec2 boltStart = vec2(tip.x + (lf_hash(seed * 11.0) - 0.5) * 0.5, 0.85);
            // Sample the bolt path as a series of jagged segments
            float minBoltD = 1e9;
            vec2 prev = boltStart;
            for (int s = 1; s <= 12; s++) {
                float fs = float(s);
                float stepT = fs / 12.0;
                vec2 linPt = mix(boltStart, tip, stepT);
                // Jagged horizontal offset
                float jag = (lf_hash(seed + fs * 0.3 + floor(x_Time * 2.0)) - 0.5) * 0.08 * (1.0 - stepT);
                vec2 boltPt = linPt + vec2(jag, 0.0);
                vec2 ab = boltPt - prev;
                vec2 pa = p - prev;
                float lenSq = dot(ab, ab);
                if (lenSq > 1e-6) {
                    float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
                    float d = length(pa - ab * h);
                    if (d < minBoltD) minBoltD = d;
                }
                prev = boltPt;
            }
            float boltMask = exp(-minBoltD * minBoltD * 40000.0);
            col += vec3(1.0, 0.95, 1.0) * boltMask * flashEnv * 2.0;
            col += vec3(0.55, 0.75, 1.0) * exp(-minBoltD * minBoltD * 2000.0) * flashEnv * 0.6;
            // Sky flash brightness boost
            if (p.y > HORIZON_Y) {
                col += lf_pal(0.5) * flashEnv * 0.08;
            }
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
