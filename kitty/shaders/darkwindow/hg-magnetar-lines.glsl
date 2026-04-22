// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Magnetar dipole field — visualize magnetic field lines as streaming curved arcs

const int   FIELD_SAMPS  = 48;
const int   FIELD_LINES  = 18;
const float INTENSITY    = 0.55;

vec3 ma_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.95, 0.25, 0.55); // magenta
    vec3 d = vec3(0.96, 0.85, 0.40); // gold
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Dipole magnetic field at position p relative to a dipole along z-axis
// In 2D XY plane, field = (3cos²θ - 1)/r³ radial + 3sin(θ)cos(θ)/r³ perpendicular
vec2 dipoleField(vec2 p) {
    float r = length(p);
    if (r < 0.01) return vec2(0.0);
    float theta = atan(p.y, p.x);
    // Standard dipole in 2D slice
    float rPow = 1.0 / (r * r * r + 0.01);
    float Br = 2.0 * cos(theta) * rPow;
    float Bt = sin(theta) * rPow;
    // Convert (Br, Bt) polar → cartesian
    vec2 rhat = vec2(cos(theta), sin(theta));
    vec2 that = vec2(-sin(theta), cos(theta));
    return Br * rhat + Bt * that;
}

// Trace a field line forward/backward from a seed point
float fieldLineDist(vec2 p, vec2 seed, float t) {
    float minD = 1e9;
    vec2 pos = seed;
    float step = 0.015;
    // Forward trace
    for (int i = 0; i < FIELD_SAMPS / 2; i++) {
        vec2 v = dipoleField(pos);
        float vLen = length(v);
        if (vLen < 0.001) break;
        pos += v / vLen * step;
        if (length(pos) > 1.0) break;
        float d = length(p - pos);
        minD = min(minD, d);
    }
    pos = seed;
    // Backward trace
    for (int i = 0; i < FIELD_SAMPS / 2; i++) {
        vec2 v = dipoleField(pos);
        float vLen = length(v);
        if (vLen < 0.001) break;
        pos -= v / vLen * step;
        if (length(pos) > 1.0) break;
        float d = length(p - pos);
        minD = min(minD, d);
    }
    return minD;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Rotating dipole axis
    float dipoleAngle = x_Time * 0.15;
    float cr = cos(dipoleAngle), sr = sin(dipoleAngle);
    mat2 rot = mat2(cr, -sr, sr, cr);
    vec2 pr = rot * p;

    vec3 col = vec3(0.0);

    // Dense magnetar star at center
    float r = length(pr);
    float core = exp(-r * r * 300.0) * 1.5;
    col += vec3(1.0, 0.98, 0.9) * core;
    col += ma_pal(0.0) * exp(-r * r * 60.0) * 0.4;

    // Draw field lines — seed points along different L-shells
    for (int i = 0; i < FIELD_LINES; i++) {
        float fi = float(i);
        // Seed at varying radius along equator (perpendicular to dipole axis)
        float seedR = 0.15 + fi * 0.025;
        vec2 seed = vec2(seedR, 0.0);
        // One line per side
        for (int sideI = 0; sideI < 2; sideI++) {
            vec2 actualSeed = (sideI == 0) ? seed : -seed;
            float lineD = fieldLineDist(pr, actualSeed, x_Time);
            float lineCore = smoothstep(0.006, 0.0, lineD);
            float lineGlow = exp(-lineD * lineD * 1500.0) * 0.3;

            // Streaming pulse — traveling along line
            float pulsePhase = fract(x_Time * 0.6 + fi * 0.13 + float(sideI) * 0.5);
            // Approximate "along" by distance from seed + some angle encoding
            float along = length(pr - actualSeed) * 2.0;
            float pulseDist = abs(fract(along) - pulsePhase);
            float pulse = exp(-pulseDist * pulseDist * 80.0) * lineCore;

            vec3 lc = ma_pal(fract(fi * 0.08 + x_Time * 0.04));
            col += lc * (lineCore * 0.7 + lineGlow);
            col += vec3(1.0, 0.95, 0.9) * pulse * 0.7;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
