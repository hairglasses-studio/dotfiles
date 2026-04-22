// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Spiral staircase — 3D helical steps descending into depth with neon rim lighting

const int   STEPS   = 20;
const float INTENSITY = 0.55;

vec3 ss_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.95, 0.30, 0.70);
    vec3 d = vec3(0.96, 0.80, 0.35);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

vec2 project3to2(vec3 v) {
    float zDist = 1.5;
    float z = 1.0 / (zDist - v.z);
    return v.xy * z;
}

// Rectangle on plane SDF at center, with width/height
float sdRectAt(vec2 p, vec2 center, vec2 halfSize) {
    vec2 d = abs(p - center) - halfSize;
    return length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.03);

    // Camera descends down the staircase
    float t = x_Time * 0.3;

    // Each step is at a specific height and angular position
    float minD = 1e9;
    int nearestStep = 0;
    for (int i = 0; i < STEPS; i++) {
        float fi = float(i);
        // Step y: from -0.5 at top to descending
        float stepY = 0.3 - fi * 0.08 - t * 0.4;
        // Normalize onto visible range
        float cycleY = mod(stepY, 3.0);
        stepY = 0.5 - cycleY;   // wraps

        // Angular position
        float stepAngle = fi * 0.8 + t;
        // Step position in 3D (radius = 0.4, varying y and angle)
        vec3 stepPos3 = vec3(cos(stepAngle) * 0.4, stepY, sin(stepAngle) * 0.4);

        // Only draw front-facing steps (z > camera z)
        if (stepPos3.z < -1.3) continue;

        vec2 stepPos2 = project3to2(stepPos3);
        // Step is rectangular
        vec2 halfSize = vec2(0.05, 0.015) / (1.5 - stepPos3.z);
        float d = sdRectAt(p, stepPos2, halfSize);
        if (d < minD) {
            minD = d;
            nearestStep = i;
        }

        // Render step
        float stepCore = smoothstep(halfSize.x * 0.1, 0.0, d);
        float stepEdge = smoothstep(halfSize.x * 0.05, 0.0, abs(d)) * 1.5;
        float depthFade = exp(stepPos3.z * 0.4);
        vec3 stepCol = ss_pal(fract(fi * 0.08 + x_Time * 0.04));
        // Fill
        if (d < 0.0) col = mix(col, stepCol * 0.3, 0.6 * depthFade);
        // Edge
        col += stepCol * stepEdge * depthFade * 0.8;
        // Glow
        col += stepCol * exp(-d * d * 800.0) * depthFade * 0.2;
    }

    // Central vertical axis
    float axisD = abs(p.x);
    if (axisD < 0.005) {
        col += ss_pal(0.0) * smoothstep(0.005, 0.0, axisD) * 0.4;
    }

    // Descent motion haze — vertical streaks suggesting motion
    float streak = 0.5 + 0.5 * sin(p.y * 30.0 + t * 15.0);
    streak = pow(streak, 30.0);
    col += ss_pal(fract(x_Time * 0.05)) * streak * 0.04;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
