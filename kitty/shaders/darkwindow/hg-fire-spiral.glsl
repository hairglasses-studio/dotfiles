// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Fire spiral — flames twisted into a vortex pattern with cool bluish edges + hot white core

const int   OCTAVES = 6;
const float INTENSITY = 0.55;

vec3 fs_col(float heat) {
    vec3 blue = vec3(0.25, 0.55, 0.98);
    vec3 vio  = vec3(0.55, 0.30, 0.95);
    vec3 mag  = vec3(0.95, 0.30, 0.70);
    vec3 orange = vec3(1.00, 0.55, 0.20);
    vec3 yellow = vec3(1.00, 0.90, 0.45);
    vec3 white  = vec3(1.00, 0.98, 0.90);
    if (heat < 0.2)      return mix(blue, vio, heat * 5.0);
    else if (heat < 0.4) return mix(vio, mag, (heat - 0.2) * 5.0);
    else if (heat < 0.6) return mix(mag, orange, (heat - 0.4) * 5.0);
    else if (heat < 0.8) return mix(orange, yellow, (heat - 0.6) * 5.0);
    else                 return mix(yellow, white, (heat - 0.8) * 5.0);
}

float fs_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float fs_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(fs_hash(i), fs_hash(i + vec2(1,0)), u.x),
               mix(fs_hash(i + vec2(0,1)), fs_hash(i + vec2(1,1)), u.x), u.y);
}

float fs_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * fs_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    float r = length(p);
    float a = atan(p.y, p.x);

    // Swirl angle increases with radius (inverse for tight spiral)
    float swirl = a + log(r + 0.05) * 2.0 + x_Time * 0.5;
    vec2 spiralP = vec2(cos(swirl), sin(swirl)) * r * 6.0;

    // FBM heat field in swirled coords
    float heat = fs_fbm(spiralP + vec2(0.0, x_Time * 2.0));
    heat = smoothstep(0.35, 0.85, heat);

    // Radius-based heat falloff
    heat *= (1.0 - smoothstep(0.0, 0.5, r));
    // Boost near core
    heat += exp(-r * r * 40.0) * 0.4;

    // Apply vortex pulse
    heat *= 0.85 + 0.2 * sin(x_Time * 0.8);

    vec3 col = fs_col(heat);

    // Embers at edges of spiral
    float embers = fs_hash(vec2(floor(p.x * 80.0), floor(p.y * 80.0) + floor(x_Time * 4.0)));
    if (embers > 0.98 && heat > 0.2) {
        col += vec3(1.0) * (embers - 0.98) * 50.0 * heat;
    }

    // Cool rim glow at outer vortex
    float rim = exp(-(r - 0.4) * (r - 0.4) * 80.0);
    col += fs_col(0.1) * rim * 0.3;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
