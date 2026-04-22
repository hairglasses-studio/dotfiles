// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Splash crown — droplet impact creates radial crown of fluid spikes + dome + secondary droplets

const int   CROWN_POINTS = 18;
const int   DROPLETS = 30;
const float INTENSITY = 0.55;

vec3 sc_pal(float t) {
    vec3 white = vec3(0.95, 0.98, 1.00);
    vec3 cyan  = vec3(0.20, 0.85, 0.95);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 mag   = vec3(0.95, 0.30, 0.70);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(white, cyan, s);
    else if (s < 2.0) return mix(cyan, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else              return mix(mag, white, s - 3.0);
}

float sc_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  sc_hash2(float n) { return vec2(sc_hash(n), sc_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Dark water-ish backdrop
    vec3 bg = mix(vec3(0.02, 0.05, 0.12), vec3(0.01, 0.02, 0.06), 1.0 - uv.y);
    vec3 col = bg;

    // Splash cycle
    float cycle = 3.0;
    float phase = mod(x_Time, cycle) / cycle;
    float cycleID = floor(x_Time / cycle);

    // Impact location this cycle
    vec2 impact = sc_hash2(cycleID * 5.1) * 0.6 - 0.3;

    vec2 toImpact = p - impact;
    float r = length(toImpact);

    // Crown expansion envelope (0-0.4 expand, 0.4-1 fade)
    float crownPhase = smoothstep(0.0, 0.3, phase) * (1.0 - smoothstep(0.5, 1.0, phase));
    float crownR = phase * 0.3;

    // Crown points — radial spikes from base of crown
    for (int k = 0; k < CROWN_POINTS; k++) {
        float fk = float(k);
        float ang = fk / float(CROWN_POINTS) * 6.28;
        vec2 spikeBase = impact + vec2(cos(ang), sin(ang)) * crownR;
        // Spike direction — outward + slightly up
        vec2 spikeDir = normalize(vec2(cos(ang), sin(ang) + 0.5));
        float spikeLen = 0.04 * crownPhase * (0.7 + sc_hash(fk + cycleID * 13.0) * 0.6);
        vec2 spikeEnd = spikeBase + spikeDir * spikeLen;

        // Distance to spike line
        vec2 pa = p - spikeBase;
        vec2 ba = spikeEnd - spikeBase;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        float d = length(pa - ba * h);
        float spikeMask = exp(-d * d * 20000.0);
        col += vec3(0.9, 0.95, 1.0) * spikeMask * crownPhase * 1.2;
        // Droplet at tip (forms droplet when spike extends)
        float tipD = length(p - spikeEnd);
        col += sc_pal(fract(fk * 0.07)) * exp(-tipD * tipD * 50000.0) * crownPhase * 0.9;
    }

    // Central column (water rising from impact)
    if (phase < 0.6) {
        float centralH = phase * 0.2;
        vec2 centralDiff = p - impact - vec2(0.0, centralH * 0.5);
        if (abs(centralDiff.x) < 0.015 && abs(centralDiff.y) < centralH) {
            float centralMask = exp(-centralDiff.x * centralDiff.x * 3000.0);
            col += vec3(0.8, 0.9, 0.98) * centralMask * (1.0 - phase / 0.6);
        }
    }

    // Secondary droplets flying outward
    for (int i = 0; i < DROPLETS; i++) {
        float fi = float(i);
        float seed = fi * 7.31 + cycleID * 11.3;
        float dropAng = fi / float(DROPLETS) * 6.28 + sc_hash(seed) * 0.5;
        float dropSpeed = 0.5 + sc_hash(seed * 3.1) * 0.3;
        vec2 dropPos = impact + vec2(cos(dropAng), sin(dropAng) + 0.5) * phase * dropSpeed;
        // Gravity on y
        dropPos.y -= phase * phase * 0.4;
        float dd = length(p - dropPos);
        float dropMask = exp(-dd * dd * 60000.0);
        col += sc_pal(fract(seed * 0.02)) * dropMask * (1.0 - phase) * 1.1;
    }

    // Expanding water ring at base of crown
    float ringDist = abs(r - crownR);
    float ringMask = exp(-ringDist * ringDist * 2000.0) * crownPhase;
    col += vec3(0.6, 0.85, 0.95) * ringMask * 0.5;

    // Impact flash
    if (phase < 0.08) {
        float flash = 1.0 - phase / 0.08;
        col += vec3(1.0) * exp(-r * r * 400.0) * flash * 1.5;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
