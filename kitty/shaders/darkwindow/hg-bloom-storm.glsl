// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Bloom storm — explosive radial flashes with ring aftershocks

const int   BLOOMS       = 5;
const float BLOOM_PERIOD = 2.5;    // seconds between blooms
const float INTENSITY    = 0.55;

vec3 bs_pal(float t) {
    vec3 hot  = vec3(1.00, 0.95, 0.60);
    vec3 mag  = vec3(0.95, 0.30, 0.70);
    vec3 cyan = vec3(0.10, 0.82, 0.92);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(hot, mag, s);
    else if (s < 2.0) return mix(mag, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, vio, s - 2.0);
    else              return mix(vio, hot, s - 3.0);
}

float bs_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  bs_hash2(float n) { return vec2(bs_hash1(n), bs_hash1(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    for (int b = 0; b < BLOOMS; b++) {
        float fb = float(b);
        // Each bloom has its own phase offset + period + position
        float period = BLOOM_PERIOD + fb * 0.4;
        float phase = mod(x_Time + fb * period * 0.3, period) / period;   // [0,1]

        // Bloom center — random per cycle
        float cycleID = floor((x_Time + fb * period * 0.3) / period);
        float seed = fb * 7.3 + cycleID * 13.7;
        vec2 center = bs_hash2(seed) * 1.4 - 0.7;
        center.x *= x_WindowSize.x / x_WindowSize.y;

        float d = length(p - center);

        // Initial flash + expanding ring
        float flashR = phase * 0.8;   // expansion radius
        float ringD = abs(d - flashR);
        float ringWidth = 0.005 + phase * 0.04;  // ring thickens as it expands
        float ring = exp(-ringD * ringD / (ringWidth * ringWidth) * 2.0);

        // Intensity fades as ring expands
        float intensity = pow(1.0 - phase, 2.0);

        vec3 bloomCol = bs_pal(fract(seed * 0.03 + x_Time * 0.05));
        col += bloomCol * ring * intensity * 1.2;

        // Central bright core at start
        if (phase < 0.15) {
            float core = exp(-d * d * 200.0) * (1.0 - phase / 0.15) * 2.0;
            col += bloomCol * core;
            col += vec3(1.0, 0.95, 0.85) * core * 0.5;
        }

        // Aftershock: second smaller ring trailing main
        if (phase > 0.2 && phase < 0.7) {
            float afterR = (phase - 0.2) * 0.5;
            float afterD = abs(d - afterR);
            float afterRing = exp(-afterD * afterD * 2000.0) * intensity * 0.4;
            col += bloomCol * afterRing;
        }

        // Radial spokes in early phase
        if (phase < 0.4) {
            float angle = atan(p.y - center.y, p.x - center.x);
            float spokes = 0.5 + 0.5 * cos(angle * 16.0 + seed);
            spokes = pow(spokes, 8.0);
            float spokeMask = smoothstep(0.0, 0.1, d) * smoothstep(0.6, 0.1, d);
            col += bloomCol * spokes * spokeMask * intensity * 0.3;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
