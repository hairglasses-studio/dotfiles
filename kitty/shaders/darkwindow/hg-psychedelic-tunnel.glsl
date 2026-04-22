// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Psychedelic tunnel — twisted rainbow rings with chromatic shift + audio-like pulsation

const float TUNNEL_TWIST = 0.8;    // radians of twist per unit depth
const float PULSE_SPEED  = 1.2;
const float RING_COUNT   = 18.0;
const float INTENSITY    = 0.55;

vec3 pt_pal(float t) {
    // Wider rainbow sweep for the full psychedelic range
    vec3 a = vec3(1.00, 0.20, 0.50); // pink
    vec3 b = vec3(1.00, 0.70, 0.15); // orange
    vec3 c = vec3(0.70, 0.95, 0.25); // lime
    vec3 d = vec3(0.15, 0.90, 0.80); // teal
    vec3 e = vec3(0.40, 0.35, 0.95); // indigo
    vec3 f = vec3(0.95, 0.30, 0.85); // magenta
    float s = mod(t * 6.0, 6.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else if (s < 5.0) return mix(e, f, s - 4.0);
    else              return mix(f, a, s - 5.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Polar coords
    float r = length(p);
    float a = atan(p.y, p.x);

    // Depth (1/r perspective) — smaller r means deeper
    float depth = 1.0 / max(r, 0.05);

    // Twist: angle shifts with depth (kaleidoscopic rotation)
    float twistedAngle = a + depth * TUNNEL_TWIST + x_Time * 0.3;

    // Pulsating scale — "audio-like" beat
    float pulse = 0.9 + 0.15 * sin(x_Time * PULSE_SPEED)
                      + 0.1 * sin(x_Time * PULSE_SPEED * 2.3);

    // Concentric rings — "moving" inward
    float ringCoord = depth * pulse - x_Time * 1.1;
    float ringPhase = fract(ringCoord);
    float ringEdge = smoothstep(0.0, 0.05, ringPhase) * smoothstep(0.3, 0.25, ringPhase);

    // Angular stripes (radial spokes that rotate)
    float spokeFreq = 12.0;
    float spokes = 0.5 + 0.5 * cos(twistedAngle * spokeFreq);

    // Color drifts through rainbow with radius, depth, and time
    float hue = fract(depth * 0.08 + twistedAngle * 0.02 + x_Time * 0.1);
    vec3 base = pt_pal(hue);

    // Chromatic aberration: sample palette at 3 offsets
    vec3 chrom;
    chrom.r = pt_pal(fract(hue + 0.01)).r;
    chrom.g = pt_pal(hue).g;
    chrom.b = pt_pal(fract(hue - 0.01)).b;

    // Combine ring + spoke patterns
    float pattern = ringEdge * (0.7 + spokes * 0.3);

    vec3 col = chrom * pattern;

    // Extra brightness boost near center (tunnel mouth)
    float mouth = exp(-depth * 0.2) * 0.3;
    col += pt_pal(fract(x_Time * 0.07)) * mouth;

    // Vignette — fade edges
    float vignette = smoothstep(0.0, 0.1, r) * smoothstep(1.2, 0.5, r);
    col *= vignette;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
