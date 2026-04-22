// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Möbius strip — parametric surface rendered as glowing band with flowing highlights

const int   BAND_SAMPS  = 128;
const float BAND_RAD    = 0.3;
const float BAND_WIDTH  = 0.08;
const float INTENSITY   = 0.55;

vec3 mo_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.25, 0.60);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Parametric Möbius strip: position at angle u and width v
vec3 moebiusPoint(float u, float v) {
    float cosU = cos(u);
    float sinU = sin(u);
    float cosHalfU = cos(u * 0.5);
    float sinHalfU = sin(u * 0.5);
    float r = BAND_RAD + v * cosHalfU * 0.5;
    return vec3(
        r * cosU,
        v * sinHalfU * 0.5,
        r * sinU
    );
}

// 3D → 2D project
vec2 project3to2(vec3 v) {
    float zDist = 2.0;
    float z = 1.0 / (zDist - v.z);
    return v.xy * z;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.02, 0.02, 0.05);

    // Rotate the whole strip
    float t = x_Time * 0.25;
    float cr = cos(t), sr = sin(t);
    mat3 rotY = mat3(cr, 0.0, -sr, 0.0, 1.0, 0.0, sr, 0.0, cr);
    float cp = cos(t * 0.6), sp = sin(t * 0.6);
    mat3 rotX = mat3(1.0, 0.0, 0.0, 0.0, cp, -sp, 0.0, sp, cp);
    mat3 rot = rotX * rotY;

    // Sample many points on the strip and compute 2D distance
    float minD = 1e9;
    float closestU = 0.0;
    float closestV = 0.0;
    for (int i = 0; i < BAND_SAMPS; i++) {
        float u = float(i) / float(BAND_SAMPS) * 6.28318;
        // Sample along width at v=0 (center of band)
        for (int j = -2; j <= 2; j++) {
            float v = float(j) / 2.0 * BAND_WIDTH;
            vec3 pt = rot * moebiusPoint(u, v);
            if (pt.z < -0.5) continue;
            vec2 proj = project3to2(pt);
            float d = length(p - proj);
            if (d < minD) {
                minD = d;
                closestU = u;
                closestV = v;
            }
        }
    }

    // Render band
    float lineCore = smoothstep(0.005, 0.0, minD);
    float lineGlow = exp(-minD * minD * 500.0) * 0.4;
    // Color depends on where on the strip (u position)
    vec3 bandCol = mo_pal(fract(closestU / 6.28 + x_Time * 0.05));
    col += bandCol * (lineCore * 1.2 + lineGlow);

    // Flowing highlight — traveling bright bit along strip
    float highlightPhase = fract(x_Time * 0.3);
    float highlightDist = abs(closestU / 6.28318 - highlightPhase);
    float highlight = exp(-highlightDist * highlightDist * 50.0) * lineCore;
    col += vec3(1.0, 0.95, 0.9) * highlight * 0.9;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
