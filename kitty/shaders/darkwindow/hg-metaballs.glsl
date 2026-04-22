// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Neon metaballs — 12 merging blobs, smooth union, glow edges

const int   BALLS      = 12;
const float THRESH     = 1.8;         // isosurface threshold
const float EDGE_BAND  = 0.25;
const float INTENSITY  = 0.55;

vec3 mb_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.30, 0.70);
    vec3 d = vec3(0.20, 0.95, 0.60);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float mb_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  mb_hash2(float n) { return vec2(mb_hash(n), mb_hash(n * 1.37 + 11.0)); }

// Ball position — wandering lissajous
vec2 ballPos(int i, float t) {
    float fi = float(i);
    vec2 center = mb_hash2(fi * 3.7) * 0.8 - 0.4;
    center.x *= x_WindowSize.x / x_WindowSize.y;
    float f1 = 0.25 + mb_hash(fi * 5.1) * 0.3;
    float f2 = 0.25 + mb_hash(fi * 7.3) * 0.3;
    center += 0.2 * vec2(sin(t * f1 + fi), cos(t * f2 + fi * 1.3));
    return center;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Accumulate field — sum of 1/r^2 kernels
    float field = 0.0;
    vec3 colorSum = vec3(0.0);
    float colorWeight = 0.0;
    for (int i = 0; i < BALLS; i++) {
        vec2 bp = ballPos(i, x_Time);
        float d = length(p - bp);
        float r = 0.08 + 0.05 * mb_hash(float(i) * 3.7);
        float contrib = r * r / (d * d + 0.001);
        field += contrib;
        // Color contribution weighted by proximity
        vec3 bc = mb_pal(fract(float(i) * 0.08 + x_Time * 0.04));
        colorSum += bc * contrib;
        colorWeight += contrib;
    }
    if (colorWeight > 0.001) colorSum /= colorWeight;

    vec3 col = vec3(0.0);

    // Inside isosurface: filled color
    if (field > THRESH) {
        col = colorSum * smoothstep(THRESH, THRESH + 0.3, field);
    }

    // Edge glow: where field crosses threshold
    float edge = smoothstep(THRESH - EDGE_BAND, THRESH, field) * smoothstep(THRESH + EDGE_BAND, THRESH, field);
    col += colorSum * edge * 1.3;

    // Outer halo
    float halo = smoothstep(THRESH - 0.8, THRESH, field) * 0.3;
    col += colorSum * halo * 0.4;

    // Soft inner highlight (fake spec — where field is very high)
    float hi = smoothstep(THRESH + 0.5, THRESH + 2.0, field);
    col = mix(col, vec3(1.0), hi * 0.4);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
