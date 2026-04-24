// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Oil film — iridescent thin-film interference rainbow on oil-slicked water. Film thickness varies spatially via FBM; interference hue = fract(thickness·k + time); ripple-driven surface displacement; rainbow highlight on wave crests.

const int   FBM_OCT = 5;
const float INTENSITY = 0.55;

// HSV to RGB conversion for rainbow hue output
vec3 of_hsv(float h, float s, float v) {
    float c = v * s;
    float hp = fract(h) * 6.0;
    float x = c * (1.0 - abs(mod(hp, 2.0) - 1.0));
    vec3 rgb;
    if      (hp < 1.0) rgb = vec3(c, x, 0.0);
    else if (hp < 2.0) rgb = vec3(x, c, 0.0);
    else if (hp < 3.0) rgb = vec3(0.0, c, x);
    else if (hp < 4.0) rgb = vec3(0.0, x, c);
    else if (hp < 5.0) rgb = vec3(x, 0.0, c);
    else               rgb = vec3(c, 0.0, x);
    return rgb + (v - c);
}

float of_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float of_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(of_hash(i), of_hash(i + vec2(1, 0)), u.x),
               mix(of_hash(i + vec2(0, 1)), of_hash(i + vec2(1, 1)), u.x), u.y);
}

float of_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < FBM_OCT; i++) {
        v += a * of_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Animated ripple distortion of sample position
    vec2 rippleDist = vec2(
        sin(p.y * 12.0 + x_Time * 0.6) * 0.015,
        cos(p.x * 15.0 - x_Time * 0.5) * 0.015
    );
    vec2 sampleP = p + rippleDist;

    // Film thickness — FBM field, slowly evolving
    float thickness = of_fbm(sampleP * 2.5 + vec2(x_Time * 0.04, 0.0));
    // Sharpen thickness variation
    thickness = pow(thickness, 0.8);

    // Secondary finer thickness variation
    float fineThick = of_fbm(sampleP * 8.0 + vec2(0.0, x_Time * 0.02));
    float filmT = thickness * 0.7 + fineThick * 0.3;

    // Thin-film interference hue: hue cycles ~N times per unit thickness
    // Multiply by k=5 for ~5 rainbow bands across the range
    float hue = fract(filmT * 5.0 + x_Time * 0.05);
    float sat = 0.85 - fineThick * 0.2;
    float val = 0.7 + fineThick * 0.3;

    vec3 filmCol = of_hsv(hue, sat, val);

    // Darker base where no oil (low thickness)
    float coverage = smoothstep(0.1, 0.4, thickness);
    vec3 darkWater = vec3(0.03, 0.05, 0.08);

    vec3 col = mix(darkWater, filmCol, coverage);

    // Wave crest highlights (grad of the ripple field)
    float wave = sin(p.y * 12.0 + x_Time * 0.6) * cos(p.x * 15.0 - x_Time * 0.5);
    float crest = smoothstep(0.7, 0.9, abs(wave));
    col += filmCol * crest * 0.35;

    // Bright specular sparkle at high-frequency noise maxima
    float sparkle = smoothstep(0.85, 0.95, fineThick);
    col += vec3(1.0, 0.98, 0.90) * sparkle * 0.3;

    // Darker edges of the oil patch (variation in coverage)
    col *= 0.85 + coverage * 0.3;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
