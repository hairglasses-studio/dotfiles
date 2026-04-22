// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Zen garden — raked concentric sand patterns around rocks with contemplative palette

const int   ROCKS = 4;
const float INTENSITY = 0.55;

vec3 zg_pal(float t) {
    vec3 warm_sand = vec3(0.92, 0.78, 0.55);
    vec3 shadow = vec3(0.25, 0.18, 0.20);
    vec3 rock = vec3(0.45, 0.35, 0.35);
    vec3 moss = vec3(0.30, 0.55, 0.35);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(warm_sand, shadow, s);
    else if (s < 2.0) return mix(shadow, rock, s - 1.0);
    else if (s < 3.0) return mix(rock, moss, s - 2.0);
    else              return mix(moss, warm_sand, s - 3.0);
}

float zg_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  zg_hash2(float n) { return vec2(zg_hash(n), zg_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Sand background
    vec3 col = zg_pal(0.0);

    // Concentric rake patterns around each rock
    for (int i = 0; i < ROCKS; i++) {
        float fi = float(i);
        vec2 rockCenter = (zg_hash2(fi * 3.71) - 0.5) * 1.4;
        rockCenter.x *= x_WindowSize.x / x_WindowSize.y;

        float rockSize = 0.04 + zg_hash(fi * 3.7) * 0.04;
        float d = length(p - rockCenter);

        // Rock body
        if (d < rockSize) {
            col = zg_pal(0.45);
            // Rock shading
            float rockShade = smoothstep(rockSize * 1.1, rockSize * 0.6, d);
            col = mix(col, zg_pal(0.25), 1.0 - rockShade);
            // Moss spots on some rocks
            if (mod(fi, 2.0) < 1.0) {
                float moss = sin(p.x * 50.0) * sin(p.y * 50.0);
                if (moss > 0.5) col = mix(col, zg_pal(0.75), moss * 0.6);
            }
        } else {
            // Concentric raked rings around rock
            float ringSpacing = 0.025;
            float ringFromEdge = (d - rockSize);
            float ringPhase = ringFromEdge / ringSpacing;
            float ringDist = abs(fract(ringPhase) - 0.5);
            // Shadow line
            if (ringDist < 0.1) {
                col = mix(col, zg_pal(0.2), (0.1 - ringDist) * 5.0 * 0.4);
            }
        }
    }

    // Horizontal straight raking (across whole garden, outside ring patterns)
    // Offset by proximity to rocks (rakes go around)
    float horizRake = sin(p.y * 40.0);
    float minRockD = 1e9;
    for (int i = 0; i < ROCKS; i++) {
        float fi = float(i);
        vec2 rockCenter = (zg_hash2(fi * 3.71) - 0.5) * 1.4;
        rockCenter.x *= x_WindowSize.x / x_WindowSize.y;
        minRockD = min(minRockD, length(p - rockCenter));
    }
    float horizFade = smoothstep(0.15, 0.5, minRockD);  // only far from rocks
    if (horizFade > 0.5 && horizRake > 0.6) {
        col = mix(col, zg_pal(0.2), (horizRake - 0.6) * 2.0 * 0.3 * horizFade);
    }

    // Falling leaves drifting across
    for (int L = 0; L < 6; L++) {
        float fL = float(L);
        float seed = fL * 7.31;
        float driftSpeed = 0.03 + zg_hash(seed * 3.7) * 0.02;
        float dPhase = fract(x_Time * driftSpeed + zg_hash(seed));
        float leafX = -1.5 + dPhase * 3.0;
        leafX *= x_WindowSize.x / x_WindowSize.y;
        float leafY = 0.3 + (zg_hash(seed * 5.1) - 0.5) * 0.4 + 0.1 * sin(x_Time + fL);
        vec2 leafP = vec2(leafX, leafY);
        float ld = length(p - leafP);
        if (ld < 0.015) {
            col = mix(col, zg_pal(0.55), smoothstep(0.015, 0.005, ld) * 0.6);
        }
    }

    // Soft morning light vignette
    col *= 0.85 + 0.15 * smoothstep(1.0, 0.3, length(p));

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.55);
    vec3 result = mix(terminal.rgb, col, visibility * 0.85);

    _wShaderOut = vec4(result, 1.0);
}
