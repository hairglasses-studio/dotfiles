// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Net runner — first-person dive through a neon cyberspace wireframe corridor with rushing ground+ceiling grid, side wall verticals, passing data-block obstacles, bright target node at vanishing point, and radial data-stream trails

const int   DATA_BLOCKS = 10;
const int   DATA_STREAMS = 30;
const float INTENSITY = 0.55;

vec3 nr_pal(float t) {
    vec3 cyan   = vec3(0.25, 0.90, 1.00);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 amber  = vec3(1.00, 0.70, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cyan, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, amber, s - 2.0);
    else if (s < 4.0) return mix(amber, mint, s - 3.0);
    else              return mix(mint, cyan, s - 4.0);
}

float nr_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float nr_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.002, 0.004, 0.012);

    // Vanishing point at center
    float r = length(p);
    float ang = atan(p.y, p.x);

    // === Ground grid (below y = 0) perspective ===
    // Inverse perspective: gy (distance below horizon) maps to depth z
    float horizonY = 0.0;
    if (p.y < horizonY) {
        float gy = horizonY - p.y;
        float z = gy / (0.9 - gy);  // asymptotic
        // Horizontal lines scrolling toward viewer
        float lineZ = fract(z * 3.0 - x_Time * 2.0);
        float hLine = smoothstep(0.03, 0.0, abs(lineZ - 0.5) - 0.48);
        // Vertical lines (fanning out from vanishing point)
        float vGrid = abs(p.x) / (gy + 0.25);
        float vLine = smoothstep(0.012, 0.0, abs(fract(vGrid * 5.0) - 0.5) - 0.48);
        // Depth-based fade
        float gFade = smoothstep(horizonY, -0.55, p.y);
        col += vec3(0.30, 0.95, 1.00) * (hLine * 0.6 + vLine * 0.5) * gFade;
    }
    // === Ceiling grid (above y = 0) mirror the ground ===
    else {
        float gy = p.y - horizonY;
        float z = gy / (0.9 - gy);
        float lineZ = fract(z * 3.0 - x_Time * 2.0);
        float hLine = smoothstep(0.03, 0.0, abs(lineZ - 0.5) - 0.48);
        float vGrid = abs(p.x) / (gy + 0.25);
        float vLine = smoothstep(0.012, 0.0, abs(fract(vGrid * 5.0) - 0.5) - 0.48);
        float cFade = smoothstep(horizonY, 0.55, p.y);
        col += vec3(0.95, 0.30, 0.70) * (hLine * 0.5 + vLine * 0.45) * cFade;
    }

    // === Bright target node at vanishing point ===
    float targPulse = 0.7 + 0.3 * sin(x_Time * 3.0);
    col += vec3(1.0, 0.95, 0.80) * exp(-r * r * 1500.0) * targPulse * 1.5;
    col += vec3(1.0, 0.8, 0.35) * exp(-r * r * 30.0) * 0.35;

    // Small cross-hair markers around target
    float crossMark = smoothstep(0.003, 0.0, abs(p.x)) * smoothstep(0.08, 0.06, abs(p.y))
                    + smoothstep(0.003, 0.0, abs(p.y)) * smoothstep(0.08, 0.06, abs(p.x));
    col += vec3(0.35, 1.00, 0.95) * crossMark * 0.7;

    // === Data-block obstacles flying past (diamond outlines in perspective) ===
    for (int i = 0; i < DATA_BLOCKS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float speed = 0.6 + nr_hash(seed) * 0.6;
        float phase = fract((x_Time + nr_hash(seed * 3.1) * 2.0) * speed);
        float z = phase * 2.0;
        // Screen position: place at (offX, offY) in far-world, scaled by 1/z near
        float sc = z / (z + 0.4);
        // Block off-center position (will grow from small to large with z)
        float offX = (nr_hash(seed * 5.1) * 2.0 - 1.0) * 0.45;
        float offY = (nr_hash(seed * 7.3) * 2.0 - 1.0) * 0.35;
        vec2 center = vec2(offX, offY) * (0.5 + sc * 2.0);
        // Diamond SDF
        vec2 localP = p - center;
        float dSize = 0.015 + sc * 0.08;
        float diamond = abs(localP.x) + abs(localP.y);
        float ringMask = smoothstep(0.005, 0.0, abs(diamond - dSize));
        vec3 blockCol = nr_pal(fract(seed * 0.15 + x_Time * 0.05));
        col += blockCol * ringMask * (0.5 + sc * 0.8);
        // Small filled node at center
        float nodeD = length(localP);
        col += blockCol * exp(-nodeD * nodeD * 3000.0) * sc * 0.7;
    }

    // === Radial data streams (fast lines from vanishing point outward) ===
    for (int i = 0; i < DATA_STREAMS; i++) {
        float fi = float(i);
        float seed = fi * 11.1;
        float streamAng = nr_hash(seed) * 6.28;
        // Stream phase (progresses from center outward)
        float sp = fract((x_Time + nr_hash(seed * 3.1)) * (1.5 + nr_hash(seed * 5.1) * 1.5));
        // Current head position along this ray
        float headR = sp * 1.3;
        vec2 rayDir = vec2(cos(streamAng), sin(streamAng));
        vec2 head = rayDir * headR;
        vec2 tail = rayDir * max(headR - 0.06, 0.02);
        vec2 ab = head - tail;
        vec2 ap = p - tail;
        float h = clamp(dot(ap, ab) / dot(ab, ab), 0.0, 1.0);
        float d = length(ap - ab * h);
        float streamMask = exp(-d * d * 150000.0);
        // Fade out at edges
        float edgeFade = smoothstep(1.4, 0.6, headR);
        col += nr_pal(fract(seed * 0.2 + x_Time * 0.1)) * streamMask * edgeFade * 0.85;
    }

    // === Side wall verticals — bright lines at fixed angles suggesting a corridor ===
    for (int w = 0; w < 6; w++) {
        float fw = float(w);
        // Angle for this wall line
        float wallAng = (fw / 6.0) * 6.28;
        vec2 wallDir = vec2(cos(wallAng), sin(wallAng));
        // Distance from current point to the ray through origin in wallDir direction
        float cross2D = abs(p.x * wallDir.y - p.y * wallDir.x);
        float along2 = dot(p, wallDir);
        if (along2 > 0.0 && along2 < 1.2) {
            float wallMask = exp(-cross2D * cross2D * 3000.0) * (1.0 - along2 / 1.2);
            col += nr_pal(fract(fw * 0.15)) * wallMask * 0.3;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
