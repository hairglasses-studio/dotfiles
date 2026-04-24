// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Arcade cabinet row — 4 CRT screens each running a different attract-mode: DVD-logo ball, tetris falling blocks, starfield shooter, pong demo. Dark arcade room with neon floor glow.

const int   CAB_COUNT = 4;
const float CAB_W = 0.40;       // cabinet logical width in normalized coords
const float INTENSITY = 0.55;

vec3 arc_pal(float t) {
    vec3 cyan   = vec3(0.20, 0.85, 1.00);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 amber  = vec3(1.00, 0.75, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cyan, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, amber, s - 2.0);
    else if (s < 4.0) return mix(amber, mint, s - 3.0);
    else              return mix(mint, cyan, s - 4.0);
}

float arc_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float arc_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

// DVD-logo bouncing ball attract mode.
// screenUV is [-1,1] within the CRT rectangle.
vec3 attract_bouncing_ball(vec2 screenUV, float t) {
    // Ball position reflects off walls; using triangle wave pattern
    float px = abs(fract(t * 0.12 + 0.25) - 0.5) * 2.0 - 0.7; // -0.7..0.3
    float py = abs(fract(t * 0.09 + 0.1) - 0.5) * 2.0 - 0.6;
    vec2 ballPos = vec2(px * 0.8, py * 0.7);
    float bd = length(screenUV - ballPos);
    vec3 ballCol = arc_pal(fract(t * 0.2));
    vec3 c = vec3(0.0);
    c += ballCol * exp(-bd * bd * 400.0) * 1.5;
    c += ballCol * exp(-bd * bd * 30.0) * 0.3;
    return c;
}

// Tetris falling blocks attract mode.
vec3 attract_tetris(vec2 screenUV, float t) {
    vec3 c = vec3(0.0);
    // Grid cells
    vec2 grid = floor(screenUV * vec2(5.0, 8.0) + vec2(5.0, 8.0)) / vec2(5.0, 8.0);
    // Falling pieces: for each column, piece spawns at top and falls
    for (int col = 0; col < 10; col++) {
        float fc = float(col);
        float seed = fc * 3.7;
        float spawnT = fract(t * 0.3 + arc_hash(seed));
        float fallY = 0.9 - spawnT * 1.8;   // from 0.9 down to -0.9
        float xPos = (arc_hash(seed * 5.1) * 2.0 - 1.0) * 0.8;
        float pieceW = 0.18;
        float pieceH = 0.12;
        vec2 d = abs(screenUV - vec2(xPos, fallY));
        if (d.x < pieceW * 0.5 && d.y < pieceH * 0.5) {
            // Inside piece: check 2x2 cell fill pattern
            float cellHash = arc_hash(seed * 7.3 + floor(d.x * 10.0) + floor(d.y * 10.0));
            if (cellHash > 0.3) {
                c += arc_pal(fract(seed * 0.5 + t * 0.1)) * 0.9;
            }
        }
    }
    // Settled blocks at the bottom (pile up)
    float pileTop = -0.3 + 0.2 * sin(screenUV.x * 8.0 + t * 0.5);
    if (screenUV.y < pileTop) {
        float pileHash = arc_hash2(floor(screenUV * 15.0));
        if (pileHash > 0.5) {
            c += arc_pal(fract(pileHash + t * 0.05)) * 0.6;
        }
    }
    return c;
}

// Shmup starfield + player ship attract
vec3 attract_shmup(vec2 screenUV, float t) {
    vec3 c = vec3(0.0);
    // Scrolling starfield
    for (int i = 0; i < 35; i++) {
        float fi = float(i);
        float seed = fi * 3.7;
        float speed = 0.3 + arc_hash(seed) * 0.8;
        float sx = arc_hash(seed * 5.1) * 2.0 - 1.0;
        float sy = fract(arc_hash(seed * 7.3) + t * speed) * 2.0 - 1.0;
        vec2 stP = vec2(sx, sy);
        float d = length(screenUV - stP);
        c += vec3(0.85, 0.9, 1.0) * exp(-d * d * 3000.0) * (0.4 + speed * 0.4);
    }
    // Player ship (triangle-ish at bottom-center)
    vec2 shipPos = vec2(sin(t * 1.5) * 0.5, -0.6);
    vec2 rel = screenUV - shipPos;
    if (abs(rel.x) < 0.08 && rel.y > 0.0 && rel.y < 0.12 && abs(rel.x) < 0.08 - rel.y * 0.6) {
        c += vec3(0.4, 0.9, 1.0) * 1.0;
    }
    // Shot trail
    float shotY = mod(t * 2.0, 1.5) - 0.6;
    if (abs(screenUV.x - shipPos.x) < 0.01 && screenUV.y > -0.5 && screenUV.y < shotY) {
        c += vec3(1.0, 0.9, 0.3) * 1.2;
    }
    return c;
}

