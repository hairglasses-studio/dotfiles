// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Deep sea hydrothermal vent — dark water + rising black smoke + tube worms + mineral glow

const int   WORMS = 12;
const int   OCTAVES = 5;
const float INTENSITY = 0.55;

vec3 dsv_pal(float t) {
    vec3 warm_orange = vec3(1.00, 0.50, 0.20);
    vec3 mag  = vec3(0.90, 0.25, 0.55);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 cyan = vec3(0.15, 0.85, 0.95);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(warm_orange, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, cyan, s - 2.0);
    else              return mix(cyan, warm_orange, s - 3.0);
}

float dsv_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float dsv_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }

float dsv_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(dsv_hash(i), dsv_hash(i + vec2(1,0)), u.x),
               mix(dsv_hash(i + vec2(0,1)), dsv_hash(i + vec2(1,1)), u.x), u.y);
}

float dsv_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * dsv_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Very dark deep ocean
    vec3 bg = mix(vec3(0.01, 0.02, 0.05), vec3(0.0, 0.01, 0.02), 1.0 - uv.y);
    vec3 col = bg;

    // Vent chimney at bottom
    float chimneyY = -0.3;
    float chimneyW = 0.08;
    if (p.y < chimneyY && abs(p.x) < chimneyW) {
        col = vec3(0.04, 0.03, 0.02);
        // Cracks with mineral glow
        float cracks = dsv_fbm(p * 15.0);
        if (cracks > 0.55) {
            col += dsv_pal(0.3) * (cracks - 0.55) * 4.0 * 0.9;
        }
    }

    // Black smoker plume rising
    if (p.y > chimneyY && abs(p.x) < chimneyW * 2.5) {
        float heightAbove = p.y - chimneyY;
        float plumeW = chimneyW + heightAbove * 0.2;
        float xDist = abs(p.x);
        if (xDist < plumeW) {
            // Turbulent smoke
            float plumeTurb = dsv_fbm(vec2(p.x * 6.0, p.y * 4.0 - x_Time * 1.5));
            float plumeMask = exp(-xDist * xDist / (plumeW * plumeW) * 1.5);
            float densityAtHeight = 1.0 - smoothstep(0.0, 1.5, heightAbove);
            // Color: reddish near vent, darkens to black higher
            vec3 smokeCol = mix(dsv_pal(0.1), vec3(0.02, 0.02, 0.03), heightAbove * 0.5);
            float smokeDensity = plumeMask * plumeTurb * densityAtHeight;
            col = mix(col, smokeCol, smokeDensity * 0.8);
            // Hot glow near base
            float baseGlow = exp(-heightAbove * 5.0) * plumeMask * 0.5;
            col += dsv_pal(0.0) * baseGlow;
        }
    }

    // Tube worms growing from chimney sides
    for (int i = 0; i < WORMS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        bool leftSide = mod(fi, 2.0) < 1.0;
        float wormSide = leftSide ? -1.0 : 1.0;
        float baseY = chimneyY + dsv_hash1(seed * 3.7) * 0.2;
        vec2 wormBase = vec2(wormSide * (chimneyW + 0.005), baseY);
        float wormLen = 0.08 + dsv_hash1(seed * 5.1) * 0.05;
        // Sway
        float swayAng = sin(x_Time * 0.5 + seed) * 0.4 - wormSide * 1.0;
        vec2 wormTip = wormBase + vec2(cos(swayAng), sin(swayAng) + 0.3) * wormLen;

        vec2 pa = p - wormBase;
        vec2 ba = wormTip - wormBase;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        float d = length(pa - ba * h);
        if (d < 0.01) {
            // Worm body
            float bodyMask = smoothstep(0.01, 0.003, d);
            col = mix(col, vec3(0.9, 0.85, 0.80), bodyMask * 0.6);
            // Red plume on top
            if (h > 0.7) {
                col += vec3(0.95, 0.25, 0.30) * bodyMask * 0.8;
            }
        }
    }

    // Mineral glow at base of chimney
    float baseD = length(p - vec2(0.0, chimneyY));
    col += dsv_pal(0.0) * exp(-baseD * baseD * 80.0) * 0.3;

    // Bioluminescent plankton drifting
    for (int b = 0; b < 20; b++) {
        float fb = float(b);
        float bphase = fract(x_Time * 0.1 * (0.5 + dsv_hash1(fb)) + dsv_hash1(fb * 3.0));
        vec2 bp = vec2(
            (dsv_hash1(fb * 5.0) - 0.5) * 2.0 * x_WindowSize.x / x_WindowSize.y,
            -0.5 + bphase * 1.2
        );
        float bd = length(p - bp);
        col += vec3(0.3, 0.85, 0.95) * exp(-bd * bd * 4000.0) * 0.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
