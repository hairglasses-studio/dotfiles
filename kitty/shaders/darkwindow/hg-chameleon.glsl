// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Chameleon iridescence — slow hue wave traveling across screen + viewing-angle shift

const int   OCTAVES = 5;
const float INTENSITY = 0.5;

vec3 ch_pal(float t) {
    vec3 a = vec3(0.20, 0.90, 0.50);
    vec3 b = vec3(0.10, 0.70, 0.95);
    vec3 c = vec3(0.55, 0.30, 0.98);
    vec3 d = vec3(0.95, 0.25, 0.65);
    vec3 e = vec3(0.95, 0.65, 0.25);
    vec3 f = vec3(0.96, 0.90, 0.40);
    float s = mod(t * 6.0, 6.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else if (s < 5.0) return mix(e, f, s - 4.0);
    else              return mix(f, a, s - 5.0);
}

float ch_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float ch_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(ch_hash(i), ch_hash(i + vec2(1,0)), u.x),
               mix(ch_hash(i + vec2(0,1)), ch_hash(i + vec2(1,1)), u.x), u.y);
}

float ch_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * ch_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Large, slow-moving wave
    float waveFreq = 2.0;
    float wavePhase = dot(p, normalize(vec2(1.0, 0.3))) * waveFreq + x_Time * 0.15;

    // FBM perturbation for texture
    float fbm = ch_fbm(p * 3.0 + x_Time * 0.05);

    // Viewing-angle approximation via cursor position
    vec2 cursorNorm = vec2(
        x_CursorPos.x / x_WindowSize.x * 2.0 - 1.0,
        x_CursorPos.y / x_WindowSize.y * 2.0 - 1.0
    );
    float viewShift = dot(cursorNorm, normalize(p)) * 0.2;

    // Hue drifts with wave + viewing + fbm
    float hue = fract(wavePhase + fbm * 0.3 + viewShift + x_Time * 0.03);

    vec3 col = ch_pal(hue);

    // Scale pattern — tiny polygonal cells (scales)
    vec2 scaleP = p * 30.0;
    vec2 scaleId = floor(scaleP);
    // Hex-like offset per row
    scaleId.x += mod(scaleId.y, 2.0) * 0.5;
    vec2 scaleCenter = scaleId + 0.5;
    vec2 scaleDiff = scaleP - scaleCenter;
    float scaleDist = length(scaleDiff);
    // Scale shape — soft circle with edge
    float scaleMask = smoothstep(0.45, 0.35, scaleDist);
    float scaleEdge = exp(-scaleDist * scaleDist * 30.0) * 0.2;

    // Per-scale hue variation
    float scaleSeed = ch_hash(scaleId);
    vec3 scaleCol = ch_pal(fract(hue + scaleSeed * 0.1));
    col = mix(col, scaleCol, scaleMask * 0.7);
    col += scaleCol * scaleEdge;

    // Subtle sheen — bright specular along wave crest
    float sheen = smoothstep(0.45, 0.55, fract(wavePhase));
    col += vec3(1.0) * sheen * 0.3;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility * 0.85);

    _wShaderOut = vec4(result, 1.0);
}
