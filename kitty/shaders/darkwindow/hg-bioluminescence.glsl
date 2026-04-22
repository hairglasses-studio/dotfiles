// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Bioluminescent tendrils — 12 organic curling tentacles with pulse glow

const int   TENDRILS   = 12;
const int   TENDRIL_PTS = 8;  // control points per tendril
const float GLOW_RADIUS = 0.025;
const float INTENSITY   = 0.5;

vec3 bl_pal(float t) {
    vec3 a = vec3(0.15, 0.95, 0.70); // bioluminescent mint-green
    vec3 b = vec3(0.10, 0.82, 0.92); // cyan
    vec3 c = vec3(0.55, 0.30, 0.98); // violet
    vec3 d = vec3(0.90, 0.30, 0.70); // magenta
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float bl_hash(float n) { return fract(sin(n * 127.1) * 43758.5); }

// Tendril shape: anchor at a random position, curls outward with sinusoidal motion
vec2 tendrilPoint(int tendril, int pt, float t) {
    float ft = float(tendril);
    float fp = float(pt);
    float u = fp / float(TENDRIL_PTS - 1);    // 0 at root, 1 at tip

    // Anchor
    vec2 anchor = vec2(
        (bl_hash(ft * 3.1) - 0.5) * 1.4,
        (bl_hash(ft * 5.7) - 0.5) * 1.0
    );
    anchor.x *= x_WindowSize.x / x_WindowSize.y;

    // Base direction
    float baseAngle = bl_hash(ft * 7.3) * 6.28318;
    vec2 baseDir = vec2(cos(baseAngle), sin(baseAngle));

    // Curling: angle changes along the tendril
    float curl = sin(u * 3.0 + t * 0.5 + ft) * 1.2;
    float angle = baseAngle + curl;
    vec2 dir = vec2(cos(angle), sin(angle));

    // Length grows with u, plus wiggle
    float len = u * 0.35;
    // Add perpendicular wiggle
    vec2 perp = vec2(-dir.y, dir.x);
    float wiggle = sin(u * 8.0 + t * 1.5 + ft * 2.7) * 0.04;

    return anchor + baseDir * len + perp * wiggle;
}

// Distance to a bezier-like tendril sampled at segments
float tendrilDist(vec2 p, int tendril, float t, out float alongT) {
    float minDist = 1e9;
    float minAlong = 0.0;
    for (int i = 0; i < TENDRIL_PTS - 1; i++) {
        vec2 a = tendrilPoint(tendril, i, t);
        vec2 b = tendrilPoint(tendril, i + 1, t);
        vec2 ab = b - a;
        vec2 ap = p - a;
        float h = clamp(dot(ap, ab) / dot(ab, ab), 0.0, 1.0);
        float d = length(ap - ab * h);
        if (d < minDist) {
            minDist = d;
            minAlong = (float(i) + h) / float(TENDRIL_PTS - 1);
        }
    }
    alongT = minAlong;
    return minDist;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    for (int i = 0; i < TENDRILS; i++) {
        float fi = float(i);
        float alongT;
        float d = tendrilDist(p, i, x_Time, alongT);

        // Core: thin bright line
        float r = GLOW_RADIUS * (1.0 - alongT * 0.6);  // thins toward tip
        float core = exp(-d * d / (r * r) * 4.0);
        // Outer halo
        float halo = exp(-d * d * 80.0) * 0.4;

        // Pulse traveling along tendril
        float pulsePhase = fract(x_Time * 0.8 + fi * 0.3);
        float pulseD = abs(alongT - pulsePhase);
        float pulse = exp(-pulseD * pulseD * 80.0) * core * 1.5;

        // Color varies per tendril
        vec3 tc = bl_pal(fract(fi * 0.085 + x_Time * 0.04));
        col += tc * (core * 0.9 + halo + pulse);
    }

    // Ambient base glow — seabed feel
    float ambient = 0.05 + 0.02 * sin(x_Time * 0.2);
    col += bl_pal(0.0) * ambient * 0.15;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
