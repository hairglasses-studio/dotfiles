// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Pool ripples — expanding concentric waves from 8 random drop points, wave sum interference

const int   DROPS = 8;
const float DROP_LIFE = 3.0;    // seconds from drop to dissipation
const float WAVE_SPEED = 0.25;
const float WAVE_FREQ  = 55.0;
const float INTENSITY  = 0.5;

vec3 pr_pal(float t) {
    vec3 a = vec3(0.10, 0.50, 0.92); // cool blue
    vec3 b = vec3(0.20, 0.85, 0.95); // cyan
    vec3 c = vec3(0.55, 0.30, 0.98); // violet
    vec3 d = vec3(0.95, 0.75, 0.45); // warm highlight
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float pr_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  pr_hash2(float n) { return vec2(pr_hash(n), pr_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Water-deep background
    vec3 bg = mix(vec3(0.04, 0.10, 0.25), vec3(0.02, 0.05, 0.18), uv.y);
    vec3 col = bg;

    float waveSum = 0.0;

    for (int i = 0; i < DROPS; i++) {
        float fi = float(i);
        // Each drop has its own cycle + position resampled per cycle
        float phase = mod(x_Time + fi * DROP_LIFE * 0.25, DROP_LIFE);
        float cycleID = floor((x_Time + fi * DROP_LIFE * 0.25) / DROP_LIFE);
        float cyclePhase = phase / DROP_LIFE;
        float seed = fi * 17.3 + cycleID * 5.7;
        vec2 dropCenter = pr_hash2(seed) * 1.4 - 0.7;
        dropCenter.x *= x_WindowSize.x / x_WindowSize.y;

        float d = length(p - dropCenter);
        // Ripple front expands at WAVE_SPEED
        float rippleR = phase * WAVE_SPEED;

        // Only active within expanding front zone
        if (d < rippleR + 0.1 && d < 0.8) {
            // Wave: sin at frequency * (d - rippleR) — oscillating at the ripple front
            float wave = sin((d - rippleR) * WAVE_FREQ * 6.28318 / WAVE_SPEED);
            // Attenuate outside front (behind it is valid oscillation, ahead fades quickly)
            float envelope = exp(-pow(max(0.0, d - rippleR), 2.0) * 400.0);
            // Radial decay
            float decay = 1.0 / (0.2 + d * 4.0);
            // Time fade
            float timeFade = 1.0 - cyclePhase * 0.8;
            float contribution = wave * envelope * decay * timeFade;
            waveSum += contribution;

            // Main ring — brighter exactly at the ripple front
            float ringWidth = 0.005;
            float ringD = abs(d - rippleR);
            float ring = exp(-ringD * ringD / (ringWidth * ringWidth) * 4.0);
            // Drop splash flash at the very start
            if (cyclePhase < 0.08) {
                float splash = exp(-d * d * 400.0) * (1.0 - cyclePhase / 0.08);
                col += pr_pal(0.3) * splash;
            }
            // Wave ring color
            col += pr_pal(fract(fi * 0.1 + x_Time * 0.04)) * ring * decay * 0.4;
        }
    }

    // Sum of wave displacements → creates highlights at crests
    float crest = pow(max(0.0, waveSum * 0.5), 2.0);
    col += pr_pal(fract(x_Time * 0.07)) * crest * 0.4;
    col += vec3(0.9, 0.95, 1.0) * crest * 0.3;

    // Highlights at troughs (dark vignettes)
    float trough = pow(max(0.0, -waveSum * 0.3), 2.0);
    col -= vec3(0.03, 0.05, 0.08) * trough;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
