// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Nebula swirl — slowly rotating nebula with multi-arm structure + star forming regions

const int   OCTAVES = 7;
const int   ARMS    = 3;
const float INTENSITY = 0.55;

vec3 nsw_pal(float t) {
    vec3 deep   = vec3(0.02, 0.04, 0.18);
    vec3 vio    = vec3(0.40, 0.15, 0.80);
    vec3 mag    = vec3(0.90, 0.25, 0.65);
    vec3 orange = vec3(1.00, 0.55, 0.25);
    vec3 gold   = vec3(1.00, 0.85, 0.40);
    vec3 cyan   = vec3(0.20, 0.85, 0.95);
    float s = mod(t * 6.0, 6.0);
    if (s < 1.0)      return mix(deep, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, orange, s - 2.0);
    else if (s < 4.0) return mix(orange, gold, s - 3.0);
    else if (s < 5.0) return mix(gold, cyan, s - 4.0);
    else              return mix(cyan, deep, s - 5.0);
}

float nsw_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float nsw_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  nsw_hash2(float n) { return vec2(nsw_hash1(n), nsw_hash1(n * 1.37 + 11.0)); }

float nsw_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(nsw_hash(i), nsw_hash(i + vec2(1,0)), u.x),
               mix(nsw_hash(i + vec2(0,1)), nsw_hash(i + vec2(1,1)), u.x), u.y);
}

float nsw_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * nsw_noise(p);
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

    vec3 col = vec3(0.0);

    // Slow rotation of arm structure
    float rotA = a + x_Time * 0.04;
    // Multi-arm logarithmic spiral
    float armStrength = 0.0;
    for (int arm = 0; arm < ARMS; arm++) {
        float armOffset = float(arm) / float(ARMS) * 6.28318;
        float armAng = rotA - log(r + 0.05) * 0.8 + armOffset;
        armAng = mod(armAng, 6.28318);
        float armDist = min(armAng, 6.28318 - armAng);
        float armMask = exp(-armDist * armDist * 8.0);
        armStrength = max(armStrength, armMask);
    }

    // Density FBM  with swirl
    vec2 swirlP = vec2(cos(rotA), sin(rotA)) * r * 5.0;
    float density = nsw_fbm(swirlP) * armStrength;
    density = pow(density, 1.5);

    // Radial falloff
    density *= smoothstep(1.2, 0.05, r);

    vec3 nebCol = nsw_pal(fract(density * 1.5 + r * 0.3 + x_Time * 0.02));
    col += nebCol * density * 1.5;

    // Bright filaments where density has steep gradient
    float dx = nsw_fbm(swirlP + vec2(0.01, 0.0)) - nsw_fbm(swirlP);
    float dy = nsw_fbm(swirlP + vec2(0.0, 0.01)) - nsw_fbm(swirlP);
    float gradMag = length(vec2(dx, dy)) * 100.0;
    col += nsw_pal(fract(density + 0.3)) * smoothstep(0.5, 1.2, gradMag) * density * 0.8;

    // Star-forming regions — bright pink spots
    for (int k = 0; k < 8; k++) {
        float fk = float(k);
        vec2 regPos = nsw_hash2(fk * 3.71) * 0.8 - 0.4;
        float regD = length(p - regPos);
        float reg = exp(-regD * regD * 1500.0);
        col += vec3(0.95, 0.35, 0.60) * reg * 1.2 * armStrength;
    }

    // Individual stars
    for (int s = 0; s < 40; s++) {
        float fs = float(s);
        vec2 sp = nsw_hash2(fs * 7.3) * 2.0 - 1.0;
        float sd = length(p - sp);
        col += vec3(0.9, 0.9, 1.0) * exp(-sd * sd * 30000.0) * 0.8;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
