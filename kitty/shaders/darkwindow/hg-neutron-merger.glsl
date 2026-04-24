// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Neutron star merger — inspiral chirp → kilonova flash → r-process ejecta shell

const int   EJECTA = 60;
const float INTENSITY = 0.55;
const float CYCLE = 8.0;

vec3 nm_pal(float t) {
    vec3 deep   = vec3(0.05, 0.05, 0.20);
    vec3 blue   = vec3(0.15, 0.45, 0.95);
    vec3 cyan   = vec3(0.25, 0.85, 1.00);
    vec3 white  = vec3(1.00, 0.98, 0.95);
    vec3 yellow = vec3(1.00, 0.90, 0.45);
    vec3 orange = vec3(1.00, 0.50, 0.20);
    vec3 mag    = vec3(0.95, 0.30, 0.55);
    float s = mod(t * 6.0, 6.0);
    if (s < 1.0)      return mix(deep, blue, s);
    else if (s < 2.0) return mix(blue, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, white, s - 2.0);
    else if (s < 4.0) return mix(white, yellow, s - 3.0);
    else if (s < 5.0) return mix(yellow, orange, s - 4.0);
    else              return mix(orange, mag, s - 5.0);
}

float nm_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float nm_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);

    vec3 col = vec3(0.005, 0.01, 0.03);

    // Background starfield (sparse)
    vec2 sg = floor(p * 140.0);
    float sh = nm_hash2(sg);
    if (sh > 0.996) col += vec3(0.85, 0.9, 1.0) * (sh - 0.996) * 200.0 * 0.3;

    // Cycle phase: 0 → 1 over CYCLE seconds
    float cycleT = mod(x_Time, CYCLE) / CYCLE;

    float inspiralEnd = 0.60;
    float mergerEnd   = 0.68;

    // === Phase 1: Inspiral — two stars with chirp + gravitational waves ===
    if (cycleT < inspiralEnd) {
        float iT = cycleT / inspiralEnd;           // [0,1]
        // Separation shrinks as GW emission drains energy
        float sep = 0.25 * pow(1.0 - iT, 0.6) + 0.012;
        // Orbital frequency increases quadratically (chirp)
        float orbAng = iT * iT * 70.0 + x_Time * 2.0;
        vec2 a = vec2(cos(orbAng), sin(orbAng)) * sep;
        vec2 b = -a;

        float dA = length(p - a);
        float dB = length(p - b);

        // NS cores: hot compact blue-white
        col += vec3(0.75, 0.88, 1.0) * exp(-dA * dA * 40000.0) * 1.3;
        col += vec3(0.75, 0.88, 1.0) * exp(-dB * dB * 40000.0) * 1.3;
        // Soft halos around cores
        col += vec3(0.30, 0.55, 0.95) * exp(-dA * dA * 800.0) * 0.45;
        col += vec3(0.30, 0.55, 0.95) * exp(-dB * dB * 800.0) * 0.45;

        // Gravitational wave ripples — radial cosine, wavelength contracts
        float wavelen = 0.06 * (1.0 - iT) + 0.008;
        float gwPhase = (r - x_Time * 0.4) / wavelen;
        float gwAmp = 0.25 * iT * iT;              // amplitude grows during inspiral
        float gwRipple = cos(gwPhase * 6.28) * 0.5 + 0.5;
        col += nm_pal(fract(r * 0.5 + x_Time * 0.05)) * gwRipple * gwAmp * exp(-r * 1.5);

        // Short trailing wisps behind each star (velocity perpendicular to radial)
        vec2 velA = vec2(-sin(orbAng), cos(orbAng));
        float tailA = dot(p - a, -velA);
        vec2 perpVA = (p - a) + velA * tailA;
        float perpA = length(perpVA);
        if (tailA > 0.0 && tailA < 0.045) {
            col += vec3(0.5, 0.75, 1.0) * exp(-perpA * perpA * 40000.0) * (1.0 - tailA / 0.045) * 0.55;
        }
        float tailB = dot(p - b, velA);
        vec2 perpVB = (p - b) - velA * tailB;
        float perpB = length(perpVB);
        if (tailB > 0.0 && tailB < 0.045) {
            col += vec3(0.95, 0.55, 0.85) * exp(-perpB * perpB * 40000.0) * (1.0 - tailB / 0.045) * 0.55;
        }
    }
    // === Phase 2: Merger flash — kilonova ===
    else if (cycleT < mergerEnd) {
        float fT = (cycleT - inspiralEnd) / (mergerEnd - inspiralEnd); // [0,1]
        // Expand-then-contract flash disk
        float flashR = 0.15 * sin(fT * 3.14159);
        float flashD = abs(r - flashR * 0.3);
        float flashCore = exp(-flashD * flashD * 300.0);
        float flashGlow = exp(-r * r * 50.0) * (1.0 - fT * 0.6);
        col += vec3(1.0, 0.96, 0.88) * (flashCore + flashGlow) * (1.6 - fT * 0.5) * 1.5;
        // Radial stellar spikes from the flash
        float ang = atan(p.y, p.x);
        float spikes = abs(sin(ang * 16.0)) * exp(-r * 6.0);
        col += vec3(1.0, 0.86, 0.55) * spikes * (1.0 - fT) * 0.8;
    }
    // === Phase 3: Post-merger — remnant + expanding ejecta shell ===
    else {
        float pT = (cycleT - mergerEnd) / (1.0 - mergerEnd); // [0,1]

        // Hot central remnant (hypermassive NS → BH)
        float remR = 0.04;
        col += vec3(1.0, 0.90, 0.70) * exp(-r * r / (remR * remR) * 0.8) * (1.25 - pT * 0.5);
        // Swirling accretion-like veil
        float swirlAng = atan(p.y, p.x);
        float swirl = sin(swirlAng * 5.0 - x_Time * 4.0 + r * 30.0) * 0.5 + 0.5;
        col += nm_pal(fract(swirlAng * 0.2 + x_Time * 0.1)) * swirl * exp(-r * r * 80.0) * 0.7;

        // Expanding kilonova shell (ring)
        float shellR = pT * 0.9;
        float shellD = abs(r - shellR);
        float shellMask = exp(-shellD * shellD * 600.0) * (1.0 - pT * 0.6);
        col += nm_pal(fract(shellR * 0.5 + x_Time * 0.05)) * shellMask * 1.3;

        // Ejecta clumps — r-process enriched with bimodal color coding
        for (int i = 0; i < EJECTA; i++) {
            float fi = float(i);
            float seed = fi * 7.31;
            float ejAng = nm_hash(seed) * 6.28;
            float ejSpeed = 0.3 + nm_hash(seed * 3.7) * 0.6;
            float ejR = pT * ejSpeed;
            vec2 ejPos = vec2(cos(ejAng), sin(ejAng)) * ejR;
            float ed = length(p - ejPos);
            float ecore = exp(-ed * ed * 20000.0);
            // Heavy r-process (lanthanides/actinides) → blue-violet; light alpha → red-orange
            float heavy = nm_hash(seed * 5.1);
            vec3 ejCol = heavy > 0.5
                ? mix(vec3(0.20, 0.30, 0.95), vec3(0.60, 0.40, 0.98), nm_hash(seed * 7.3))
                : mix(vec3(1.0, 0.50, 0.20), vec3(0.98, 0.30, 0.50), nm_hash(seed * 11.1));
            col += ejCol * ecore * (1.0 - pT * 0.5) * 1.1;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
