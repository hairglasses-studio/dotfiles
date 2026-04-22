// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Mandala flower — 12-petal layered flower with iterative petaling + rotating rings

const int   PETAL_LAYERS = 4;
const int   PETALS_BASE  = 12;
const float INTENSITY = 0.55;

vec3 mf_pal(float t) {
    vec3 a = vec3(0.90, 0.30, 0.70);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.20, 0.95, 0.60);
    vec3 d = vec3(0.96, 0.85, 0.40);
    vec3 e = vec3(0.10, 0.82, 0.92);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

// Petal shape SDF at origin
float sdPetal(vec2 p, float size, float aspect) {
    p.y /= aspect;
    return length(p) - size;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    float r = length(p);
    float a = atan(p.y, p.x);

    vec3 col = vec3(0.01, 0.01, 0.03);

    // Central core
    col += mf_pal(fract(x_Time * 0.04)) * exp(-r * r * 600.0) * 1.5;

    // Multiple layers of petals at increasing radii
    for (int L = 0; L < PETAL_LAYERS; L++) {
        float fL = float(L);
        int petals = PETALS_BASE + int(fL * 2.0);   // more petals per outer layer
        float layerR = 0.05 + fL * 0.08;
        float petalSize = 0.04 + fL * 0.012;
        float layerRot = x_Time * (0.1 + fL * 0.05) * (fL + 1.0 > 3.0 ? -1.0 : 1.0);   // alternating direction

        // Only consider fragments near this layer's radius
        if (abs(r - layerR) > petalSize * 3.0) continue;

        // Modular petal angle
        float petalAng = mod(a + layerRot, 6.28318 / float(petals)) - 3.14159 / float(petals);
        // Petal origin (radius outward from center)
        vec2 petalOrigin = vec2(cos(petalAng) * layerR, sin(petalAng) * layerR);
        // Pixel relative to that petal origin, in rotated frame
        float petalCosAng = cos(petalAng + 3.14159 * 0.5);
        float petalSinAng = sin(petalAng + 3.14159 * 0.5);
        vec2 petalLocal = mat2(petalCosAng, -petalSinAng, petalSinAng, petalCosAng) * (vec2(r * cos(petalAng), r * sin(petalAng)) - petalOrigin);

        float d = sdPetal(petalLocal, petalSize, 2.2);
        float body = smoothstep(0.0, -petalSize * 0.3, d);
        float edge = exp(-d * d * 2000.0) * 0.4;

        vec3 layerCol = mf_pal(fract(fL * 0.18 + x_Time * 0.04));
        col = mix(col, layerCol * 0.6, body * 0.5);
        col += layerCol * edge;
    }

    // Concentric bright ring
    float ringD = abs(r - 0.42);
    float ring = exp(-ringD * ringD * 5000.0);
    col += mf_pal(fract(x_Time * 0.06)) * ring * 0.5;

    // Emanating pulses
    float pulse = sin(r * 30.0 - x_Time * 2.0) * 0.5 + 0.5;
    pulse = pow(pulse, 6.0);
    col += mf_pal(fract(a / 6.28 + x_Time * 0.02)) * pulse * 0.1 * smoothstep(0.0, 0.5, r);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
