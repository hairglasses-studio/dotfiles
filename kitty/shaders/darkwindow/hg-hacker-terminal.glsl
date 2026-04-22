// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Hacker terminal — fast scrolling code with glyph corruption, bright cursor, status bars

const float LINE_HEIGHT  = 0.022;
const float CHAR_WIDTH   = 0.013;
const float SCROLL_SPEED = 0.25;
const float INTENSITY    = 0.55;

const vec3 GREEN_HI = vec3(0.25, 0.95, 0.55);
const vec3 GREEN_LO = vec3(0.08, 0.40, 0.20);
const vec3 AMBER    = vec3(1.00, 0.70, 0.20);
const vec3 RED      = vec3(1.00, 0.25, 0.25);

float ht_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// Render a glyph cell — binary pattern based on seed
float htGlyph(vec2 uvC, float seed) {
    vec2 pix = floor(uvC * vec2(5.0, 7.0));
    if (pix.x < 0.0 || pix.x > 4.0 || pix.y < 0.0 || pix.y > 6.0) return 0.0;
    float h = ht_hash(pix + vec2(seed, 0.0));
    return step(0.55, h);
}

// Line type based on hash — controls color, density, and format
vec3 lineColor(float lineSeed) {
    if (lineSeed > 0.95) return RED;         // warning
    if (lineSeed > 0.85) return AMBER;        // status
    if (lineSeed > 0.50) return GREEN_HI;     // active code
    return GREEN_LO;                           // old/comment
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    // Scroll uv.y
    float scrollY = uv.y + x_Time * SCROLL_SPEED;
    float lineIdx = floor(scrollY / LINE_HEIGHT);
    float inLineY = fract(scrollY / LINE_HEIGHT);

    float charIdx = floor(uv.x / CHAR_WIDTH);
    float inCharX = fract(uv.x / CHAR_WIDTH);

    // Determine line density (how much of line is filled with characters)
    float lineSeed = ht_hash(vec2(lineIdx, 0.0));
    float density = 0.3 + lineSeed * 0.6;

    // Per-char hash: is this position filled?
    float charSeed = ht_hash(vec2(charIdx, lineIdx));
    float hasChar = step(1.0 - density, charSeed);

    // Glyph pattern (refreshes slowly for realism)
    float gSeed = charSeed + floor(x_Time * 0.5) * 0.1;
    float glyph = htGlyph(vec2(inCharX, inLineY), gSeed);

    // Occasional glyph corruption (intense flicker)
    float corruption = 0.0;
    if (ht_hash(vec2(lineIdx, floor(x_Time * 4.0))) > 0.97) {
        corruption = step(0.5, fract(ht_hash(vec2(charIdx * 0.3, x_Time * 10.0)) * 5.0));
    }
    glyph = mix(glyph, 1.0 - glyph, corruption);

    vec3 col = vec3(0.0);
    vec3 lc = lineColor(lineSeed);
    if (hasChar > 0.5 && glyph > 0.5) {
        col = lc;
    }

    // Cursor — blinking block at a specific line
    float cursorLine = floor(fract(x_Time * 0.1) * 30.0);  // cursor wanders
    float cursorCol = floor(ht_hash(vec2(cursorLine, 13.7)) * 40.0);
    float cursorBlink = step(0.5, fract(x_Time * 2.0));
    if (abs(lineIdx - cursorLine) < 0.5 && abs(charIdx - cursorCol) < 0.5 && cursorBlink > 0.5) {
        col = GREEN_HI * 1.5;
    }

    // Top status bar — persistent, inverted colors
    if (uv.y > 0.97) {
        float barChar = floor(uv.x / CHAR_WIDTH);
        float barSeed = ht_hash(vec2(barChar, floor(x_Time * 1.5)));
        float bg = barSeed > 0.5 ? 1.0 : 0.3;
        col = mix(vec3(0.02, 0.15, 0.08), GREEN_HI * bg, 0.9);
        float statGlyph = htGlyph(vec2(fract(uv.x / CHAR_WIDTH), fract(uv.y * 40.0)), barSeed);
        if (statGlyph > 0.5) col = vec3(0.05);
    }

    // Bottom prompt bar — fixed "shell prompt" line
    if (uv.y < 0.02) {
        col = AMBER * 0.9;
    }

    // Phosphor glow
    float phosphor = 0.0;
    if (hasChar > 0.5 && glyph > 0.5) phosphor = 1.0;
    // Subtle halo from neighbors via sin-modulated field
    float halo = 0.08 * sin(uv.x * 400.0) * sin(uv.y * 600.0);
    col += lc * abs(halo) * 0.1 * hasChar;

    // CRT scanlines
    float scan = 0.88 + 0.12 * sin(x_PixelPos.y * 2.0);
    col *= scan;

    // Flicker
    col *= 0.95 + 0.05 * ht_hash(vec2(x_PixelPos.y * 0.3, floor(x_Time * 50.0)));

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.55);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
