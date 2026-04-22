// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Liquid metal — reflective blob with chromatic reflections + flowing surface waves

const int   OCTAVES = 6;
const float INTENSITY = 0.55;

vec3 lm_pal(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(0.95, 0.25, 0.65);
    vec3 d = vec3(1.00, 0.85, 0.45);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float lm_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float lm_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(lm_hash(i), lm_hash(i + vec2(1,0)), u.x),
               mix(lm_hash(i + vec2(0,1)), lm_hash(i + vec2(1,1)), u.x), u.y);
}

float lm_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < OCTAVES; i++) {
        v += a * lm_noise(p);
        p = rot * p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    float t = x_Time * 0.2;

    // "Height" field — domain warped FBM
    vec2 q = vec2(lm_fbm(p + vec2(0.0, t)), lm_fbm(p + vec2(5.2, t)));
    float h = lm_fbm(p + q * 2.0 + t);

    // Normal approximation from height gradient
    float dx = lm_fbm(p + vec2(0.01, 0.0) + q * 2.0 + t) - h;
    float dy = lm_fbm(p + vec2(0.0, 0.01) + q * 2.0 + t) - h;
    vec3 normal = normalize(vec3(-dx * 100.0, -dy * 100.0, 1.0));

    // Simulated environment reflection — view dir is (0,0,-1), reflect around normal
    vec3 viewDir = vec3(0.0, 0.0, -1.0);
    vec3 reflDir = reflect(viewDir, normal);

    // Sample "environment" palette at reflection direction (cheap faux env map)
    // Encode as (angle, height)
    float reflAngle = atan(reflDir.y, reflDir.x) / 6.28 + 0.5;
    float reflHeight = reflDir.z * 0.5 + 0.5;

    // Chromatic reflections: offset palette sample per channel
    vec3 col;
    col.r = lm_pal(fract(reflAngle + 0.03 + t * 0.1)).r;
    col.g = lm_pal(fract(reflAngle + t * 0.1)).g;
    col.b = lm_pal(fract(reflAngle - 0.03 + t * 0.1)).b;

    // Modulate with height for depth
    col *= 0.4 + reflHeight * 0.7;

    // Specular highlights
    vec3 lightDir = normalize(vec3(0.5, 0.7, 0.5));
    float spec = pow(max(0.0, dot(normal, lightDir)), 48.0);
    col += vec3(1.0, 0.98, 0.9) * spec * 0.5;

    // Fresnel-like rim at horizontal edges
    float fresnel = pow(1.0 - normal.z, 2.0);
    col += lm_pal(fract(t * 0.3 + 0.5)) * fresnel * 0.3;

    // Soft vignette
    col *= 1.0 - length(p) * 0.2;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility * 0.9);

    _wShaderOut = vec4(result, 1.0);
}
