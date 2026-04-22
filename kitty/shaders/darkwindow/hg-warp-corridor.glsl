// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Warp-speed light streaks through a neon corridor

const int   STAR_LAYERS  = 4;       // parallax layers of streaks
const int   STARS_PER    = 80;      // streaks per layer
const int   MOTION_BLUR  = 6;       // temporal samples per streak
const float WARP_SPEED   = 2.2;
const float CORRIDOR_FOV = 0.7;     // how steep the tunnel convergence feels
const float INTENSITY    = 0.55;

vec3 wc_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.90, 0.18, 0.60); // magenta
    vec3 c = vec3(0.55, 0.30, 0.98); // violet
    vec3 d = vec3(0.96, 0.70, 0.25); // gold
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float wc_hash(float n) { return fract(sin(n * 43.37) * 43758.5453); }
vec2  wc_hash2(float n) { return vec2(wc_hash(n), wc_hash(n * 1.37 + 11.0)); }

// Single streak: origin angle + radial speed, projected via 1/z perspective
float streak(vec2 p, float seed, float t, float speed, float layerDepth) {
    // Random spawn angle + phase
    vec2 h = wc_hash2(seed);
    float ang = h.x * 6.28318 + t * 0.05;
    vec2 dir = vec2(cos(ang), sin(ang));

    // Streak lifecycle: emerges from near-center, streams outward
    float phase = fract(h.y + t * speed);
    // Depth-dependent radius: small near center, big at exit
    float r = phase * phase;     // quadratic ease-out for warp feel

    // Motion-blur: accumulate MOTION_BLUR samples along the streak's path
    float accum = 0.0;
    float sampleLen = 0.025 * layerDepth;   // streak length in UV space
    for (int i = 0; i < MOTION_BLUR; i++) {
        float dt = float(i) / float(MOTION_BLUR);
        float rr = r - dt * sampleLen;
        if (rr < 0.0) continue;
        vec2 sp = dir * rr;
        float d = length(p - sp);
        // Thin streak kernel — sharper at head, softer at tail
        float thick = 0.003 * (0.3 + layerDepth * 0.7) * (1.0 - dt * 0.6);
        float k = exp(-d * d / (thick * thick));
        accum += k * (1.0 - dt * 0.7);
    }
    // Fade at both extremes of lifecycle
    float life = smoothstep(0.0, 0.08, phase) * (1.0 - smoothstep(0.85, 1.0, phase));
    return accum * life / float(MOTION_BLUR) * 4.0;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    // Centered + aspect-correct; scale so corridor extends to screen edge
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    p /= CORRIDOR_FOV;

    vec3 col = vec3(0.0);

    // Parallax layers — each layer at a different depth, warping at slightly
    // different speeds to give a volumetric feel
    for (int L = 0; L < STAR_LAYERS; L++) {
        float layer = float(L);
        float layerDepth = 0.4 + 0.2 * layer;           // closer layers lag behind
        float layerSpeed = WARP_SPEED * (0.6 + 0.2 * layer);
        float layerSeed  = 100.0 * (layer + 1.0);
        for (int s = 0; s < STARS_PER; s++) {
            float sd = float(s) * 3.137 + layerSeed;
            float k = streak(p, sd, x_Time, layerSpeed, layerDepth);
            if (k > 0.001) {
                vec3 sc = wc_pal(fract(sd * 0.017 + x_Time * 0.04));
                col += sc * k;
            }
        }
    }

    // Vignette + central tunnel haze
    float r = length(p);
    float vignette = smoothstep(1.5, 0.5, r);
    float haze = exp(-r * r * 3.0) * 0.25;
    col *= vignette;
    col += wc_pal(fract(x_Time * 0.03)) * haze;

    // Composite under text — text sits in the bright tunnel mouth, stays readable
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
