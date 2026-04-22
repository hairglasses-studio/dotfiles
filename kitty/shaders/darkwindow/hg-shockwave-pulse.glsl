// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Shockwave pulse — expanding circular shockwaves refracting the terminal through the wave front

const int   PULSES    = 5;
const float CYCLE     = 2.5;
const float INTENSITY = 0.55;

vec3 sw_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.25, 0.60);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float sw_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  sw_hash2(float n) { return vec2(sw_hash(n), sw_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);
    vec2 sampleUV = uv;

    for (int i = 0; i < PULSES; i++) {
        float fi = float(i);
        float phase = fract((x_Time + sw_hash(fi) * CYCLE) / CYCLE);
        float cycleID = floor((x_Time + sw_hash(fi) * CYCLE) / CYCLE);
        // Center per pulse
        vec2 center = sw_hash2(fi * 3.7 + cycleID * 7.3) * 1.2 - 0.6;
        center.x *= x_WindowSize.x / x_WindowSize.y;

        float r = length(p - center);
        float waveR = phase * 1.2;
        float wd = abs(r - waveR);
        float ringW = 0.015 + phase * 0.04;

        // Wavefront ring brightness
        float ring = exp(-wd * wd / (ringW * ringW) * 2.0) * (1.0 - phase);
        vec3 ringCol = sw_pal(fract(fi * 0.15 + x_Time * 0.05));
        col += ringCol * ring * 0.6;

        // Refraction at wavefront: shift terminal sample radially
        if (ring > 0.05) {
            vec2 refractDir = normalize(p - center);
            // Compression: sample closer inside, farther outside
            float refractStrength = ring * 0.01;
            sampleUV += refractDir * refractStrength * sign(r - waveR) * 0.3;
        }
    }

    // Sample terminal through refracted UV
    vec4 terminal = x_Texture(clamp(sampleUV, 0.0, 1.0));

    // Composite
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.5);
    vec3 result = mix(terminal.rgb, terminal.rgb + col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
