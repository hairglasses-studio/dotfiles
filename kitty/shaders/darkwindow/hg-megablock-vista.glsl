// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Megablock vista — layered cyberpunk city silhouette with three depth planes, lit neon windows, two big pulsing neon signs, searchlight sweeps, smog gradient at horizon, and blinking airship beacons

const int   BLOCKS_NEAR = 7;
const int   BLOCKS_MID  = 11;
const int   BLOCKS_FAR  = 16;
const int   WIND_COLS   = 6;
const int   WIND_ROWS   = 14;
const int   AIRSHIPS    = 4;
const float HORIZON = -0.12;   // where the smog floor sits
const float INTENSITY = 0.55;

vec3 mb_pal(float t) {
    vec3 cyan   = vec3(0.20, 0.85, 1.00);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 amber  = vec3(1.00, 0.70, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, amber, s - 2.0);
    else              return mix(amber, mint, s - 3.0);
}

float mb_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float mb_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

// Returns block silhouette top y (in p.y space) for pixel column p.x within
// a given depth layer. Layers: 0 = near, 1 = mid, 2 = far.
// Blocks are rectangles with x-center, half-width, and top height.
// Using ascending index + per-block random width & position.
float blockSkyline(float px, int layer) {
    int blocks;
    float ySpan;       // max height above HORIZON
    float zDepth;      // depth factor
    if (layer == 0) { blocks = BLOCKS_NEAR; ySpan = 0.7; zDepth = 0.0; }
    else if (layer == 1) { blocks = BLOCKS_MID; ySpan = 0.50; zDepth = 0.4; }
    else { blocks = BLOCKS_FAR; ySpan = 0.30; zDepth = 0.8; }

    float skyline = HORIZON;  // default: no block → sky
    for (int i = 0; i < blocks; i++) {
        float fi = float(i);
        float seed = fi * 7.31 + float(layer) * 97.0;
        float cx = mb_hash(seed) * 2.2 - 1.1;   // block center x
        float hw = 0.045 + mb_hash(seed * 3.7) * 0.065 * (1.0 - zDepth * 0.5);  // half-width
        float h  = 0.10 + mb_hash(seed * 5.1) * ySpan;
        // Is p.x within the block column?
        if (abs(px - cx) < hw) {
            float blockTop = HORIZON + h;
            if (blockTop > skyline) skyline = blockTop;
        }
    }
    return skyline;
}

