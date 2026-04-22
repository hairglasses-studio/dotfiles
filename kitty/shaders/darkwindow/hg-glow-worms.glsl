// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Glow worms on cave ceiling — thousands of tiny light points + hanging silk threads

const int   WORMS  = 200;
const int   SILK   = 15;
const float INTENSITY = 0.55;

vec3 gw_pal(float t) {
    vec3 cyan  = vec3(0.30, 0.95, 0.98);
    vec3 mint  = vec3(0.40, 0.95, 0.70);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 white = vec3(0.95, 0.98, 1.00);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, mint, s);
    else if (s < 2.0) return mix(mint, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, white, s - 2.0);
    else              return mix(white, cyan, s - 3.0);
}

float gw_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  gw_hash2(float n) { return vec2(gw_hash(n), gw_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Cave ceiling + water at bottom
    vec3 col = mix(vec3(0.03, 0.02, 0.08), vec3(0.01, 0.03, 0.10), uv.y);

    // Worms: concentrated near top of screen
    for (int i = 0; i < WORMS; i++) {
        float fi = float(i);
        float seed = fi * 3.71;
        vec2 pos = gw_hash2(seed) * vec2(2.4, 0.8);
        pos.x -= 1.2;
        pos.x *= x_WindowSize.x / x_WindowSize.y;
        pos.y = 0.1 + pos.y;   // Upper 80%

        // Subtle drift
        pos += 0.01 * vec2(sin(x_Time * 0.3 + fi), cos(x_Time * 0.25 + fi * 1.3));

        float d = length(p - pos);
        if (d > 0.03) continue;

        // Blink pattern
        float blinkF = 0.3 + gw_hash(seed * 5.1) * 0.5;
        float blinkPhase = fract(x_Time * blinkF + gw_hash(seed * 3.7));
        float blink = pow(1.0 - abs(blinkPhase - 0.5) * 2.0, 3.0);

        float size = 0.0015;
        float core = exp(-d * d / (size * size) * 2.0);
        float halo = exp(-d * d * 5000.0) * 0.3;
        vec3 wormCol = gw_pal(fract(seed * 0.03 + x_Time * 0.02));
        col += wormCol * (core * 1.3 + halo) * blink;
        col += vec3(1.0) * core * blink * 0.5;
    }

    // Silk threads hanging down from some worms
    for (int s = 0; s < SILK; s++) {
        float fs = float(s);
        float seed = fs * 7.31;
        vec2 anchor = gw_hash2(seed) * vec2(2.4, 0.4);
        anchor.x -= 1.2;
        anchor.x *= x_WindowSize.x / x_WindowSize.y;
        anchor.y = 0.3 + anchor.y;   // Upper region
        float threadLen = 0.25 + gw_hash(seed * 3.7) * 0.3;
        float swing = 0.01 * sin(x_Time * 0.5 + seed * 3.0);
        vec2 bottom = anchor + vec2(swing, -threadLen);

        // Distance from p to this vertical segment
        vec2 pa = p - anchor;
        vec2 ba = bottom - anchor;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        float d = length(pa - ba * h);

        if (d < 0.003) {
            float thread = smoothstep(0.003, 0.0, d);
            col += gw_pal(fract(fs * 0.1 + x_Time * 0.03)) * thread * 0.4;
        }

        // Bottom end droplet
        float drop_d = length(p - bottom);
        if (drop_d < 0.01) {
            col += gw_pal(fract(fs * 0.1)) * exp(-drop_d * drop_d * 6000.0) * 0.8;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
