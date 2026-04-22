// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Smoke rings — rising vortex rings with FBM turbulence and neon palette drift

const int   RINGS   = 5;
const float RISE_SPEED = 0.15;
const float INTENSITY = 0.5;

vec3 sr_pal(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(0.90, 0.25, 0.60);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float sr_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float sr_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }

float sr_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(sr_hash(i), sr_hash(i + vec2(1,0)), u.x),
               mix(sr_hash(i + vec2(0,1)), sr_hash(i + vec2(1,1)), u.x), u.y);
}

float sr_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 4; i++) {
        v += a * sr_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.03);

    for (int i = 0; i < RINGS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        // Each ring at different cycle offset
        float cycle = 5.0;
        float phase = fract((x_Time + sr_hash1(seed) * cycle) / cycle);
        float cycleID = floor((x_Time + sr_hash1(seed) * cycle) / cycle);

        // Ring center rises over phase
        float centerY = -0.7 + phase * 1.8;
        // Slight horizontal wobble
        float centerX = (sr_hash1(seed + cycleID * 13.7) - 0.5) * 0.4;

        // Ring expands as it rises
        float ringR = 0.08 + phase * 0.15;
        // Ring thickness also grows
        float ringW = 0.015 + phase * 0.03;

        vec2 rp = p - vec2(centerX, centerY);
        float rd = length(rp);
        float ringDist = abs(rd - ringR);

        // Turbulence distortion
        float turbPhase = atan(rp.y, rp.x) + phase * 6.28;
        float turbAmount = 0.5 * sr_fbm(vec2(turbPhase, rd * 8.0 + phase * 4.0)) - 0.25;
        ringDist -= turbAmount * 0.015;

        // Ring body
        float ringBody = exp(-ringDist * ringDist / (ringW * ringW) * 2.0);
        // Interior hole visible
        float holeMask = smoothstep(ringR * 0.7, ringR * 0.9, rd);

        // Fade over lifetime
        float fade = 1.0 - phase * 0.7;

        vec3 rc = sr_pal(fract(fi * 0.15 + phase * 0.5 + cycleID * 0.1));
        col += rc * ringBody * holeMask * fade * 0.45;

        // Central swirl through ring
        if (rd < ringR * 0.9 && abs(ringDist) < ringR * 0.5) {
            float swirlAng = atan(rp.y, rp.x);
            float swirl = 0.5 + 0.5 * cos(swirlAng * 6.0 + phase * 10.0);
            col += rc * ringBody * swirl * 0.15;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
