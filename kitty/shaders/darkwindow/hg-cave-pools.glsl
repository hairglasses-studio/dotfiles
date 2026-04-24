// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Cave pools — dark limestone cavern interior with stalactites hanging from the ceiling, bioluminescent puddles at the floor emitting cyan/teal glow, periodic drips from stalactite tips creating expanding ripples where they hit the water

const int   STALACTITES = 6;
const int   POOLS = 4;
const int   RIPPLES = 8;
const float INTENSITY = 0.55;

vec3 cp_pal(float t) {
    vec3 deep   = vec3(0.02, 0.08, 0.12);
    vec3 teal   = vec3(0.15, 0.55, 0.60);
    vec3 cyan   = vec3(0.30, 0.90, 0.95);
    vec3 mint   = vec3(0.35, 1.00, 0.70);
    vec3 violet = vec3(0.55, 0.30, 0.95);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(deep, teal, s);
    else if (s < 2.0) return mix(teal, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, mint, s - 2.0);
    else              return mix(mint, violet, s - 3.0);
}

float cp_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float cp_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float cp_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(cp_hash2(i), cp_hash2(i + vec2(1, 0)), u.x),
               mix(cp_hash2(i + vec2(0, 1)), cp_hash2(i + vec2(1, 1)), u.x), u.y);
}

// Stalactite tip at (cx, ctipY). Body is a triangle from ceiling down to tip.
vec2 stalactitePos(int i) {
    float fi = float(i);
    float seed = fi * 11.1;
    float cx = (fi + 0.5) / float(STALACTITES) * 2.0 - 1.0 + (cp_hash(seed) - 0.5) * 0.08;
    float ctipY = 0.30 - cp_hash(seed * 3.7) * 0.15;
    return vec2(cx, ctipY);
}

