// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Shatter impact — central strike creates radial cracks + concentric fracture rings + flying shards

const int   RADIAL_CRACKS = 24;
const int   SHARDS        = 40;
const float INTENSITY     = 0.55;

vec3 si_pal(float t) {
    vec3 a = vec3(0.95, 0.98, 1.00);
    vec3 b = vec3(0.20, 0.85, 0.95);
    vec3 c = vec3(0.95, 0.30, 0.60);
    vec3 d = vec3(0.55, 0.30, 0.98);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float si_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  si_hash2(float n) { return vec2(si_hash(n), si_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Impact cycle — explosion + fade + reset
    float cycle = 4.0;
    float phase = mod(x_Time, cycle) / cycle;
    float cycleID = floor(x_Time / cycle);

    // Impact point — different each cycle
    vec2 impact = si_hash2(cycleID * 7.3) * 0.4 - 0.2;

    vec2 toImpact = p - impact;
    float r = length(toImpact);
    float ang = atan(toImpact.y, toImpact.x);

    vec3 col = vec3(0.02, 0.02, 0.05);

    // Radial cracks — thin lines from impact point outward
    float radialMax = 0.0 + phase * 1.0;
    if (r < radialMax) {
        // Use angle to find nearest crack direction
        float crackWedge = 6.28318 / float(RADIAL_CRACKS);
        float wedgeA = mod(ang + si_hash(cycleID + 1.0), crackWedge);
        float angToNearestCrack = min(wedgeA, crackWedge - wedgeA);
        float crackWidth = 0.003 + 0.002 * (1.0 - r / radialMax);
        float crackMask = exp(-angToNearestCrack * angToNearestCrack * 2000.0 / r);
        // Random gaps (not every crack is present)
        float crackIdx = floor((ang + 3.14159) / crackWedge);
        float crackPresent = si_hash(crackIdx + cycleID * 31.0);
        if (crackPresent > 0.4) {
            vec3 crackCol = si_pal(fract(crackIdx * 0.03 + x_Time * 0.05));
            col += crackCol * crackMask * 0.8;
        }
    }

    // Concentric fracture rings (expanding)
    for (int ring = 0; ring < 4; ring++) {
        float fring = float(ring);
        float ringR = phase * 0.8 * (fring + 1.0) * 0.3;
        float ringDist = abs(r - ringR);
        float ringWidth = 0.003 + phase * 0.01;
        float ring_ = exp(-ringDist * ringDist / (ringWidth * ringWidth) * 2.0);
        float ringFade = 1.0 - phase;
        col += si_pal(fract(fring * 0.2 + x_Time * 0.04)) * ring_ * ringFade * 0.4;
    }

    // Flying shards — triangular pieces scattering outward
    for (int s = 0; s < SHARDS; s++) {
        float fs = float(s);
        float seed = fs * 7.31 + cycleID * 11.3;
        float ang_s = fs / float(SHARDS) * 6.28 + si_hash(seed) * 0.3;
        vec2 dir = vec2(cos(ang_s), sin(ang_s));
        float speed = 0.3 + si_hash(seed * 3.7) * 0.4;
        vec2 shardPos = impact + dir * phase * speed;
        // Gravity
        shardPos.y -= phase * phase * 0.4;

        float sd = length(p - shardPos);
        float size = 0.004 * (1.0 - phase * 0.5);
        float shardMask = exp(-sd * sd / (size * size) * 3.0);

        vec3 sc = si_pal(fract(seed * 0.03));
        col += sc * shardMask * (1.0 - phase);
    }

    // Impact flash at phase 0
    if (phase < 0.1) {
        float flash = (1.0 - phase / 0.1);
        col += vec3(1.0) * exp(-r * r * 200.0) * flash * 1.5;
        col += si_pal(0.0) * exp(-r * r * 15.0) * flash * 0.5;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
