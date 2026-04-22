// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Raymarched volumetric plasma globe with interior electric filaments

const int   STEPS        = 48;       // raymarch steps through globe interior
const float GLOBE_RADIUS = 0.42;
const float FIL_SCALE    = 3.5;
const float GLOW_FALLOFF = 1.6;
const float INTENSITY    = 0.55;
const float COLOR_SPEED  = 0.07;

vec3 pg_palette(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98); // violet
    vec3 b = vec3(0.90, 0.18, 0.60); // magenta
    vec3 c = vec3(0.10, 0.82, 0.92); // cyan
    vec3 d = vec3(0.20, 0.95, 0.60); // mint highlight
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Cheap 3D hash + value noise
float pg_hash(vec3 p) {
    p = fract(p * vec3(443.8975, 397.2973, 491.1871));
    p += dot(p.yxz, p.xyz + 19.19);
    return fract(p.x * p.y * p.z);
}

float pg_noise3(vec3 p) {
    vec3 i = floor(p);
    vec3 f = fract(p);
    vec3 u = f * f * (3.0 - 2.0 * f);
    return mix(
        mix(
            mix(pg_hash(i + vec3(0,0,0)), pg_hash(i + vec3(1,0,0)), u.x),
            mix(pg_hash(i + vec3(0,1,0)), pg_hash(i + vec3(1,1,0)), u.x),
            u.y),
        mix(
            mix(pg_hash(i + vec3(0,0,1)), pg_hash(i + vec3(1,0,1)), u.x),
            mix(pg_hash(i + vec3(0,1,1)), pg_hash(i + vec3(1,1,1)), u.x),
            u.y),
        u.z);
}

// 6-octave 3D FBM for interior density
float pg_fbm3(vec3 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 6; i++) {
        v += a * pg_noise3(p);
        p = p * 2.1 + vec3(0.13, 0.27, 0.51);
        a *= 0.5;
    }
    return v;
}

// Density field — filaments concentrated near radial lines
float pg_density(vec3 p, float t) {
    vec3 q = p * FIL_SCALE + vec3(0.0, 0.0, t * 0.4);
    float n = pg_fbm3(q);
    // Bias toward filament-like structures — anisotropic high-frequency ridges
    float radial = 1.0 - smoothstep(0.15, 0.6, length(p.xy));
    float ridge = abs(n - 0.5) * 2.0;                        // [0,1], 0 at ridges
    float fil = pow(1.0 - ridge, 3.0);                        // sharp filaments
    return fil * (0.3 + 0.7 * radial);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    // Centered, aspect-corrected
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // 2D outer glow — far-field halo
    float r2 = length(p);
    float halo = exp(-pow(max(0.0, r2 - GLOBE_RADIUS) / 0.35, GLOW_FALLOFF)) * 0.9;

    vec3 col = vec3(0.0);

    // Ray setup: orthographic cam down +Z, globe at origin
    vec3 ro = vec3(p, -1.0);
    vec3 rd = vec3(0.0, 0.0, 1.0);

    // Analytic sphere intersection (unit sphere scaled to GLOBE_RADIUS)
    float b = dot(ro, rd);
    float c = dot(ro, ro) - GLOBE_RADIUS * GLOBE_RADIUS;
    float h = b * b - c;
    if (h > 0.0) {
        h = sqrt(h);
        float tNear = -b - h;
        float tFar  = -b + h;
        float stepLen = (tFar - tNear) / float(STEPS);
        float t = tNear;
        float accum = 0.0;
        vec3 colAccum = vec3(0.0);
        // Rotate the volume over time
        float ang = x_Time * 0.15;
        mat2 rotXZ = mat2(cos(ang), -sin(ang), sin(ang), cos(ang));
        mat2 rotYZ = mat2(cos(ang * 0.7), -sin(ang * 0.7), sin(ang * 0.7), cos(ang * 0.7));
        for (int i = 0; i < STEPS; i++) {
            vec3 sp = ro + rd * t;
            vec3 rp = sp;
            rp.xz = rotXZ * rp.xz;
            rp.yz = rotYZ * rp.yz;
            float d = pg_density(rp, x_Time);
            // Color drifts with depth + time
            vec3 stopColor = pg_palette(fract(length(rp) * 0.8 + x_Time * COLOR_SPEED + float(i) * 0.004));
            colAccum += stopColor * d * stepLen;
            accum   += d * stepLen;
            t += stepLen;
        }
        // Interior glow — brighter for denser paths
        col += colAccum * 2.2;
        // Rim highlight
        float rim = smoothstep(GLOBE_RADIUS * 0.95, GLOBE_RADIUS, length(p));
        col += pg_palette(fract(x_Time * COLOR_SPEED)) * rim * 0.8;
    }

    // Outer halo
    col += pg_palette(fract(x_Time * COLOR_SPEED + 0.15)) * halo * 0.35;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