// Test whether a pixel is inside any block of the given layer, return the
// block's index-hash and top (so we can light windows).
// Returns (inBlock, blockSeed, blockTopY, blockCenterX, blockHalfWidth)
vec4 blockInfo(vec2 p, int layer) {
    int blocks;
    float ySpan;
    if (layer == 0) { blocks = BLOCKS_NEAR; ySpan = 0.7; }
    else if (layer == 1) { blocks = BLOCKS_MID; ySpan = 0.50; }
    else { blocks = BLOCKS_FAR; ySpan = 0.30; }

    float bestBlockSeed = -1.0;
    float bestTop = HORIZON;
    float bestCX = 0.0;
    float bestHW = 0.0;
    for (int i = 0; i < blocks; i++) {
        float fi = float(i);
        float seed = fi * 7.31 + float(layer) * 97.0;
        float cx = mb_hash(seed) * 2.2 - 1.1;
        float hw = 0.045 + mb_hash(seed * 3.7) * 0.065 * (1.0 - float(layer) * 0.25);
        float h  = 0.10 + mb_hash(seed * 5.1) * ySpan;
        if (abs(p.x - cx) < hw) {
            float blockTop = HORIZON + h;
            if (p.y < blockTop && p.y > HORIZON && blockTop > bestTop) {
                bestBlockSeed = seed;
                bestTop = blockTop;
                bestCX = cx;
                bestHW = hw;
            }
        }
    }
    return vec4(bestBlockSeed, bestTop, bestCX, bestHW);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.003, 0.006, 0.020);

    // === Sky gradient (dark at top, smoggy near horizon) ===
    float altAbove = max(p.y - HORIZON, 0.0);
    vec3 skyTop  = vec3(0.015, 0.020, 0.060);
    vec3 skyHor  = vec3(0.12, 0.08, 0.15) + mb_pal(fract(x_Time * 0.02)) * 0.08;
    float skyMix = exp(-altAbove * 2.5);
    col = mix(skyTop, skyHor, skyMix);

    // Subtle smog haze at horizon
    float hazeBand = exp(-pow(p.y - HORIZON - 0.04, 2.0) * 80.0) * 0.4;
    col += vec3(0.25, 0.15, 0.30) * hazeBand;

    // === Searchlight sweeps in the upper sky ===
    for (int i = 0; i < 3; i++) {
        float fi = float(i);
        float seed = fi * 13.1;
        vec2 slOrigin = vec2(-1.2 + fi * 0.8, HORIZON + 0.02);
        float slAng = 1.5 + 0.8 * sin(x_Time * (0.15 + mb_hash(seed) * 0.1) + fi);
        vec2 slDir = vec2(cos(slAng), sin(slAng));
        vec2 rel = p - slOrigin;
        float along = dot(rel, slDir);
        float perp = abs(rel.x * slDir.y - rel.y * slDir.x);
        if (along > 0.0 && along < 1.5 && p.y > HORIZON) {
            float width = 0.025 + along * 0.05;
            float beam = exp(-perp * perp / (width * width) * 2.0) * (1.0 - along / 1.5);
            col += mb_pal(fract(0.3 + fi * 0.2)) * beam * 0.25;
        }
    }

    // === Blinking airships (high sky) ===
    for (int i = 0; i < AIRSHIPS; i++) {
        float fi = float(i);
        float seed = fi * 29.1;
        float speed = 0.03 + mb_hash(seed) * 0.05;
        float phase = mod(x_Time * speed + mb_hash(seed * 3.1), 1.0);
        vec2 apos = vec2(-1.2 + phase * 2.4, 0.55 + mb_hash(seed * 5.1) * 0.30);
        float ad = length(p - apos);
        float blink = step(0.5, sin(x_Time * 2.0 + fi * 3.7));
        col += mb_pal(fract(fi * 0.25)) * exp(-ad * ad * 30000.0) * blink * 0.7;
    }

    // === Megablock layers: far → mid → near ===
    // For each pixel, check layers from far to near; near wins.
    // Use depth to colorize blocks and lit windows.

    // Below HORIZON: smog floor (no buildings rendered below)
    if (p.y < HORIZON) {
        float floorDepth = (HORIZON - p.y) / 0.5;
        vec3 floorCol = mix(vec3(0.08, 0.04, 0.10), vec3(0.02, 0.015, 0.04), clamp(floorDepth, 0.0, 1.0));
        col = mix(col, floorCol, 0.8);
    } else {
        // Far layer
        vec4 farInfo = blockInfo(p, 2);
        if (farInfo.x >= 0.0) {
            vec3 farSil = vec3(0.07, 0.06, 0.12);
            col = mix(col, farSil, 0.85);
            // Tiny sparse far-window lights
            float lit = step(0.92, mb_hash2(floor(p * 100.0) + farInfo.x));
            col += vec3(0.8, 0.55, 0.30) * lit * 0.25;
        }
        // Mid layer
        vec4 midInfo = blockInfo(p, 1);
        if (midInfo.x >= 0.0) {
            vec3 midSil = vec3(0.03, 0.025, 0.055);
            col = mix(col, midSil, 0.95);
            // Mid-layer windows
            vec2 cell = floor(p * 40.0);
            float litSeed = mb_hash2(cell + midInfo.x * 3.1);
            if (litSeed > 0.72) {
                float flick = 0.7 + 0.3 * sin(x_Time * (3.0 + litSeed * 5.0) + litSeed * 13.0);
                col += mb_pal(fract(litSeed * 1.3 + x_Time * 0.05)) * flick * 0.35;
            }
        }
        // Near layer
        vec4 nearInfo = blockInfo(p, 0);
        if (nearInfo.x >= 0.0) {
            col = mix(col, vec3(0.005, 0.003, 0.012), 0.97);

            // Near-layer bright windows on a fine grid
            vec2 cell = floor(p * 60.0);
            float litSeed = mb_hash2(cell + nearInfo.x * 7.3);
            if (litSeed > 0.6) {
                float flick = 0.6 + 0.4 * sin(x_Time * (5.0 + litSeed * 7.0) + cell.x * 1.3 + cell.y * 0.7);
                col += mb_pal(fract(litSeed * 1.7 + x_Time * 0.07)) * flick * 0.8;
            }

            // Big neon sign on the 2 largest near blocks (identified by larger seed positioned high)
            // Render a bright horizontal band about 0.55 up the block
            float bandY = HORIZON + (nearInfo.y - HORIZON) * 0.65;
            float signBand = exp(-pow(p.y - bandY, 2.0) * 600.0);
            float signEdge = smoothstep(nearInfo.w * 0.95, nearInfo.w * 0.85, abs(p.x - nearInfo.z));
            if (signBand > 0.15 && signEdge > 0.1) {
                float pulse = 0.6 + 0.4 * sin(x_Time * 2.0 + nearInfo.x);
                col += mb_pal(fract(nearInfo.x * 0.5 + x_Time * 0.1)) * signBand * signEdge * pulse * 1.5;
            }
        }
    }

    // Bottom street glow (smog lit from below by streetlights)
    float streetGlow = exp(-pow(p.y + 0.35, 2.0) * 20.0) * 0.2;
    col += mb_pal(fract(x_Time * 0.03 + 0.3)) * streetGlow;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
