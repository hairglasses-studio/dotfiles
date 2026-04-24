// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Desert oasis — golden-hour sunset over sand dunes with central water pool, 4 palm-tree silhouettes (curved trunk + 6 radial fronds), sky reflection in water, mirage shimmer near horizon, and wind-blown sand particles

const int   PALMS = 4;
const int   SAND_PARTICLES = 40;
const int   FBM_OCT = 4;
const float HORIZON = 0.08;
const float POOL_Y = -0.32;
const float INTENSITY = 0.55;

vec3 oa_pal(float t) {
    vec3 deep   = vec3(0.18, 0.06, 0.12);
    vec3 mag    = vec3(0.95, 0.35, 0.25);
    vec3 orange = vec3(1.00, 0.55, 0.15);
    vec3 amber  = vec3(1.00, 0.80, 0.35);
    vec3 cream  = vec3(1.00, 0.95, 0.70);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(deep, mag, s);
    else if (s < 2.0) return mix(mag, orange, s - 1.0);
    else if (s < 3.0) return mix(orange, amber, s - 2.0);
    else              return mix(amber, cream, s - 3.0);
}

float oa_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float oa_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float oa_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(oa_hash2(i), oa_hash2(i + vec2(1, 0)), u.x),
               mix(oa_hash2(i + vec2(0, 1)), oa_hash2(i + vec2(1, 1)), u.x), u.y);
}

