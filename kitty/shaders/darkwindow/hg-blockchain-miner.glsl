// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Blockchain miner — horizontal chain of 7 blocks scrolling left; newest block on the right rapidly scrambles hash cells as the nonce is searched, then commits at cycle boundary and chain advances. Chain links connect blocks with a traveling pulse.

const int   BLOCKS = 7;
const int   HASH_COLS = 8;
const int   HASH_ROWS = 4;
const float BLOCK_CYCLE = 4.0;   // time to mine a new block
const float INTENSITY = 0.55;

vec3 bm_pal(float t) {
    vec3 cyan   = vec3(0.25, 0.90, 1.00);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 amber  = vec3(1.00, 0.70, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, amber, s - 2.0);
    else if (s < 4.0) return mix(amber, mint, s - 3.0);
    else              return mix(mint, cyan, s - 4.0);
}

float bm_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float bm_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.008, 0.022);

    // Faint gridlines
    vec2 gp = fract(p * 12.0);
    float gx = smoothstep(0.015, 0.0, abs(gp.x - 0.5) - 0.48);
    float gy = smoothstep(0.015, 0.0, abs(gp.y - 0.5) - 0.48);
    col += vec3(0.08, 0.10, 0.18) * max(gx, gy) * 0.18;

    // Blockchain cycle progress
    float cycT = mod(x_Time, BLOCK_CYCLE) / BLOCK_CYCLE;
    // Continuous scroll offset: as cycle progresses, chain scrolls leftward
    // When cycle completes, a new block is added and existing blocks shift
    float scrollOff = cycT * 0.28;  // smooth scroll fraction over cycle

    float blockW = 0.22;
    float blockH = 0.28;
    float blockSpacing = 0.26;
    float blockY = 0.0;

    // === Chain links between blocks ===
    for (int i = 0; i < BLOCKS; i++) {
        float fi = float(i);
        float bx = (float(BLOCKS - 1 - i)) * blockSpacing - 0.72 - scrollOff;
        float bxPrev = (float(BLOCKS - 2 - i)) * blockSpacing - 0.72 - scrollOff;
        if (i < BLOCKS - 1) {
            // Link from block[i] right edge to block[i+1] left edge
            vec2 linkStart = vec2(bx + blockW * 0.5, blockY);
            vec2 linkEnd = vec2(bxPrev - blockW * 0.5, blockY);
            // only render if meaningful segment
            if (linkEnd.x > linkStart.x) {
                vec2 ab = linkEnd - linkStart;
                vec2 pa = p - linkStart;
                float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
                float d = length(pa - ab * h);
                col += bm_pal(0.7) * exp(-d * d * 80000.0) * 0.7;
                // Pulse along link
                float pulseT = fract(x_Time * 1.5 + fi * 0.2);
                vec2 pulsePos = mix(linkStart, linkEnd, pulseT);
                float pd = length(p - pulsePos);
                col += vec3(1.0, 0.95, 0.80) * exp(-pd * pd * 30000.0) * 1.1;
            }
        }
    }

    // === Blocks ===
    for (int i = 0; i < BLOCKS; i++) {
        float fi = float(i);
        // i=0 is newest (rightmost, actively mining)
        float bx = (float(BLOCKS - 1 - i)) * blockSpacing - 0.72 - scrollOff;
        // Block rectangle centered at (bx, blockY), width=blockW, height=blockH
        vec2 rel = p - vec2(bx, blockY);
        if (abs(rel.x) < blockW * 0.5 && abs(rel.y) < blockH * 0.5) {
            // Inside block
            bool isNewest = (i == 0);
            bool isMined = (i > 0);

            // Background: dark translucent
            col = mix(col, vec3(0.015, 0.020, 0.030), 0.92);

            // Block border
            float rim = min(blockW * 0.5 - abs(rel.x), blockH * 0.5 - abs(rel.y));
            float rimMask = smoothstep(0.005, 0.0, rim);
            vec3 rimCol = isNewest ? vec3(1.0, 0.95, 0.50) : bm_pal(fract(fi * 0.15 + 0.3));
            col += rimCol * rimMask * 1.1;

            // Hash grid cells (8x4)
            vec2 gridRel = (rel / vec2(blockW * 0.5, blockH * 0.5));  // ∈ [-1, 1]
            gridRel = (gridRel + 1.0) * 0.5;  // ∈ [0, 1]
            vec2 cellCoord = floor(gridRel * vec2(float(HASH_COLS), float(HASH_ROWS)));

            float cellHash;
            if (isNewest) {
                // Actively scrambling: hash changes very rapidly
                float scrambleT = floor(x_Time * 30.0);  // 30 scrambles/s
                cellHash = bm_hash2(cellCoord + vec2(scrambleT, fi));
            } else {
                // Mined block: stable hash (per-block seed)
                float blockSeed = fi * 13.7 + floor((x_Time - cycT * BLOCK_CYCLE) / BLOCK_CYCLE);
                cellHash = bm_hash2(cellCoord + vec2(blockSeed * 3.7, 0.0));
            }

            // Render hash cell
            vec2 cellFrac = fract(gridRel * vec2(float(HASH_COLS), float(HASH_ROWS)));
            // Cell content: a small bar whose height encodes the hash value
            float barH = cellHash * 0.7 + 0.1;
            if (cellFrac.y < barH && cellFrac.y > 0.1) {
                vec3 hashCol = isNewest ? vec3(1.0, 0.95, 0.50) : bm_pal(fract(cellHash + fi * 0.1));
                col += hashCol * 0.55;
            }

            // Block height label bar at bottom
            if (rel.y < -blockH * 0.4 && rel.y > -blockH * 0.5 + 0.005) {
                col += rimCol * 0.3;
            }
        }
    }

    // === Leftmost block fades as it scrolls off screen ===
    // (Handled implicitly by position)

    // === Mining progress bar at bottom ===
    {
        float barY = -0.42;
        float barH = 0.025;
        if (p.y > barY - barH * 0.5 && p.y < barY + barH * 0.5) {
            float barT = (p.x + 0.9) / 1.8;
            if (barT < cycT) {
                col = mix(col, bm_pal(0.3), 0.8);
                // Bright edge at progress front
                float edgeD = abs(barT - cycT);
                if (edgeD < 0.02) {
                    col += vec3(1.0, 0.95, 0.70) * (1.0 - edgeD / 0.02) * 0.9;
                }
            } else {
                col = mix(col, vec3(0.05, 0.07, 0.12), 0.6);
            }
        }
    }

    // Commit flash: briefly brighten whole frame when cycle rolls over
    if (cycT < 0.03) {
        float flashInt = 1.0 - cycT / 0.03;
        col += bm_pal(0.5) * flashInt * 0.3;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
