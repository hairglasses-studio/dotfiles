// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Black ice — cracked frozen surface with refractive bright crack-lights + sparkle

const int   SEEDS = 50;
const float INTENSITY = 0.5;

vec3 bi_pal(float t) {
    vec3 a = vec3(0.20, 0.85, 0.95);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.90, 1.00);
    vec3 d = vec3(0.95, 0.30, 0.70);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float bi_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  bi_hash2(float n) { return vec2(bi_hash(n), bi_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Voronoi — plate edges = cracks
    float d1 = 1e9, d2 = 1e9;
    int i1 = 0;
    for (int i = 0; i < SEEDS; i++) {
        vec2 sp = bi_hash2(float(i) * 3.71) * 2.0 - 1.0;
        sp.x *= x_WindowSize.x / x_WindowSize.y;
        sp += 0.008 * vec2(sin(x_Time * 0.05 + float(i)), cos(x_Time * 0.04 + float(i) * 1.3));
        float d = length(p - sp);
        if (d < d1) { d2 = d1; d1 = d; i1 = i; }
        else if (d < d2) d2 = d;
    }
    float edgeDist = d2 - d1;

    // Very dark base
    vec3 col = vec3(0.02, 0.03, 0.04);

    // Each plate has a slight hue variation (thin ice tinted)
    vec3 plateCol = bi_pal(fract(float(i1) * 0.07 + x_Time * 0.02));
    col += plateCol * 0.05;

    // Subtle refraction — sample terminal with perturbation based on plate ID
    vec2 refractOff = (bi_hash2(float(i1) * 5.3) - 0.5) * 0.004;
    vec4 terminal = x_Texture(uv + refractOff);

    // Cracks: bright glowing lines where plates meet
    float crackMask = smoothstep(0.015, 0.0, edgeDist);
    float crackGlow = exp(-edgeDist * edgeDist * 3000.0) * 0.4;

    // Crack color changes slowly
    vec3 crackCol = bi_pal(fract(float(i1) * 0.05 + x_Time * 0.04));
    col += crackCol * (crackMask * 1.4 + crackGlow);

    // Sparkles on ice (random bright points)
    vec2 sparkG = floor(p * 200.0);
    float sparkH = bi_hash(sparkG.x * 31.0 + sparkG.y + floor(x_Time * 3.0));
    if (sparkH > 0.995) {
        col += vec3(0.9, 0.95, 1.0) * (sparkH - 0.995) * 200.0 * 0.7;
    }

    // Light source — makes one side of ice brighter (sunrise/sunset glint)
    vec2 lightPos = vec2(0.4, 0.4);
    float lightDist = length(p - lightPos);
    float lightGlint = exp(-lightDist * lightDist * 3.0) * 0.15;
    col += vec3(0.95, 0.85, 0.70) * lightGlint;

    // Specular highlight on crack near light
    if (edgeDist < 0.02) {
        float specD = length(p - lightPos);
        col += vec3(1.0) * exp(-specD * specD * 30.0) * crackMask * 0.5;
    }

    // Composite (terminal already refracted)
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * 0.7 * (1.0 - termLuma * 0.4);
    vec3 result = mix(terminal.rgb, terminal.rgb + col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
