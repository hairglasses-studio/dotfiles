// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Fractal feather — recursive feather-vane pattern with iridescent shimmer

const int   FEATHER_LVLS = 4;
const int   FEATHERS     = 5;
const float INTENSITY    = 0.55;

vec3 ff_pal(float t) {
    vec3 blue_deep = vec3(0.10, 0.25, 0.95);
    vec3 cyan      = vec3(0.20, 0.90, 0.98);
    vec3 mint      = vec3(0.25, 0.95, 0.65);
    vec3 gold      = vec3(0.95, 0.80, 0.35);
    vec3 mag       = vec3(0.90, 0.30, 0.60);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(blue_deep, cyan, s);
    else if (s < 2.0) return mix(cyan, mint, s - 1.0);
    else if (s < 3.0) return mix(mint, gold, s - 2.0);
    else if (s < 4.0) return mix(gold, mag, s - 3.0);
    else              return mix(mag, blue_deep, s - 4.0);
}

// Distance to a feather "barb" (shorter segment) at angle branchAng from axis
float barbDist(vec2 p, vec2 axisStart, vec2 axisDir, float axisLen, float branchAngle, float branchLen) {
    // Pick anchor along axis
    // Multiple barbs — each anchor at different t
    float minD = 1e9;
    for (int i = 1; i < 8; i++) {
        float t = float(i) / 8.0;
        vec2 anchor = axisStart + axisDir * axisLen * t;
        float cr = cos(branchAngle), sr = sin(branchAngle);
        vec2 perp = vec2(-axisDir.y, axisDir.x);
        vec2 branchDir = axisDir * cr + perp * sr;
        vec2 tipR = anchor + branchDir * branchLen;
        // Also mirror branch on other side
        vec2 branchDirL = axisDir * cr - perp * sr;
        vec2 tipL = anchor + branchDirL * branchLen;

        vec2 pa = p - anchor;
        vec2 ab = tipR - anchor;
        float hR = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
        float dR = length(pa - ab * hR);
        minD = min(minD, dR);

        vec2 abL = tipL - anchor;
        float hL = clamp(dot(pa, abL) / dot(abL, abL), 0.0, 1.0);
        float dL = length(pa - abL * hL);
        minD = min(minD, dL);
    }
    return minD;
}

// Main feather: a quill plus a field of barbs
float featherDist(vec2 p, vec2 quillStart, vec2 quillEnd) {
    vec2 axisDir = normalize(quillEnd - quillStart);
    float axisLen = length(quillEnd - quillStart);

    // Quill line
    vec2 pa = p - quillStart;
    float h = clamp(dot(pa, axisDir) / axisLen, 0.0, 1.0);
    vec2 proj = quillStart + axisDir * axisLen * h;
    float quillD = length(p - proj);

    // Barbs at ~60° (feather-like)
    float barb1 = barbDist(p, quillStart, axisDir, axisLen, 0.52, axisLen * 0.25);
    float barb2 = barbDist(p, quillStart, axisDir, axisLen, 0.7, axisLen * 0.12);

    return min(quillD, min(barb1, barb2));
}

float ff_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.03);

    float minD = 1e9;
    float colorHue = 0.0;

    // Multiple feathers radiating from center
    for (int f = 0; f < FEATHERS; f++) {
        float ff = float(f);
        float baseAngle = ff / float(FEATHERS) * 6.28318 + x_Time * 0.08;
        vec2 dir = vec2(cos(baseAngle), sin(baseAngle));
        vec2 quillEnd = dir * 0.45;
        float d = featherDist(p, vec2(0.0), quillEnd);
        if (d < minD) {
            minD = d;
            colorHue = ff / float(FEATHERS);
        }
    }

    // Feather body with iridescent color
    float core = smoothstep(0.003, 0.0, minD);
    float glow = exp(-minD * minD * 1500.0) * 0.4;
    vec3 featherCol = ff_pal(fract(colorHue + minD * 0.5 + x_Time * 0.05));
    col += featherCol * (core * 1.2 + glow);

    // Central pulse
    float center = exp(-length(p) * length(p) * 200.0) * 0.8;
    col += ff_pal(0.0) * center;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
