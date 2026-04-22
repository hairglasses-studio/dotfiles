// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk) — Retrowave warp tunnel — concentric neon rings spiraling inward

const float TUNNEL_SPEED    = 0.8;    // rings per second moving outward
const float SPIRAL_STRENGTH = 0.35;   // how much polar angle twists with depth
const float RING_DENSITY    = 5.5;    // rings per screen at u=1
const float RING_WIDTH      = 0.18;   // ring thickness (fraction of period)
const float INTENSITY       = 0.45;

vec3 neonCycle(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.90, 0.18, 0.60); // magenta
    vec3 c = vec3(0.55, 0.30, 0.98); // violet
    float s = mod(t * 3.0, 3.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else              return mix(c, a, s - 2.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    // Centered, aspect-corrected polar coords
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);
    float a = atan(p.y, p.x);

    // Tunnel coordinate: 1/r gives perspective (smaller r → deeper in)
    float depth = 1.0 / max(r, 0.04);
    float spiral = a * SPIRAL_STRENGTH + depth * 0.6;

    // Rings marching outward
    float ringCoord = depth - x_Time * TUNNEL_SPEED;
    float ringPhase = fract(ringCoord);
    float ring = smoothstep(0.5 - RING_WIDTH * 0.5, 0.5, ringPhase)
               * (1.0 - smoothstep(0.5, 0.5 + RING_WIDTH * 0.5, ringPhase));

    // 8-fold radial slices create "rib" accents
    float slice = 0.5 + 0.5 * cos(a * 8.0 + spiral);
    ring *= mix(0.6, 1.0, slice);

    // Color drifts with depth
    vec3 tun = neonCycle(fract(depth * 0.05 + x_Time * 0.05)) * ring;

    // Vignette: bright toward tunnel mouth (outer edge), fade at extreme center
    float vignette = smoothstep(0.02, 0.2, r);
    tun *= vignette;

    // Faint ambient glow from inside the tunnel
    vec3 glow = neonCycle(x_Time * 0.05) * 0.15 * (1.0 - smoothstep(0.0, 0.35, r));
    tun += glow;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.7);
    vec3 result = mix(terminal.rgb, tun, visibility * clamp(length(tun), 0.0, 1.0));

    _wShaderOut = vec4(result, 1.0);
}
