// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk) — Classic plasma field with cycling neon palette

// Classic sum-of-sines plasma.
// Palette cycles through a deep-neon wheel: magenta → cyan → violet → gold.

const float INTENSITY = 0.35;   // overall opacity over terminal
const float SPEED     = 0.35;   // palette + motion speed

vec3 pal(float t) {
    // 5-stop neon cycle. t in [0,1).
    vec3 a = vec3(0.90, 0.18, 0.60); // magenta
    vec3 b = vec3(0.10, 0.80, 0.92); // cyan
    vec3 c = vec3(0.55, 0.30, 0.98); // violet
    vec3 d = vec3(0.96, 0.70, 0.25); // gold
    vec3 e = vec3(0.20, 0.95, 0.60); // mint
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

    // Aspect-corrected centered coords
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float t = x_Time * SPEED;

    // Four interfering waves
    float v = 0.0;
    v += sin(p.x * 8.0 + t * 1.3);
    v += sin(p.y * 9.0 + t * 1.7);
    v += sin((p.x + p.y) * 6.0 + t * 1.1);
    v += sin(length(p) * 14.0 - t * 2.3);
    v = v * 0.25 + 0.5; // normalize to roughly [0,1]

    // Slow palette drift
    vec3 plasma = pal(fract(v + t * 0.07));

    // Composite — keep text readable
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.7);
    vec3 result = mix(terminal.rgb, plasma, visibility);

    _wShaderOut = vec4(result, 1.0);
}
