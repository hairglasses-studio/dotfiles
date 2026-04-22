// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Heavy matrix glitch — cascade of green code with intense distortion, block swaps, code rewrites

const float COL_COUNT = 60.0;
const float ROW_COUNT = 45.0;
const float INTENSITY = 0.55;

vec3 mx_pal(float t) {
    vec3 green = vec3(0.25, 0.98, 0.55);
    vec3 mint  = vec3(0.20, 0.85, 0.45);
    vec3 cyan  = vec3(0.20, 0.80, 0.85);
    vec3 red   = vec3(0.95, 0.20, 0.35);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(green, mint, s);
    else if (s < 2.0) return mix(mint, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, red, s - 2.0);
    else              return mix(red, green, s - 3.0);
}

float mx_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float mxGlyph(vec2 uvC, float seed) {
    vec2 pix = floor(uvC * 5.0);
    if (pix.x < 0.0 || pix.x > 4.0 || pix.y < 0.0 || pix.y > 4.0) return 0.0;
    float h = mx_hash(pix + vec2(seed, 0.0));
    return step(0.5, h);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    // Horizontal block displacement (massive corruption)
    float aspect = x_WindowSize.x / x_WindowSize.y;
    vec2 blockSize = vec2(0.08, 0.02);
    vec2 blockId = floor(uv / blockSize);
    float blockHash = mx_hash(blockId + vec2(floor(x_Time * 8.0), 0.0));

    vec2 shiftUV = uv;
    if (blockHash > 0.8) {
        shiftUV.x += (blockHash - 0.8) * 3.0;
        shiftUV.x = fract(shiftUV.x);
    }

    // Vertical pixel shift (tearing)
    float vhash = mx_hash(vec2(floor(shiftUV.x * 30.0), floor(x_Time * 6.0)));
    if (vhash > 0.92) shiftUV.y += (vhash - 0.92) * 0.5;
    shiftUV.y = fract(shiftUV.y);

    vec2 gridCoord = shiftUV * vec2(COL_COUNT * aspect, ROW_COUNT);
    vec2 cell = floor(gridCoord);
    vec2 cellF = fract(gridCoord);

    float colIdx = cell.x;
    float colSeed = mx_hash(vec2(colIdx, 0.0));
    float speed = 0.6 + colSeed * 0.8;
    float headY = fract(colIdx * 0.113 + x_Time * speed) * ROW_COUNT;
    float dist = headY - cell.y;
    if (dist < 0.0) dist += ROW_COUNT;

    float trailLen = 15.0 + colSeed * 15.0;
    float trail = smoothstep(trailLen, 0.0, dist);
    float headMask = smoothstep(1.2, 0.0, dist);

    // Glyph — cycles rapidly
    float glyphSeed = colIdx * 7.3 + cell.y + floor(x_Time * 15.0);
    float glyph = mxGlyph(cellF, glyphSeed);

    vec3 col = vec3(0.0);
    if (glyph > 0.5) {
        vec3 cc = mx_pal(fract(colSeed * 0.5 + x_Time * 0.08));
        col = mix(cc, vec3(1.0), headMask * 0.8) * trail;
    }

    // Occasional full-row bright flash
    float rowFlashHash = mx_hash(vec2(floor(cell.y), floor(x_Time * 4.0)));
    if (rowFlashHash > 0.95) {
        col += vec3(0.7, 1.0, 0.8) * 0.2;
    }

    // Chromatic split on intense glitches
    float chromaticIntensity = blockHash > 0.9 ? 0.008 : 0.002;
    vec4 termRGB;
    termRGB.r = x_Texture(shiftUV + vec2(chromaticIntensity, 0.0)).r;
    termRGB.g = x_Texture(shiftUV).g;
    termRGB.b = x_Texture(shiftUV - vec2(chromaticIntensity, 0.0)).b;

    vec3 finalCol = col;
    // If no glyph, show terminal through the glitch-shifted UV
    if (glyph < 0.5) {
        // Dim base
        finalCol = termRGB.rgb * 0.6;
    }

    // Random code injection — whole-screen bit-shift at rare intervals
    float injectHash = mx_hash(vec2(floor(x_Time * 3.0), 0.0));
    if (injectHash > 0.95) {
        finalCol += mx_pal(fract(uv.x + x_Time)) * 0.1;
    }

    // Scanlines
    finalCol *= 0.88 + 0.12 * sin(x_PixelPos.y * 2.0);

    _wShaderOut = vec4(finalCol, 1.0);
}
