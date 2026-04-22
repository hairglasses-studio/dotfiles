// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Spirograph — multi-gear hypotrochoid curves forming intricate flower-like patterns

const int   ROLLS = 4;
const int   SAMPLES = 256;
const float INTENSITY = 0.55;

vec3 sg_pal(float t) {
    vec3 a = vec3(0.95, 0.30, 0.70);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.10, 0.82, 0.92);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Hypotrochoid: small circle rolls inside big circle
// R = big radius, r = small radius, d = pen distance from small circle center
vec2 hypotrochoid(float theta, float R, float r, float d) {
    float rotRatio = (R - r) / r;
    return vec2(
        (R - r) * cos(theta) + d * cos(rotRatio * theta),
        (R - r) * sin(theta) - d * sin(rotRatio * theta)
    );
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.03);

    // Multiple hypotrochoids at different parameters
    for (int L = 0; L < ROLLS; L++) {
        float fL = float(L);
        float R = 0.35 - fL * 0.04;
        float r = 0.07 + fL * 0.02;
        float d = 0.10 + fL * 0.015;
        // Animate phase
        float phase = x_Time * (0.3 + fL * 0.1);

        // Find nearest segment on curve
        float minD = 1e9;
        float closestTheta = 0.0;
        // Sample along θ from 0 to 2π * R / gcd(R,r) — use long range
        int samps = SAMPLES;
        for (int i = 0; i < samps; i++) {
            float theta = float(i) / float(samps) * 50.0 + phase;
            vec2 pt = hypotrochoid(theta, R, r, d);
            float dd = length(p - pt);
            if (dd < minD) {
                minD = dd;
                closestTheta = theta;
            }
        }

        float lineCore = smoothstep(0.003, 0.0, minD);
        float glow = exp(-minD * minD * 1000.0) * 0.3;
        vec3 lc = sg_pal(fract(fL * 0.2 + closestTheta * 0.1 + x_Time * 0.03));
        col += lc * (lineCore * 0.8 + glow);
    }

    // Center gear dot
    col += sg_pal(fract(x_Time * 0.05)) * exp(-length(p) * length(p) * 800.0) * 0.8;

    // Outer ring (R = biggest)
    float outerD = abs(length(p) - 0.4);
    col += vec3(0.2, 0.25, 0.35) * exp(-outerD * outerD * 2000.0) * 0.4;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