// Pool center + radius
vec2 poolPos(int i) {
    float fi = float(i);
    float seed = fi * 7.31;
    float cx = (fi + 0.5) / float(POOLS) * 2.0 - 1.0 + (cp_hash(seed) - 0.5) * 0.15;
    float cy = -0.40 + cp_hash(seed * 3.7) * 0.05;
    return vec2(cx, cy);
}
float poolRadius(int i) {
    return 0.08 + cp_hash(float(i) * 5.1) * 0.06;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Cave wall texture (dark brownish, FBM)
    float wall = cp_noise(p * 4.0) * 0.5 + cp_noise(p * 12.0) * 0.25;
    vec3 wallCol = mix(vec3(0.02, 0.025, 0.04), vec3(0.08, 0.07, 0.08), wall);
    vec3 col = wallCol;

    // === Ceiling shading (darker at top) ===
    if (p.y > 0.45) {
        col = mix(col, vec3(0.005, 0.008, 0.015), smoothstep(0.45, 0.85, p.y));
    }

    // === Pools ===
    for (int pi = 0; pi < POOLS; pi++) {
        vec2 pc = poolPos(pi);
        float pr = poolRadius(pi);
        // Squash y for elliptical puddle (seen from above-ish angle)
        vec2 pRel = vec2(p.x - pc.x, (p.y - pc.y) / 0.45);
        float pd = length(pRel);

        if (pd < pr) {
            // Inside pool: glow + bioluminescent bacteria
            float edge = smoothstep(pr, pr * 0.5, pd);
            // Bacteria pattern: small sparkles at various positions
            float bact = cp_noise(pRel * 40.0 + vec2(0.0, x_Time * 0.3));
            bact = smoothstep(0.7, 1.0, bact);
            vec3 poolCore = cp_pal(0.55) * (0.8 + edge * 0.5);
            poolCore += vec3(0.60, 1.00, 0.85) * bact * 1.2;
            col = mix(col, poolCore, 0.85);
            // Slight shimmer on surface (lighter interior)
            float shimmer = 0.7 + 0.3 * sin(pRel.x * 30.0 - x_Time * 2.0 + pRel.y * 20.0);
            col *= shimmer;
        }

        // Outer pool glow (escapes into cave)
        if (pd >= pr && pd < pr * 3.0) {
            float glowD = pd - pr;
            float glow = exp(-glowD * glowD * 100.0) * 0.35;
            col += cp_pal(0.5) * glow;
        }
    }

    // === Stalactites ===
    for (int i = 0; i < STALACTITES; i++) {
        vec2 tip = stalactitePos(i);
        // Stalactite body: from ceiling (y=0.85) to tip at (tip.x, tip.y)
        // Width decreases linearly from 0.04 at top to 0 at tip
        float yFrac = (p.y - tip.y) / (0.85 - tip.y);
        if (yFrac > 0.0 && yFrac < 1.0) {
            float bodyWidth = 0.03 * yFrac + 0.006;
            float xOffset = tip.x + (0.03 - 0.03) * (1.0 - yFrac);  // narrow at bottom
            if (abs(p.x - tip.x) < bodyWidth) {
                col = mix(col, vec3(0.06, 0.05, 0.08), 0.85);
                // Vertical striations
                float stripe = 0.9 + 0.1 * sin(p.y * 60.0);
                col *= stripe;
            }
        }

        // Faint wet glow at tip from hanging drop
        float tipD = length(p - tip);
        if (tipD < 0.015) {
            col += vec3(0.30, 0.65, 0.70) * exp(-tipD * tipD * 12000.0) * 0.5;
        }

        // Drip: small droplet falling at periodic intervals
        float dripCycle = 1.8 + cp_hash(float(i) * 1.7) * 1.5;
        float dripT = fract((x_Time + cp_hash(float(i) * 3.1) * dripCycle) / dripCycle);
        // Fall distance: accelerating with gravity
        float fallY = -0.7 * dripT * dripT;  // accelerating down
        float dropY = tip.y + fallY;
        // Stop at pool surface (-0.38)
        float poolSurface = -0.38;
        bool hitSurface = dropY <= poolSurface;
        if (!hitSurface) {
            vec2 dropPos = vec2(tip.x, dropY);
            float dd = length(p - dropPos);
            col += vec3(0.40, 0.80, 0.85) * exp(-dd * dd * 25000.0) * 1.2;
            // Trailing blur
            vec2 trailPos = vec2(tip.x, dropY + 0.015 * dripT);
            float td = length(p - trailPos);
            col += vec3(0.30, 0.65, 0.70) * exp(-td * td * 8000.0) * 0.4;
        }
    }

    // === Ripples on pool surfaces (expanding rings from drop impacts) ===
    for (int r = 0; r < RIPPLES; r++) {
        float fr = float(r);
        float seed = fr * 7.31;
        float rCycle = 2.0 + cp_hash(seed) * 2.0;
        float rT = fract((x_Time + cp_hash(seed * 3.1) * rCycle) / rCycle);
        // Which pool to ripple on
        int poolIdx = int(cp_hash(seed * 5.1) * float(POOLS));
        poolIdx = poolIdx % POOLS;
        vec2 pc = poolPos(poolIdx);
        float pr2 = poolRadius(poolIdx);
        // Impact point inside pool
        float impX = pc.x + (cp_hash(seed * 7.3) - 0.5) * pr2 * 0.6;
        float impY = pc.y + (cp_hash(seed * 11.1) - 0.5) * pr2 * 0.4;
        vec2 rippleC = vec2(impX, impY);
        // Squash for elliptical perspective
        vec2 rippleRel = vec2(p.x - rippleC.x, (p.y - rippleC.y) / 0.45);
        float rd = length(rippleRel);
        float rR = rT * pr2 * 1.3;
        float ringD = abs(rd - rR);
        float ringMask = exp(-ringD * ringD * 8000.0);
        float rFade = 1.0 - rT;
        col += vec3(0.70, 1.00, 0.90) * ringMask * rFade * 0.7;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
