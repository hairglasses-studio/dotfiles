// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Growing crystal dendrites — branching 6-fold pattern expanding from center

const int   BRANCH_DEPTH = 5;
const float GROWTH_PERIOD = 6.0;
const float INTENSITY    = 0.55;

vec3 cg_pal(float t) {
    vec3 a = vec3(0.85, 0.95, 1.00);
    vec3 b = vec3(0.20, 0.75, 0.95);
    vec3 c = vec3(0.55, 0.30, 0.98);
    vec3 d = vec3(0.96, 0.85, 0.45);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float cg_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Snowflake-style branching: recursively fold into a wedge, draw branches
float crystalDist(vec2 p, float growth, float t) {
    float minD = 1e9;
    // 6-fold symmetry
    float ang = atan(p.y, p.x);
    float wedge = 6.28318 / 6.0;
    float fa = mod(ang, wedge) - wedge * 0.5;
    float r = length(p);
    vec2 wp = vec2(cos(fa), sin(fa)) * r;

    // Main trunk — line from origin outward
    float trunkY = wp.y;
    float trunkX = wp.x;
    // Only within growth radius
    if (trunkX > growth) return 1e9;
    float trunkD = abs(trunkY);
    if (trunkX > 0.0 && trunkX < growth) {
        minD = min(minD, trunkD);
    }

    // Branch arms at 4 locations along the trunk
    for (int b = 0; b < BRANCH_DEPTH; b++) {
        float fb = float(b);
        float branchT = 0.15 + fb * 0.15;
        float branchX = branchT;
        if (branchX > growth) continue;
        // Branch grows outward when trunk reaches this point
        float branchGrowth = clamp((growth - branchX) * 4.0, 0.0, 0.08 + fb * 0.01);
        // Two branches per side, ~60° angle
        for (int s = 0; s < 2; s++) {
            float signM = (s == 0 ? 1.0 : -1.0);
            vec2 branchStart = vec2(branchX, 0.0);
            vec2 branchDir = vec2(cos(0.5236), sin(0.5236) * signM);  // 30°
            vec2 branchEnd = branchStart + branchDir * branchGrowth;
            vec2 bp = wp;
            vec2 ba = branchEnd - branchStart;
            vec2 pa = bp - branchStart;
            float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
            float bd = length(pa - ba * h);
            minD = min(minD, bd);
        }
    }

    return minD;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Growth cycle
    float cyclePhase = mod(x_Time, GROWTH_PERIOD) / GROWTH_PERIOD;
    float growth = pow(cyclePhase, 0.7) * 0.5;

    // Slow rotation
    float rotAng = x_Time * 0.05;
    float cr = cos(rotAng), sr = sin(rotAng);
    vec2 rp = mat2(cr, -sr, sr, cr) * p;

    float d = crystalDist(rp, growth, x_Time);

    vec3 col = vec3(0.0);

    // Crystal edge (thin)
    float core = smoothstep(0.004, 0.0, d);
    float glow = exp(-d * d * 5000.0) * 0.4;
    float outerHalo = exp(-d * d * 500.0) * 0.15;

    vec3 cc = cg_pal(fract(growth + x_Time * 0.04));
    col += cc * (core * 1.4 + glow + outerHalo);

    // Bright tip where growth frontier is
    float tipDist = length(p) - growth;
    if (abs(tipDist) < 0.02 && d < 0.01) {
        col += vec3(1.0, 0.98, 0.95) * 1.2;
    }

    // Ambient background (cold dark)
    col += vec3(0.02, 0.04, 0.08);

    // Dust particles freezing onto crystal (sparkles near edges)
    float sparkle = 0.0;
    if (d < 0.02 && d > 0.005) {
        sparkle = step(0.98, cg_hash(floor(p.x * 100.0) * 31.0 + floor(p.y * 100.0) + floor(x_Time * 4.0)));
    }
    col += vec3(1.0) * sparkle * 0.6;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
