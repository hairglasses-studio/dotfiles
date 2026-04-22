// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Gravitational waves — expanding ripples distorting space-time grid + dual central objects

const int   RINGS = 12;
const float INTENSITY = 0.55;

vec3 gw_pal(float t) {
    vec3 deep = vec3(0.05, 0.02, 0.15);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 mag  = vec3(0.95, 0.30, 0.65);
    vec3 white = vec3(0.95, 0.98, 1.00);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(deep, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, white, s - 2.0);
    else              return mix(white, deep, s - 3.0);
}

float gw_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.01, 0.03);

    // Two compact objects orbiting center — binary pulsar
    float orbitT = x_Time * 0.8;
    float orbitSep = 0.08 + 0.01 * sin(x_Time * 0.3);
    vec2 obj1 = vec2(cos(orbitT), sin(orbitT)) * orbitSep;
    vec2 obj2 = -obj1;

    // Object rendering
    float d1 = length(p - obj1);
    float d2 = length(p - obj2);
    col += vec3(1.0, 0.9, 0.85) * exp(-d1 * d1 * 8000.0) * 1.5;
    col += vec3(1.0, 0.9, 0.85) * exp(-d2 * d2 * 8000.0) * 1.5;
    col += gw_pal(0.2) * exp(-d1 * d1 * 300.0) * 0.4;
    col += gw_pal(0.2) * exp(-d2 * d2 * 300.0) * 0.4;

    // Space-time grid — lines warped by waves
    vec2 waveP = p;
    // Sum of outgoing gravitational waves (each ring is one)
    float r = length(p);
    for (int ring = 0; ring < RINGS; ring++) {
        float fr = float(ring);
        float ringAge = fract(x_Time * 0.4 + fr * 0.15);
        float ringR = ringAge * 1.3;
        // Wave amplitude near this ring
        float rDiff = r - ringR;
        // Quadrupole pattern: amplitude varies with cos(2*theta)
        float theta = atan(p.y, p.x);
        float quadrupole = cos(theta * 2.0 - orbitT * 2.0);
        float amplitudeEnv = exp(-rDiff * rDiff * 80.0);
        float waveAmp = amplitudeEnv * (1.0 - ringAge) * quadrupole * 0.03;
        // Displace grid position
        waveP += normalize(p) * waveAmp;

        // Ring visibility
        float ringMask = exp(-rDiff * rDiff * 300.0) * (1.0 - ringAge);
        col += gw_pal(fract(fr * 0.1 + x_Time * 0.04)) * ringMask * 0.3;
    }

    // Draw the warped grid
    vec2 grid = fract(waveP * 8.0);
    vec2 gridDist = abs(grid - 0.5);
    float gridLine = smoothstep(0.48, 0.5, max(gridDist.x, gridDist.y));
    col += gw_pal(0.4) * gridLine * 0.3;

    // Background stars
    vec2 sg = floor(p * 100.0);
    float sh = gw_hash(sg);
    if (sh > 0.995) {
        col += vec3(0.9, 0.9, 1.0) * (sh - 0.995) * 200.0 * 0.5;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
