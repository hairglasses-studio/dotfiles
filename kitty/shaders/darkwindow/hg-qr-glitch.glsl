// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — QR-code glitch — 2D bitmap pattern with corruption + chromatic bit-flicker + scan lines

const float QR_CELLS = 30.0;
const float INTENSITY = 0.55;

vec3 qr_pal(float t) {
    vec3 a = vec3(0.20, 0.90, 0.55);
    vec3 b = vec3(0.90, 0.25, 0.70);
    vec3 c = vec3(0.55, 0.30, 0.98);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float qr_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // QR grid centered on screen
    vec2 qrSize = vec2(0.7);
    vec2 qrOffset = -qrSize * 0.5;
    vec2 qrUV = (p - qrOffset) / qrSize;
    if (qrUV.x < 0.0 || qrUV.x > 1.0 || qrUV.y < 0.0 || qrUV.y > 1.0) {
        _wShaderOut = terminal;
        return;
    }

    vec2 cellId = floor(qrUV * QR_CELLS);
    vec2 cellF = fract(qrUV * QR_CELLS);

    vec3 col = vec3(0.03, 0.04, 0.05);

    // Cell "on"/"off" based on hash + animated (slow bit flipping)
    float cellHash = qr_hash(cellId + vec2(floor(x_Time * 0.3), 0.0));
    float onMask = step(0.5, cellHash);

    // Three finder patterns — corners (top-left, top-right, bottom-left)
    vec2 finderPositions[3];
    finderPositions[0] = vec2(0.0, 0.0);
    finderPositions[1] = vec2(QR_CELLS - 7.0, 0.0);
    finderPositions[2] = vec2(0.0, QR_CELLS - 7.0);

    bool inFinder = false;
    for (int f = 0; f < 3; f++) {
        vec2 finderDiff = cellId - finderPositions[f];
        if (finderDiff.x >= 0.0 && finderDiff.x < 7.0 && finderDiff.y >= 0.0 && finderDiff.y < 7.0) {
            inFinder = true;
            // Finder ring pattern
            float maxAbs = max(abs(finderDiff.x - 3.0), abs(finderDiff.y - 3.0));
            if (maxAbs <= 1.0 || maxAbs == 3.0) {
                onMask = 1.0;
            } else {
                onMask = 0.0;
            }
        }
    }

    // Chromatic flicker — occasional RGB offset
    float chromHash = qr_hash(vec2(floor(x_Time * 4.0), 0.0));
    vec2 chromOffset = vec2(chromHash - 0.5) * 0.02;

    // Render bit
    if (onMask > 0.5) {
        vec3 cellCol = vec3(0.9, 0.95, 1.0);
        // Occasional glitch: shift cell color
        if (cellHash > 0.97) {
            cellCol = qr_pal(fract(cellHash + x_Time));
        }
        // Cell fill with small inset
        vec2 inset = abs(cellF - 0.5);
        if (max(inset.x, inset.y) < 0.45) {
            col = cellCol;
        }
    }

    // Corruption: rows shift horizontally
    float rowCorrupt = qr_hash(vec2(cellId.y, floor(x_Time * 4.0)));
    if (rowCorrupt > 0.96) {
        col = qr_pal(fract(cellId.x / QR_CELLS + x_Time));
    }

    // Scanline overlay
    col *= 0.88 + 0.12 * sin(qrUV.y * QR_CELLS * 6.28 + x_Time * 5.0);

    // Bright outer border frame
    float borderD = min(min(qrUV.x, 1.0 - qrUV.x), min(qrUV.y, 1.0 - qrUV.y));
    if (borderD < 0.015) {
        col = qr_pal(fract(x_Time * 0.1));
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.55);
    vec3 result = mix(terminal.rgb, col, visibility * 0.85);

    _wShaderOut = vec4(result, 1.0);
}
