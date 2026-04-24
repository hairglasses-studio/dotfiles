// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Calabi-Yau slice — 2D Hanson-style projection of nested 5-fold symmetric complex iso-contours evoking a cross-section of a 6D Calabi-Yau manifold. 12 layered fractional-power curves with per-layer phase, hue, and rotation.

const int   LAYERS = 12;
const int   N_POWER = 5;              // 5-fold symmetry (quintic Calabi-Yau)
const int   SAMPS_PER_LAYER = 180;
const float INTENSITY = 0.55;

vec3 cy_pal(float t) {
    vec3 cyan   = vec3(0.25, 0.90, 1.00);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 amber  = vec3(1.00, 0.70, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, amber, s - 2.0);
    else if (s < 4.0) return mix(amber, mint, s - 3.0);
    else              return mix(mint, cyan, s - 4.0);
}

// Polar → cartesian with 5-fold modulating radius. Layer-specific phase + scale.
vec2 layerPoint(int layer, int k, float globalRot) {
    float fl = float(layer);
    float fk = float(k);
    float theta = fk / float(SAMPS_PER_LAYER) * 6.28318 + globalRot + fl * 0.18;
    // Radial modulation: r = baseR * (1 + A cos(N θ + phaseLayer))
    float baseR = 0.15 + fl * 0.06;
    float amp = 0.18 + 0.15 * sin(fl * 0.4);
    float phaseLayer = fl * 0.37;
    float r = baseR * (1.0 + amp * cos(float(N_POWER) * theta + phaseLayer));
    // Secondary harmonic (for richer CY-like fluting)
    r += 0.015 * cos(float(N_POWER * 2) * theta + fl * 0.7);
    return vec2(cos(theta), sin(theta)) * r;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.003, 0.005, 0.018);

    // Background radial vignette
    float rdist = length(p);
    col += vec3(0.05, 0.02, 0.08) * (1.0 - smoothstep(0.0, 1.1, rdist)) * 0.5;

    float globalRot = x_Time * 0.08;

    float minD = 1e9;
    int closestLayer = 0;
    float closestSampK = 0.0;

    // For each layer, sample SAMPS_PER_LAYER points around the curve, find
    // closest segment to pixel.
    for (int l = 0; l < LAYERS; l++) {
        vec2 prev = layerPoint(l, 0, globalRot);
        for (int k = 1; k <= SAMPS_PER_LAYER; k++) {
            vec2 cur;
            if (k == SAMPS_PER_LAYER) cur = layerPoint(l, 0, globalRot); // close loop
            else cur = layerPoint(l, k, globalRot);
            vec2 ab = cur - prev;
            vec2 pa = p - prev;
            float lenSq = dot(ab, ab);
            if (lenSq > 1e-7) {
                float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
                float d = length(pa - ab * h);
                if (d < minD) {
                    minD = d;
                    closestLayer = l;
                    closestSampK = (float(k - 1) + h) / float(SAMPS_PER_LAYER);
                }
            }
            prev = cur;
        }
    }

    // Line rendering
    float thickness = 0.002;
    float lineMask = exp(-minD * minD / (thickness * thickness) * 1.5);
    vec3 lineCol = cy_pal(fract(float(closestLayer) * 0.09 + closestSampK * 0.4 + x_Time * 0.03));
    col += lineCol * lineMask * 1.3;
    // Halo
    col += lineCol * exp(-minD * minD * 800.0) * 0.22;

    // Center accent — bright star/dot at origin where all layers converge
    float centerGlow = exp(-rdist * rdist * 250.0);
    col += vec3(1.0, 0.95, 0.80) * centerGlow * 0.9;
    col += cy_pal(0.35) * exp(-rdist * rdist * 20.0) * 0.25;

    // 5 outer "ray" markers (singular points at 5-fold lattice)
    for (int i = 0; i < 5; i++) {
        float fi = float(i);
        float ang = fi / 5.0 * 6.28318 + globalRot;
        vec2 outer = vec2(cos(ang), sin(ang)) * 0.85;
        float od = length(p - outer);
        col += cy_pal(fract(fi * 0.2 + x_Time * 0.04)) * exp(-od * od * 3000.0) * 0.7;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
