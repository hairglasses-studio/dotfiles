// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Tokamak fusion reactor — toroidal plasma confinement with bright magnetic rings + hot core

const int   SAMPLES = 64;
const float TORUS_R = 0.25;    // major radius
const float TORUS_r = 0.07;    // minor radius
const float INTENSITY = 0.55;

vec3 tk_pal(float t) {
    vec3 vio   = vec3(0.55, 0.25, 0.95);
    vec3 mag   = vec3(0.95, 0.30, 0.65);
    vec3 hot   = vec3(1.00, 0.65, 0.20);
    vec3 white = vec3(1.00, 0.98, 0.85);
    if (t < 0.33)      return mix(vio, mag, t * 3.0);
    else if (t < 0.66) return mix(mag, hot, (t - 0.33) * 3.0);
    else               return mix(hot, white, (t - 0.66) * 3.0);
}

float tk_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.04);

    // Distance from torus ring (2D projection)
    float r = length(p);
    float majorDist = abs(r - TORUS_R);

    // Plasma if near the ring
    if (majorDist < TORUS_r * 2.0) {
        // Angle around torus
        float theta = atan(p.y, p.x);
        // Position within cross-section
        float minorR = majorDist / TORUS_r;

        // Plasma turbulence — animated
        float turb = 0.0;
        for (int i = 0; i < 5; i++) {
            float fi = float(i);
            turb += sin(theta * (3.0 + fi * 2.0) + x_Time * (1.0 + fi * 0.3)) / (fi + 2.0);
        }
        turb = turb * 0.5 + 0.5;

        // Heat: brightest near center of ring
        float heat = exp(-minorR * minorR * 3.0) * (0.6 + turb * 0.4);

        vec3 plasmaCol = tk_pal(heat);
        col += plasmaCol * heat * 1.2;

        // Bright center line of torus
        float centerLine = exp(-majorDist * majorDist * 800.0);
        col += tk_pal(0.85) * centerLine * 0.6;
    }

    // Magnetic field coils — bright rings at toroidal positions
    for (int k = 0; k < 8; k++) {
        float fk = float(k);
        float coilAng = fk / 8.0 * 6.28318 + x_Time * 0.2;
        vec2 coilPos = vec2(cos(coilAng), sin(coilAng)) * TORUS_R;
        float coilD = length(p - coilPos);
        float ringD = abs(coilD - TORUS_r * 1.3);
        float ring = exp(-ringD * ringD * 2000.0);
        col += vec3(0.3, 0.5, 0.95) * ring * 0.7;
    }

    // Outer vessel wall — dim ring at major radius
    float wallD = abs(r - TORUS_R * 1.5);
    col += vec3(0.15, 0.18, 0.25) * smoothstep(0.005, 0.0, wallD) * 0.6;

    // Particle flux — spirals along toroidal direction
    for (int s = 0; s < 40; s++) {
        float fs = float(s);
        float particlePhase = fract(x_Time * 0.6 + fs * 0.025);
        float pTheta = particlePhase * 6.28318 + fs * 0.5;
        float pOffset = (tk_hash(vec2(fs, 0.0)) - 0.5) * TORUS_r * 0.8;
        float pRadius = TORUS_R + pOffset;
        vec2 pPos = vec2(cos(pTheta), sin(pTheta)) * pRadius;
        float pd = length(p - pPos);
        col += vec3(1.0, 0.95, 0.85) * exp(-pd * pd * 15000.0) * 0.7;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
