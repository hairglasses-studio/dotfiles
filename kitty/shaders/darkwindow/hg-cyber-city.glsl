// Shader attribution: hairglasses (original)
// Technique inspired by: Shadertoy ssSSDG Cyberpunk City (z0rg) — SDF skyline with neon window pattern.
// License: MIT
// (Cyberpunk — showcase/heavy) — Neon skyline with lit windows, rain reflections, atmospheric haze

const int   BUILDING_LAYERS = 4;
const float HORIZON = 0.55;
const float INTENSITY = 0.55;

vec3 cc_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);  // cyan window
    vec3 b = vec3(0.90, 0.20, 0.55);  // magenta sign
    vec3 c = vec3(0.55, 0.30, 0.98);  // violet glow
    vec3 d = vec3(0.96, 0.85, 0.45);  // warm gold
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float cc_hash(float x) { return fract(sin(x * 127.1) * 43758.5); }
float cc_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// Layered city silhouette — sum of randomly-placed rectangles
float cityHeight(float x, float layer) {
    float h = 0.0;
    float seed = layer * 17.31;
    // Sample a few building "blocks" around this x
    for (int i = 0; i < 12; i++) {
        float fi = float(i);
        float bx = (fi + 0.5) / 12.0 * 3.0 - 1.5;   // spread across [-1.5, 1.5]
        bx += (cc_hash(fi + seed) - 0.5) * 0.08;
        float bw = 0.06 + 0.06 * cc_hash(fi + seed + 1.0);
        float bh = 0.03 + 0.18 * cc_hash(fi + seed + 2.0);
        // Distance from this x to the building center (in a tiling sense)
        float dx = mod(x - bx + 1.5, 3.0) - 1.5;
        float inside = step(abs(dx), bw);
        h = max(h, inside * bh);
    }
    return h;
}

// Lit-window pattern inside a building
float windowPattern(vec2 p, float layer) {
    // Grid of windows
    vec2 g = floor(p * vec2(20.0, 30.0));
    float lit = cc_hash2(g + vec2(layer, 0.0));
    // Many windows off, some on, flicker a few
    float onThreshold = 0.65 + layer * 0.05;
    float flicker = sin(x_Time * 3.0 + lit * 20.0) * 0.5 + 0.5;
    float onMask = step(onThreshold, lit) + step(0.95, lit) * flicker * 0.5;
    // Cell fill — slightly smaller than cell
    vec2 gf = fract(p * vec2(20.0, 30.0));
    float cell = step(0.15, gf.x) * step(gf.x, 0.85) * step(0.2, gf.y) * step(gf.y, 0.8);
    return onMask * cell;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec3 bg = vec3(0.02, 0.02, 0.06);

    // Sky gradient (purple → dark)
    if (uv.y > HORIZON) {
        float skyT = (uv.y - HORIZON) / (1.0 - HORIZON);
        bg = mix(vec3(0.24, 0.04, 0.32), vec3(0.03, 0.01, 0.10), smoothstep(0.0, 0.85, skyT));
        // Distant atmospheric haze near horizon
        float haze = 1.0 - smoothstep(0.0, 0.2, skyT);
        bg += vec3(0.35, 0.08, 0.30) * haze * 0.35;
    }

    // Layered buildings — parallax: closer layers scroll faster
    float cityMask = 0.0;
    vec3 cityCol = vec3(0.0);
    for (int L = 0; L < BUILDING_LAYERS; L++) {
        float fL = float(L);
        float scrollSpeed = 0.015 * (fL + 1.0);
        float xScroll = uv.x + x_Time * scrollSpeed;
        float layerHoriz = HORIZON - 0.02 - fL * 0.015;
        float h = cityHeight(xScroll * (3.0 + fL), fL);
        float buildingTop = layerHoriz + h;
        if (uv.y < buildingTop && uv.y > HORIZON - 0.3) {
            // We're inside a building of this layer
            float depthShade = 0.15 + fL * 0.10;
            vec3 bldgBase = vec3(0.05, 0.04, 0.10) * depthShade;
            // Windows
            vec2 wp = vec2(xScroll * (3.0 + fL), (buildingTop - uv.y) * 10.0);
            float win = windowPattern(wp, fL);
            vec3 winCol = cc_pal(fract(cc_hash2(floor(wp * vec2(20.0, 30.0)) + fL) + x_Time * 0.02));
            vec3 totalCol = mix(bldgBase, winCol, win * 0.9);
            cityCol = mix(cityCol, totalCol, 1.0 - cityMask);
            cityMask = max(cityMask, 1.0 - fL * 0.15);
        }
    }
    bg = mix(bg, cityCol, cityMask);

    // Ground: wet pavement + sign reflections
    if (uv.y < HORIZON - 0.3) {
        float groundT = (HORIZON - 0.3 - uv.y) / (HORIZON - 0.3);
        // Reflected sign glow — vertical streak at each "sign" x position
        vec3 reflection = vec3(0.0);
        for (int i = 0; i < 4; i++) {
            float fi = float(i);
            float sx = 0.5 + (fi - 1.5) * 0.22 + sin(x_Time * 0.5 + fi * 2.0) * 0.04;
            float streakW = 0.015 + 0.02 * cc_hash(fi);
            float streak = exp(-pow((uv.x - sx) / streakW, 2.0));
            streak *= (1.0 - groundT * 0.8);   // fade toward bottom
            vec3 sigCol = cc_pal(fract(fi * 0.25 + x_Time * 0.1));
            reflection += sigCol * streak * 0.4;
        }
        bg = mix(bg, vec3(0.01, 0.01, 0.03), smoothstep(0.0, 0.3, groundT)) + reflection;
    }

    // Rain — sparse diagonal streaks, animated
    float rainSeed = floor(uv.x * 60.0) + floor(x_Time * 10.0);
    float rainHash = cc_hash(rainSeed);
    if (rainHash > 0.97) {
        float yPhase = fract(uv.y * 8.0 - x_Time * 3.0);
        float streak = smoothstep(0.0, 0.2, yPhase) * smoothstep(1.0, 0.9, yPhase);
        bg += vec3(0.4, 0.45, 0.55) * streak * 0.12;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, bg, visibility);

    _wShaderOut = vec4(result, 1.0);
}
