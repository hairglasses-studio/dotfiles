// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Tracer round — muzzle at lower-left corner firing streaking tracer bullets with bright motion-blur trails, periodic muzzle flashes, ejecting shell casings, and impact sparks on the far wall

const int   TRACERS = 14;
const int   SHELLS = 6;
const int   SPARKS_PER_IMPACT = 6;
const float INTENSITY = 0.55;

vec3 trc_pal(float t) {
    vec3 amber  = vec3(1.00, 0.70, 0.20);
    vec3 orange = vec3(1.00, 0.45, 0.10);
    vec3 red    = vec3(1.00, 0.20, 0.20);
    vec3 mag    = vec3(0.95, 0.30, 0.55);
    vec3 cream  = vec3(1.00, 0.95, 0.80);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(amber, orange, s);
    else if (s < 2.0) return mix(orange, red, s - 1.0);
    else if (s < 3.0) return mix(red, mag, s - 2.0);
    else              return mix(mag, cream, s - 3.0);
}

float trc_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Muzzle position fixed at lower-left
vec2 muzzlePos() { return vec2(-0.78, -0.45); }

// Tracer trajectory: angle and travel distance.
// Parameterize tracer i by phase in [0, 1]; at phase=0 it's at muzzle, phase=1
// it impacts the wall.
struct Tracer {
    vec2 origin;
    vec2 dir;        // unit vector
    float travel;    // total travel distance
    float phase;     // [0,1]
    float speed;
    float seed;
};

Tracer getTracer(int i, float t) {
    float fi = float(i);
    float seed = fi * 7.31;
    Tracer T;
    T.origin = muzzlePos();
    // Angle: mostly upward-right, fan out
    float ang = 0.35 + trc_hash(seed) * 0.55;     // ~20..50 degrees upward from horizontal
    T.dir = vec2(cos(ang), sin(ang));
    T.travel = 1.8 + trc_hash(seed * 3.7) * 0.5;
    T.speed = 0.9 + trc_hash(seed * 5.1) * 0.7;
    // Phase cycles; each tracer has own start offset
    float cycle = 1.0 / T.speed;
    T.phase = fract((t + trc_hash(seed * 11.0) * cycle) / cycle);
    T.seed = seed;
    return T;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.008, 0.010, 0.020);

    // Faint smoky volumetric haze
    float haze = 0.5 + 0.5 * sin(p.x * 3.0 + x_Time * 0.3 + sin(p.y * 2.0 + x_Time * 0.2));
    col += vec3(0.10, 0.05, 0.08) * haze * 0.15;

    // "Wall" at the upper right — slightly brighter band as the target surface
    if (p.x > 0.75 || p.y > 0.55) {
        float wallDist = max(p.x - 0.75, p.y - 0.55);
        col += vec3(0.08, 0.05, 0.06) * exp(-wallDist * 4.0);
    }

    vec2 mPos = muzzlePos();

    // Muzzle flash — triggered whenever any tracer has phase near 0
    float muzzleFlashAccum = 0.0;
    for (int i = 0; i < TRACERS; i++) {
        Tracer T = getTracer(i, x_Time);
        if (T.phase < 0.05) {
            muzzleFlashAccum += (1.0 - T.phase / 0.05);
        }
    }
    float mD = length(p - mPos);
    // Flash burst
    col += vec3(1.0, 0.95, 0.75) * exp(-mD * mD * 250.0) * muzzleFlashAccum * 1.5;
    col += trc_pal(0.2) * exp(-mD * mD * 30.0) * muzzleFlashAccum * 0.4;

    // Muzzle glow (persistent, lower)
    col += vec3(0.9, 0.35, 0.15) * exp(-mD * mD * 1800.0) * 0.45;

    // === Tracers: render as line segments with motion-blur trails ===
    for (int i = 0; i < TRACERS; i++) {
        Tracer T = getTracer(i, x_Time);
        if (T.phase < 0.01) continue; // still in muzzle
        // Current bullet head position
        vec2 head = T.origin + T.dir * T.travel * T.phase;
        // Trail start: previous position (scaled by phase-lag = trail length)
        float trailLen = 0.12;
        float trailStartPhase = max(0.0, T.phase - trailLen / T.travel);
        vec2 trailStart = T.origin + T.dir * T.travel * trailStartPhase;

        // Point-to-segment distance
        vec2 ab = head - trailStart;
        vec2 ap = p - trailStart;
        float h = clamp(dot(ap, ab) / dot(ab, ab), 0.0, 1.0);
        float dSeg = length(ap - ab * h);

        // Trail thickness tapers toward start
        float thickness = 0.002 + (1.0 - h) * 0.0015;
        float trailMask = exp(-dSeg * dSeg / (thickness * thickness) * 1.8);
        // Intensity fades toward trail start
        float fade = smoothstep(0.0, 0.3, h) * h;
        // Head is brightest
        float headBoost = exp(-length(p - head) * length(p - head) * 30000.0) * 1.5;
        vec3 tCol = trc_pal(fract(T.seed * 0.1 + x_Time * 0.05));
        col += tCol * trailMask * (0.7 + fade * 0.5) * 1.2;
        col += tCol * headBoost;

        // Soft halo around head
        col += tCol * exp(-length(p - head) * length(p - head) * 400.0) * 0.3;

        // Impact spark at end (when phase > 0.95)
        if (T.phase > 0.95) {
            vec2 impact = T.origin + T.dir * T.travel;
            float sAge = (T.phase - 0.95) / 0.05;
            for (int k = 0; k < SPARKS_PER_IMPACT; k++) {
                float fk = float(k);
                float sparkAng = trc_hash(T.seed * 3.1 + fk) * 6.28;
                float sparkDist = sAge * 0.06 * (0.5 + trc_hash(T.seed * 5.1 + fk) * 0.5);
                vec2 sparkPos = impact + vec2(cos(sparkAng), sin(sparkAng)) * sparkDist;
                float sd = length(p - sparkPos);
                col += trc_pal(fk * 0.15) * exp(-sd * sd * 25000.0) * (1.0 - sAge) * 0.9;
            }
            // Main impact flash
            float impactD = length(p - impact);
            col += vec3(1.0, 0.95, 0.75) * exp(-impactD * impactD * 2000.0) * (1.0 - sAge) * 1.4;
        }
    }

    // === Shell casings ejecting from muzzle ===
    for (int i = 0; i < SHELLS; i++) {
        float fi = float(i);
        float seed = fi * 11.1;
        float ejectCycle = 0.4;
        float sPhase = fract((x_Time + trc_hash(seed) * ejectCycle) / ejectCycle);
        // Parabolic trajectory: initial velocity up-left, gravity pulls down
        float vx = -0.15 - trc_hash(seed * 3.1) * 0.1;
        float vy = 0.18 + trc_hash(seed * 5.1) * 0.08;
        float g = -0.4;
        vec2 shellPos = mPos + vec2(vx, vy) * sPhase + vec2(0.0, 0.5 * g * sPhase * sPhase);
        float sd = length(p - shellPos);
        col += vec3(0.85, 0.75, 0.45) * exp(-sd * sd * 8000.0) * (0.8 - sPhase * 0.5);
        // Shell glint
        col += vec3(1.0, 0.95, 0.80) * exp(-sd * sd * 80000.0) * 0.8;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
