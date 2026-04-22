// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Solar system — sun + 6 orbiting planets with orbit trails + texture hints

const int   PLANETS = 6;
const float INTENSITY = 0.55;

vec3 ss_pal(float t) {
    vec3 sun = vec3(1.00, 0.85, 0.35);
    vec3 mag = vec3(0.95, 0.25, 0.55);
    vec3 cyan = vec3(0.20, 0.80, 0.95);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(sun, mag, s);
    else if (s < 2.0) return mix(mag, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, vio, s - 2.0);
    else              return mix(vio, sun, s - 3.0);
}

float ss_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.01, 0.03);

    // Sun
    float r = length(p);
    float sunR = 0.06;
    float sunCore = exp(-r * r / (sunR * sunR) * 2.0);
    float sunHalo = exp(-r * r * 50.0) * 0.4;
    vec3 sunCol = vec3(1.0, 0.92, 0.6);
    col += sunCol * sunCore * 1.8;
    col += ss_pal(0.0) * sunHalo;

    // Planets + orbits
    for (int i = 0; i < PLANETS; i++) {
        float fi = float(i);
        // Orbital radius
        float orbitR = 0.12 + fi * 0.065;
        // Orbit period (inner = faster, Kepler-ish)
        float period = pow(orbitR, 1.5) * 4.0;
        float theta = x_Time / period + fi * 1.3;
        vec2 planetPos = vec2(cos(theta), sin(theta) * 0.5) * orbitR;  // slight elliptical tilt

        // Orbit ring — subtle
        float orbitDist = abs(length(p * vec2(1.0, 2.0)) - orbitR);
        float orbitRing = exp(-orbitDist * orbitDist * 8000.0) * 0.2;
        col += vec3(0.15, 0.18, 0.22) * orbitRing;

        // Planet body
        float planetR = 0.008 + 0.006 * ss_hash(fi * 3.1);
        float pd = length(p - planetPos);

        // Planet color — varied, with simple banding
        vec3 planetCol = ss_pal(fract(fi * 0.15 + x_Time * 0.02));
        // Simple surface feature: horizontal bands (like Jupiter)
        if (pd < planetR) {
            float yOff = (p.y - planetPos.y) / planetR;
            float bands = sin(yOff * 8.0 + fi) * 0.5 + 0.5;
            planetCol *= 0.7 + bands * 0.3;
        }
        float core = exp(-pd * pd / (planetR * planetR) * 3.0);
        float halo = exp(-pd * pd * 3000.0) * 0.2;
        col += planetCol * core * 1.4;
        col += planetCol * halo;

        // Moon for the 3rd planet
        if (i == 2) {
            float moonTheta = x_Time * 3.0;
            vec2 moonPos = planetPos + vec2(cos(moonTheta), sin(moonTheta)) * 0.025;
            float md = length(p - moonPos);
            col += vec3(0.85, 0.85, 0.9) * exp(-md * md * 20000.0) * 0.8;
        }

        // Saturn-like ring for 4th
        if (i == 4) {
            vec2 toPlanet = p - planetPos;
            // Rotated ring
            float ringAng = sin(x_Time * 0.1);
            vec2 ringP = mat2(cos(ringAng), -sin(ringAng), sin(ringAng), cos(ringAng)) * toPlanet;
            ringP.y *= 3.0;  // flatten
            float ringD = abs(length(ringP) - planetR * 2.2);
            float ringMask = exp(-ringD * ringD * 1800.0);
            col += ss_pal(0.5) * ringMask * 0.5;
        }
    }

    // Distant stars
    vec2 sg = floor(p * 130.0);
    float sh = ss_hash(sg.x * 31.0 + sg.y);
    if (sh > 0.996) {
        col += vec3(0.9, 0.9, 1.0) * (sh - 0.996) * 200.0 * 0.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
