// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Ring galaxy — Cartwheel-style collisional ring galaxy with expanding star-forming ring, nucleus bulge, H-II sparkles, and a passing intruder companion

const int   HII_SPARKS = 36;
const int   BG_STARS = 140;
const float NUCLEUS_R = 0.05;
const float DISK_R    = 0.42;
const float TILT      = 0.42;   // axis ratio (y-squash for viewing tilt)
const float INTENSITY = 0.55;

vec3 rng_pal(float t) {
    vec3 deep   = vec3(0.08, 0.04, 0.20);
    vec3 blue   = vec3(0.25, 0.45, 0.98);
    vec3 cyan   = vec3(0.30, 0.90, 1.00);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 amber  = vec3(1.00, 0.75, 0.35);
    vec3 cream  = vec3(1.00, 0.96, 0.85);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(deep, blue, s);
    else if (s < 2.0) return mix(blue, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, mag, s - 2.0);
    else if (s < 4.0) return mix(mag, amber, s - 3.0);
    else              return mix(amber, cream, s - 4.0);
}

float rng_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float rng_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.01, 0.03);

    // Sparse background starfield
    for (int i = 0; i < BG_STARS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        vec2 sp = vec2(rng_hash(seed) * 2.0 - 1.0, rng_hash(seed * 3.7) * 2.0 - 1.0);
        sp *= 0.95;
        float sd = length(p - sp);
        float mag = 0.4 + rng_hash(seed * 5.1) * 0.6;
        float twinkle = 0.7 + 0.3 * sin(x_Time * (1.0 + rng_hash(seed * 11.0)) + seed);
        col += vec3(0.85, 0.9, 1.0) * exp(-sd * sd * 35000.0) * mag * twinkle * 0.35;
    }

    // Galaxy frame: tilted coordinates (y squashed)
    vec2 g = vec2(p.x, p.y / TILT);
    float gr = length(g);
    float gang = atan(g.y, g.x);

    // Slow disk rotation
    float diskRot = x_Time * 0.05;
    float gangRot = gang + diskRot;

    // === Nucleus bulge ===
    float nucD = gr;
    if (nucD < NUCLEUS_R * 3.0) {
        float nucCore = exp(-nucD * nucD / (NUCLEUS_R * NUCLEUS_R) * 1.8);
        float nucFlick = 0.9 + 0.1 * sin(x_Time * 1.7);
        col += vec3(1.0, 0.85, 0.45) * nucCore * 1.2 * nucFlick;
        col += vec3(1.0, 0.55, 0.20) * exp(-nucD * nucD * 60.0) * 0.5;
    }

    // === Expanding ring wave (post-collision, propagates outward) ===
    // Cycle: ring starts at center, expands to DISK_R over CYCLE seconds, then restarts
    float CYCLE = 14.0;
    float cycT = fract(x_Time / CYCLE);
    float ringR = NUCLEUS_R + cycT * (DISK_R - NUCLEUS_R);
    float ringD = abs(gr - ringR);
    float ringThick = 0.022 + cycT * 0.018;   // thickens as it ages
    float ringMask = exp(-ringD * ringD / (ringThick * ringThick) * 1.2);
    // Ring fades toward outer edge + age
    float ringFade = (1.0 - cycT * 0.4);
    // Bright blue (young massive stars)
    col += rng_pal(fract(0.15 + cycT * 0.5)) * ringMask * ringFade * 1.2;

    // Trailing older ring (previous wave, dimmer)
    float ringR2 = NUCLEUS_R + fract(cycT + 0.6) * (DISK_R - NUCLEUS_R);
    float ringD2 = abs(gr - ringR2);
    float ringMask2 = exp(-ringD2 * ringD2 / (ringThick * ringThick) * 1.4);
    col += rng_pal(0.7) * ringMask2 * 0.35;

    // === Spokes — faint radial spokes between rings (old Cartwheel structure) ===
    if (gr > NUCLEUS_R * 1.5 && gr < DISK_R) {
        float spokeN = 8.0;
        float spokePhase = gangRot * spokeN;
        float spoke = pow(abs(cos(spokePhase)), 18.0);
        float spokeFade = smoothstep(DISK_R, NUCLEUS_R * 2.0, gr);
        col += vec3(0.55, 0.45, 0.75) * spoke * spokeFade * 0.25;
    }

    // === Faint disk envelope ===
    if (gr > NUCLEUS_R && gr < DISK_R * 1.1) {
        float diskFade = smoothstep(DISK_R * 1.1, DISK_R * 0.6, gr);
        float diskInner = smoothstep(NUCLEUS_R, NUCLEUS_R * 3.0, gr);
        col += vec3(0.25, 0.20, 0.55) * diskFade * diskInner * 0.18;
    }

    // === H-II region sparkles along the current ring ===
    for (int i = 0; i < HII_SPARKS; i++) {
        float fi = float(i);
        float seed = fi * 7.31 + floor(x_Time / CYCLE) * 13.0;  // re-randomize each cycle
        float sparkAng = rng_hash(seed) * 6.28;
        float sparkR = ringR + (rng_hash(seed * 3.7) - 0.5) * ringThick * 2.0;
        vec2 sparkPos = vec2(cos(sparkAng + diskRot), sin(sparkAng + diskRot) * TILT) * sparkR;
        float sparkD = length(p - sparkPos);
        float sparkCore = exp(-sparkD * sparkD * 25000.0);
        // Pulse: each H-II region flickers on a unique phase
        float pulse = 0.6 + 0.4 * sin(x_Time * (2.0 + rng_hash(seed * 5.1) * 3.0) + seed);
        col += vec3(0.9, 0.6, 1.0) * sparkCore * pulse * 0.9;
    }

    // === Intruder companion galaxy — small offset compact ===
    float intT = x_Time * 0.07;
    vec2 intPos = vec2(0.55 + 0.05 * sin(intT), -0.30 + 0.03 * cos(intT * 1.3));
    float intR = length(p - intPos);
    if (intR < 0.08) {
        float intCore = exp(-intR * intR * 600.0);
        col += vec3(1.0, 0.85, 0.55) * intCore * 0.9;
        // Small elliptical disk
        vec2 intLocal = p - intPos;
        vec2 intDisk = vec2(intLocal.x, intLocal.y / 0.55);
        float intDiskD = length(intDisk);
        if (intDiskD < 0.06) {
            col += vec3(0.85, 0.60, 0.40) * (1.0 - intDiskD / 0.06) * 0.35;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
