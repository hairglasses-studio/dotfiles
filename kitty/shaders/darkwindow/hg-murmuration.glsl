// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Murmuration — 150 starlings flocking in coherent parametric motion (mimicking boids cohesion/alignment) against a dusk sky gradient. Flock center drifts, birds oscillate with multiple harmonics, density sculpts visible shapes.

const int   BIRDS = 150;
const int   FBM_OCT = 3;
const float INTENSITY = 0.55;

vec3 mur_pal(float t) {
    vec3 dusk1 = vec3(0.65, 0.25, 0.38);
    vec3 dusk2 = vec3(1.00, 0.55, 0.30);
    vec3 dusk3 = vec3(1.00, 0.80, 0.55);
    vec3 dusk4 = vec3(0.25, 0.15, 0.40);
    float s = mod(t * 3.0, 3.0);
    if (s < 1.0)      return mix(dusk4, dusk1, s);
    else if (s < 2.0) return mix(dusk1, dusk2, s - 1.0);
    else              return mix(dusk2, dusk3, s - 2.0);
}

float mur_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float mur_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float mur_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(mur_hash2(i), mur_hash2(i + vec2(1, 0)), u.x),
               mix(mur_hash2(i + vec2(0, 1)), mur_hash2(i + vec2(1, 1)), u.x), u.y);
}

// Flock center drifts in a figure-8-ish path
vec2 flockCenter(float t) {
    return vec2(0.4 * sin(t * 0.25),
                0.18 * sin(t * 0.37 + 0.7));
}

// Birds's individual position = flock center + individual orbital offset
vec2 birdPos(int i, float t) {
    float fi = float(i);
    float seed = fi * 7.31;
    vec2 center = flockCenter(t);
    // Individual orbital radius + angular phase + different speeds → coherent swarm
    float baseR = 0.1 + mur_hash(seed) * 0.3;
    float ang1 = mur_hash(seed * 3.7) * 6.28 + t * (0.5 + mur_hash(seed * 5.1) * 0.7);
    float ang2 = t * (1.3 + mur_hash(seed * 7.3) * 0.5) + seed;

    vec2 offset1 = vec2(cos(ang1), sin(ang1)) * baseR;
    vec2 offset2 = vec2(cos(ang2), sin(ang2)) * 0.04;

    // Coherent turning: every few seconds the whole flock pivots
    float pivotPhase = sin(t * 0.15) * 0.5;
    float c = cos(pivotPhase), s = sin(pivotPhase);
    offset1 = vec2(offset1.x * c - offset1.y * s, offset1.x * s + offset1.y * c);

    return center + offset1 + offset2;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Dusk sky: gradient from top (violet) to horizon (warm amber) with clouds
    float skyT = (p.y + 0.9) / 1.8;
    vec3 skyBase = mix(vec3(0.08, 0.05, 0.18),
                       vec3(0.85, 0.45, 0.30),
                       clamp(skyT, 0.0, 1.0));
    // Subtle cloud FBM
    float cloud = mur_noise(p * 2.0 + vec2(x_Time * 0.05, 0.0)) * 0.4;
    cloud += mur_noise(p * 5.0 + 2.7) * 0.25;
    skyBase += mur_pal(fract(cloud + 0.5)) * cloud * 0.25;
    // Horizon sun glow
    float horY = -0.35;
    float sunGlow = exp(-pow(p.y - horY, 2.0) * 20.0) * 0.6;
    skyBase += vec3(1.0, 0.6, 0.30) * sunGlow;
    vec3 col = skyBase;

    // === Bird positions ===
    // Accumulate darkness contribution per pixel (each nearby bird darkens)
    float darknessAccum = 0.0;
    for (int i = 0; i < BIRDS; i++) {
        vec2 bp = birdPos(i, x_Time);
        float d = length(p - bp);
        float birdSize = 0.004 + mur_hash(float(i) * 11.0) * 0.003;
        // Bird is a small dark dot
        float core = exp(-d * d / (birdSize * birdSize) * 1.5);
        darknessAccum += core;
        // Bird halo (local density sculpts)
        float halo = exp(-d * d * 2500.0) * 0.15;
        darknessAccum += halo;
    }
    // Darken the sky where birds are
    darknessAccum = clamp(darknessAccum, 0.0, 0.95);
    col = mix(col, vec3(0.02, 0.02, 0.05), darknessAccum);

    // Faint silhouette of hills at bottom
    if (p.y < -0.38) {
        float hills = mur_noise(vec2(p.x * 3.0, 0.0)) * 0.05;
        float hillLine = -0.38 - hills;
        if (p.y < hillLine) {
            col = mix(col, vec3(0.05, 0.03, 0.08), smoothstep(0.0, -0.05, p.y - hillLine));
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
