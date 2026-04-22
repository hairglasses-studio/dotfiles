// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Plasma ball — glass sphere with internal lightning tendrils reaching outward

const int   TENDRILS   = 7;
const int   TENDRIL_SEG = 10;
const float BALL_R     = 0.32;
const float CORE_R     = 0.05;
const float INTENSITY  = 0.55;

vec3 pb_pal(float t) {
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 cyan  = vec3(0.20, 0.85, 0.98);
    vec3 pink  = vec3(0.95, 0.30, 0.75);
    vec3 white = vec3(0.95, 0.98, 1.00);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(vio, cyan, s);
    else if (s < 2.0) return mix(cyan, pink, s - 1.0);
    else if (s < 3.0) return mix(pink, white, s - 2.0);
    else              return mix(white, vio, s - 3.0);
}

float pb_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Tendril segment point
vec2 tendrilPoint(int tend, int seg, float t) {
    float ft = float(tend);
    float fs = float(seg) / float(TENDRIL_SEG - 1);
    float seed = ft * 13.7;
    // Target outer point (wanders on sphere surface)
    float baseAng = ft * 6.28318 / float(TENDRILS) + t * 0.1;
    float ang = baseAng + sin(t * 0.3 + ft) * 0.5;
    vec2 outerTarget = vec2(cos(ang), sin(ang)) * BALL_R * 0.95;
    // Current position along tendril — bends randomly
    vec2 base = vec2(0.0) + fs * outerTarget;
    // Jitter per segment — more at midsection
    float jitEnv = sin(fs * 3.14);
    float jit = (pb_hash(seed + float(seg) * 3.7 + floor(t * 3.0)) - 0.5) * 0.08 * jitEnv;
    vec2 perp = vec2(-outerTarget.y, outerTarget.x) / length(outerTarget);
    return base + perp * jit;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    float r = length(p);

    // Outside ball = nothing
    // Inside ball: bright background tint + tendrils + core
    if (r < BALL_R) {
        // Background of ball
        col = vec3(0.08, 0.05, 0.12) * (1.0 - r / BALL_R);

        // Tendrils
        float minTendD = 1e9;
        float tendColorHue = 0.0;
        for (int t = 0; t < TENDRILS; t++) {
            float minD = 1e9;
            for (int seg = 0; seg < TENDRIL_SEG - 1; seg++) {
                vec2 a = tendrilPoint(t, seg, x_Time);
                vec2 b = tendrilPoint(t, seg + 1, x_Time);
                vec2 pa = p - a;
                vec2 ba = b - a;
                float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
                float d = length(pa - ba * h);
                minD = min(minD, d);
            }
            if (minD < minTendD) {
                minTendD = minD;
                tendColorHue = float(t) / float(TENDRILS);
            }
        }

        float core = exp(-minTendD * minTendD * 8000.0);
        float glow = exp(-minTendD * minTendD * 400.0) * 0.3;
        vec3 tendCol = pb_pal(fract(tendColorHue + x_Time * 0.05));
        col += tendCol * (core * 1.5 + glow);

        // Central bright electrode
        float coreMask = exp(-r * r / (CORE_R * CORE_R) * 2.0);
        col += vec3(1.0, 0.95, 0.9) * coreMask * 1.5;
        col += pb_pal(0.0) * exp(-r * r / (CORE_R * CORE_R) * 0.3) * 0.4;

        // Glass rim (inside edge of ball)
        float rimDist = BALL_R - r;
        float rim = smoothstep(0.01, 0.0, rimDist);
        col += vec3(0.7, 0.85, 0.95) * rim * 0.4;
    } else {
        // Outside glass — terminal shows through with slight tint
        // Outer halo
        float haloDist = r - BALL_R;
        float halo = exp(-haloDist * 4.0) * 0.2;
        col += pb_pal(fract(x_Time * 0.05)) * halo;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
