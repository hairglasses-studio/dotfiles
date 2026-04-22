// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Ion trap — single ion confined between 4 RF electrodes with wobbling motion + field lines

const int   ELECTRODES = 4;
const int   FIELD_LINES = 32;
const float INTENSITY = 0.55;

vec3 it_pal(float t) {
    vec3 cyan = vec3(0.20, 0.85, 0.95);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 white = vec3(0.95, 0.98, 1.00);
    vec3 mag  = vec3(0.95, 0.30, 0.70);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, white, s - 1.0);
    else if (s < 3.0) return mix(white, mag, s - 2.0);
    else              return mix(mag, cyan, s - 3.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.02, 0.05);

    // 4 electrodes at corners of square
    for (int e = 0; e < ELECTRODES; e++) {
        float fe = float(e);
        float ang = fe * 1.5708 + 0.785;  // π/4 offset
        vec2 pos = vec2(cos(ang), sin(ang)) * 0.35;
        float d = length(p - pos);
        // Electrode body (metal tip)
        if (d < 0.04) {
            col = vec3(0.5, 0.55, 0.65);
            // Polarity indicator (RF alternates)
            float polarity = sin(x_Time * 10.0 + fe * 1.5708);
            col += (polarity > 0.0 ? vec3(0.5, 0.1, 0.1) : vec3(0.1, 0.3, 0.5)) * smoothstep(0.04, 0.02, d);
        }
        // Glow around electrode
        col += it_pal(0.0) * exp(-d * d * 100.0) * 0.2;
    }

    // Field lines between electrodes — curved paths
    for (int i = 0; i < FIELD_LINES; i++) {
        float fi = float(i);
        float srcAng = fi / float(FIELD_LINES) * 6.28;
        vec2 srcPoint = vec2(cos(srcAng), sin(srcAng)) * 0.35;
        // Target = opposite electrode
        vec2 dstPoint = -srcPoint;

        // Approximate curved path with midpoint displacement (bends toward center)
        vec2 midPoint = (srcPoint + dstPoint) * 0.5;
        // Add perpendicular bend toward center
        float bend = 0.15;
        vec2 perp = normalize(vec2(-srcPoint.y, srcPoint.x));
        vec2 curved = midPoint + perp * bend * sin(fi * 0.3);

        // Sample along bezier-like curve (3 straight segments)
        float minD = 1e9;
        vec2 a = srcPoint;
        vec2 b = curved;
        vec2 c = dstPoint;
        // Segment 1
        vec2 pa = p - a;
        vec2 ba = b - a;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        minD = min(minD, length(pa - ba * h));
        // Segment 2
        pa = p - b;
        ba = c - b;
        h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        minD = min(minD, length(pa - ba * h));

        if (minD < 0.002) {
            // Traveling pulse
            float pulsePhase = fract(x_Time * 0.8 + fi * 0.05);
            float pulseD = length(p - mix(a, c, pulsePhase));
            float pulseMask = exp(-pulseD * pulseD * 3000.0);
            col += it_pal(fract(fi * 0.03 + x_Time * 0.04)) * smoothstep(0.002, 0.0, minD) * 0.35;
            col += vec3(1.0) * pulseMask * 0.9;
        }
    }

    // Confined ion at center — micromotion (tiny wobble)
    vec2 ionPos = vec2(
        0.01 * sin(x_Time * 50.0),
        0.01 * cos(x_Time * 47.3)
    );
    float ionD = length(p - ionPos);
    float ionCore = exp(-ionD * ionD * 15000.0);
    float ionHalo = exp(-ionD * ionD * 500.0) * 0.4;
    col += vec3(1.0, 0.95, 0.85) * ionCore * 2.0;
    col += it_pal(fract(x_Time * 0.1)) * ionHalo;

    // Laser cooling beam (horizontal)
    float laserD = abs(p.y) * 1000.0;
    float laser = exp(-laserD * laserD) * smoothstep(0.4, 0.0, abs(p.x));
    col += vec3(0.3, 0.8, 0.95) * laser * 0.2;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
