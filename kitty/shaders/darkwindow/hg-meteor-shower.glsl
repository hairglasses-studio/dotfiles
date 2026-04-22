// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Meteor shower — streaking meteors with motion-blurred trails, flash burnouts, starfield

const int   METEORS   = 24;
const int   TRAIL_TAPS = 12;
const float INTENSITY = 0.55;

vec3 ms_pal(float t) {
    vec3 gold = vec3(1.00, 0.85, 0.40);
    vec3 orange = vec3(1.00, 0.55, 0.25);
    vec3 mag = vec3(0.95, 0.25, 0.60);
    vec3 cyan = vec3(0.20, 0.90, 0.95);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(gold, orange, s);
    else if (s < 2.0) return mix(orange, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, cyan, s - 2.0);
    else              return mix(cyan, gold, s - 3.0);
}

float ms_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  ms_hash2(float n) { return vec2(ms_hash(n), ms_hash(n * 1.37 + 11.0)); }

// Meteor position at time t, given seed. Life ∈ [0,1].
// Returns (position, life, angle, speed)
vec4 meteorState(int i, float t) {
    float fi = float(i);
    float seed = fi * 7.31;
    float lifeCycle = 2.5;
    float phase = fract((t + ms_hash(seed) * lifeCycle) / lifeCycle);
    float cycleID = floor((t + ms_hash(seed) * lifeCycle) / lifeCycle);
    float cycleSeed = seed + cycleID * 13.7;

    // Start position (top edge or random)
    vec2 startPos = ms_hash2(cycleSeed) * 2.0 - 1.0;
    startPos.y += 1.2;  // bias to top
    startPos.x *= x_WindowSize.x / x_WindowSize.y;

    // Direction — mostly downward-diagonal
    float angle = -1.5708 + (ms_hash(cycleSeed * 3.7) - 0.5) * 1.4;
    float speed = 1.0 + ms_hash(cycleSeed * 5.3) * 0.8;

    // Current position
    vec2 pos = startPos + vec2(cos(angle), sin(angle)) * phase * speed;

    return vec4(pos, phase, angle);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Night sky backdrop
    vec3 bg = mix(vec3(0.02, 0.04, 0.12), vec3(0.01, 0.02, 0.06), uv.y);
    vec3 col = bg;

    // Sparse starfield
    vec2 sg = floor(p * 100.0);
    float sh = ms_hash(sg.x * 31.0 + sg.y);
    if (sh > 0.993) {
        float tw = 0.4 + 0.6 * sin(x_Time * (1.5 + sh * 3.0) + sh * 20.0);
        col += vec3(0.85, 0.9, 1.0) * (sh - 0.993) * 150.0 * tw;
    }

    // Meteors
    for (int i = 0; i < METEORS; i++) {
        vec4 state = meteorState(i, x_Time);
        vec2 headPos = state.xy;
        float phase = state.z;
        float angle = state.w;

        vec2 dir = vec2(cos(angle), sin(angle));

        // Color for this meteor
        float seed = float(i) * 7.31 + floor((x_Time + ms_hash(float(i) * 7.31) * 2.5) / 2.5) * 13.7;
        vec3 mc = ms_pal(fract(ms_hash(seed * 1.1) + x_Time * 0.05));

        // Render motion-blurred trail — multiple samples along the path
        for (int k = 0; k < TRAIL_TAPS; k++) {
            float fk = float(k);
            float tailT = fk / float(TRAIL_TAPS);
            vec2 tailPos = headPos - dir * tailT * 0.15;
            float d = length(p - tailPos);
            float coreSize = 0.002 * (1.0 - tailT * 0.7);
            float core = exp(-d * d / (coreSize * coreSize) * 2.0);
            float fade = pow(1.0 - tailT, 2.0);
            // Near-burnout flash at start of meteor life (head)
            if (k == 0 && phase < 0.1) {
                float flash = exp(-d * d * 800.0) * (1.0 - phase / 0.1);
                col += vec3(1.0, 0.95, 0.85) * flash * 0.8;
            }
            col += mc * core * fade * 0.6;
            col += mc * exp(-d * d * 500.0) * fade * 0.15;
        }

        // Near-burnout: meteor bursts at end of life
        if (phase > 0.85) {
            float burstPhase = (phase - 0.85) / 0.15;
            float burstD = length(p - headPos);
            float burst = exp(-burstD * burstD / (0.01 * 0.01) * 2.0) * (1.0 - burstPhase) * (1.0 - burstPhase);
            col += mc * burst * 2.0;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
