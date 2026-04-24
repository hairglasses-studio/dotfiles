// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Magma vein — sinuous glowing molten vein running across dark cracked rock, with hot white-yellow core → orange halo → dark-red crust, heat shimmer, branch veins, popping spark droplets, periodic flare pulses

const int   VEIN_SAMPS = 60;
const int   BRANCHES = 3;
const int   SPARKS = 18;
const int   FBM_OCT = 4;
const float INTENSITY = 0.55;

vec3 mv_pal(float heat) {
    vec3 charcoal = vec3(0.10, 0.03, 0.02);
    vec3 darkRed  = vec3(0.55, 0.08, 0.05);
    vec3 red      = vec3(0.95, 0.25, 0.10);
    vec3 orange   = vec3(1.00, 0.55, 0.15);
    vec3 yellow   = vec3(1.00, 0.85, 0.40);
    vec3 white    = vec3(1.00, 0.98, 0.85);
    if (heat < 0.2)      return mix(charcoal, darkRed, heat * 5.0);
    else if (heat < 0.4) return mix(darkRed, red, (heat - 0.2) * 5.0);
    else if (heat < 0.6) return mix(red, orange, (heat - 0.4) * 5.0);
    else if (heat < 0.8) return mix(orange, yellow, (heat - 0.6) * 5.0);
    else                 return mix(yellow, white, (heat - 0.8) * 5.0);
}

float mv_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float mv_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float mv_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(mv_hash2(i), mv_hash2(i + vec2(1, 0)), u.x),
               mix(mv_hash2(i + vec2(0, 1)), mv_hash2(i + vec2(1, 1)), u.x), u.y);
}

float mv_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < FBM_OCT; i++) {
        v += a * mv_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

// Main vein path — parametric curve running from upper-left to lower-right with sinuous shape
vec2 veinPoint(float s, float t) {
    // s ∈ [0, 1]
    float x = -1.0 + s * 2.0 + 0.15 * sin(s * 14.0 + t * 0.3);
    float y = 0.35 - s * 0.7 + 0.25 * sin(s * 6.0 + t * 0.2) + 0.08 * sin(s * 20.0);
    return vec2(x, y);
}

// Branch vein — small offshoot
vec2 branchPoint(int b, float s, float t) {
    float fb = float(b);
    float seed = fb * 7.31;
    // Branch attaches at some s0 on main vein
    float s0 = 0.25 + mv_hash(seed) * 0.5;
    vec2 anchor = veinPoint(s0, t);
    // Branch direction and spread
    float ang = mv_hash(seed * 3.7) * 6.28;
    float length_ = 0.25 + mv_hash(seed * 5.1) * 0.15;
    return anchor + vec2(cos(ang), sin(ang)) * s * length_
           + vec2(0.04 * sin(s * 12.0 + t * 0.5), 0.02 * cos(s * 8.0));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Rocky background FBM
    float rockNoise = mv_fbm(p * 3.0);
    vec3 rockCol = mix(vec3(0.04, 0.03, 0.05), vec3(0.12, 0.08, 0.08), rockNoise);
    // Darker cracks
    float crack = smoothstep(0.65, 0.75, abs(mv_fbm(p * 8.0 + 2.0) - 0.5) * 2.0);
    rockCol *= 1.0 - crack * 0.4;
    vec3 col = rockCol;

    // Main vein distance
    float minD = 1e9;
    float closestS = 0.0;
    for (int i = 0; i < VEIN_SAMPS - 1; i++) {
        float s1 = float(i) / float(VEIN_SAMPS - 1);
        float s2 = float(i + 1) / float(VEIN_SAMPS - 1);
        vec2 a = veinPoint(s1, x_Time);
        vec2 b = veinPoint(s2, x_Time);
        vec2 ab = b - a;
        vec2 pa = p - a;
        float lenSq = dot(ab, ab);
        if (lenSq > 1e-6) {
            float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
            float d = length(pa - ab * h);
            if (d < minD) { minD = d; closestS = mix(s1, s2, h); }
        }
    }

    // Branch veins
    float minBD = 1e9;
    float closestBS = 0.0;
    for (int b = 0; b < BRANCHES; b++) {
        for (int i = 0; i < 15; i++) {
            float s1 = float(i) / 14.0;
            float s2 = float(i + 1) / 14.0;
            vec2 a = branchPoint(b, s1, x_Time);
            vec2 bp = branchPoint(b, s2, x_Time);
            vec2 ab = bp - a;
            vec2 pa = p - a;
            float lenSq = dot(ab, ab);
            if (lenSq > 1e-6) {
                float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
                float d = length(pa - ab * h);
                if (d < minBD) { minBD = d; closestBS = s1; }
            }
        }
    }

    // Heat shimmer: add small FBM-based displacement
    float shimmer = mv_fbm(p * 12.0 + vec2(0.0, x_Time * 1.5)) * 0.005;
    minD = max(minD - shimmer, 0.0);

    // Render magma: inner core (thin, bright white-yellow), middle (orange), outer (dark red)
    float coreThick = 0.012;
    float halo = 0.05;
    float beyond = 0.12;

    // Core
    float coreMask = exp(-minD * minD / (coreThick * coreThick) * 1.5);
    // Add small flicker variation
    float flicker = 0.85 + 0.15 * mv_fbm(p * 8.0 + vec2(0.0, x_Time * 3.0));
    col = mix(col, mv_pal(0.92) * flicker * 1.5, coreMask);
    // Halo (orange)
    float haloMask = exp(-minD * minD / (halo * halo) * 1.5) * 0.6;
    col += mv_pal(0.55) * haloMask;
    // Beyond (dark red crust)
    float beyondMask = exp(-minD * minD / (beyond * beyond) * 1.2) * 0.3;
    col += mv_pal(0.25) * beyondMask;

    // Branches: only render as thinner, dimmer veins
    float branchMask = exp(-minBD * minBD / (coreThick * coreThick * 0.6) * 2.0);
    col = mix(col, mv_pal(0.75), branchMask * 0.85);
    col += mv_pal(0.45) * exp(-minBD * minBD / (halo * halo * 0.5) * 1.5) * 0.35;

    // Periodic pulse: brighten the core every 5s
    float pulse = 0.7 + 0.3 * sin(x_Time * 1.25);
    col += mv_pal(0.95) * coreMask * pulse * 0.3;

    // Sparks popping from the vein
    for (int i = 0; i < SPARKS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float sparkCycle = 1.5 + mv_hash(seed) * 1.5;
        float sparkT = fract((x_Time + mv_hash(seed * 3.1) * sparkCycle) / sparkCycle);
        // Spawn position: random s along main vein
        float s_spawn = mv_hash(seed * 5.1);
        vec2 origin = veinPoint(s_spawn, x_Time);
        // Launch upward + slight sideways
        float launchAng = 1.3 + (mv_hash(seed * 7.3) - 0.5) * 0.8;
        float gravity = -0.7;
        float v0 = 0.35 + mv_hash(seed * 11.0) * 0.15;
        vec2 sparkPos = origin + vec2(cos(launchAng), sin(launchAng)) * v0 * sparkT
                      + vec2(0.0, 0.5 * gravity * sparkT * sparkT);
        float sd = length(p - sparkPos);
        float sparkCore = exp(-sd * sd * 15000.0);
        col += mv_pal(0.85 - sparkT * 0.5) * sparkCore * (1.0 - sparkT) * 1.1;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
