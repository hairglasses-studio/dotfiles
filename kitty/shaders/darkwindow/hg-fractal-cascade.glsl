// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Domain-warped FBM cascade — 3 nested levels of warping for organic fluid pattern

const int   OCTAVES    = 7;
const float WARP_A     = 3.5;
const float WARP_B     = 2.8;
const float INTENSITY  = 0.55;

vec3 fc_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.96, 0.70, 0.25); // gold
    vec3 d = vec3(0.90, 0.25, 0.60); // magenta
    vec3 e = vec3(0.20, 0.95, 0.60); // mint
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

float fc_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5);
}

float fc_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(fc_hash(i), fc_hash(i + vec2(1,0)), u.x),
               mix(fc_hash(i + vec2(0,1)), fc_hash(i + vec2(1,1)), u.x), u.y);
}

float fc_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < OCTAVES; i++) {
        v += a * fc_noise(p);
        p = rot * p * 2.09 + 0.17;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = x_PixelPos / x_WindowSize.y;
    float t = x_Time * 0.07;

    // Level 1 warp
    vec2 q = vec2(
        fc_fbm(p + vec2(0.0, t * 2.0)),
        fc_fbm(p + vec2(5.2, 1.3 + t))
    );
    // Level 2 warp — uses level 1 as offset
    vec2 r = vec2(
        fc_fbm(p + WARP_A * q + vec2(1.7, -t)),
        fc_fbm(p + WARP_A * q + vec2(8.3, t * 0.8))
    );
    // Level 3 warp — uses level 2
    vec2 s2 = vec2(
        fc_fbm(p + WARP_B * r + vec2(-t * 0.5, 4.2)),
        fc_fbm(p + WARP_B * r + vec2(2.1, -t * 0.3))
    );
    // Final density
    float d = fc_fbm(p + WARP_A * r + WARP_B * s2);

    // Two-color blending — vorticity-like magnitudes
    float q_mag = length(q - 0.5);
    float r_mag = length(r - 0.5);

    vec3 base = fc_pal(fract(d * 1.5 + t * 3.0));
    vec3 layer2 = fc_pal(fract(q_mag * 2.0 + t * 2.0 + 0.3));
    vec3 layer3 = fc_pal(fract(r_mag * 3.0 + t * 1.5 + 0.6));

    vec3 col = base * d;
    col = mix(col, layer2, q_mag * 0.8);
    col = mix(col, layer3, r_mag * 0.6);

    // Sharpen
    col = pow(col, vec3(1.3));

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility * 0.9);

    _wShaderOut = vec4(result, 1.0);
}
