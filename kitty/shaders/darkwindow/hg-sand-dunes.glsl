// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Sand dunes — rolling dune silhouettes against sunset sky + drifting sand particles

const int   DUNE_LAYERS = 5;
const int   PARTICLES = 80;
const int   OCTAVES = 5;
const float INTENSITY = 0.55;

vec3 sd_pal(float t) {
    vec3 warm_orange = vec3(1.00, 0.55, 0.20);
    vec3 mag   = vec3(0.85, 0.25, 0.55);
    vec3 vio   = vec3(0.45, 0.20, 0.75);
    vec3 gold  = vec3(0.95, 0.75, 0.35);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(warm_orange, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, gold, s - 2.0);
    else              return mix(gold, warm_orange, s - 3.0);
}

float sd_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float sd_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }

float sd_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(sd_hash(i), sd_hash(i + vec2(1,0)), u.x),
               mix(sd_hash(i + vec2(0,1)), sd_hash(i + vec2(1,1)), u.x), u.y);
}

float sd_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * sd_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Sky gradient
    vec3 bg;
    if (uv.y > 0.5) {
        float skyT = (uv.y - 0.5) / 0.5;
        bg = mix(sd_pal(0.0), sd_pal(0.4), skyT);
    } else {
        bg = mix(sd_pal(0.3), vec3(0.1, 0.05, 0.15), 0.5 - uv.y);
    }
    vec3 col = bg;

    // Sun
    vec2 sunPos = vec2(0.1, 0.0);
    float sunD = length(p - sunPos);
    col += vec3(1.0, 0.9, 0.6) * smoothstep(0.08, 0.06, sunD) * 0.9;
    col += sd_pal(0.1) * exp(-sunD * sunD * 15.0) * 0.4;

    // Dune layers — each layer is a silhouette defined by FBM
    for (int L = 0; L < DUNE_LAYERS; L++) {
        float fL = float(L);
        float layerY = -0.05 - fL * 0.06;
        float scrollSpeed = 0.01 * (1.0 + fL * 0.3);
        float scale = 2.0 + fL * 0.5;
        float duneAmp = 0.06 + fL * 0.02;
        // Dune height at this x
        float duneHeight = sd_fbm(vec2(p.x * scale + x_Time * scrollSpeed, fL)) * duneAmp;
        float duneY = layerY + duneHeight;

        if (uv.y < duneY + 0.5) {
            // Layer silhouette — darker with depth
            float depthShade = 0.15 + fL * 0.1;
            vec3 duneCol = mix(sd_pal(0.5), vec3(0.02, 0.01, 0.04), depthShade);
            // Sunset side-light brighter
            if (p.x > 0.0) duneCol = mix(duneCol, sd_pal(0.0), 0.3 * (1.0 - depthShade));
            col = duneCol;
        }
    }

    // Sand ripples on nearest dunes
    float rippleY = -0.3;
    if (uv.y < 0.3) {
        float rippleFbm = sd_fbm(p * 40.0);
        float rippleMask = smoothstep(0.3, 0.5, rippleFbm);
        col += sd_pal(0.3) * rippleMask * 0.1 * smoothstep(0.0, 0.3, uv.y);
    }

    // Drifting sand particles
    for (int pi = 0; pi < PARTICLES; pi++) {
        float fpi = float(pi);
        float speed = 0.05 + sd_hash1(fpi) * 0.1;
        float phase = fract(x_Time * speed + sd_hash1(fpi * 3.7));
        float partX = -1.3 + phase * 2.6;
        partX *= x_WindowSize.x / x_WindowSize.y;
        float partY = (sd_hash1(fpi * 5.1) - 0.5) * 0.3 - 0.1;
        partY += 0.02 * sin(x_Time * 3.0 + fpi);
        vec2 pp = vec2(partX, partY);
        float pd = length(p - pp);
        col += sd_pal(0.25) * exp(-pd * pd * 60000.0) * 0.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
