// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Coral reef — branching SDF coral + drifting particles + glowing fish silhouettes

const int   CORAL_BRANCHES = 5;
const int   FISH_COUNT     = 8;
const int   PARTICLES      = 40;
const float INTENSITY      = 0.5;

vec3 cr_pal(float t) {
    vec3 mag   = vec3(0.95, 0.30, 0.55);  // pink coral
    vec3 orange = vec3(1.00, 0.55, 0.20);
    vec3 cyan  = vec3(0.15, 0.85, 0.95);
    vec3 mint  = vec3(0.20, 0.95, 0.70);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(mag, orange, s);
    else if (s < 2.0) return mix(orange, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, mint, s - 2.0);
    else if (s < 4.0) return mix(mint, vio, s - 3.0);
    else              return mix(vio, mag, s - 4.0);
}

float cr_hash(float n) { return fract(sin(n * 127.1) * 43758.5); }

// Branching coral SDF — recursive tree
// Each coral branch starts at (x, bottom) and grows upward with children
float coralDist(vec2 p, float rootX, float seed, float sway) {
    float minD = 1e9;
    // Main trunk (vertical) — base at bottom
    vec2 a = vec2(rootX, -0.5);
    vec2 b = a + vec2(0.02 * sway, 0.35);
    float d = abs(dot(p - a, vec2(1.0, 0.0)) - (p.y - a.y) * 0.02 * sway * 3.0);
    vec2 pa = p - a;
    vec2 ba = b - a;
    float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
    minD = min(minD, length(pa - ba * h));

    // Branches emerging at varying heights
    for (int br = 0; br < CORAL_BRANCHES; br++) {
        float fbr = float(br);
        float branchT = 0.3 + fbr * 0.15;
        vec2 branchRoot = mix(a, b, branchT);
        float branchSide = (fract(seed * 3.7 + fbr) > 0.5) ? 1.0 : -1.0;
        float branchAngle = branchSide * (0.4 + 0.2 * cr_hash(seed + fbr));
        vec2 branchTip = branchRoot + vec2(
            cos(1.5708 - branchAngle),
            sin(1.5708 - branchAngle)
        ) * 0.15 + vec2(0.015 * sway, 0.0);
        // Branch distance
        vec2 bpa = p - branchRoot;
        vec2 bba = branchTip - branchRoot;
        float bh = clamp(dot(bpa, bba) / dot(bba, bba), 0.0, 1.0);
        minD = min(minD, length(bpa - bba * bh));

        // Sub-branches (twigs)
        for (int tw = 0; tw < 2; tw++) {
            float ftw = float(tw);
            float twigT = 0.5 + ftw * 0.3;
            vec2 twigRoot = mix(branchRoot, branchTip, twigT);
            float twigSide = (fract(seed * 5.1 + fbr + ftw) > 0.5) ? 1.0 : -1.0;
            vec2 twigTip = twigRoot + vec2(twigSide * 0.05, 0.04);
            vec2 tpa = p - twigRoot;
            vec2 tba = twigTip - twigRoot;
            float th = clamp(dot(tpa, tba) / dot(tba, tba), 0.0, 1.0);
            minD = min(minD, length(tpa - tba * th));
        }
    }
    return minD;
}

// Fish silhouette — elongated ellipse moving horizontally
float fishDist(vec2 p, int i, float t) {
    float fi = float(i);
    float seed = fi * 7.31;
    float y = cr_hash(seed * 3.7) * 0.8 - 0.1;
    float speed = 0.08 + cr_hash(seed * 5.1) * 0.05;
    float dir = (fract(seed * 11.0) > 0.5) ? 1.0 : -1.0;
    float x = mod(t * speed * dir + cr_hash(seed), 2.0) - 1.0;
    x *= x_WindowSize.x / x_WindowSize.y;
    // Wobble
    y += 0.02 * sin(t * 2.0 + fi);
    vec2 fishP = vec2(x, y);
    vec2 d = (p - fishP) * vec2(0.7, 2.0);   // elongated horizontally
    return length(d) - 0.02;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Underwater blue-green gradient
    vec3 water = mix(vec3(0.02, 0.12, 0.20), vec3(0.04, 0.25, 0.35), uv.y);
    vec3 col = water;

    // Coral silhouettes
    float sway = sin(x_Time * 0.5) * 0.6;
    float minCoralD = 1e9;
    float coralHue = 0.0;
    for (int c = 0; c < 4; c++) {
        float fc = float(c);
        float rootX = (fc - 1.5) * 0.28 * x_WindowSize.x / x_WindowSize.y / 2.0;
        float d = coralDist(p, rootX, fc * 3.7, sway + fc * 0.3);
        if (d < minCoralD) { minCoralD = d; coralHue = fract(fc * 0.17 + x_Time * 0.02); }
    }
    float coralMask = smoothstep(0.01, 0.0, minCoralD);
    float coralGlow = exp(-minCoralD * minCoralD * 1500.0) * 0.4;
    vec3 coralCol = cr_pal(coralHue);
    col = mix(col, coralCol * 0.9, coralMask);
    col += coralCol * coralGlow;

    // Floating particles / plankton
    for (int pi = 0; pi < PARTICLES; pi++) {
        float fpi = float(pi);
        float seed = fpi * 3.17;
        vec2 pp = vec2(
            cr_hash(seed) * 2.0 - 1.0,
            mod(cr_hash(seed * 5.1) + x_Time * 0.03, 1.5) - 0.75
        );
        pp.x *= x_WindowSize.x / x_WindowSize.y;
        pp.x += 0.01 * sin(x_Time + fpi * 1.3);
        float pd = length(p - pp);
        float pm = exp(-pd * pd * 30000.0);
        col += vec3(0.7, 0.9, 1.0) * pm * 0.6;
    }

    // Fish glow silhouettes
    for (int f = 0; f < FISH_COUNT; f++) {
        float fd = fishDist(p, f, x_Time);
        float fm = smoothstep(0.015, 0.0, fd);
        vec3 fishCol = cr_pal(fract(float(f) * 0.12 + x_Time * 0.03));
        col = mix(col, fishCol * 0.7, fm);
        col += fishCol * exp(-fd * fd * 300.0) * 0.25;
    }

    // Surface caustic tint (subtle upward brighten)
    float surfHint = smoothstep(0.3, 0.6, uv.y) * 0.15;
    col += vec3(0.85, 0.95, 1.0) * surfHint;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
