// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Binary star system — 2 orbiting suns with plasma streams + L1 Lagrangian point

const int   PLASMA_SAMPLES = 40;
const float ORBIT_RADIUS   = 0.18;
const float ORBIT_SPEED    = 0.4;
const float STAR_RADIUS    = 0.07;
const float INTENSITY      = 0.55;

vec3 star_col(float heat) {
    vec3 deep   = vec3(0.35, 0.02, 0.05);
    vec3 orange = vec3(1.00, 0.45, 0.12);
    vec3 yellow = vec3(1.00, 0.90, 0.45);
    vec3 white  = vec3(1.00, 0.98, 0.85);
    if (heat < 0.33)      return mix(deep, orange, heat * 3.0);
    else if (heat < 0.66) return mix(orange, yellow, (heat - 0.33) * 3.0);
    else                  return mix(yellow, white, (heat - 0.66) * 3.0);
}

vec3 plasma_col(float t) {
    vec3 a = vec3(1.00, 0.50, 0.20);
    vec3 b = vec3(0.95, 0.30, 0.70);
    vec3 c = vec3(0.55, 0.30, 0.98);
    vec3 d = vec3(0.10, 0.82, 0.92);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float ts_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Two stars orbiting center of mass (origin)
    float theta = x_Time * ORBIT_SPEED;
    vec2 star1 = vec2(cos(theta), sin(theta)) * ORBIT_RADIUS;
    vec2 star2 = -star1;

    vec3 col = vec3(0.0);

    // Star surfaces
    float d1 = length(p - star1);
    float d2 = length(p - star2);
    // Surface activity: FBM-like via angle
    float ang1 = atan(p.y - star1.y, p.x - star1.x);
    float ang2 = atan(p.y - star2.y, p.x - star2.x);
    float activity1 = sin(ang1 * 6.0 + x_Time * 3.0) * sin(ang1 * 11.0 + x_Time * 2.3);
    float activity2 = cos(ang2 * 5.0 + x_Time * 2.7) * sin(ang2 * 9.0 + x_Time * 3.1);

    float surf1 = smoothstep(STAR_RADIUS, STAR_RADIUS * 0.98, d1);
    float surf2 = smoothstep(STAR_RADIUS, STAR_RADIUS * 0.98, d2);
    float heat1 = 0.8 + activity1 * 0.15;
    float heat2 = 0.7 + activity2 * 0.15;

    col = mix(col, star_col(heat1), surf1);
    col = mix(col, star_col(heat2), surf2);

    // Halos around stars
    col += star_col(0.6) * exp(-d1 * d1 * 80.0) * 0.5;
    col += star_col(0.5) * exp(-d2 * d2 * 80.0) * 0.5;

    // Plasma stream along the line between stars — L1 Lagrangian mass transfer
    vec2 axisVec = star2 - star1;
    float axisLen = length(axisVec);
    vec2 axisDir = axisVec / axisLen;
    vec2 perp = vec2(-axisDir.y, axisDir.x);
    // Project p onto axis
    vec2 toP = p - star1;
    float along = dot(toP, axisDir);
    float perpDist = abs(dot(toP, perp));

    // Only draw plasma between stars + slightly outside
    if (along > STAR_RADIUS && along < axisLen - STAR_RADIUS) {
        float alongT = along / axisLen;   // [0,1]
        // Accumulate plasma "filaments" as sum of wavy displacement samples
        float plasmaSum = 0.0;
        for (int s = 0; s < PLASMA_SAMPLES; s++) {
            float fs = float(s);
            float sT = fs / float(PLASMA_SAMPLES);
            // Wavy offset at this sample point
            float wave = sin(sT * 10.0 + x_Time * 2.0 + fs * 0.3) * 0.018
                       * sin(alongT * 3.14);  // max displacement mid-stream
            float sampleDist = abs(perpDist - abs(wave));
            // Distance from fragment to this sample  point
            plasmaSum += exp(-sampleDist * sampleDist * 2500.0);
        }
        plasmaSum /= float(PLASMA_SAMPLES);
        // Brightness peaks toward center of stream (mass transfer narrows)
        plasmaSum *= sin(alongT * 3.14);
        vec3 streamCol = plasma_col(fract(alongT * 1.5 + x_Time * 0.08));
        col += streamCol * plasmaSum * 1.4;
    }

    // Background starfield (sparse)
    vec2 sg = floor(p * 120.0);
    float sh = ts_hash(sg);
    if (sh > 0.996) {
        col += vec3(0.9, 0.9, 1.0) * (sh - 0.996) * 200.0 * 0.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
