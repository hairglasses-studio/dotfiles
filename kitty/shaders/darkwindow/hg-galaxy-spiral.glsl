// Shader attribution: hairglasses (original)
// Technique inspired by: Shadertoy tdKSRz (community galaxy raymarch).
// License: MIT
// (Cyberpunk — showcase/heavy) — Raymarched spiral galaxy with dust lanes and neon core glow

const int   STEPS       = 48;
const float CORE_RADIUS = 0.08;
const float ARM_COUNT   = 2.0;      // spiral arm pairs
const float ARM_PITCH   = 0.45;     // radians/ln(r) — how tightly wound
const float INTENSITY   = 0.55;

vec3 gal_pal(float t) {
    vec3 core = vec3(1.00, 0.92, 0.55);  // hot-white core
    vec3 gold = vec3(0.96, 0.65, 0.25);
    vec3 mag  = vec3(0.90, 0.20, 0.55);
    vec3 vio  = vec3(0.40, 0.22, 0.92);
    vec3 cyan = vec3(0.10, 0.75, 0.95);
    if (t < 0.25) return mix(core, gold, t * 4.0);
    if (t < 0.5)  return mix(gold, mag, (t - 0.25) * 4.0);
    if (t < 0.75) return mix(mag, vio, (t - 0.5) * 4.0);
    return mix(vio, cyan, (t - 0.75) * 4.0);
}

float gal_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
}

float gal_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(
        mix(gal_hash(i), gal_hash(i + vec2(1,0)), u.x),
        mix(gal_hash(i + vec2(0,1)), gal_hash(i + vec2(1,1)), u.x),
        u.y);
}

// 5-octave FBM
float gal_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < 5; i++) {
        v += a * gal_noise(p);
        p = rot * p * 2.07 + 0.11;
        a *= 0.5;
    }
    return v;
}

// Spiral density — dust lanes follow a logarithmic spiral
float spiralDensity(vec2 p, float t) {
    float r = length(p);
    if (r < 0.001) return 1.0;
    float a = atan(p.y, p.x);
    // Logarithmic spiral arms: a - log(r)/pitch forms "rail", repeat ARM_COUNT times
    float arm = fract((a - log(r + 0.01) / ARM_PITCH + t * 0.1) * ARM_COUNT / 6.28318 + 0.5);
    arm = abs(arm - 0.5) * 2.0;                 // [0,1], 0 on arm rail
    // Arm sharpness falls off with radius — arms vanish near rim
    float armTightness = 0.4 + 0.4 * smoothstep(0.45, 0.05, r);
    float armMask = pow(1.0 - arm, 3.0 * armTightness);
    // Stars per pixel noise  — FBM modulated
    float noise = gal_fbm(p * 22.0 + t * 0.05);
    float stars = smoothstep(0.55, 0.8, noise);
    // Dust lanes: darker region offset from arm ridge
    float dust = smoothstep(0.88, 1.0, arm);
    // Disc falloff from center
    float disc = exp(-r * r * 8.0);
    return (armMask * 0.8 + stars * 0.4) * (1.0 - dust * 0.6) * disc;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Slight slow rotation to animate the whole galaxy
    float tilt = x_Time * 0.03;
    float cr = cos(tilt), sr = sin(tilt);
    p = mat2(cr, -sr, sr, cr) * p;

    // Flatten along y to give slight edge-on perspective
    p.y /= 0.88;

    // Bright core
    float r = length(p);
    float coreMask = exp(-r * r / (CORE_RADIUS * CORE_RADIUS) * 3.0);
    vec3 core = gal_pal(0.0) * coreMask * 1.6;

    // Accumulate radial layers of spiral density
    vec3 col = core;
    for (int i = 0; i < STEPS; i++) {
        float fi = float(i) / float(STEPS);
        // Blur arm at different scales for depth
        vec2 pp = p * (1.0 + fi * 0.02);
        float d = spiralDensity(pp, x_Time);
        vec3 c = gal_pal(clamp(r * 1.6 + fi * 0.05, 0.0, 0.99));
        col += c * d * 0.035;
    }

    // Outer halo
    float halo = exp(-r * r * 2.5) * 0.4;
    col += gal_pal(0.45) * halo;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.8, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
