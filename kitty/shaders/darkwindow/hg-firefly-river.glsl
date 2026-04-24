// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Firefly river — ~90 fireflies drift along a sinuous river path with individual blink cycles + perpendicular wander, set against dark forest with silhouetted tree trunks and shimmering water band

const int   FIREFLIES = 90;
const int   TREES = 10;
const int   FBM_OCT = 3;
const float INTENSITY = 0.55;

vec3 ff_pal(float t) {
    vec3 amber  = vec3(1.00, 0.85, 0.35);
    vec3 green  = vec3(0.55, 1.00, 0.55);
    vec3 cream  = vec3(1.00, 0.95, 0.75);
    vec3 mint   = vec3(0.45, 0.95, 0.70);
    float s = mod(t * 3.0, 3.0);
    if (s < 1.0)      return mix(amber, green, s);
    else if (s < 2.0) return mix(green, cream, s - 1.0);
    else              return mix(cream, mint, s - 2.0);
}

float ff_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float ff_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float ff_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(ff_hash2(i), ff_hash2(i + vec2(1, 0)), u.x),
               mix(ff_hash2(i + vec2(0, 1)), ff_hash2(i + vec2(1, 1)), u.x), u.y);
}

float ff_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < FBM_OCT; i++) {
        v += a * ff_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

// River path — parametric S-curve. s ∈ [0, 1]. Returns point on the river centerline.
vec2 riverCenter(float s) {
    float x = -1.1 + s * 2.3 + 0.15 * sin(s * 5.0);
    float y = -0.35 + 0.4 * sin(s * 3.2) - 0.1 * sin(s * 1.5);
    return vec2(x, y);
}

// River tangent (for perpendicular wandering)
vec2 riverTangent(float s) {
    float dt = 0.01;
    return normalize(riverCenter(s + dt) - riverCenter(s - dt));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.004, 0.008, 0.015);

    // Night forest floor gradient (darker toward top, hint of glow at river line)
    float skyT = (p.y + 0.9) / 1.8;
    col = mix(vec3(0.01, 0.015, 0.025), vec3(0.025, 0.03, 0.04), clamp(skyT, 0.0, 1.0));

    // === River band — sample distance to river centerline, render as shimmering water ===
    float minRiverD = 1e9;
    float closestS = 0.0;
    for (int i = 0; i < 30; i++) {
        float s1 = float(i) / 29.0;
        float s2 = float(i + 1) / 29.0;
        vec2 a = riverCenter(s1);
        vec2 b = riverCenter(s2);
        vec2 ab = b - a;
        vec2 pa = p - a;
        float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
        float d = length(pa - ab * h);
        if (d < minRiverD) {
            minRiverD = d;
            closestS = mix(s1, s2, h);
        }
    }

    // River width
    float riverW = 0.08 + 0.015 * sin(closestS * 6.0);
    if (minRiverD < riverW) {
        // Water: very dark blue-green with subtle shimmer
        float shimmer = ff_fbm(vec2(closestS * 20.0 - x_Time * 0.3, minRiverD * 40.0));
        vec3 waterCol = mix(vec3(0.02, 0.05, 0.06), vec3(0.04, 0.08, 0.10), shimmer);
        col = mix(col, waterCol, 0.9);
        // Bank highlight
        float bank = smoothstep(riverW, riverW * 0.85, minRiverD);
        col += vec3(0.05, 0.08, 0.03) * bank * 0.5;
    }

    // === Tree silhouettes (tall vertical dark strips scattered beyond the river) ===
    for (int t = 0; t < TREES; t++) {
        float ft = float(t);
        float seed = ft * 7.31;
        float treeX = (ff_hash(seed) - 0.5) * 2.2;
        float treeW = 0.020 + ff_hash(seed * 3.7) * 0.015;
        float treeBase = -0.3 + ff_hash(seed * 5.1) * 0.3;  // base y at this tree
        if (abs(p.x - treeX) < treeW && p.y > treeBase && p.y < treeBase + 0.75) {
            // Skip trees where the river passes through
            vec2 treeCenterline = vec2(treeX, treeBase + 0.35);
            // Tree silhouette
            col = mix(col, vec3(0.01, 0.015, 0.02), 0.9);
            // Canopy wider at top
            if (p.y > treeBase + 0.4) {
                float canopyR = treeW * 2.0 + 0.01 * sin(p.y * 40.0 + ft);
                if (abs(p.x - treeX) < canopyR) {
                    col = mix(col, vec3(0.015, 0.02, 0.025), 0.85);
                }
            }
        }
    }

    // === Fireflies along the river ===
    for (int i = 0; i < FIREFLIES; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        // Each firefly has a base position along the river + perpendicular offset
        float baseS = ff_hash(seed);
        // Slow drift forward along river (with loop-back)
        float driftSpeed = 0.015 + ff_hash(seed * 3.7) * 0.015;
        float s_now = fract(baseS + x_Time * driftSpeed);
        vec2 rc = riverCenter(s_now);
        vec2 rt = riverTangent(s_now);
        vec2 rn = vec2(-rt.y, rt.x);  // normal to river

        // Perpendicular wander (random walk approx via combined sines)
        float wanderAmp = 0.08 + ff_hash(seed * 5.1) * 0.08;
        float wander = wanderAmp * sin(x_Time * (0.8 + ff_hash(seed * 7.3)) + seed * 11.0);
        wander += 0.03 * sin(x_Time * 2.5 + seed * 5.7);
        vec2 ffPos = rc + rn * wander;

        float d = length(p - ffPos);

        // Blink cycle
        float blinkCycle = 1.2 + ff_hash(seed * 11.0) * 2.0;
        float blinkT = fract((x_Time + ff_hash(seed * 13.0) * blinkCycle) / blinkCycle);
        float blink = exp(-pow(blinkT - 0.2, 2.0) * 30.0);

        // Core
        float core = exp(-d * d * 30000.0);
        // Glow
        float glow = exp(-d * d * 500.0) * 0.35;
        vec3 ffCol = ff_pal(fract(seed * 0.1 + x_Time * 0.05));
        col += ffCol * (core + glow) * blink * 1.3;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
