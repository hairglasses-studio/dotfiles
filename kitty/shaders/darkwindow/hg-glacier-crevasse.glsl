// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Glacier crevasse — looking straight down into a jagged ice crack with characteristic blue-depth gradient (blue light penetrates ice deepest). Ice walls with stratified layers + FBM texture, snow at top, occasional falling ice chunks, volumetric depth haze

const int   ICE_CHUNKS = 8;
const int   FBM_OCT = 4;
const float INTENSITY = 0.55;

vec3 cv_pal(float t) {
    vec3 deep    = vec3(0.01, 0.04, 0.12);
    vec3 darkBlue = vec3(0.05, 0.20, 0.45);
    vec3 teal    = vec3(0.20, 0.55, 0.75);
    vec3 cyan    = vec3(0.45, 0.80, 0.90);
    vec3 ice     = vec3(0.82, 0.92, 0.98);
    vec3 snow    = vec3(0.97, 0.99, 1.00);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(deep, darkBlue, s);
    else if (s < 2.0) return mix(darkBlue, teal, s - 1.0);
    else if (s < 3.0) return mix(teal, cyan, s - 2.0);
    else if (s < 4.0) return mix(cyan, ice, s - 3.0);
    else              return mix(ice, snow, s - 4.0);
}

float cv_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float cv_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float cv_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(cv_hash2(i), cv_hash2(i + vec2(1, 0)), u.x),
               mix(cv_hash2(i + vec2(0, 1)), cv_hash2(i + vec2(1, 1)), u.x), u.y);
}

float cv_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < FBM_OCT; i++) {
        v += a * cv_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

// Crevasse opening varies by y (depth). At top (y=0.45) it's wide, narrows
// as we go down. Returns the half-width of the opening at vertical position y.
float crevasseHalfWidth(float y) {
    // y ∈ [-1, 0.5]. Top wide, bottom narrow.
    float depth = (0.45 - y) / 1.4;
    depth = clamp(depth, 0.0, 1.0);
    // Half-width shrinks with depth, jittered slightly
    float base = 0.35 * (1.0 - depth * 0.7);
    float jitter = 0.04 * sin(y * 12.0) + 0.025 * cv_noise(vec2(y * 5.0, 0.0));
    return max(0.04, base + jitter);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.010, 0.025);

    // Crevasse half-width at this y
    float halfW = crevasseHalfWidth(p.y);
    float xDistToWall = abs(p.x) - halfW;

    // === Inside the crevasse (|x| < halfW) → blue depth gradient ===
    if (xDistToWall < 0.0) {
        // Depth parameter: 0 at top, 1 at deepest
        float depthT = (0.45 - p.y) / 1.4;
        depthT = clamp(depthT, 0.0, 1.0);
        // Color transitions from pale ice at top to deep dark blue
        vec3 depthCol = mix(cv_pal(0.75), cv_pal(0.0), pow(depthT, 0.8));
        col = depthCol;

        // Volumetric depth haze with subtle vertical streaks (water droplets, mist)
        float mist = cv_noise(p * vec2(4.0, 2.0) + vec2(0.0, x_Time * 0.2));
        col += cv_pal(0.35) * mist * 0.15 * (1.0 - depthT);

        // Horizontal ice stratification layers visible on the opposite walls
        float layer = sin(p.y * 30.0 + cv_fbm(p * 2.0)) * 0.5 + 0.5;
        col += cv_pal(0.55) * layer * 0.08 * (1.0 - depthT * 0.5);
    }

    // === Ice walls (jagged, textured) ===
    if (xDistToWall > -0.08 && xDistToWall < 0.3) {
        float wallT = smoothstep(-0.06, 0.0, xDistToWall);  // bright at outer edge of wall
        float ice1 = cv_fbm(p * 7.0);
        float ice2 = cv_fbm(p * 15.0 + 3.0);
        // Stratified layers (horizontal bands)
        float strat = smoothstep(0.65, 0.72, cv_fbm(vec2(0.0, p.y) * vec2(1.0, 12.0) + 2.0));

        vec3 wallCol = mix(vec3(0.35, 0.55, 0.70), vec3(0.75, 0.88, 0.95), ice1 * 0.6 + ice2 * 0.3);
        wallCol = mix(wallCol, cv_pal(0.75), strat * 0.4);
        // Crack highlights
        float crack = smoothstep(0.7, 0.85, cv_fbm(p * 25.0));
        wallCol *= 1.0 - crack * 0.4;

        // Blend with inside if we're near the wall edge from inside
        if (xDistToWall < 0.0) {
            col = mix(col, wallCol, wallT);
        } else {
            col = mix(col, wallCol * (0.5 + wallT * 0.8), 0.85);
        }
    }

    // === Snow banks at the top edges (if p.y > 0.40 and near wall) ===
    if (p.y > 0.38) {
        float topDepth = p.y - 0.38;
        float snowRise = 0.05 * cv_noise(vec2(p.x * 6.0, 0.0));
        if (topDepth > 0.0 && abs(p.x) > halfW - 0.03) {
            // Snow layer
            if (topDepth < 0.07 + snowRise) {
                col = mix(col, vec3(0.95, 0.97, 1.00), 0.92);
                // Sparkle
                float sparkle = cv_hash2(floor(p * 300.0));
                if (sparkle > 0.995) col += vec3(0.6, 0.8, 1.0) * (sparkle - 0.995) * 200.0;
            }
        }
    }

    // === Falling ice chunks inside crevasse ===
    for (int i = 0; i < ICE_CHUNKS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float cycle = 3.0 + cv_hash(seed) * 2.0;
        float fallT = fract((x_Time + cv_hash(seed * 3.1) * cycle) / cycle);
        float startX = (cv_hash(seed * 5.1) - 0.5) * 0.4;
        float yStart = 0.40;
        float yEnd = -0.85;
        float chunkY = mix(yStart, yEnd, fallT);
        // Horizontal oscillation (tumbling)
        float chunkX = startX + 0.02 * sin(fallT * 20.0 + seed);
        // Only render if chunk is inside crevasse (within halfW at that y)
        float chunkHW = crevasseHalfWidth(chunkY);
        if (abs(chunkX) < chunkHW * 0.9) {
            vec2 chunkPos = vec2(chunkX, chunkY);
            float cd = length(p - chunkPos);
            float chunkSize = 0.008 + cv_hash(seed * 11.0) * 0.01;
            float chunkMask = exp(-cd * cd / (chunkSize * chunkSize) * 1.5);
            col += vec3(0.85, 0.95, 1.00) * chunkMask * (1.0 - fallT * 0.4) * 0.9;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
