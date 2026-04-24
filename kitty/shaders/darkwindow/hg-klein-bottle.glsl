// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Klein bottle — parametric figure-8 Klein-bottle immersion R³ with u- and v- isocurves sampled as a wireframe grid, rotating in 3D with time. Self-intersecting surface renders as interlocking tubes.

const int   U_LINES = 8;
const int   V_LINES = 8;
const int   SAMPS = 16;       // samples per iso-line
const float INTENSITY = 0.55;

vec3 klein_pal(float t) {
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

// 3D rotation around Y then X
vec3 rotate3D(vec3 v, float ax, float ay) {
    float cy = cos(ay), sy = sin(ay);
    vec3 w = vec3(v.x * cy - v.z * sy, v.y, v.x * sy + v.z * cy);
    float cx = cos(ax), sx = sin(ax);
    return vec3(w.x, w.y * cx - w.z * sx, w.y * sx + w.z * cx);
}

// Klein bottle parametric immersion (figure-8 form).
// u ∈ [0, 2π] (around the big loop), v ∈ [0, 2π] (around the figure-8)
vec3 kleinBottle(float u, float v) {
    float R = 2.0;
    float hu = u * 0.5;
    float cu = cos(hu), su = sin(hu);
    float sv = sin(v), s2v = sin(2.0 * v);
    // Width of ring at parameter u
    float ringOff = cu * sv - su * s2v;
    float ringR = R + ringOff;
    return vec3(ringR * cos(u),
                ringR * sin(u),
                su * sv + cu * s2v);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv_screen = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv_screen);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.004, 0.006, 0.020);

    // Background soft radial wash
    float rdist = length(p);
    col += vec3(0.04, 0.03, 0.06) * (1.0 - smoothstep(0.0, 1.2, rdist)) * 0.4;

    float rotX = 0.4 + 0.3 * sin(x_Time * 0.15);
    float rotY = x_Time * 0.18;
    float scale = 0.12;

    float minD = 1e9;
    float closestFeatureKey = 0.0;
    float closestDepth = 0.0;

    // === u-iso-lines: fix u, vary v ===
    for (int i = 0; i < U_LINES; i++) {
        float fi = float(i);
        float u = fi / float(U_LINES) * 6.28318;
        vec3 prev = rotate3D(kleinBottle(u, 0.0), rotX, rotY);
        for (int k = 1; k <= SAMPS; k++) {
            float fk = float(k);
            float v = fk / float(SAMPS) * 6.28318;
            vec3 cur = rotate3D(kleinBottle(u, v), rotX, rotY);
            vec2 a = prev.xy * scale;
            vec2 b = cur.xy * scale;
            vec2 ab = b - a;
            vec2 pa = p - a;
            float lenSq = dot(ab, ab);
            if (lenSq > 1e-6) {
                float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
                float d = length(pa - ab * h);
                if (d < minD) {
                    minD = d;
                    closestFeatureKey = fi * 0.11;
                    closestDepth = mix(prev.z, cur.z, h);
                }
            }
            prev = cur;
        }
    }

    // === v-iso-lines: fix v, vary u ===
    for (int j = 0; j < V_LINES; j++) {
        float fj = float(j);
        float v = fj / float(V_LINES) * 6.28318;
        vec3 prev = rotate3D(kleinBottle(0.0, v), rotX, rotY);
        for (int k = 1; k <= SAMPS; k++) {
            float fk = float(k);
            float u = fk / float(SAMPS) * 6.28318;
            vec3 cur = rotate3D(kleinBottle(u, v), rotX, rotY);
            vec2 a = prev.xy * scale;
            vec2 b = cur.xy * scale;
            vec2 ab = b - a;
            vec2 pa = p - a;
            float lenSq = dot(ab, ab);
            if (lenSq > 1e-6) {
                float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
                float d = length(pa - ab * h);
                if (d < minD) {
                    minD = d;
                    closestFeatureKey = 0.5 + fj * 0.11;  // offset to distinguish from u-lines
                    closestDepth = mix(prev.z, cur.z, h);
                }
            }
            prev = cur;
        }
    }

    // Render closest line
    float thickness = 0.004;
    float lineMask = exp(-minD * minD / (thickness * thickness) * 1.5);
    float depthFade = 1.0 / (1.0 + abs(closestDepth) * 0.15);
    vec3 lineCol = klein_pal(fract(closestFeatureKey + x_Time * 0.03));
    col += lineCol * lineMask * depthFade * 1.25;
    // Halo
    col += lineCol * exp(-minD * minD * 600.0) * depthFade * 0.2;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
