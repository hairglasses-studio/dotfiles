// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — <SHADER-GENRE> — <ONE-LINE-DESCRIPTION>

const int   OCTAVES = 5;              // tune per shader
const float INTENSITY = 0.55;         // terminal composite strength

// Per-shader palette — cycle through neon cyberpunk palette over time.
// Rename prefix (e.g. mg_ / pg_ / bc_) to avoid collisions with helpers in other shaders.
vec3 XX_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.90, 0.25, 0.65); // magenta
    vec3 d = vec3(0.96, 0.85, 0.40); // gold
    vec3 e = vec3(0.20, 0.95, 0.60); // mint
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

// Per-shader hash/noise helpers (always prefix XX_ to avoid collisions).
float XX_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float XX_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(XX_hash(i), XX_hash(i + vec2(1, 0)), u.x),
               mix(XX_hash(i + vec2(0, 1)), XX_hash(i + vec2(1, 1)), u.x), u.y);
}

float XX_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < OCTAVES; i++) {
        v += a * XX_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    // Standard setup: uv for terminal sampling, p for centered aspect-correct
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // === SHADER BODY ===
    // Build 3+ distinct visual layers:
    //   1. Background / base color
    //   2. Mid-layer structure (raymarched, fractal, particle field, etc.)
    //   3. Accent highlights (pulses, sparkles, rim glows, traveling features)
    //
    // Animation: reference x_Time for motion. Cursor-reactive elements: use
    // x_CursorPos. Do NOT reference Ghostty-specific uniforms like iPreviousCursor
    // or iTimeCursorChange — they don't exist in DarkWindow.

    vec3 col = vec3(0.0);

    // Example: palette-cycled radial gradient
    float r = length(p);
    col += XX_pal(fract(r + x_Time * 0.05)) * exp(-r * r * 8.0) * 0.6;

    // === COMPOSITE WITH TERMINAL (mandatory) ===
    // termLuma-modulated visibility keeps bright text legible.
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);        // or use a scene-specific mask
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
