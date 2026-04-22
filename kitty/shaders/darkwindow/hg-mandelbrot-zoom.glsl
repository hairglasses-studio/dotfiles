// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Animated infinite Mandelbrot zoom, smooth escape-time coloring, 256 max iterations

const int   MAX_ITER = 256;
const float ESCAPE   = 16.0;
const float INTENSITY = 0.55;

// Target coordinates with interesting local structure
// Classic "elephant valley" / seahorse area
const vec2 ZOOM_TARGET = vec2(-0.7436438870, 0.1318259042);

vec3 mz_pal(float t) {
    vec3 a = vec3(0.02, 0.02, 0.15);   // deep blue (inside set)
    vec3 b = vec3(0.10, 0.80, 0.92);   // cyan
    vec3 c = vec3(0.55, 0.30, 0.98);   // violet
    vec3 d = vec3(0.90, 0.20, 0.55);   // magenta
    vec3 e = vec3(0.96, 0.85, 0.40);   // gold edge
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

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Animated zoom: linear zoom-in then reset, periodic
    float zoomCycle = 30.0;  // seconds per full zoom cycle
    float zoomPhase = mod(x_Time, zoomCycle) / zoomCycle;
    float zoomLevel = exp(zoomPhase * 8.0);  // 1x → ~3000x over 30s

    vec2 c = ZOOM_TARGET + p / zoomLevel * 2.0;

    vec2 z = vec2(0.0);
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
        // Inside set — deep blue with slow internal pulse
        col = mz_pal(0.0) * (0.8 + 0.2 * sin(x_Time * 0.5));
    } else {
        // Smooth escape-time coloring (continuous iteration count)
        float smoothIter = iter + 1.0 - log2(log2(zr2) * 0.5);
        float t = fract(smoothIter * 0.02 + x_Time * 0.04);
        col = mz_pal(t);
        // Bright halo at escape edge
        float edgeGlow = smoothstep(10.0, 3.0, abs(iter - float(MAX_ITER) + 1.0));
        col += vec3(0.9, 0.95, 1.0) * edgeGlow * 0.1;
    }

    // Slow zoom-distance vignette so deep levels don't look flat
    float vignette = 1.0 - length(p) * 0.25;
    col *= vignette;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.85, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
