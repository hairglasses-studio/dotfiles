// Shader attribution: hairglasses (original)
// Technique inspired by: Kali's "Star Nest" (Shadertoy XlfGRj) — Kaliset fractal.
// License: MIT
// (Cyberpunk — showcase/heavy) — Kaliset fractal warp-starfield, 20-iter density, volumetric fog

const int   ITER         = 20;
const int   VOLSTEPS     = 16;
const float FRACT_SCALE  = 2.15;
const float FOLD_OFFSET  = -0.5;
const float BRIGHTNESS   = 0.0015;
const float DISTFADING   = 0.73;
const float SATURATION   = 0.82;
const float INTENSITY    = 0.55;

vec3 sn_palette(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.96, 0.70, 0.25); // gold
    vec3 d = vec3(0.90, 0.18, 0.60); // magenta
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    // Aspect-correct
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    vec3 dir = vec3(p, 1.0);

    // Camera drift through fractal space
    float time = x_Time * 0.07;
    vec3 from = vec3(1.0, 0.5, 0.5) + vec3(time * 2.0, time, -2.0);

    // Subtle rotation
    float a1 = 0.5 + sin(time * 0.3) * 0.15;
    float a2 = 0.8 + cos(time * 0.25) * 0.10;
    mat2 rot1 = mat2(cos(a1), sin(a1), -sin(a1), cos(a1));
    mat2 rot2 = mat2(cos(a2), sin(a2), -sin(a2), cos(a2));
    dir.xz = rot1 * dir.xz;
    dir.xy = rot2 * dir.xy;
    from.xz = rot1 * from.xz;
    from.xy = rot2 * from.xy;

    // Volumetric march — accumulate density at each step
    float s = 0.1, fade = 1.0;
    vec3 col = vec3(0.0);
    for (int r = 0; r < VOLSTEPS; r++) {
        vec3 q = from + s * dir * 0.5;
        q = abs(vec3(FRACT_SCALE) - mod(q, vec3(FRACT_SCALE * 2.0)));

        float pa = 0.0, a = 0.0;
        for (int i = 0; i < ITER; i++) {
            q = abs(q) / dot(q, q) + FOLD_OFFSET;
            float d = length(q);
            a += abs(d - pa);
            pa = d;
        }
        float dm = max(0.0, 1.0 - 0.0085 * a * a);
        if (r > 6) fade *= 1.0 - dm;

        // Color drifts with iteration + time
        vec3 layerCol = sn_palette(fract(float(r) * 0.055 + time * 0.9));
        col += fade * dm * layerCol;

        fade *= DISTFADING;
        s += 0.1;
    }

    col = mix(vec3(length(col)), col, SATURATION) * BRIGHTNESS * 350.0;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
