// Shader attribution: fielding (https://github.com/fielding/ghostty-shader-adventures)
// License: no upstream LICENSE — personal/non-commercial use, attribution preserved.
// Ported to DarkWindow by hairglasses — iTimeCursorChange stubbed, iFrame derived from x_Time.
// (Post-FX) — Cursor-localized digital interference (scanline tear + RGB split + static)
// Stack after any base shader; at a distance from cursor it passes through unchanged.

const float GLITCH_RADIUS  = 0.15;
const float GLITCH_BLOOM   = 0.10;
const float TEAR_STRENGTH  = 0.04;
const float RGB_SPLIT_MAX  = 0.005;
const float STATIC_DENSITY = 0.03;

float cg_hash11(float p) {
    return fract(sin(p * 127.1) * 43758.5453123);
}

float cg_hash21(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453123);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;

    // Stub iTimeCursorChange → perpetual mid-heat pulse (sinusoidal), gives a living glitch
    // instead of a keystroke-tied one. 0.3–0.9 range keeps it visible but not overwhelming.
    float heat = 0.6 + 0.3 * sin(x_Time * 2.3);

    // Derive iFrame from x_Time (60fps approximation)
    int frame = int(x_Time * 60.0);

    // Cursor distance (normalized by height)
    float curDist = length(x_PixelPos - x_CursorPos) / x_WindowSize.y;
    float radius = GLITCH_RADIUS + GLITCH_BLOOM * heat;
    float intensity = smoothstep(radius, radius * 0.1, curDist) * heat;

    vec2 displaced = uv;

    if (intensity > 0.01) {
        // Scanline tear — coherent per line, refreshed ~15 times/sec
        float scanSeed = float(int(x_PixelPos.y) * 13 + frame / 4);
        float lineHash = cg_hash11(scanSeed);
        if (lineHash < intensity * 0.5) {
            float tearDir = (lineHash < intensity * 0.25) ? 1.0 : -1.0;
            displaced.x += tearDir * intensity * TEAR_STRENGTH * (0.3 + lineHash * 2.0);
        }
        // 8px block jump
        float blockSeed = floor(x_PixelPos.y / 8.0) * 7.13 + float(frame / 6);
        float blockHash = cg_hash11(blockSeed);
        if (blockHash < intensity * 0.1) {
            displaced.y += (blockHash * 2.0 - 1.0) * intensity * 0.03;
        }
    }

    displaced = clamp(displaced, 0.0, 1.0);

    // RGB channel separation
    float split = intensity * RGB_SPLIT_MAX;
    vec3 color;
    color.r = x_Texture(clamp(displaced + vec2(split, 0.0), 0.0, 1.0)).r;
    color.g = x_Texture(displaced).g;
    color.b = x_Texture(clamp(displaced - vec2(split, 0.0), 0.0, 1.0)).b;

    // Sparse digital static
    if (intensity > 0.05) {
        float noise = cg_hash21(x_PixelPos + float(frame) * 0.173);
        float threshold = 1.0 - intensity * STATIC_DENSITY;
        if (noise > threshold) {
            float bright = step(0.5, fract(noise * 7.0));
            color = mix(color, vec3(bright), 0.6 * intensity);
        }
    }

    _wShaderOut = vec4(color, 1.0);
}
