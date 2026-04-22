// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Traversable wormhole — polar warp, streaming rings, deep core glow

const int   RING_LAYERS  = 12;
const float CORE_RADIUS  = 0.06;
const float TUNNEL_SPEED = 2.0;
const float ROTATION_SPD = 0.3;
const float INTENSITY    = 0.55;

vec3 wh_pal(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98); // violet
    vec3 b = vec3(0.10, 0.82, 0.92); // cyan
    vec3 c = vec3(0.90, 0.30, 0.70); // magenta
    vec3 d = vec3(0.20, 0.95, 0.60); // mint
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float wh_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);
    float a = atan(p.y, p.x);

    // Wormhole stretch: depth = log(1/r) (hyperbolic). Angle rotates with depth.
    float depth = -log(max(r, 0.02));
    float rotAngle = a + depth * ROTATION_SPD * sign(sin(x_Time * 0.1));

    vec3 col = vec3(0.0);

    // Streaming rings — each ring moves inward at TUNNEL_SPEED rings/sec
    for (int i = 0; i < RING_LAYERS; i++) {
        float fi = float(i);
        float phase = fract(depth * 1.5 - x_Time * TUNNEL_SPEED * (1.0 + fi * 0.05) + fi * 0.15);
        float ringMask = smoothstep(0.0, 0.1, phase) * smoothstep(0.3, 0.25, phase);

        // Angular ribs per ring (mouths cells)
        float ribs = 0.5 + 0.5 * cos(rotAngle * 12.0 + fi * 0.7);
        ribs = pow(ribs, 4.0);

        // Layer color
        vec3 rc = wh_pal(fract(fi * 0.08 + depth * 0.05 + x_Time * 0.04));
        // Depth-driven brightness: deeper rings brighter (core glow)
        float depthBright = smoothstep(0.5, 2.5, depth);
        col += rc * ringMask * (0.3 + ribs * 0.7) * depthBright * 0.3;
    }

    // Event-horizon-like bright core
    float coreGlow = exp(-r * r / (CORE_RADIUS * CORE_RADIUS) * 3.0);
    col += wh_pal(fract(x_Time * 0.06)) * coreGlow * 1.8;

    // Thin photon ring at ~2×core
    float photonR = CORE_RADIUS * 2.0;
    float photonRing = exp(-abs(r - photonR) * 600.0) * 1.2;
    col += vec3(0.95, 0.95, 1.0) * photonRing;

    // Star streaks drawn past the mouth (outside wormhole throat)
    if (r > 0.4) {
        // Radial streak sampling
        vec2 starP = floor(vec2(a * 8.0, depth * 2.0) + x_Time * 0.5);
        float starH = wh_hash(starP);
        if (starH > 0.97) {
            col += vec3(0.9, 0.92, 1.0) * 0.35 * (starH - 0.97) * 30.0;
        }
    }

    // Vignette around throat
    float vignette = smoothstep(1.5, 0.2, r);
    col *= vignette;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
