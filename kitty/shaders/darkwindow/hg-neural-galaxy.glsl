// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Neural galaxy hybrid — spiral arms of star-neurons with pulse-traveling synapses

const int   STARS      = 32;
const float SPIRAL_PITCH = 0.55;
const float ARM_WIDTH  = 0.1;
const float INTENSITY  = 0.55;

vec3 ng_pal(float t) {
    vec3 hot  = vec3(1.00, 0.92, 0.55);
    vec3 mag  = vec3(0.90, 0.25, 0.60);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 cyan = vec3(0.15, 0.85, 0.95);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(hot, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, cyan, s - 2.0);
    else              return mix(cyan, hot, s - 3.0);
}

float ng_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// A star-neuron position: anchored to a logarithmic spiral arm, slight drift
vec2 starPos(int i, float t) {
    float fi = float(i);
    // Pick one of 2 spiral arms
    float armIdx = mod(fi, 2.0);
    // Parameterize along arm: base radius grows with i
    float armT = fract(fi * 0.073) * 2.0;                  // [0, 2]
    float baseR = 0.05 + armT * 0.3;
    // Spiral angle at this radius (logarithmic)
    float angle = armIdx * 3.14159 + log(baseR) / SPIRAL_PITCH + t * 0.15;
    // Scatter perpendicular to arm
    float scatter = (ng_hash(fi * 3.1) - 0.5) * ARM_WIDTH;
    vec2 center = vec2(cos(angle), sin(angle)) * baseR;
    vec2 perpDir = vec2(-sin(angle), cos(angle));
    center += perpDir * scatter;
    // Slight drift
    center += 0.01 * vec2(sin(t * 0.3 + fi), cos(t * 0.25 + fi * 1.3));
    return center;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    // Bright galactic core
    float r = length(p);
    float coreMask = exp(-r * r * 200.0) * 1.2;
    col += ng_pal(0.0) * coreMask;
    // Core halo
    col += ng_pal(0.05) * exp(-r * r * 20.0) * 0.3;

    // Star-neurons
    vec2 starsPos[32];
    for (int i = 0; i < STARS; i++) {
        starsPos[i] = starPos(i, x_Time);
        float d = length(p - starsPos[i]);
        float seed = float(i) * 3.1;
        float size = 0.006 + 0.004 * ng_hash(seed);
        float core = exp(-d * d / (size * size) * 2.0);
        float halo = exp(-d * d * 1200.0) * 0.3;
        // Firing pulse — periodic brightness peak
        float pulse = pow(fract(x_Time * 0.5 + ng_hash(seed * 3.7) * 5.0), 8.0);
        vec3 sc = ng_pal(fract(seed * 0.05 + x_Time * 0.04));
        col += sc * (core * (0.7 + pulse * 1.5) + halo);
    }

    // Synapse connections — for stars close in both space and "sequence" (i, i+n)
    for (int a = 0; a < STARS; a++) {
        for (int off = 1; off <= 3; off++) {
            int b = (a + off) % STARS;
            vec2 dp = starsPos[b] - starsPos[a];
            float distance = length(dp);
            if (distance > 0.35 || distance < 0.03) continue;
            // Segment distance
            vec2 pa = p - starsPos[a];
            float h = clamp(dot(pa, dp) / dot(dp, dp), 0.0, 1.0);
            float d = length(pa - dp * h);
            float edgeMask = exp(-d * d * 6000.0);
            // Traveling pulse on this synapse
            float pulsePhase = fract(x_Time * 0.8 + float(a) * 0.1 + float(off) * 0.3);
            float pulseDist = abs(h - pulsePhase);
            float pulse = exp(-pulseDist * pulseDist * 120.0);
            vec3 synCol = ng_pal(fract(float(a) * 0.02 + x_Time * 0.03));
            col += synCol * edgeMask * 0.25;
            col += vec3(1.0, 0.95, 0.9) * pulse * edgeMask * 0.8;
        }
    }

    // Subtle dust lanes
    float dust = sin(r * 30.0 - atan(p.y, p.x) * 3.0 + x_Time * 0.1) * 0.5 + 0.5;
    col *= 0.7 + dust * 0.3;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
