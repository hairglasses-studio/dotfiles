// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Aggressive glitch-matrix — bit-shift errors, block corruption, RGB tear, scan corrupt

const float INTENSITY = 0.5;

vec3 gm_pal(float t) {
    vec3 green = vec3(0.20, 0.95, 0.55);
    vec3 red   = vec3(1.00, 0.15, 0.25);
    vec3 cyan  = vec3(0.20, 0.85, 0.95);
    vec3 mag   = vec3(0.95, 0.20, 0.75);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(green, red, s);
    else if (s < 2.0) return mix(red, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, mag, s - 2.0);
    else              return mix(mag, green, s - 3.0);
}

float gm_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;

    // Block corruption — divide screen into blocks, some blocks get shifted
    vec2 blockSize = vec2(0.12, 0.04);  // wide horizontal blocks
    vec2 blockId = floor(uv / blockSize);
    float blockSeed = gm_hash(blockId + vec2(floor(x_Time * 6.0), 0.0));

    vec2 shiftUV = uv;
    if (blockSeed > 0.88) {
        // Block corrupt — horizontal shift
        shiftUV.x += (blockSeed - 0.88) * 2.0;
        shiftUV.x = fract(shiftUV.x);
    }

    // Line tear — full-width single-line horizontal shifts
    float lineSeed = gm_hash(vec2(floor(uv.y * 120.0), floor(x_Time * 10.0)));
    if (lineSeed > 0.95) {
        shiftUV.x += (lineSeed - 0.95) * 8.0;
        shiftUV.x = fract(shiftUV.x);
    }

    // Sample terminal with RGB split
    vec3 col;
    float chromaticIntensity = 0.003 * (0.8 + 0.6 * gm_hash(vec2(floor(x_Time * 3.0), 0.0)));
    col.r = x_Texture(shiftUV + vec2(chromaticIntensity, 0.0)).r;
    col.g = x_Texture(shiftUV).g;
    col.b = x_Texture(shiftUV - vec2(chromaticIntensity, 0.0)).b;

    // Bit-shift: replace pixel with random high-saturation color occasionally
    float bitHash = gm_hash(x_PixelPos + vec2(floor(x_Time * 60.0), 0.0));
    if (bitHash > 0.995) {
        col = gm_pal(fract(bitHash * 10.0 + x_Time));
    }

    // Scan corrupt — occasional full horizontal line of corruption
    float scanCorrupt = gm_hash(vec2(floor(uv.y * 80.0), floor(x_Time * 3.0)));
    if (scanCorrupt > 0.97) {
        float corruptPattern = fract(uv.x * 50.0 + x_Time * 10.0);
        col = mix(col, gm_pal(fract(corruptPattern + x_Time)), 0.7);
    }

    // Data overlay — sparse glyph pattern
    vec2 glyphId = floor(uv * vec2(80.0, 40.0));
    float glyphHash = gm_hash(glyphId + vec2(floor(x_Time * 2.0), 0.0));
    if (glyphHash > 0.92) {
        vec2 glyphF = fract(uv * vec2(80.0, 40.0));
        float glyphPattern = step(0.5, gm_hash(floor(glyphF * 5.0) + glyphId));
        col = mix(col, gm_pal(0.0), glyphPattern * 0.3);
    }

    // Frame jitter
    float jitterSeed = gm_hash(vec2(floor(x_Time * 5.0), 0.0));
    if (jitterSeed > 0.9) {
        col += vec3(0.2);
    }

    // Scanlines
    col *= 0.85 + 0.15 * sin(x_PixelPos.y * 2.0);

    _wShaderOut = vec4(col, 1.0);
}
