// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Truchet tiles — randomly rotated quarter-arc tiles forming connected neon paths

const float TILE_SIZE = 0.14;
const float LINE_WIDTH = 0.028;
const float INTENSITY  = 0.55;

vec3 tr_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);
    vec3 b = vec3(0.90, 0.30, 0.70);
    vec3 c = vec3(0.55, 0.30, 0.98);
    vec3 d = vec3(0.96, 0.70, 0.25);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float tr_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5);
}

// Distance to a quarter circle at position c with radius r
float sdQuarterArc(vec2 p, vec2 c, float r) {
    return abs(length(p - c) - r);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Tile coordinates (+ time-based drift)
    vec2 tp = p / TILE_SIZE + vec2(x_Time * 0.08, x_Time * 0.05);
    vec2 tileId = floor(tp);
    vec2 tileF = fract(tp);  // [0,1] in tile

    // Choose tile orientation based on hash
    float h = tr_hash(tileId);
    bool rotate = h > 0.5;

    // Truchet tile: two quarter arcs connecting opposite corners
    // Orientation A: TL↔BR (centers at (0,0) and (1,1), radius 0.5)
    // Orientation B: TR↔BL (centers at (1,0) and (0,1), radius 0.5)
    float arcDist;
    if (!rotate) {
        float d1 = sdQuarterArc(tileF, vec2(0.0, 0.0), 0.5);
        float d2 = sdQuarterArc(tileF, vec2(1.0, 1.0), 0.5);
        arcDist = min(d1, d2);
    } else {
        float d1 = sdQuarterArc(tileF, vec2(1.0, 0.0), 0.5);
        float d2 = sdQuarterArc(tileF, vec2(0.0, 1.0), 0.5);
        arcDist = min(d1, d2);
    }

    // Bright line on arc
    float core = smoothstep(LINE_WIDTH * 0.5, 0.0, arcDist);
    // Outer glow
    float glow = exp(-arcDist * arcDist * 300.0) * 0.4;

    // Color per tile — hash-based offset
    vec3 lineCol = tr_pal(fract(h + x_Time * 0.04));

    // Traveling pulse on arc — animate a bright spot along the line
    float pulse = 0.0;
    if (arcDist < LINE_WIDTH) {
        // Parameterize along arc — use angle from one center
        float pulseT;
        vec2 localP = tileF;
        if (!rotate) pulseT = atan(localP.y, localP.x) / 1.5708;  // 0..1 over quarter arc
        else         pulseT = atan(localP.y, 1.0 - localP.x) / 1.5708;
        float pulsePhase = fract(x_Time * 0.5 + h * 10.0);
        float pd = abs(pulseT - pulsePhase);
        pulse = exp(-pd * pd * 80.0);
    }

    vec3 col = lineCol * (core * 1.2 + glow);
    col += vec3(1.0, 0.95, 0.9) * pulse * 0.6;

    // Subtle tile corner indicator dot
    float cornerD = min(length(tileF), min(length(tileF - vec2(1.0, 0.0)),
                        min(length(tileF - vec2(0.0, 1.0)), length(tileF - vec2(1.0, 1.0)))));
    float cornerDot = exp(-cornerD * cornerD * 6000.0);
    col += tr_pal(fract(h * 1.7 + x_Time * 0.06)) * cornerDot * 0.4;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.8, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
