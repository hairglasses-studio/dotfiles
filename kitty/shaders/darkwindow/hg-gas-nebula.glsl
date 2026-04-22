// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Dense gas nebula — thick H II clouds with dark absorption lanes + embedded hot stars

const int   OCTAVES = 7;
const int   STARS = 22;
const float INTENSITY = 0.55;

vec3 gn_pal(float t) {
    vec3 deep_red = vec3(0.70, 0.15, 0.20);
    vec3 mag      = vec3(0.95, 0.30, 0.55);
    vec3 pink     = vec3(1.00, 0.55, 0.70);
    vec3 gold     = vec3(1.00, 0.80, 0.30);
    vec3 cyan     = vec3(0.20, 0.85, 0.95);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(deep_red, mag, s);
    else if (s < 2.0) return mix(mag, pink, s - 1.0);
    else if (s < 3.0) return mix(pink, gold, s - 2.0);
    else if (s < 4.0) return mix(gold, cyan, s - 3.0);
    else              return mix(cyan, deep_red, s - 4.0);
}

float gn_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float gn_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  gn_hash2(float n) { return vec2(gn_hash1(n), gn_hash1(n * 1.37 + 11.0)); }

float gn_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(gn_hash(i), gn_hash(i + vec2(1,0)), u.x),
               mix(gn_hash(i + vec2(0,1)), gn_hash(i + vec2(1,1)), u.x), u.y);
}

float gn_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * gn_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.02, 0.01, 0.04);

    // Main emission cloud — dense pinkish gas
    vec2 fbmP = p * 2.0 + vec2(x_Time * 0.03, x_Time * 0.02);
    float density = gn_fbm(fbmP);
    density = pow(density, 1.3);
    col += gn_pal(fract(density * 1.5 + x_Time * 0.02)) * density * 1.3;

    // Secondary layer — different scale
    vec2 fbmP2 = p * 1.2 + vec2(x_Time * 0.025);
    float secondLayer = gn_fbm(fbmP2 + vec2(5.3, 2.1));
    col += gn_pal(fract(secondLayer + 0.3)) * secondLayer * 0.5;

    // Dark absorption lanes — inverse FBM
    float darkLaneFbm = gn_fbm(p * 3.0 + vec2(10.0));
    float darkLane = smoothstep(0.6, 0.4, darkLaneFbm);
    col *= 1.0 - darkLane * 0.6;

    // Embedded hot young stars
    for (int s = 0; s < STARS; s++) {
        float fs = float(s);
        vec2 starP = gn_hash2(fs * 3.71) * 2.0 - 1.0;
        starP.x *= x_WindowSize.x / x_WindowSize.y;
        float sd = length(p - starP);

        float brightness = 0.5 + gn_hash1(fs * 7.3) * 0.8;
        float twinkle = 0.6 + 0.4 * sin(x_Time * (2.0 + gn_hash1(fs)) + fs);

        float core = exp(-sd * sd * 8000.0) * brightness;
        float halo = exp(-sd * sd * 200.0) * 0.2;
        float farHalo = exp(-sd * sd * 20.0) * 0.05;

        col += vec3(1.0, 0.95, 0.85) * (core * (0.8 + twinkle * 0.4) + halo + farHalo);

        // Diffraction spikes on brightest stars
        if (brightness > 1.0) {
            float spikeX = exp(-pow(p.x - starP.x, 2.0) * 10000.0) * step(abs(p.y - starP.y), 0.025);
            float spikeY = exp(-pow(p.y - starP.y, 2.0) * 10000.0) * step(abs(p.x - starP.x), 0.025);
            col += vec3(0.95, 0.9, 1.0) * (spikeX + spikeY) * 0.2;
        }
    }

    // Bright central core
    float centerD = length(p);
    col += gn_pal(0.3) * exp(-centerD * centerD * 12.0) * 0.4;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
