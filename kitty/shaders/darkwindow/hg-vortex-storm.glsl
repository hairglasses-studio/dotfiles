// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Vortex storm — cyclonic cloud bands with eye + lightning flashes inside

const int   OCTAVES = 6;
const float EYE_R   = 0.05;
const float INTENSITY = 0.5;

vec3 vs_pal(float t) {
    vec3 dark   = vec3(0.10, 0.08, 0.18);
    vec3 vio    = vec3(0.45, 0.20, 0.75);
    vec3 mag    = vec3(0.85, 0.30, 0.65);
    vec3 gold   = vec3(0.95, 0.75, 0.30);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(dark, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, gold, s - 2.0);
    else              return mix(gold, dark, s - 3.0);
}

float vs_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float vs_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(vs_hash(i), vs_hash(i + vec2(1,0)), u.x),
               mix(vs_hash(i + vec2(0,1)), vs_hash(i + vec2(1,1)), u.x), u.y);
}

float vs_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * vs_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    float r = length(p);
    float a = atan(p.y, p.x);

    // Spiraling texture: sample FBM in spiral coords
    float swirlAng = a + log(r + 0.05) * 1.5 + x_Time * 0.15;
    vec2 spiralP = vec2(cos(swirlAng), sin(swirlAng)) * r * 6.0;
    float cloud = vs_fbm(spiralP);

    // Cloud bands — alternating light/dark
    float bandMask = smoothstep(0.35, 0.65, cloud);

    vec3 col = vec3(0.03, 0.02, 0.08);

    // Overall storm disc
    float discMask = smoothstep(0.9, 0.4, r);

    // Base cloud color — darker near eye, lighter farther
    vec3 cloudCol = vs_pal(fract(r + x_Time * 0.03));
    col = mix(col, cloudCol * 0.7, bandMask * discMask);

    // Eye — clear dark hole
    if (r < EYE_R) {
        col = mix(col, vec3(0.02, 0.01, 0.05), 0.85);
    }

    // Eyewall — brightest region just outside eye
    float eyewallDist = abs(r - EYE_R * 1.5);
    float eyewall = exp(-eyewallDist * eyewallDist * 600.0) * 0.5;
    col += vs_pal(0.8) * eyewall * bandMask;

    // Rain bands — periodic darker radial bands
    float rainPhase = cos(swirlAng * 4.0 + x_Time * 0.4);
    float rain = smoothstep(0.4, 0.8, rainPhase) * discMask * 0.2;
    col *= 1.0 - rain;

    // Lightning flashes — inside eyewall region
    float lightningPhase = fract(x_Time * 2.0);
    float lightningTrigger = vs_hash(vec2(floor(x_Time * 2.0), 0.0));
    if (lightningTrigger > 0.7 && lightningPhase < 0.15) {
        float flashFade = 1.0 - lightningPhase / 0.15;
        // Randomly positioned flash
        vec2 flashPos = vec2(
            cos(lightningTrigger * 6.28) * EYE_R * 2.0,
            sin(lightningTrigger * 6.28) * EYE_R * 2.0
        );
        float flashD = length(p - flashPos);
        col += vec3(0.9, 0.95, 1.0) * exp(-flashD * flashD * 200.0) * flashFade * 0.9;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
