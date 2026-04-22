// Shader attribution: fielding (https://github.com/fielding/ghostty-shader-adventures)
// License: no upstream LICENSE — personal/non-commercial use, attribution preserved.
// Ported to DarkWindow by hairglasses — iTimeCursorChange stubbed as fixed offset (steady heat).
// (Background) — Soft parallax FBM clouds drifting behind terminal text

// --- Tuning ---
const float CLOUD_OPACITY = 0.18;
const float CLOUD_SCALE   = 3.0;
const float DRIFT_SPEED   = 0.04;
const float TURBULENCE    = 0.02;
const float CURSOR_GLOW   = 0.12;
const float CURSOR_RANGE  = 0.2;

// --- Palette ---
const vec3 COL_PINK   = vec3(0.906, 0.204, 0.612);
const vec3 COL_CYAN   = vec3(0.102, 0.816, 0.839);
const vec3 COL_PURPLE = vec3(0.596, 0.443, 0.996);
const vec3 COL_GOLD   = vec3(0.949, 0.651, 0.200);
const vec3 COL_BLUE   = vec3(0.271, 0.541, 0.886);

float clouds_hash21(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453123);
}

float clouds_vnoise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(
        mix(clouds_hash21(i), clouds_hash21(i + vec2(1.0, 0.0)), u.x),
        mix(clouds_hash21(i + vec2(0.0, 1.0)), clouds_hash21(i + vec2(1.0, 1.0)), u.x),
        u.y
    );
}

float clouds_fbm(vec2 p, int octaves) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < 8; i++) {
        if (i >= octaves) break;
        v += a * clouds_vnoise(p);
        p = rot * p * 2.0 + 0.1;
        a *= 0.5;
    }
    return v;
}

float cloudLayer(vec2 uv, float t, float speed, float scale, int detail) {
    vec2 drift = vec2(t * speed, t * speed * 0.4);
    float n = clouds_fbm((uv + drift) * scale, detail);
    return smoothstep(0.35, 0.65, n);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 cloudUV = x_PixelPos / vec2(x_WindowSize.y, x_WindowSize.y);

    // Stub iTimeCursorChange → steady mid-heat (no typing reactivity available in DarkWindow)
    float timeSinceType = 1.5;
    float heat = smoothstep(3.0, 0.05, timeSinceType);

    // Layered clouds (back → front)
    float c1 = cloudLayer(cloudUV, x_Time, DRIFT_SPEED * 0.5, CLOUD_SCALE * 0.6, 6);
    vec3 col1 = mix(COL_PURPLE, COL_BLUE, 0.5) * c1;

    float c2 = cloudLayer(cloudUV, x_Time, DRIFT_SPEED * 0.8, CLOUD_SCALE * 0.9, 5);
    vec3 col2 = mix(COL_CYAN, COL_BLUE, 0.6) * c2;

    float c3 = cloudLayer(cloudUV, x_Time, DRIFT_SPEED * 1.2, CLOUD_SCALE * 1.4, 4);
    vec3 col3 = mix(COL_PINK, COL_GOLD, 0.4) * c3;

    vec3 clouds = col1 * 0.4 + col2 * 0.5 + col3 * 0.3;
    float cloudAlpha = max(max(c1 * 0.4, c2 * 0.5), c3 * 0.3);

    // Cursor interaction (DarkWindow provides x_CursorPos only, no size info)
    vec2 curPos = x_CursorPos;
    vec2 curOffset = (x_PixelPos - curPos) / x_WindowSize.y;
    float curDist = length(curOffset);

    float proximity = 1.0 - smoothstep(0.0, CURSOR_RANGE, curDist);
    clouds *= 1.0 + proximity * heat * 1.5;

    float glow = exp(-curDist * curDist * 80.0) * CURSOR_GLOW * heat;
    vec3 glowCol = mix(COL_GOLD, COL_PINK, sin(x_Time * 0.3) * 0.5 + 0.5);

    // Composite — keep text readable
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = CLOUD_OPACITY * (1.0 - termLuma * 0.8);

    vec3 result = mix(terminal.rgb, clouds, cloudAlpha * visibility);
    result += glowCol * glow;

    _wShaderOut = vec4(result, 1.0);
}
