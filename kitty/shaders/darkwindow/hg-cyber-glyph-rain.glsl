// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Cyber glyph rain — 3D pseudo-stereoscopic falling neon glyphs with depth layers

const float COL_COUNT = 45.0;
const float ROW_COUNT = 32.0;
const int   LAYERS    = 3;
const float INTENSITY = 0.55;

vec3 cg_pal(float t) {
    vec3 mint = vec3(0.25, 0.95, 0.55);
    vec3 cyan = vec3(0.10, 0.82, 0.92);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 mag  = vec3(0.95, 0.30, 0.75);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(mint, cyan, s);
    else if (s < 2.0) return mix(cyan, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else              return mix(mag, mint, s - 3.0);
}

float cg_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// Stylized glyph
float cgGlyph(vec2 uvC, float seed) {
    vec2 pix = floor(uvC * 5.0);
    if (pix.x < 0.0 || pix.x > 4.0 || pix.y < 0.0 || pix.y > 4.0) return 0.0;
    return step(0.5, cg_hash(pix + vec2(seed, 0.0)));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    float aspect = x_WindowSize.x / x_WindowSize.y;
    vec3 col = vec3(0.01, 0.01, 0.02);

    // Multiple depth layers — each layer offset and scaled differently
    for (int L = 0; L < LAYERS; L++) {
        float fL = float(L);
        float depth = 1.0 - fL * 0.3;   // closer for later L
        float scale = 1.0 + fL * 0.3;

        // Layer offset (parallax) — based on center
        vec2 layerUV = (uv - 0.5) * scale + 0.5;
        // Wrap
        layerUV.x = mod(layerUV.x, 1.0);
        layerUV.y = mod(layerUV.y, 1.0);

        vec2 gridCoord = layerUV * vec2(COL_COUNT * aspect, ROW_COUNT);
        vec2 cell = floor(gridCoord);
        vec2 cellF = fract(gridCoord);

        float colIdx = cell.x;
        float colSeed = cg_hash(vec2(colIdx, fL));
        float speed = 0.3 + colSeed * 0.7;
        float headY = fract(colIdx * 0.113 + x_Time * speed + fL * 0.3) * ROW_COUNT;
        float dist = headY - cell.y;
        if (dist < 0.0) dist += ROW_COUNT;

        float trailLen = 8.0 + colSeed * 15.0;
        float trail = smoothstep(trailLen, 0.0, dist);
        float headMask = smoothstep(1.2, 0.0, dist);

        float glyphSeed = colIdx * 7.3 + cell.y + floor(x_Time * 4.0);
        float glyph = cgGlyph(cellF, glyphSeed);

        vec3 layerCol = cg_pal(fract(colSeed * 0.3 + fL * 0.2 + x_Time * 0.03));

        // Apply layer brightness
        float layerAlpha = depth;  // closer = brighter

        if (glyph > 0.5) {
            col += layerCol * trail * headMask * layerAlpha * 0.4;
            col += mix(layerCol, vec3(1.0), headMask * 0.5) * trail * headMask * layerAlpha * 0.7;
        }
    }

    // Vignette
    col *= 1.0 - length((uv - 0.5) * vec2(aspect, 1.0)) * 0.1;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
