// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Ripple drops — occasional large drops creating clean expanding rings with reflection

const int   DROPS = 5;
const float INTENSITY = 0.55;

vec3 rd_pal(float t) {
    vec3 cyan = vec3(0.20, 0.85, 0.95);
    vec3 white = vec3(0.95, 0.98, 1.00);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 mag  = vec3(0.95, 0.30, 0.70);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, white, s);
    else if (s < 2.0) return mix(white, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else              return mix(mag, cyan, s - 3.0);
}

float rdr_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  rdr_hash2(float n) { return vec2(rdr_hash(n), rdr_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Dark water background with gradient
    vec3 bg = mix(vec3(0.04, 0.10, 0.25), vec3(0.01, 0.03, 0.12), 1.0 - uv.y);

    // Sum of wave displacements for refraction
    vec2 totalDisplacement = vec2(0.0);
    vec3 waveCol = vec3(0.0);

    for (int i = 0; i < DROPS; i++) {
        float fi = float(i);
        float cycle = 4.0 + rdr_hash(fi * 3.7) * 2.0;
        float phase = fract((x_Time + rdr_hash(fi) * cycle) / cycle);
        float cycleID = floor((x_Time + rdr_hash(fi) * cycle) / cycle);

        vec2 dropCenter = rdr_hash2(fi + cycleID * 13.7) * 1.4 - 0.7;
        dropCenter.x *= x_WindowSize.x / x_WindowSize.y;

        float d = length(p - dropCenter);

        // Main expanding ring — grows with phase
        float ringR = phase * 0.9;
        float ringDist = abs(d - ringR);
        float ringW = 0.005 + phase * 0.02;

        // Ring wave amplitude (decays outward)
        float fade = 1.0 - phase;
        float wave = exp(-ringDist * ringDist / (ringW * ringW) * 2.0) * fade;

        // Displacement toward ring (causes lensing)
        if (d > 0.01 && wave > 0.1) {
            vec2 radialDir = (p - dropCenter) / d;
            totalDisplacement += radialDir * wave * 0.012 * sign(d - ringR);
        }

        // Secondary ring (echo)
        if (phase > 0.2) {
            float echoR = (phase - 0.2) * 0.5;
            float echoDist = abs(d - echoR);
            float echo = exp(-echoDist * echoDist * 2000.0) * (1.0 - phase) * 0.4;
            waveCol += rd_pal(fract(fi * 0.15 + x_Time * 0.05)) * echo * 0.5;
        }

        // Ring color
        waveCol += rd_pal(fract(fi * 0.15 + x_Time * 0.04)) * wave * 0.6;

        // Central splash flash
        if (phase < 0.1) {
            float splashFade = 1.0 - phase / 0.1;
            waveCol += vec3(1.0, 0.98, 0.95) * exp(-d * d * 1000.0) * splashFade * 1.2;
        }
    }

    // Apply refraction to terminal sample
    vec4 terminal = x_Texture(uv + totalDisplacement);

    vec3 col = mix(terminal.rgb, bg, 0.3);
    col += waveCol;

    // Composite
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.4);
    vec3 result = mix(terminal.rgb, col, visibility * 0.75);

    _wShaderOut = vec4(result, 1.0);
}
