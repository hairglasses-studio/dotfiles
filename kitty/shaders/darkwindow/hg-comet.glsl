// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Comet — bright nucleus + long ion + dust tails + coma with streaming particles

const int   DUST_PARTICLES = 40;
const int   ION_SAMPLES = 20;
const float INTENSITY = 0.55;

vec3 cmt_pal(float t) {
    vec3 white = vec3(1.00, 0.98, 0.90);
    vec3 cyan  = vec3(0.30, 0.85, 0.98);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 gold  = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(white, cyan, s);
    else if (s < 2.0) return mix(cyan, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, gold, s - 2.0);
    else              return mix(gold, white, s - 3.0);
}

float cmt_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.01, 0.03);

    // Comet orbits slowly
    float orbit = x_Time * 0.15;
    vec2 cometPos = 0.4 * vec2(cos(orbit), sin(orbit) * 0.6);
    // Tail direction — away from sun (anti-radial)
    vec2 tailDir = -normalize(cometPos);

    // Nucleus — bright point
    float nucD = length(p - cometPos);
    col += vec3(1.0, 0.98, 0.9) * exp(-nucD * nucD * 10000.0) * 2.0;
    col += cmt_pal(0.0) * exp(-nucD * nucD * 400.0) * 0.6;

    // Coma — large soft halo around nucleus
    col += cmt_pal(0.2) * exp(-nucD * nucD * 30.0) * 0.3;

    // Ion tail — long, narrow, bluish, straight-back
    // "Project" p onto tail axis
    vec2 toP = p - cometPos;
    float along = dot(toP, tailDir);
    float perp = abs(toP.x * tailDir.y - toP.y * tailDir.x);
    if (along > 0.0 && along < 1.0) {
        float tailW = 0.01 + along * 0.02;
        float ionMask = exp(-perp * perp / (tailW * tailW) * 2.0) * exp(-along * 1.5);

        // Streaming particles in ion tail
        float ionStreak = 0.0;
        for (int i = 0; i < ION_SAMPLES; i++) {
            float fi = float(i);
            float ionPos = fract(x_Time * 0.5 + fi * 0.05);
            float ionD = abs(along - ionPos);
            ionStreak += exp(-ionD * ionD * 500.0) * exp(-perp * perp * 5000.0);
        }
        col += vec3(0.3, 0.8, 0.98) * ionMask * 0.8;
        col += vec3(0.9, 0.95, 1.0) * ionStreak * 0.15;
    }

    // Dust tail — curved back (slightly different direction), yellowish
    vec2 dustDir = normalize(tailDir - vec2(tailDir.y, -tailDir.x) * 0.3);
    float dustAlong = dot(toP, dustDir);
    float dustPerp = abs(toP.x * dustDir.y - toP.y * dustDir.x);
    if (dustAlong > 0.0 && dustAlong < 0.7) {
        float dustW = 0.02 + dustAlong * 0.04;
        float dustMask = exp(-dustPerp * dustPerp / (dustW * dustW) * 1.5) * exp(-dustAlong * 2.0);
        col += vec3(0.95, 0.85, 0.65) * dustMask * 0.5;
    }

    // Dust particles floating along trail
    for (int i = 0; i < DUST_PARTICLES; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float life = fract(x_Time * 0.3 + cmt_hash(seed));
        vec2 dustPos = cometPos + dustDir * life * 0.6;
        // Perpendicular scatter
        dustPos += vec2(-dustDir.y, dustDir.x) * (cmt_hash(seed * 3.1) - 0.5) * 0.04 * life;
        float pd = length(p - dustPos);
        float core = exp(-pd * pd * 25000.0);
        col += vec3(0.95, 0.85, 0.55) * core * (1.0 - life) * 0.9;
    }

    // Background stars
    vec2 sg = floor(p * 120.0);
    float sh = fract(sin(dot(sg, vec2(127.1, 311.7))) * 43758.5);
    if (sh > 0.996) {
        col += vec3(0.85, 0.85, 1.0) * (sh - 0.996) * 200.0 * 0.5;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
