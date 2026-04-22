// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Nested matryoshka rings — 8 concentric rotating rings with counter-rotation + phase-shifted color

const int   RING_COUNT = 8;
const int   SEGMENTS = 16;
const float INTENSITY = 0.55;

vec3 mr_pal(float t) {
    vec3 a = vec3(0.95, 0.30, 0.70);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(0.55, 0.30, 0.98);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    float r = length(p);
    float a = atan(p.y, p.x);

    vec3 col = vec3(0.01, 0.01, 0.04);

    for (int i = 0; i < RING_COUNT; i++) {
        float fi = float(i);
        float ringR = 0.06 + fi * 0.05;
        float ringW = 0.012;
        float rDist = abs(r - ringR);
        if (rDist > ringW * 2.0) continue;

        // Counter-rotation: odd rings go opposite direction
        float rotSign = mod(fi, 2.0) < 1.0 ? 1.0 : -1.0;
        float rotSpeed = rotSign * (0.1 + fi * 0.05);
        float segAngle = 6.28318 / float(SEGMENTS);
        float rotAng = a + x_Time * rotSpeed;
        float seg = mod(rotAng, segAngle);
        float segCenter = segAngle * 0.5;
        float segDist = abs(seg - segCenter);

        // Gap pattern (some segments missing)
        float segIdx = floor(rotAng / segAngle);
        float segHash = fract(sin(segIdx * 12.9898 + fi * 78.233) * 43758.5);
        float presence = step(0.3, segHash);

        float segMask = smoothstep(segAngle * 0.4, segAngle * 0.2, segDist) * presence;
        float ringMask = exp(-rDist * rDist / (ringW * ringW) * 2.0) * segMask;
        float ringGlow = exp(-rDist * rDist * 800.0) * 0.2 * segMask;

        vec3 ringCol = mr_pal(fract(fi * 0.13 + segIdx * 0.02 + x_Time * 0.03));
        col += ringCol * (ringMask * 0.8 + ringGlow);
    }

    // Central bright pulse
    float pulse = 0.6 + 0.4 * sin(x_Time * 1.0);
    col += mr_pal(fract(x_Time * 0.05)) * exp(-r * r * 400.0) * pulse * 1.0;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
