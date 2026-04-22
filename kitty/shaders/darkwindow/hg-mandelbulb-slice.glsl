// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Animated 2D slice through 3D Mandelbulb — power-8 iteration with neon coloring

const int   MAX_ITER = 8;
const float POWER    = 8.0;
const float BAILOUT  = 2.0;
const float INTENSITY = 0.55;

vec3 mbu_pal(float t) {
    vec3 a = vec3(0.02, 0.02, 0.15);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.25, 0.60);
    vec3 d = vec3(0.96, 0.85, 0.40);
    vec3 e = vec3(0.10, 0.82, 0.92);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

// Mandelbulb DE (distance estimator). Computes escape iteration + smooth color.
float mandelbulbIter(vec3 pos) {
    vec3 z = pos;
    float dr = 1.0;
    float r = 0.0;
    float iter = 0.0;
    for (int i = 0; i < MAX_ITER; i++) {
        r = length(z);
        if (r > BAILOUT) break;
        float theta = acos(clamp(z.z / r, -1.0, 1.0));
        float phi = atan(z.y, z.x);
        dr = pow(r, POWER - 1.0) * POWER * dr + 1.0;
        float zr = pow(r, POWER);
        theta *= POWER;
        phi *= POWER;
        z = zr * vec3(sin(theta) * cos(phi), sin(phi) * sin(theta), cos(theta));
        z += pos;
        iter += 1.0;
    }
    return iter;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y * 1.8;

    // Slow rotation + z-animation (slice moves through 3D bulb)
    float t = x_Time * 0.15;
    float cr = cos(t), sr = sin(t);
    p = mat2(cr, -sr, sr, cr) * p;
    float z = 0.4 * sin(x_Time * 0.1);

    // Sample bulb at this 3D point
    float iter = mandelbulbIter(vec3(p, z));

    vec3 col;
    if (iter >= float(MAX_ITER)) {
        // Inside
        col = mbu_pal(0.0) * 0.9;
    } else {
        // Escape-time coloring
        float smoothIter = iter / float(MAX_ITER);
        col = mbu_pal(fract(smoothIter * 1.5 + x_Time * 0.05));
        // Edge highlight
        float edgeBoost = smoothstep(0.2, 0.8, smoothIter);
        col += vec3(0.9, 0.95, 1.0) * (1.0 - edgeBoost) * 0.1;
    }

    // Vignette
    float r = length(p);
    col *= smoothstep(1.8, 0.2, r);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.85, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
