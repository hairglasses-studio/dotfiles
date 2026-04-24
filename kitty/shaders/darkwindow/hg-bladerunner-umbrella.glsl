// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Bladerunner umbrella — BR2049-style glowing umbrellas drifting through rainy night street, LED-lit handles, figure silhouettes below, diagonal rain streaks, occasional lightning flash, fogged neon horizon

const int   UMBRELLAS = 8;
const int   RAIN_DROPS = 160;
const float INTENSITY = 0.55;

vec3 br_pal(float t) {
    vec3 cyan   = vec3(0.20, 0.85, 1.00);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 amber  = vec3(1.00, 0.65, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cyan, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, amber, s - 2.0);
    else if (s < 4.0) return mix(amber, mint, s - 3.0);
    else              return mix(mint, cyan, s - 4.0);
}

float br_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Umbrella at (cx, cy), dome radius R, handle length H
struct Umbrella {
    vec2 center;
    float r;
    float h;
    float seed;
};

Umbrella getUmbrella(int i, float t) {
    Umbrella U;
    float fi = float(i);
    U.seed = fi * 7.31;
    // Column position across frame
    float col = (fi + 0.5) / float(UMBRELLAS) * 2.0 - 1.0;
    // Slow horizontal sway from walking
    float sway = sin(t * 0.4 + fi * 1.3) * 0.05;
    // Slight vertical bob from head step
    float bob = sin(t * 1.4 + fi * 2.1) * 0.008;
    U.center = vec2(col + sway, 0.05 + bob);
    U.r = 0.08 + br_hash(U.seed) * 0.03;
    U.h = 0.22 + br_hash(U.seed * 3.7) * 0.03;
    return U;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.006, 0.010, 0.025);

    // === Lightning flash (periodic, brief) ===
    float lightCycle = 10.0;
    float lightT = fract(x_Time / lightCycle);
    float lightning = 0.0;
    if (lightT < 0.015) {
        lightning = 1.0 - lightT / 0.015;
    } else if (lightT > 0.03 && lightT < 0.05) {
        // Secondary flicker
        lightning = (lightT - 0.03) / 0.02;
        lightning = lightning * (1.0 - lightning) * 4.0;
    }
    col += vec3(0.35, 0.38, 0.50) * lightning * 0.6;

    // === Background sky / horizon neon fog ===
    if (p.y > -0.1) {
        float skyT = (p.y + 0.1) / 1.0;
        vec3 skyTop = vec3(0.015, 0.02, 0.06);
        vec3 skyLow = vec3(0.10, 0.06, 0.15);
        col = mix(col + skyLow, skyTop, smoothstep(0.0, 0.8, skyT));
        // Distant neon signs smear (mag/cyan)
        float neonY = 0.08;
        float neonBand = exp(-pow(p.y - neonY, 2.0) * 90.0);
        col += br_pal(fract(p.x * 0.4 + x_Time * 0.05)) * neonBand * 0.3;
    }

    // === Wet ground reflections (below y = -0.25) ===
    if (p.y < -0.25) {
        float wetness = (-0.25 - p.y) / 0.5;
        // Reflection of neon: sample with flipped y
        vec2 refP = vec2(p.x + sin(p.x * 4.0 + x_Time * 0.5) * 0.01,
                         -p.y - 0.5 + sin(x_Time * 0.7) * 0.005);
        // Simplified: just add smeared cyan+mag bands moving slowly
        float refNeon = exp(-pow(refP.y - 0.08, 2.0) * 20.0);
        col += br_pal(fract(refP.x * 0.3 + x_Time * 0.04)) * refNeon * 0.5 * wetness;
        // Dark ground tone
        col = mix(col, vec3(0.01, 0.015, 0.03), wetness * 0.3);
    }

    // === Umbrellas ===
    for (int i = 0; i < UMBRELLAS; i++) {
        Umbrella U = getUmbrella(i, x_Time);
        vec2 rel = p - U.center;

        // Dome: top half-circle bright glow (radius U.r)
        if (rel.y > -0.01 && length(rel) < U.r * 1.15) {
            float domeD = length(rel);
            float domeMask = smoothstep(U.r, U.r * 0.7, domeD);
            // Fade at bottom (dome is upper half)
            float topFade = smoothstep(-0.01, U.r * 0.4, rel.y);
            vec3 umbCol = br_pal(fract(U.seed * 0.1 + x_Time * 0.02));
            col += umbCol * domeMask * topFade * 0.65;
            // Rim highlight on the dome edge
            float rim = exp(-pow(domeD - U.r * 0.92, 2.0) * 15000.0);
            col += umbCol * rim * topFade * 1.2;
        }

        // Handle: vertical bar below dome
        vec2 hRel = p - (U.center - vec2(0.0, U.h * 0.5));
        if (abs(hRel.x) < 0.003 && abs(hRel.y) < U.h * 0.5) {
            vec3 hCol = br_pal(fract(U.seed * 0.1 + x_Time * 0.02));
            col += hCol * 1.1;
        }

        // Figure silhouette (small triangle below handle)
        vec2 fRel = p - (U.center - vec2(0.0, U.h + 0.08));
        if (fRel.y > -0.08 && fRel.y < 0.04 && abs(fRel.x) < 0.04 + fRel.y * 0.3) {
            col = mix(col, vec3(0.005, 0.005, 0.015), 0.92);
        }

        // Light bleed under umbrella (soft glow on the figure)
        float bleedD = length(p - (U.center - vec2(0.0, U.h * 0.7)));
        col += br_pal(fract(U.seed * 0.1 + x_Time * 0.02)) * exp(-bleedD * bleedD * 120.0) * 0.12;
    }

    // === Rain streaks ===
    for (int i = 0; i < RAIN_DROPS; i++) {
        float fi = float(i);
        float seed = fi * 3.7;
        float speed = 2.0 + br_hash(seed) * 1.5;
        float phase = fract(x_Time * speed * 0.3 + br_hash(seed * 5.1));
        vec2 dropOrigin = vec2(br_hash(seed * 7.3) * 2.4 - 1.2, 1.0);
        // Falling diagonally
        vec2 dropVel = vec2(-0.15, -1.3);
        vec2 dropPos = dropOrigin + dropVel * phase;
        // Streak: from dropPos to dropPos - dropVel * dt
        vec2 tail = dropPos - dropVel * 0.03;
        vec2 ab = dropPos - tail;
        vec2 ap = p - tail;
        float h = clamp(dot(ap, ab) / dot(ab, ab), 0.0, 1.0);
        float d = length(ap - ab * h);
        float mask = exp(-d * d * 150000.0);
        // Fade out streaks below ground
        float groundFade = smoothstep(-0.5, -0.3, dropPos.y);
        col += vec3(0.5, 0.55, 0.70) * mask * (0.4 + lightning * 0.8) * groundFade * 0.6;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