float oa_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < FBM_OCT; i++) {
        v += a * oa_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

// Palm tree SDF — curved trunk + 6 fronds radiating from the crown
float palmSDF(vec2 q, float baseX, float height) {
    // Trunk: quadratic Bezier-like curve
    float trunkD = 999.0;
    vec2 base = vec2(baseX, POOL_Y);
    vec2 crown = vec2(baseX + 0.03 * sign(q.x - baseX), POOL_Y + height);
    vec2 ctrl = vec2(baseX + 0.015, POOL_Y + height * 0.5);
    // Sample along the curve
    for (int i = 0; i < 10; i++) {
        float t1 = float(i) / 10.0;
        float t2 = float(i + 1) / 10.0;
        // Quadratic Bezier point
        float u1 = 1.0 - t1, u2 = 1.0 - t2;
        vec2 p1 = u1 * u1 * base + 2.0 * u1 * t1 * ctrl + t1 * t1 * crown;
        vec2 p2 = u2 * u2 * base + 2.0 * u2 * t2 * ctrl + t2 * t2 * crown;
        vec2 ab = p2 - p1;
        vec2 pa = q - p1;
        float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
        float d = length(pa - ab * h) - (0.008 + (1.0 - h * 0.5) * 0.005);
        trunkD = min(trunkD, d);
    }

    // Fronds: 6 curves radiating from crown at various angles
    float fronds = 999.0;
    for (int i = 0; i < 6; i++) {
        float fi = float(i);
        float ang = -0.4 + fi * 0.5;  // angles from -0.4 to ~2.1 rad
        float frondLen = 0.10 + oa_hash(fi * 1.7) * 0.04;
        vec2 tip = crown + vec2(sin(ang), cos(ang) * 0.85) * frondLen;
        // Frond segment
        vec2 ab = tip - crown;
        vec2 pa = q - crown;
        float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
        // Sag: droop based on h
        float sag = h * (1.0 - h) * 0.03;
        vec2 pointOnFrond = crown + ab * h + vec2(0.0, -sag);
        float d = length(q - pointOnFrond) - 0.006;
        fronds = min(fronds, d);
    }

    return min(trunkD, fronds);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.02, 0.04);

    // === Sky ===
    if (p.y > HORIZON) {
        float skyT = (p.y - HORIZON) / (0.9 - HORIZON);
        vec3 skyHor = oa_pal(0.5);                          // orange horizon
        vec3 skyHigh = mix(oa_pal(0.15), vec3(0.25, 0.15, 0.30), 0.5);
        col = mix(skyHor, skyHigh, skyT);
        // Sun disc just above horizon
        vec2 sunPos = vec2(0.15, HORIZON + 0.09);
        float sunD = length(p - sunPos);
        float sunDisc = smoothstep(0.065, 0.055, sunD);
        float sunGlow = exp(-sunD * sunD * 15.0) * 0.4;
        col += vec3(1.00, 0.88, 0.55) * sunDisc * 1.0;
        col += vec3(1.00, 0.60, 0.30) * sunGlow;
    }

    // === Dune silhouettes at horizon (far side of oasis) ===
    if (p.y > POOL_Y && p.y < HORIZON + 0.03) {
        float duneLine = HORIZON - 0.01 - 0.05 * sin(p.x * 3.5) - 0.03 * oa_fbm(vec2(p.x * 4.0, 0.0));
        if (p.y < duneLine) {
            float duneT = (duneLine - p.y) / 0.5;
            col = mix(col, vec3(0.25, 0.10, 0.10), 0.9 - duneT * 0.3);
        }
    }

    // === Water pool (elliptical) centered at (0, POOL_Y) ===
    vec2 poolRel = vec2(p.x, (p.y - POOL_Y) / 0.35);
    float poolD = length(poolRel);
    bool inPool = poolD < 0.42;

    if (inPool && p.y < POOL_Y + 0.1) {
        // Water: reflects sky above it
        vec2 mirror = vec2(p.x, POOL_Y - (p.y - POOL_Y) * 0.55);
        // Shimmer: horizontal distortion with time
        mirror.x += 0.008 * sin(p.y * 40.0 + x_Time * 2.0);
        // Sample reflection with darker tint
        float mirrorY = mirror.y + 0.2;  // offset to sample sky
        vec3 waterCol = oa_pal(fract(0.4 + mirrorY * 0.5)) * 0.55;
        // Deeper center: darker
        float centerDepth = 1.0 - smoothstep(0.0, 0.3, length(vec2(p.x, (p.y - POOL_Y) * 2.0)));
        waterCol = mix(waterCol, vec3(0.15, 0.05, 0.08), centerDepth * 0.4);
        // Surface ripples
        float ripple = 0.5 + 0.5 * sin(length(poolRel) * 80.0 - x_Time * 3.0);
        waterCol += oa_pal(0.75) * ripple * 0.1 * (1.0 - centerDepth);
        col = waterCol;
    }

    // === Ground (sand beyond pool + dune base) ===
    if (p.y < POOL_Y && !inPool) {
        float sandT = (POOL_Y - p.y) / 0.7;
        vec3 sandCol = mix(vec3(0.75, 0.45, 0.18), vec3(0.35, 0.15, 0.08), sandT);
        // Sand grain
        sandCol += oa_fbm(p * 20.0) * 0.1 * vec3(0.30, 0.20, 0.10);
        col = sandCol;
    }

    // === Palm trees ===
    for (int i = 0; i < PALMS; i++) {
        float fi = float(i);
        float seed = fi * 11.1;
        float baseX = (fi - 1.5) * 0.3 + (oa_hash(seed) - 0.5) * 0.04;
        // Skip if too close to center (would be in pool)
        float palmH = 0.30 + oa_hash(seed * 3.7) * 0.06;
        float sdf = palmSDF(p, baseX, palmH);
        if (sdf < 0.0) {
            col = mix(col, vec3(0.05, 0.02, 0.03), 0.92);
        } else if (sdf < 0.006) {
            float edge = smoothstep(0.006, 0.0, sdf);
            col = mix(col, oa_pal(0.3), edge * 0.5);
        }
    }

    // === Mirage shimmer near horizon ===
    if (p.y > HORIZON - 0.08 && p.y < HORIZON + 0.03) {
        float shimmerOff = 0.005 * sin(p.x * 60.0 + x_Time * 3.0);
        // Apply as slight color modulation
        col += oa_pal(0.7) * abs(shimmerOff) * 10.0 * 0.15;
    }

    // === Wind-blown sand particles ===
    for (int i = 0; i < SAND_PARTICLES; i++) {
        float fi = float(i);
        float seed = fi * 3.1;
        float speed = 0.25 + oa_hash(seed) * 0.35;
        float yBase = oa_hash(seed * 5.1) * 0.8 - 0.65;
        float x0 = fract(x_Time * speed + oa_hash(seed * 7.3)) * 2.6 - 1.3;
        // Vertical oscillation
        float y = yBase + 0.02 * sin(x0 * 15.0 + seed);
        vec2 pp = vec2(x0, y);
        float d = length(p - pp);
        col += vec3(0.95, 0.80, 0.55) * exp(-d * d * 100000.0) * 0.6;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
