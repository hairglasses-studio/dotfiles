// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Fluid flow field — curl-noise velocity streamlines with color-cycling particles

const int   OCTAVES  = 5;
const int   PARTICLES = 96;
const int   TRAIL_TAPS = 6;
const float INTENSITY = 0.5;

vec3 ff_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.90, 0.25, 0.60); // magenta
    vec3 d = vec3(0.20, 0.95, 0.60); // mint
    vec3 e = vec3(0.96, 0.70, 0.25); // gold
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

float ff_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float ff_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(ff_hash(i), ff_hash(i + vec2(1,0)), u.x),
               mix(ff_hash(i + vec2(0,1)), ff_hash(i + vec2(1,1)), u.x), u.y);
}

float ff_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * ff_noise(p);
        p = p * 2.13 + 0.11;
        a *= 0.5;
    }
    return v;
}

// Curl of scalar noise = divergence-free 2D velocity
vec2 curlNoise(vec2 p, float t) {
    float eps = 0.01;
    float n1 = ff_fbm(p + vec2(0.0, eps) + vec2(t * 0.1, 0.0));
    float n2 = ff_fbm(p - vec2(0.0, eps) + vec2(t * 0.1, 0.0));
    float n3 = ff_fbm(p + vec2(eps, 0.0) + vec2(t * 0.1, 0.0));
    float n4 = ff_fbm(p - vec2(eps, 0.0) + vec2(t * 0.1, 0.0));
    return vec2((n1 - n2) / (2.0 * eps), -(n3 - n4) / (2.0 * eps));
}

// Integrate particle position backwards along the field (streamline trace)
vec2 traceBack(vec2 p, float t, float dt) {
    vec2 pos = p;
    for (int i = 0; i < TRAIL_TAPS; i++) {
        vec2 v = curlNoise(pos, t - float(i) * dt);
        pos -= v * dt * 0.3;
    }
    return pos;
}

float ff_hash1(float n) { return fract(sin(n * 127.1) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    // Background: sample velocity field, visualize vorticity
    vec2 vel = curlNoise(p * 1.5, x_Time);
    float vMag = length(vel);
    float vAng = atan(vel.y, vel.x) / 6.28318 + 0.5;

    // Subtle streamline overlay — use curl noise directly as background tint
    vec3 bgCol = ff_pal(fract(vAng + x_Time * 0.03)) * vMag * 0.15;
    col += bgCol;

    // Particle tracers — distributed uniformly, advected by flow
    for (int i = 0; i < PARTICLES; i++) {
        float fi = float(i);
        float seed = fi * 7.31;

        // Base grid position (reset every few seconds)
        float resetCycle = 6.0;
        float cyclePhase = mod(x_Time + ff_hash1(seed) * resetCycle, resetCycle);
        float cycleT = cyclePhase / resetCycle;    // [0,1]
        float cycleID = floor((x_Time + ff_hash1(seed) * resetCycle) / resetCycle);

        vec2 basePos = vec2(
            ff_hash1(seed + cycleID * 13.7) * 2.0 - 1.0,
            ff_hash1(seed + cycleID * 23.3) * 2.0 - 1.0
        );
        basePos.x *= x_WindowSize.x / x_WindowSize.y;

        // Advect particle forward by velocity * cycleT
        vec2 pos = basePos;
        for (int k = 0; k < 6; k++) {
            vec2 v = curlNoise(pos, x_Time - (1.0 - cycleT) * 0.5);
            pos += v * 0.1 * cycleT / 6.0;
        }

        // Draw particle + short trail behind it
        for (int tap = 0; tap < TRAIL_TAPS; tap++) {
            float ft = float(tap);
            vec2 samplePos = pos;
            float trailStep = 0.008 * ft;
            // Back-trace
            for (int k = 0; k < 2; k++) {
                vec2 v = curlNoise(samplePos, x_Time - ft * 0.02);
                samplePos -= v * trailStep * 0.5;
            }
            float d = length(p - samplePos);
            float fade = 1.0 - ft / float(TRAIL_TAPS);
            float size = 0.004 * fade;
            float k = exp(-d * d / (size * size) * 3.0);
            vec3 pCol = ff_pal(fract(seed * 0.01 + x_Time * 0.04));
            // Particles die at end of cycle
            float lifeFade = 1.0 - smoothstep(0.8, 1.0, cycleT);
            col += pCol * k * fade * 0.35 * lifeFade;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.8, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
