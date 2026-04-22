// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Wave pool — top-down animated water surface with sun glitter + splash highlights

const int   WAVE_DIRS = 5;
const float INTENSITY = 0.55;

vec3 wp_pal(float t) {
    vec3 shallow = vec3(0.25, 0.85, 0.95);
    vec3 mid     = vec3(0.10, 0.55, 0.85);
    vec3 deep    = vec3(0.03, 0.15, 0.45);
    vec3 sun     = vec3(1.00, 0.90, 0.45);
    vec3 mag     = vec3(0.90, 0.30, 0.60);
    if (t < 0.25)      return mix(deep, mid, t * 4.0);
    else if (t < 0.5)  return mix(mid, shallow, (t - 0.25) * 4.0);
    else if (t < 0.75) return mix(shallow, sun, (t - 0.5) * 4.0);
    else               return mix(sun, mag, (t - 0.75) * 4.0);
}

float wp_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// Sum-of-directional-sines wave
float wpHeight(vec2 p, float t) {
    float h = 0.0;
    float amp = 1.0;
    float freq = 1.8;
    float ang = 0.0;
    for (int i = 0; i < WAVE_DIRS; i++) {
        vec2 d = vec2(cos(ang), sin(ang));
        h += amp * sin(dot(p, d) * freq + t * freq * 0.7);
        amp *= 0.7;
        freq *= 1.6;
        ang += 1.4;
    }
    return h * 0.3;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y * 3.0;

    float h = wpHeight(p, x_Time);
    // Approximate normal from height
    float hx = wpHeight(p + vec2(0.01, 0.0), x_Time);
    float hy = wpHeight(p + vec2(0.0, 0.01), x_Time);
    vec3 normal = normalize(vec3((h - hx) * 80.0, (h - hy) * 80.0, 1.0));

    // Base color by height
    float heightT = 0.3 + h * 0.5;
    vec3 col = wp_pal(clamp(heightT, 0.0, 1.0));

    // Sun glitter — specular spots where normal points at "sun direction"
    vec3 sunDir = normalize(vec3(0.5, 0.5, 0.7));
    float spec = pow(max(0.0, dot(normal, sunDir)), 64.0);
    col += vec3(1.0, 0.95, 0.8) * spec * 0.9;

    // Secondary moon/neon light
    vec3 moonDir = normalize(vec3(-0.4, 0.2, 0.6));
    float spec2 = pow(max(0.0, dot(normal, moonDir)), 32.0);
    col += wp_pal(fract(x_Time * 0.05 + 0.3)) * spec2 * 0.4;

    // Foam / splash — sparse bright dots where height is at local max
    float foamThresh = 0.25;
    if (h > foamThresh) {
        float foam = pow((h - foamThresh) / 0.1, 2.0);
        col = mix(col, vec3(0.9, 0.95, 1.0), min(foam, 0.6));
    }

    // Wave crest darken
    float crestVal = pow(max(0.0, h * 2.0), 3.0);
    col += wp_pal(0.7) * crestVal * 0.1;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
