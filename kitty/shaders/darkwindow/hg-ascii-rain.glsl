// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — ASCII rain — scrolling columns of glyphs with head-cursor and color-cycle trails

const float COL_COUNT = 50.0;
const float ROW_COUNT = 40.0;
const float INTENSITY = 0.55;

vec3 ar_pal(float t) {
    vec3 mint = vec3(0.20, 0.95, 0.60);
    vec3 cyan = vec3(0.10, 0.82, 0.92);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 mag  = vec3(0.90, 0.25, 0.60);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(mint, cyan, s);
    else if (s < 2.0) return mix(cyan, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else              return mix(mag, mint, s - 3.0);
}

float ar_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// Stylized ASCII glyph from a 5x7 binary pattern per hash seed
// Approximates dense random characters
float arGlyph(vec2 uv, float seed) {
    vec2 pix = floor(uv * vec2(5.0, 7.0));
    if (pix.x < 0.0 || pix.x > 4.0 || pix.y < 0.0 || pix.y > 6.0) return 0.0;
    float h = ar_hash(pix + vec2(seed, 0.0));
    // Preserve some structural bias: central pixels more likely lit
    float centerBias = (pix.x == 2.0 || pix.y == 3.0) ? 0.1 : 0.0;
    return step(0.55 - centerBias, h);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    float aspect = x_WindowSize.x / x_WindowSize.y;
    vec2 gridCoord = uv * vec2(COL_COUNT * aspect, ROW_COUNT);
    vec2 cell = floor(gridCoord);
    vec2 cellF = fract(gridCoord);

    float colIdx = cell.x;
    float colSeed = ar_hash(vec2(colIdx, 0.0));
    float speed = 0.3 + colSeed * 0.6;
    float headY = fract(colIdx * 0.113 + x_Time * speed) * ROW_COUNT;
    // Distance from head in rows (wraps)
    float dist = headY - cell.y;
    if (dist < 0.0) dist += ROW_COUNT;

    float trailLen = 8.0 + colSeed * 14.0;
    float trail = smoothstep(trailLen, 0.0, dist);
    float headMask = smoothstep(1.2, 0.0, dist);

    // Glyph content: refreshes every few frames per cell
    float glyphSeed = colIdx * 7.3 + cell.y + floor(x_Time * 4.0);
    float glyph = arGlyph(cellF, glyphSeed);

    vec3 col = vec3(0.0);
    if (glyph > 0.5) {
        vec3 cc = ar_pal(fract(colSeed + x_Time * 0.03));
        // Head brightest, white-tinted
        col = mix(cc, vec3(1.0, 1.0, 0.95), headMask) * trail;
    }

    // Soft halo on head (independent of glyph)
    float headGlow = exp(-dist * 0.8) * 0.25;
    col += ar_pal(fract(colSeed + x_Time * 0.03)) * headGlow;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
