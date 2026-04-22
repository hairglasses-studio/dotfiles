// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk) — Fluid nebula — domain-warped FBM drifting through neon palette

const float INTENSITY = 0.42;
const float WARP_AMOUNT = 2.0;
const float FLOW_SPEED  = 0.08;

float nb_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453123);
}

float nb_vnoise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(
        mix(nb_hash(i), nb_hash(i + vec2(1.0, 0.0)), u.x),
        mix(nb_hash(i + vec2(0.0, 1.0)), nb_hash(i + vec2(1.0, 1.0)), u.x),
        u.y
    );
}

float nb_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < 5; i++) {
        v += a * nb_vnoise(p);
        p = rot * p * 2.0 + 0.13;
        a *= 0.5;
    }
    return v;
}

vec3 nbPalette(float t) {
    vec3 a = vec3(0.08, 0.02, 0.20); // deep purple
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.90, 0.18, 0.60); // magenta
    vec3 d = vec3(0.96, 0.70, 0.25); // gold highlight
    vec3 e = vec3(0.10, 0.82, 0.92); // cyan edge
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = x_PixelPos / x_WindowSize.y;
    float t = x_Time * FLOW_SPEED;

    // Domain warp: sample FBM at an offset that is itself FBM
    vec2 q = vec2(
        nb_fbm(p + vec2(0.0, t)),
        nb_fbm(p + vec2(5.2, t * 0.8))
    );
    vec2 r = vec2(
        nb_fbm(p + WARP_AMOUNT * q + vec2(1.7, -t * 0.5)),
        nb_fbm(p + WARP_AMOUNT * q + vec2(8.3,  t * 0.4))
    );
    float density = nb_fbm(p + WARP_AMOUNT * r);

    // Density → palette stop + brightness
    vec3 nebula = nbPalette(fract(density + t * 2.0));
    float brightness = smoothstep(0.3, 0.8, density);

    // Bright filaments along high-gradient regions (cheap approximation)
    float filament = smoothstep(0.55, 0.65, density);
    nebula += vec3(0.9, 0.95, 1.0) * filament * 0.25;

    vec3 effect = nebula * brightness;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.7);
    vec3 result = mix(terminal.rgb, effect, visibility * brightness);

    _wShaderOut = vec4(result, 1.0);
}