// Pong attract
vec3 attract_pong(vec2 screenUV, float t) {
    vec3 c = vec3(0.0);
    // Paddles on left and right
    float paddleHalfH = 0.25;
    float leftY = sin(t * 1.2) * 0.5;
    float rightY = sin(t * 1.1 + 1.2) * 0.5;
    if (screenUV.x < -0.82 && screenUV.x > -0.88
        && abs(screenUV.y - leftY) < paddleHalfH) {
        c += vec3(1.0, 0.95, 0.85);
    }
    if (screenUV.x > 0.82 && screenUV.x < 0.88
        && abs(screenUV.y - rightY) < paddleHalfH) {
        c += vec3(1.0, 0.95, 0.85);
    }
    // Ball bouncing
    float bx = abs(fract(t * 0.4) - 0.5) * 2.0 - 1.0; // -1..1 back and forth
    bx *= 0.82;
    float by = abs(fract(t * 0.3 + 0.3) - 0.5) * 2.0 - 0.6;
    vec2 ballPos = vec2(bx, by);
    float bd = length(screenUV - ballPos);
    c += vec3(1.0, 1.0, 0.95) * exp(-bd * bd * 1500.0) * 1.2;
    // Center dashed line
    if (abs(screenUV.x) < 0.008 && fract(screenUV.y * 10.0) < 0.6) {
        c += vec3(0.75, 0.75, 0.75) * 0.5;
    }
    return c;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.008, 0.010, 0.025);

    // === Room environment ===
    // Ceiling gradient (dark at top)
    float ceilFade = smoothstep(0.5, 0.1, p.y);
    col = mix(vec3(0.005, 0.01, 0.02), col, ceilFade);

    // Floor neon glow
    if (p.y < -0.35) {
        float depth = (-0.35 - p.y) / 0.4;
        depth = clamp(depth, 0.0, 1.0);
        vec3 floorCol = mix(vec3(0.08, 0.04, 0.12), vec3(0.02, 0.01, 0.04), depth);
        col = mix(col, floorCol, 0.85);
        // Neon strip glow
        float strip = exp(-pow(p.y + 0.35, 2.0) * 400.0);
        col += arc_pal(fract(x_Time * 0.05 + p.x * 0.3)) * strip * 0.6;
    }

    // === Cabinets (4 arranged horizontally) ===
    // Map p.x → cabinet index. Cabinets span x in [-1.0, 1.0] divided into 4.
    float pxMapped = (p.x + 1.0) * 0.5 * float(CAB_COUNT);
    int cabIdx = int(pxMapped);
    cabIdx = clamp(cabIdx, 0, CAB_COUNT - 1);
    float localX = fract(pxMapped) * 2.0 - 1.0;  // -1..1 within cabinet

    // Cabinet body: rectangles occupying roughly y in [-0.35, 0.50]
    float cabTop = 0.50;
    float cabBot = -0.35;

    if (p.y > cabBot && p.y < cabTop) {
        // Dark cabinet body
        vec3 cabSil = vec3(0.04, 0.04, 0.08);
        col = mix(col, cabSil, 0.9);

        // Cabinet side bezels (dark vertical strips at edges of each cabinet)
        if (abs(localX) > 0.90) {
            col = mix(col, vec3(0.02, 0.02, 0.04), 0.8);
        }

        // CRT screen area: centered, roughly y in [0.02, 0.42], localX in [-0.75, 0.75]
        float scrTop = 0.42, scrBot = 0.02, scrSide = 0.75;
        if (p.y > scrBot && p.y < scrTop && abs(localX) < scrSide) {
            // Map to screen-local UV in [-1, 1]
            vec2 scrUV = vec2(localX / scrSide,
                             (p.y - (scrTop + scrBot) * 0.5) / ((scrTop - scrBot) * 0.5));

            vec3 scr = vec3(0.0);
            if (cabIdx == 0)      scr = attract_bouncing_ball(scrUV, x_Time);
            else if (cabIdx == 1) scr = attract_tetris(scrUV, x_Time);
            else if (cabIdx == 2) scr = attract_shmup(scrUV, x_Time);
            else                  scr = attract_pong(scrUV, x_Time);

            // CRT scanlines
            float scanline = 0.85 + 0.15 * sin(p.y * 200.0);
            scr *= scanline;

            // Vignette
            float vig = 1.0 - dot(scrUV, scrUV) * 0.4;

            col = scr * vig * 1.2;
            // Bezel ring
            float bezelDist = min(min(p.y - scrBot, scrTop - p.y),
                                   min(localX + scrSide, scrSide - localX));
            if (bezelDist < 0.015) {
                col = mix(col, vec3(0.015, 0.015, 0.03), 0.9);
            }
        }

        // Marquee top strip — glows with title
        if (p.y > 0.42 && p.y < 0.50) {
            float marqT = fract(x_Time * 0.1 + float(cabIdx) * 0.1);
            col += arc_pal(marqT) * 0.4 * smoothstep(0.0, 0.04, p.y - 0.42);
            // Subtle text-like pattern (just stripes)
            float stripes = abs(sin(localX * 40.0 + x_Time * 2.0)) * 0.3;
            col += arc_pal(marqT) * stripes * 0.2;
        }

        // Underbezel neon glow (bottom of cabinet)
        if (p.y > -0.35 && p.y < -0.25) {
            col += arc_pal(fract(float(cabIdx) * 0.2 + x_Time * 0.08)) * 0.35
                   * smoothstep(-0.35, -0.30, p.y);
        }
    }

    // CRT glow halo escaping past cabinet boundary (bloom)
    if (p.y > 0.02 && p.y < 0.42 && abs(localX) > 0.75 && abs(localX) < 0.92) {
        float haloStrength = 1.0 - (abs(localX) - 0.75) / 0.17;
        col += arc_pal(fract(float(cabIdx) * 0.3 + x_Time * 0.05)) * haloStrength * 0.25;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
