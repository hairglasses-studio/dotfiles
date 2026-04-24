// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Drone swarm — 36 drones in coherent formation (delta / grid / circle transitions), each with red port / green starboard / white strobe nav lights, per-drone trail, and periodic strobe flash

const int   DRONES = 36;
const int   BG_STARS = 70;
const float INTENSITY = 0.55;
const float FORMATION_CYCLE = 12.0;

vec3 drn_pal(float t) {
    vec3 red    = vec3(1.00, 0.25, 0.25);
    vec3 amber  = vec3(1.00, 0.75, 0.25);
    vec3 green  = vec3(0.35, 1.00, 0.55);
    vec3 cyan   = vec3(0.30, 0.90, 1.00);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(red, amber, s);
    else if (s < 2.0) return mix(amber, green, s - 1.0);
    else if (s < 3.0) return mix(green, cyan, s - 2.0);
    else if (s < 4.0) return mix(cyan, vio, s - 3.0);
    else              return mix(vio, red, s - 4.0);
}

float drn_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float drn_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

// Formation position for drone i at time t.
// Three formations cycle: 0=grid, 1=delta, 2=circle. Smooth blend between them.
vec2 formationPos(int i, float t) {
    float fi = float(i);
    float seed = fi * 7.31;

    // Grid formation (6x6)
    int col = i % 6;
    int row = i / 6;
    vec2 grid = vec2(float(col) - 2.5, float(row) - 2.5) * 0.11;

    // Delta (triangle) formation
    // Arrange rows so row 0 has 1, row 1 has 3, row 2 has 5, etc.
    // Compute triangular-ish using index → (rowD, indexInRow)
    int rowD = 0;
    int accum = 0;
    for (int k = 0; k < 6; k++) {
        int thisRow = 2 * k + 1;
        if (i >= accum && i < accum + thisRow) { rowD = k; break; }
        accum += thisRow;
    }
    int idxInRow = i - accum;
    int rowSize = 2 * rowD + 1;
    float dxD = (float(idxInRow) - float(rowSize - 1) * 0.5) * 0.08;
    vec2 delta = vec2(dxD, 0.30 - float(rowD) * 0.08);

    // Circle formation
    float ca = float(i) / float(DRONES) * 6.28;
    vec2 circle = vec2(cos(ca), sin(ca)) * 0.32;

    // Blend by cycle phase
    float ph = fract(t / FORMATION_CYCLE) * 3.0;  // 0..3 across 3 formations
    vec2 a, b;
    float blend;
    if (ph < 1.0)      { a = grid;   b = delta;  blend = smoothstep(0.6, 1.0, ph); }
    else if (ph < 2.0) { a = delta;  b = circle; blend = smoothstep(0.6, 1.0, ph - 1.0); }
    else               { a = circle; b = grid;   blend = smoothstep(0.6, 1.0, ph - 2.0); }

    vec2 formPos = mix(a, b, blend);

    // Per-drone jitter (hover oscillation)
    float joff = sin(t * 1.3 + seed) * 0.004 + cos(t * 1.7 + seed * 3.0) * 0.004;
    formPos += vec2(joff, joff * 0.7);

    // Global formation drift: slow sway
    vec2 drift = vec2(sin(t * 0.15) * 0.15, cos(t * 0.1) * 0.08);

    return formPos + drift;
}

// Drone heading: approximate from velocity (forward difference)
vec2 droneHeading(int i, float t) {
    float dt = 0.03;
    return normalize(formationPos(i, t + dt) - formationPos(i, t - dt) + vec2(1e-5, 0.0));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.008, 0.025);

    // Faint urban horizon glow at bottom (suggestion of a city beneath)
    float hglow = exp(-(p.y + 0.55) * (p.y + 0.55) * 30.0) * 0.25;
    col += vec3(0.35, 0.18, 0.40) * hglow;

    // Background stars
    for (int i = 0; i < BG_STARS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        vec2 sp = vec2(drn_hash(seed) * 2.0 - 1.0, drn_hash(seed * 3.7) * 1.6 - 0.8);
        float sd = length(p - sp);
        float mag = 0.3 + drn_hash(seed * 5.1) * 0.3;
        col += vec3(0.85, 0.9, 1.0) * exp(-sd * sd * 40000.0) * mag * 0.28;
    }

    // === Drones ===
    for (int i = 0; i < DRONES; i++) {
        vec2 dpos = formationPos(i, x_Time);
        vec2 heading = droneHeading(i, x_Time);
        vec2 side = vec2(-heading.y, heading.x);

        // Body: small dark cluster at dpos
        float bodyD = length(p - dpos);
        if (bodyD < 0.025) {
            col = mix(col, vec3(0.015, 0.015, 0.02), exp(-bodyD * bodyD * 12000.0));
        }

        // Red port nav light (left wing tip)
        vec2 portPos = dpos - side * 0.012;
        float portD = length(p - portPos);
        col += vec3(1.0, 0.25, 0.25) * exp(-portD * portD * 50000.0) * 1.0;
        col += vec3(1.0, 0.25, 0.25) * exp(-portD * portD * 500.0) * 0.1;

        // Green starboard nav light (right wing tip)
        vec2 stbdPos = dpos + side * 0.012;
        float stbdD = length(p - stbdPos);
        col += vec3(0.3, 1.0, 0.4) * exp(-stbdD * stbdD * 50000.0) * 1.0;
        col += vec3(0.3, 1.0, 0.4) * exp(-stbdD * stbdD * 500.0) * 0.1;

        // White strobe (top) — brief flash every ~1.5s with per-drone phase
        float strobePhase = fract(x_Time / 1.5 + drn_hash(float(i) * 5.1));
        float strobe = exp(-strobePhase * strobePhase * 180.0);
        col += vec3(1.0, 1.0, 0.95) * exp(-bodyD * bodyD * 3000.0) * strobe * 1.2;

        // Trail behind drone (motion blur)
        float tailAlong = dot(p - dpos, -heading);
        float tailPerp = length(p - dpos - (-heading) * tailAlong);
        if (tailAlong > 0.0 && tailAlong < 0.06) {
            float trailW = exp(-tailPerp * tailPerp * 80000.0);
            float trailFade = 1.0 - tailAlong / 0.06;
            col += drn_pal(fract(float(i) * 0.05 + x_Time * 0.1)) * trailW * trailFade * 0.35;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
