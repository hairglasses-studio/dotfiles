// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Cosmic string — vibrating 1D topological defect through spacetime, with gravitational lensing around it

const int   STRING_SAMPS = 128;
const int   OCTAVES = 4;
const float INTENSITY = 0.55;

vec3 cs_pal(float t) {
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 mag  = vec3(0.95, 0.30, 0.70);
    vec3 white = vec3(0.95, 0.98, 1.00);
    vec3 cyan = vec3(0.20, 0.85, 0.95);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(vio, mag, s);
    else if (s < 2.0) return mix(mag, white, s - 1.0);
    else if (s < 3.0) return mix(white, cyan, s - 2.0);
    else              return mix(cyan, vio, s - 3.0);
}

float cs_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// 1D string shape: vibrating along length
vec2 stringPoint(float s, float t) {
    // Base: spans from left to right
    float x = -0.8 + s * 1.6;
    // Multiple oscillation modes superposed
    float y = 0.0;
    y += 0.08 * sin(s * 8.0 - t * 1.5);
    y += 0.04 * sin(s * 16.0 + t * 2.2);
    y += 0.02 * sin(s * 32.0 - t * 3.0);
    return vec2(x, y);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Lensing: pixels near string get offset toward it (fake gravitational lens)
    vec2 lensedUV = uv;
    float minStringD = 1e9;
    float closestS = 0.0;
    vec2 closestPoint = vec2(0.0);
    for (int i = 0; i < STRING_SAMPS - 1; i++) {
        float s1 = float(i) / float(STRING_SAMPS - 1);
        float s2 = float(i + 1) / float(STRING_SAMPS - 1);
        vec2 a = stringPoint(s1, x_Time);
        vec2 b = stringPoint(s2, x_Time);
        vec2 pa = p - a;
        vec2 ba = b - a;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        float d = length(pa - ba * h);
        if (d < minStringD) {
            minStringD = d;
            closestS = mix(s1, s2, h);
            closestPoint = a + ba * h;
        }
    }

    // Lensing offset: deflect sample position around string
    if (minStringD < 0.15) {
        vec2 toString = closestPoint - p;
        float lensStrength = 0.003 / (0.01 + minStringD * 5.0);
        lensedUV += normalize(vec2(-toString.y, toString.x)) * lensStrength;
    }

    // Sample terminal through lens
    vec4 terminal = x_Texture(lensedUV);

    vec3 col = terminal.rgb * 0.7;

    // String rendering — bright glowing line
    float stringCore = smoothstep(0.002, 0.0, minStringD);
    float stringGlow = exp(-minStringD * minStringD * 500.0) * 0.4;
    vec3 stringCol = cs_pal(fract(closestS * 2.0 + x_Time * 0.1));
    col += stringCol * (stringCore * 1.4 + stringGlow * 0.8);

    // Chromatic sparkle on bright core
    col += vec3(1.0, 0.95, 0.95) * stringCore * 0.6;

    // Soft halo around whole string
    float haloD = minStringD;
    col += cs_pal(fract(x_Time * 0.04)) * exp(-haloD * 8.0) * 0.1;

    // Composite
    _wShaderOut = vec4(col, 1.0);
}
