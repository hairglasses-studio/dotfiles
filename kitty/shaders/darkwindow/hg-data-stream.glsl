// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Vertical data stream — glyph columns with chromatic trails and packet highlights

const float COL_WIDTH    = 14.0;    // glyphs across screen
const float GLYPH_HEIGHT = 28.0;    // glyphs down screen
const float FALL_SPEED   = 0.45;
const int   COLUMNS      = 80;
const float INTENSITY    = 0.55;

vec3 ds_pal(float t) {
    vec3 a = vec3(0.25, 0.98, 0.60);  // mint — classic matrix
    vec3 b = vec3(0.10, 0.82, 0.92);  // cyan
    vec3 c = vec3(0.55, 0.30, 0.98);  // violet
    vec3 d = vec3(0.90, 0.20, 0.55);  // magenta
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float ds_hash(float n) { return fract(sin(n * 127.1) * 43758.5); }
float ds_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// A single glyph — simplified as a binary-pattern mask
float glyphMask(vec2 f, float seed) {
    // 5x5 binary glyph from hash
    vec2 g = floor(f * 5.0);
    if (g.x < 0.0 || g.x > 4.0 || g.y < 0.0 || g.y > 4.0) return 0.0;
    float bit = ds_hash2(g + vec2(seed, 0.0));
    return step(0.5, bit);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    // Scale screen to a column grid
    float aspect = x_WindowSize.x / x_WindowSize.y;
    vec2 cellGrid = vec2(COL_WIDTH * aspect, GLYPH_HEIGHT);
    vec2 cell = floor(uv * cellGrid);
    vec2 cellF = fract(uv * cellGrid);

    vec3 col = vec3(0.0);

    // Each column (floor(uv.x * COL_WIDTH*aspect)) is an independent stream
    float colIdx = cell.x;
    float colSeed = ds_hash(colIdx * 3.71);
    float colSpeed = FALL_SPEED * (0.6 + colSeed * 0.8);
    // Each column has a current "head" y-position
    float headY = fract(colIdx * 0.137 + x_Time * colSpeed) * GLYPH_HEIGHT;
    // Local distance from this fragment's cell to the head (in rows)
    float distFromHead = headY - cell.y;
    if (distFromHead < 0.0) distFromHead += GLYPH_HEIGHT;

    // Trail length varies per column
    float trailLen = 10.0 + colSeed * 14.0;
    float trail = smoothstep(trailLen, 0.0, distFromHead);

    // The head is bright white; the trail fades to column color
    float headMask = smoothstep(1.0, 0.0, distFromHead);
    vec3 colColor = ds_pal(fract(colSeed + x_Time * 0.03));

    // Glyph pattern — changes every few frames at current cell
    float glyphSeed = colIdx + cell.y + floor(x_Time * 6.0 + colSeed * 10.0) * 0.1;
    float gm = glyphMask(cellF, glyphSeed);

    // Compose: head is bright white, trail is column color
    vec3 g = mix(colColor, vec3(0.9, 1.0, 0.9), headMask) * trail * gm;

    // Glow from head — stable halo
    float headGlow = exp(-distFromHead * 0.6) * 0.3 * gm;
    g += colColor * headGlow;

    // Occasional bright "packet" — much rarer bolt that illuminates whole column
    float packetPhase = fract(colIdx * 2.13 + floor(x_Time * 0.3));
    float packetLive = 1.0 - smoothstep(0.0, 0.1, packetPhase);
    float packetY = packetLive * GLYPH_HEIGHT * 2.0;
    float packetDist = abs(packetY - cell.y);
    float packet = exp(-packetDist * packetDist * 0.02) * packetLive;
    g += vec3(1.0, 0.9, 1.0) * packet * gm * 0.5;

    col = g;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
