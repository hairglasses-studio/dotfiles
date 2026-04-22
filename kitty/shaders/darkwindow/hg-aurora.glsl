// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Aurora borealis — vertical ribbon curtains with chromatic shimmer

const int   AURORA_BANDS = 5;
const float INTENSITY    = 0.55;

vec3 au_pal(float t) {
    vec3 a = vec3(0.10, 0.95, 0.55);   // signature aurora green
    vec3 b = vec3(0.20, 0.75, 0.95);   // cyan
    vec3 c = vec3(0.55, 0.30, 0.98);   // violet
    vec3 d = vec3(0.90, 0.30, 0.65);   // pink
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float au_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
}

float au_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(
        mix(au_hash(i), au_hash(i + vec2(1,0)), u.x),
        mix(au_hash(i + vec2(0,1)), au_hash(i + vec2(1,1)), u.x),
        u.y);
}

// 5-octave FBM for curtain distortion
float au_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 5; i++) {
        v += a * au_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

// Single curtain: narrow horizontal band with vertical fringe, undulating x-axis
float auroraCurtain(vec2 p, float yCenter, float height, float phase, float t) {
    // Horizontal undulation — different frequency per band
    float xWave = au_fbm(vec2(p.x * 3.0 + phase, t * 0.3 + phase)) * 0.15;
    float yOffset = p.y - yCenter - xWave;

    // Vertical fringe fade — bright at center, tapering top/bottom
    float bandFalloff = exp(-yOffset * yOffset / (height * height) * 2.0);

    // Vertical "rays" — narrow streaks pointing down from the band
    float rayPhase = p.x * 20.0 + t * 0.5 + phase * 3.0;
    float rayFreq = au_fbm(vec2(p.x * 2.5, t * 0.2 + phase));
    float rays = 0.5 + 0.5 * sin(rayPhase + rayFreq * 6.28);
    // Ray direction bias — streaks only visible below band
    float rayBias = smoothstep(0.0, -height * 1.5, yOffset);
    float rayMask = rays * rayBias;

    return bandFalloff + rayMask * bandFalloff * 0.7;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos / x_WindowSize.y) - vec2(x_WindowSize.x / x_WindowSize.y * 0.5, 0.5);

    vec3 col = vec3(0.0);

    // Multiple bands at different heights, each with its own phase
    for (int i = 0; i < AURORA_BANDS; i++) {
        float fi = float(i);
        float yCenter = 0.1 + fi * 0.1 + 0.05 * sin(x_Time * 0.2 + fi);
        float height = 0.04 + fi * 0.015;
        float phase = fi * 2.7;
        float intensity = 1.0 - fi * 0.15;   // nearer bands brighter
        float curtain = auroraCurtain(p, yCenter, height, phase, x_Time) * intensity;

        // Color drifts across bands + time
        vec3 bc = au_pal(fract(fi * 0.18 + x_Time * 0.04));
        col += bc * curtain * 0.5;
    }

    // Subtle base sky glow (bottom-heavy so text stays readable)
    float groundGlow = smoothstep(0.0, -0.3, p.y) * 0.08;
    col += au_pal(fract(x_Time * 0.03)) * groundGlow;

    // Sparse starfield
    vec2 starP = floor(p * 80.0);
    float starH = au_hash(starP);
    if (starH > 0.99) {
        float twinkle = 0.5 + 0.5 * sin(x_Time * (1.0 + starH * 5.0) + starH * 20.0);
        col += vec3(0.9, 0.85, 1.0) * twinkle * 0.6;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.8, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
