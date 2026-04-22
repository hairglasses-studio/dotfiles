// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Animated Julia set — c-parameter orbits through interesting values, smooth escape-time coloring

const int   MAX_ITER = 128;
const float ESCAPE   = 16.0;
const float INTENSITY = 0.55;

vec3 ju_pal(float t) {
    vec3 a = vec3(0.02, 0.02, 0.20);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.25, 0.60);
    vec3 d = vec3(0.20, 0.95, 0.60);
    vec3 e = vec3(0.10, 0.82, 0.92);
    vec3 f = vec3(0.96, 0.85, 0.45);
    float s = mod(t * 6.0, 6.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else if (s < 5.0) return mix(e, f, s - 4.0);
    else              return mix(f, a, s - 5.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y * 2.0;

    // c-parameter orbits a hypotrochoid-like path through interesting Julia parameter space
    float phi = x_Time * 0.15;
    vec2 c = vec2(
        0.7885 * cos(phi * 1.0 + sin(phi * 0.3) * 0.4),
        0.7885 * sin(phi * 1.0 + sin(phi * 0.3) * 0.4)
    );
    // Add smaller wobble for more variety
    c += 0.08 * vec2(cos(phi * 2.3), sin(phi * 3.1));

    // Julia iteration: z_{n+1} = z^2 + c, starting with z = p
    vec2 z = p;
    float iter = 0.0;
    float zr2 = 0.0;
    for (int i = 0; i < MAX_ITER; i++) {
        z = vec2(z.x * z.x - z.y * z.y, 2.0 * z.x * z.y) + c;
        zr2 = dot(z, z);
        if (zr2 > ESCAPE) break;
        iter += 1.0;
    }

    vec3 col;
    if (iter >= float(MAX_ITER)) {
        // Inside set — slow-pulsing dark violet
        col = ju_pal(0.0) * (0.8 + 0.2 * sin(x_Time * 0.5));
    } else {
        // Smooth escape-time coloring
        float smoothIter = iter + 1.0 - log2(log2(zr2) * 0.5);
        float t = fract(smoothIter * 0.04 + x_Time * 0.05);
        col = ju_pal(t);
        // Extra bright at the escape edge
        float edgeBoost = smoothstep(5.0, 0.0, float(MAX_ITER) - iter);
        col += vec3(0.9, 0.95, 1.0) * edgeBoost * 0.15;
    }

    // Subtle vignette
    float r = length(p) * 0.5;
    col *= 1.0 - r * 0.15;

    // c-parameter marker: tiny cross where c currently lies (in same space as p)
    float cMark = exp(-length(p - c) * length(p - c) * 2000.0);
    col += vec3(1.0, 0.9, 0.6) * cMark * 0.8;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.85, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
