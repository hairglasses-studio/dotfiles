// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Snowflake tiling — 6-fold-symmetric snowflakes on hexagonal grid with glowing edges

const float CELL_SIZE = 0.18;
const int   FRACTAL_DEPTH = 3;
const float INTENSITY = 0.55;

vec3 sf_pal(float t) {
    vec3 ice   = vec3(0.70, 0.90, 1.00);
    vec3 cyan  = vec3(0.30, 0.85, 0.98);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 mag   = vec3(0.90, 0.35, 0.70);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(ice, cyan, s);
    else if (s < 2.0) return mix(cyan, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else              return mix(mag, ice, s - 3.0);
}

float sf_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// Distance from point to center — take into account 6-fold fold
float snowflakeDist(vec2 p, float seed) {
    // 6-fold rotation fold
    float a = atan(p.y, p.x);
    float wedge = 6.28318 / 6.0;
    float fa = mod(a, wedge) - wedge * 0.5;
    float r = length(p);
    vec2 wp = vec2(cos(fa), sin(fa)) * r;

    // Main trunk
    float minD = abs(wp.y);
    if (wp.x > 0.06 || wp.x < 0.0) minD = 1e9;

    // Branches at different heights along the trunk
    for (int b = 0; b < FRACTAL_DEPTH; b++) {
        float fb = float(b);
        float branchT = 0.15 + fb * 0.15;
        if (branchT > 0.06) continue;
        vec2 branchStart = vec2(branchT, 0.0);
        // Mirror branches on each side (60° from trunk)
        for (int s = 0; s < 2; s++) {
            float signM = (s == 0 ? 1.0 : -1.0);
            vec2 branchDir = vec2(cos(1.0472), sin(1.0472) * signM);  // 60°
            float branchLen = 0.025 + sf_hash(vec2(seed, fb)) * 0.01;
            vec2 branchEnd = branchStart + branchDir * branchLen;
            vec2 pa = wp - branchStart;
            vec2 ba = branchEnd - branchStart;
            float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
            float d = length(pa - ba * h);
            minD = min(minD, d);

            // Sub-branches
            for (int ss = 0; ss < 2; ss++) {
                float subT = 0.4 + float(ss) * 0.3;
                vec2 subStart = mix(branchStart, branchEnd, subT);
                float subSignM = (ss == 0 ? 1.0 : -1.0);
                vec2 subDir = vec2(cos(1.0472), sin(1.0472) * subSignM);
                vec2 subEnd = subStart + subDir * 0.01;
                vec2 sa = wp - subStart;
                vec2 sb = subEnd - subStart;
                float sh = clamp(dot(sa, sb) / dot(sb, sb), 0.0, 1.0);
                float sd = length(sa - sb * sh);
                minD = min(minD, sd);
            }
        }
    }

    return minD;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Hex grid
    vec2 hexAxis = vec2(p.x / CELL_SIZE, (p.y + CELL_SIZE * mod(floor(p.x / CELL_SIZE) * 0.5, 1.0)) / (CELL_SIZE * 0.866));
    vec2 cellId = floor(hexAxis);
    vec2 cellCenter = vec2(cellId.x * CELL_SIZE, (cellId.y + mod(cellId.x, 2.0) * 0.5) * CELL_SIZE * 0.866);
    vec2 cellP = p - cellCenter;

    // Each cell has its own snowflake seed (controls variation) + rotation
    float seed = sf_hash(cellId);
    float rotAng = x_Time * 0.3 + seed * 6.28;
    float cr = cos(rotAng), sr = sin(rotAng);
    vec2 rp = mat2(cr, -sr, sr, cr) * cellP;

    float d = snowflakeDist(rp, seed);

    vec3 col = vec3(0.01, 0.02, 0.05);

    float core = smoothstep(0.002, 0.0, d);
    float glow = exp(-d * d * 2000.0) * 0.4;
    vec3 sfCol = sf_pal(fract(seed + x_Time * 0.04));
    col += sfCol * (core * 1.3 + glow);

    // Sparkle at center of each snowflake
    float centerD = length(cellP);
    col += sfCol * exp(-centerD * centerD * 20000.0) * 0.6;

    // Drifting downward snow particles overlaid
    for (int sp = 0; sp < 30; sp++) {
        float fsp = float(sp);
        float sSeed = fsp * 7.3;
        float fallSpeed = 0.1 + sf_hash(vec2(sSeed, 0.0)) * 0.1;
        float phase = fract(x_Time * fallSpeed + sf_hash(vec2(sSeed, 1.0)));
        vec2 pp = vec2(
            (sf_hash(vec2(sSeed, 2.0)) - 0.5) * 2.0 * x_WindowSize.x / x_WindowSize.y,
            1.0 - phase * 2.0
        );
        pp.x += 0.03 * sin(x_Time * 0.5 + fsp);
        float pd = length(p - pp);
        col += vec3(0.9, 0.95, 1.0) * exp(-pd * pd * 20000.0) * 0.5;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
