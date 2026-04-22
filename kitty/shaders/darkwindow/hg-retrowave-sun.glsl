// Shader attribution: hairglasses (original)
// Technique inspired by: the canonical Shadertoy retrowave scene (e.g. 7lyGzh) — sun + stripes + grid.
// License: MIT
// (Cyberpunk — showcase/heavy) — Full Miami Vice scene: striped sun, mountains silhouette, perspective grid floor, starfield

const float HORIZON      = 0.48;
const float GRID_SCROLL  = 0.7;
const float GRID_SPACING = 8.0;
const float LINE_WIDTH   = 0.010;
const float SUN_RADIUS   = 0.22;
const int   MOUNTAIN_N   = 6;   // mountain ridges
const float INTENSITY    = 0.55;

const vec3 SKY_DEEP    = vec3(0.06, 0.02, 0.16);
const vec3 SKY_NEAR    = vec3(0.45, 0.08, 0.42);
const vec3 SUN_BRIGHT  = vec3(1.00, 0.92, 0.45);
const vec3 SUN_MID     = vec3(1.00, 0.40, 0.55);
const vec3 SUN_DARK    = vec3(0.58, 0.08, 0.38);
const vec3 GROUND_DARK = vec3(0.02, 0.01, 0.08);
const vec3 GRID_CYAN   = vec3(0.10, 0.85, 0.95);
const vec3 MOUNTAIN    = vec3(0.18, 0.04, 0.28);

float rw_hash(float x) { return fract(sin(x * 127.1) * 43758.5); }
float rw_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// Mountain silhouette via layered 1D noise
float mountainHeight(float x, float seed) {
    float h = 0.0;
    float amp = 1.0;
    for (int i = 0; i < 4; i++) {
        float f = pow(2.0, float(i));
        float xa = x * f + seed;
        float ix = floor(xa);
        float fx = fract(xa);
        fx = fx * fx * (3.0 - 2.0 * fx);
        h += amp * mix(rw_hash(ix + seed * 13.0), rw_hash(ix + 1.0 + seed * 13.0), fx);
        amp *= 0.5;
    }
    return h;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec3 bg;

    if (uv.y > HORIZON) {
        // --- Sky ---
        float skyT = (uv.y - HORIZON) / (1.0 - HORIZON);
        bg = mix(SKY_NEAR, SKY_DEEP, smoothstep(0.0, 0.8, skyT));

        // Starfield — sparse bright dots
        vec2 starGrid = uv * vec2(x_WindowSize.x / x_WindowSize.y * 80.0, 80.0);
        vec2 starCell = floor(starGrid);
        vec2 starF = fract(starGrid);
        float starSeed = rw_hash2(starCell);
        if (starSeed > 0.985) {
            float twinkle = 0.5 + 0.5 * sin(x_Time * (1.0 + starSeed * 5.0) + starSeed * 20.0);
            float star = smoothstep(0.4, 0.0, length(starF - 0.5));
            bg += vec3(0.9, 0.85, 1.0) * star * twinkle * skyT;
        }

        // --- Sun ---
        vec2 sunUV = uv - vec2(0.5, HORIZON + SUN_RADIUS * 0.75);
        sunUV.y *= 1.1;
        float sunDist = length(sunUV);
        float sunMask = smoothstep(SUN_RADIUS, SUN_RADIUS * 0.98, sunDist);
        // Gradient colors top→bottom
        float sunGradT = clamp((0.5 - sunUV.y / SUN_RADIUS) * 0.5 + 0.3, 0.0, 1.0);
        vec3 sunCol = mix(SUN_DARK, mix(SUN_MID, SUN_BRIGHT, sunGradT), sunGradT);
        // Horizontal stripes (gap expands near bottom)
        float stripeY = (HORIZON + SUN_RADIUS * 0.75 - uv.y) / SUN_RADIUS;
        float stripeWidth = 0.18 * clamp(stripeY, 0.0, 1.0);
        float stripeMask = step(0.5, fract(stripeY * 9.0));
        // Stripe cutout only in bottom half
        float stripeActive = smoothstep(-0.1, 0.15, stripeY);
        vec3 sunFinal = sunCol * (1.0 - (1.0 - stripeMask) * stripeActive * 0.9);
        bg = mix(bg, sunFinal, sunMask);
        // Sun halo
        float halo = exp(-sunDist * 4.0) * 0.22;
        bg += SUN_MID * halo * 0.5;

        // --- Mountains silhouette ---
        float mountMask = 0.0;
        for (int i = 0; i < MOUNTAIN_N; i++) {
            float fi = float(i);
            float yBase = HORIZON + 0.01 + fi * 0.01;
            float heightScale = 0.04 + fi * 0.008;
            float mountainY = yBase + mountainHeight(uv.x * (4.0 - fi * 0.3) + fi * 7.0, fi * 2.7) * heightScale;
            if (uv.y < mountainY) {
                float depthShade = 0.25 + fi * 0.08;
                bg = mix(bg, MOUNTAIN * depthShade, 1.0 - mountMask);
                mountMask = 1.0;
            }
        }
    } else {
        // --- Ground: perspective grid floor ---
        float groundY = (HORIZON - uv.y) / HORIZON;
        float depth = 1.0 / max(groundY, 0.002);
        // Horizontal rows
        float hRow = depth - x_Time * GRID_SCROLL;
        float hLine = 1.0 - smoothstep(0.0, LINE_WIDTH * depth * 5.0,
            abs(fract(hRow) - 0.5) * 2.0 - 1.0 + LINE_WIDTH * depth * 5.0);
        // Vertical lanes
        float xn = (uv.x - 0.5) * 2.0;
        float worldX = xn * depth;
        float vLine = 1.0 - smoothstep(0.0, LINE_WIDTH * 0.8,
            abs(fract(worldX * GRID_SPACING * 0.5 + 0.5) - 0.5));
        float distFade = smoothstep(0.0, 0.35, groundY);
        float grid = (hLine + vLine) * distFade;
        // Ground base: dark → subtle magenta under horizon
        vec3 ground = mix(SKY_NEAR * 0.3, GROUND_DARK, groundY);
        bg = mix(ground, GRID_CYAN, clamp(grid, 0.0, 1.0) * 0.9);
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.7);
    vec3 result = mix(terminal.rgb, bg, visibility);

    _wShaderOut = vec4(result, 1.0);
}
