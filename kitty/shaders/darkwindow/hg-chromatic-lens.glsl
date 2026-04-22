// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Chromatic lens — lens displaying chromatic aberration rings at varying focus

const float INTENSITY = 0.55;

vec3 cl_pal(float t) {
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

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Lens center — slowly drifting
    vec2 lensCenter = 0.1 * vec2(sin(x_Time * 0.2), cos(x_Time * 0.17));
    float lensR = 0.35;

    vec2 toLens = p - lensCenter;
    float r = length(toLens);

    vec3 col = vec3(0.0);

    if (r < lensR) {
        // Inside lens — chromatic aberration based on radial distance
        float aberrStr = r * r / (lensR * lensR) * 0.025;
        vec2 normal = toLens / max(r, 0.001);

        // Sample 3 channels at different refraction strengths (dispersion)
        vec2 uvR = uv - normal * aberrStr * 1.2;
        vec2 uvG = uv - normal * aberrStr;
        vec2 uvB = uv - normal * aberrStr * 0.8;

        col.r = x_Texture(uvR).r;
        col.g = x_Texture(uvG).g;
        col.b = x_Texture(uvB).b;

        // Concentric color rings (Airy pattern simulation)
        for (int ring = 1; ring <= 5; ring++) {
            float fring = float(ring);
            float ringR = lensR * fring * 0.15;
            float ringDist = abs(r - ringR);
            float ring_ = exp(-ringDist * ringDist * 5000.0) * 0.2;
            // Each ring a different hue
            col += cl_pal(fract(fring * 0.2 + x_Time * 0.03)) * ring_;
        }

        // Edge highlight (bright ring)
        float edgeD = lensR - r;
        float edge = smoothstep(0.01, 0.0, edgeD);
        col += cl_pal(fract(x_Time * 0.05)) * edge * 0.6;

        // Central bright spot (focused light)
        float centerD = r;
        col += vec3(1.0) * exp(-centerD * centerD * 1000.0) * 0.4;
    } else {
        // Outside lens — pass through terminal
        vec4 terminal = x_Texture(uv);
        col = terminal.rgb;
        // Soft outer halo
        float haloD = r - lensR;
        col += cl_pal(fract(x_Time * 0.04)) * exp(-haloD * haloD * 50.0) * 0.15;
    }

    _wShaderOut = vec4(col, 1.0);
}
