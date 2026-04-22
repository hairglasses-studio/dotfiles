// Shader attribution: hairglasses (original)
// Technique inspired by: Shadertoy stB3zW volumetric god-rays + classic radial blur approach.
// License: MIT
// (Cyberpunk — showcase/heavy) — Volumetric god-rays through floating dust — radial sampling, 64 taps

const int   RAY_SAMPLES = 64;       // radial samples toward light source
const float DECAY       = 0.97;
const float EXPOSURE    = 0.32;
const float WEIGHT      = 0.4;
const float DENSITY     = 1.0;
const float INTENSITY   = 0.55;

vec3 gr_pal(float t) {
    vec3 gold = vec3(1.00, 0.75, 0.30);
    vec3 amber = vec3(0.96, 0.48, 0.20);
    vec3 mag = vec3(0.90, 0.20, 0.55);
    vec3 cyan = vec3(0.10, 0.80, 0.92);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(gold, amber, s);
    else if (s < 2.0) return mix(amber, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, cyan, s - 2.0);
    else              return mix(cyan, gold, s - 3.0);
}

float gr_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
}

float gr_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(
        mix(gr_hash(i), gr_hash(i + vec2(1,0)), u.x),
        mix(gr_hash(i + vec2(0,1)), gr_hash(i + vec2(1,1)), u.x),
        u.y);
}

// 4-octave dust field
float gr_dust(vec2 p, float t) {
    float v = 0.0, a = 0.5;
    p += vec2(t * 0.1, t * 0.07);
    for (int i = 0; i < 4; i++) {
        v += a * gr_noise(p);
        p = p * 2.03 + 0.17;
        a *= 0.5;
    }
    // Bias toward dense filaments
    return pow(v, 1.6);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    // Source position: moves slowly across the top of the screen in a figure-8
    float lightT = x_Time * 0.15;
    vec2 lightUV = vec2(0.5 + 0.28 * sin(lightT), 0.18 + 0.06 * sin(lightT * 2.0));

    // Radial sampling toward light source — god rays are samples of the dust field
    // weighted by distance to the light, attenuated per step
    vec2 delta = (uv - lightUV);
    delta *= 1.0 / float(RAY_SAMPLES) * DENSITY;

    vec2 coord = uv;
    float illumDecay = 1.0;
    float sumDust = 0.0;
    for (int i = 0; i < RAY_SAMPLES; i++) {
        coord -= delta;
        // Scale sample coords to shader space with slight aspect correction
        vec2 sp = (coord - 0.5) * vec2(x_WindowSize.x / x_WindowSize.y, 1.0);
        float s = gr_dust(sp * 4.0, x_Time) * WEIGHT;
        s *= illumDecay;
        sumDust += s;
        illumDecay *= DECAY;
    }

    // Light source core + halo
    float ld = length(uv - lightUV);
    float core = exp(-ld * ld * 100.0) * 1.2;
    float halo = exp(-ld * 5.0) * 0.3;

    // Color ray by source-palette + drifting
    vec3 rayColor = gr_pal(fract(x_Time * 0.05));
    vec3 effect = rayColor * (sumDust * EXPOSURE + core + halo);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(effect), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, effect, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
