// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Sci-fi holographic HUD — concentric arcs, crosshair, tick marks, scanline flicker

const float ARC_RADIUS  = 0.35;
const int   ARC_SEGMENTS = 6;       // arcs divided into segments
const int   TICK_COUNT   = 32;
const float INTENSITY    = 0.48;

const vec3 HOLO_CYAN   = vec3(0.15, 0.92, 0.98);
const vec3 HOLO_MAGENTA = vec3(0.90, 0.25, 0.70);

float hu_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5);
}

// Distance to a circle of radius r
float sdCircle(vec2 p, float r) {
    return abs(length(p) - r);
}

// Distance to an arc spanning angle range [a1, a2] with radius r
float sdArc(vec2 p, float r, float a1, float a2) {
    float a = atan(p.y, p.x);
    if (a < 0.0) a += 6.28318;
    if (a1 < 0.0) a1 += 6.28318;
    if (a2 < 0.0) a2 += 6.28318;
    // Inside angle range → distance to circle at radius r
    // Outside → distance to nearest endpoint
    if (a1 < a2) {
        if (a >= a1 && a <= a2) return abs(length(p) - r);
    } else {
        if (a >= a1 || a <= a2) return abs(length(p) - r);
    }
    vec2 pa = vec2(cos(a1), sin(a1)) * r;
    vec2 pb = vec2(cos(a2), sin(a2)) * r;
    return min(length(p - pa), length(p - pb));
}

// Digital-looking glyph pattern (5x3 binary grid)
float hu_glyph(vec2 p, float seed) {
    vec2 g = floor(p * vec2(3.0, 5.0));
    if (g.x < 0.0 || g.x > 2.0 || g.y < 0.0 || g.y > 4.0) return 0.0;
    float h = hu_hash(g + vec2(seed, 0.0));
    return step(0.5, h);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    // Rotating main arc — three concentric radii, different rotation speeds
    for (int r = 0; r < 3; r++) {
        float fr = float(r);
        float radius = ARC_RADIUS * (0.6 + fr * 0.2);
        float rotSpeed = (fr - 1.0) * 0.2;
        // Per-ring rotation
        float phi = x_Time * rotSpeed;
        vec2 pr = mat2(cos(phi), -sin(phi), sin(phi), cos(phi)) * p;

        // Multiple segments per ring with gaps
        float a = atan(pr.y, pr.x);
        float segAngle = 6.28318 / float(ARC_SEGMENTS);
        float inSeg = fract(a / segAngle) * segAngle;
        float gapMask = step(segAngle * 0.2, inSeg);
        float ringDist = abs(length(pr) - radius);
        float ringMask = smoothstep(0.003, 0.0, ringDist) * gapMask;

        col += HOLO_CYAN * ringMask * 0.9;
    }

    // Tick marks around outer ring
    float tickRadius = ARC_RADIUS * 1.1;
    float a0 = atan(p.y, p.x);
    if (a0 < 0.0) a0 += 6.28318;
    float tickAngle = 6.28318 / float(TICK_COUNT);
    float tickPhase = mod(a0, tickAngle);
    float inTick = step(tickPhase, tickAngle * 0.08);   // tiny tick width
    float radialBand = smoothstep(0.01, 0.005, abs(length(p) - tickRadius));
    float majorTick = step(mod(a0, 6.28318 / 8.0), tickAngle * 0.05);  // 8 major ticks
    float tickIntensity = radialBand * (inTick * 0.7 + majorTick * 0.3);
    col += HOLO_CYAN * tickIntensity;

    // Crosshair in center — vertical + horizontal lines (with gap at center)
    float crossW = 0.002;
    float centerGap = 0.03;
    float hLine = smoothstep(crossW, 0.0, abs(p.y)) * step(centerGap, abs(p.x)) * step(abs(p.x), ARC_RADIUS * 0.55);
    float vLine = smoothstep(crossW, 0.0, abs(p.x)) * step(centerGap, abs(p.y)) * step(abs(p.y), ARC_RADIUS * 0.55);
    col += HOLO_CYAN * (hLine + vLine) * 0.7;

    // Target reticule dot at a wandering position
    vec2 target = vec2(0.18 * sin(x_Time * 0.3), 0.12 * cos(x_Time * 0.4));
    float targetDist = length(p - target);
    float targetMask = smoothstep(0.015, 0.012, targetDist) - smoothstep(0.012, 0.009, targetDist);
    col += HOLO_MAGENTA * targetMask * 1.2;
    float targetGlow = exp(-targetDist * targetDist * 400.0) * 0.4;
    col += HOLO_MAGENTA * targetGlow;

    // Data glyphs on the right edge — scrolling vertical column
    vec2 glyphBase = vec2(ARC_RADIUS * 1.25, p.y);
    if (p.x > glyphBase.x - 0.04 && p.x < glyphBase.x + 0.04) {
        vec2 gp = vec2(
            (p.x - glyphBase.x) * 12.0 + 0.5,
            fract(p.y * 15.0 - x_Time * 0.4)
        );
        float glyph = hu_glyph(gp, floor(p.y * 15.0 - x_Time * 0.4));
        col += HOLO_CYAN * glyph * 0.4;
    }

    // Scanline flicker — rolling horizontal noise band
    float scanY = fract(uv.y + x_Time * 0.15);
    float scanBand = smoothstep(0.0, 0.02, scanY) * smoothstep(0.05, 0.03, scanY);
    col *= 1.0 + scanBand * 0.5;

    // Horizontal scanlines (subtle, constant)
    float hScan = 0.5 + 0.5 * sin(x_PixelPos.y * 1.5);
    col *= 0.85 + hScan * 0.3;

    // Flicker noise
    float flicker = hu_hash(vec2(x_PixelPos.y * 0.1, floor(x_Time * 30.0)));
    col *= 0.92 + flicker * 0.12;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col) * 0.8, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
