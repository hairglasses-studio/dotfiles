// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Cosmic dust cloud — turbulent 3D dust with embedded stars + light-year depth

const int   STARS = 30;
const int   OCTAVES = 6;
const float INTENSITY = 0.55;

vec3 cd_pal(float t) {
    vec3 a = vec3(0.05, 0.03, 0.12);
    vec3 b = vec3(0.35, 0.15, 0.55);
    vec3 c = vec3(0.80, 0.30, 0.55);
    vec3 d = vec3(1.00, 0.65, 0.35);
    vec3 e = vec3(0.20, 0.80, 0.95);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

float cd_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float cd_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  cd_hash2(float n) { return vec2(cd_hash1(n), cd_hash1(n * 1.37 + 11.0)); }

float cd_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(cd_hash(i), cd_hash(i + vec2(1,0)), u.x),
               mix(cd_hash(i + vec2(0,1)), cd_hash(i + vec2(1,1)), u.x), u.y);
}

float cd_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * cd_noise(p);
        p = p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.03);

    // Multiple dust layers at different depths
    for (int L = 0; L < 3; L++) {
        float fL = float(L);
        float depth = 0.5 + fL * 0.3;
        float scale = 1.2 + fL * 0.6;
        vec2 dp = p * scale + vec2(x_Time * 0.03 * (fL + 1.0), 0.0);
        float density = cd_fbm(dp);
        density = pow(max(0.0, density - 0.4), 1.5);
        vec3 dc = cd_pal(fract(fL * 0.2 + density + x_Time * 0.04));
        col += dc * density * (0.3 - fL * 0.08);
    }

    // Bright filaments along high-density boundaries
    vec2 fp = p * 2.0;
    float fil = cd_fbm(fp + x_Time * 0.05);
    float filEdge = smoothstep(0.45, 0.55, fil);
    col += cd_pal(0.4) * filEdge * 0.3;

    // Embedded stars
    for (int s = 0; s < STARS; s++) {
        float fs = float(s);
        vec2 starP = cd_hash2(fs * 3.71) * 2.0 - 1.0;
        starP.x *= x_WindowSize.x / x_WindowSize.y;
        // Slight drift
        starP += 0.02 * vec2(sin(x_Time * 0.1 + fs), cos(x_Time * 0.08 + fs * 1.3));
        float sd = length(p - starP);
        float twinkle = 0.5 + 0.5 * sin(x_Time * (2.0 + cd_hash1(fs)) + fs);
        float starSize = 0.003 + cd_hash1(fs * 3.7) * 0.005;
        float core = exp(-sd * sd / (starSize * starSize) * 2.0);
        float halo = exp(-sd * sd * 800.0) * 0.15;
        vec3 sc = cd_pal(fract(cd_hash1(fs * 5.3) + x_Time * 0.04));
        col += sc * (core * (0.8 + twinkle) + halo);

        // Diffraction spikes on brightest stars
        if (twinkle > 0.8) {
            float spike = exp(-pow(p.x - starP.x, 2.0) * 10000.0) * step(abs(p.y - starP.y), 0.02);
            spike += exp(-pow(p.y - starP.y, 2.0) * 10000.0) * step(abs(p.x - starP.x), 0.02);
            col += sc * spike * twinkle * 0.3;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
