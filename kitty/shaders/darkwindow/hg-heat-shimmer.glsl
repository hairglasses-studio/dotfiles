// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Heat shimmer — horizontal refractive distortion above heated source + mirage

const int   OCTAVES = 5;
const float INTENSITY = 0.5;

vec3 hs_pal(float t) {
    vec3 dark_heat = vec3(0.25, 0.04, 0.06);
    vec3 warm      = vec3(0.95, 0.55, 0.20);
    vec3 bright    = vec3(1.00, 0.90, 0.50);
    vec3 cyan      = vec3(0.10, 0.82, 0.92);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(dark_heat, warm, s);
    else if (s < 2.0) return mix(warm, bright, s - 1.0);
    else if (s < 3.0) return mix(bright, cyan, s - 2.0);
    else              return mix(cyan, dark_heat, s - 3.0);
}

float hs_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float hs_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(hs_hash(i), hs_hash(i + vec2(1,0)), u.x),
               mix(hs_hash(i + vec2(0,1)), hs_hash(i + vec2(1,1)), u.x), u.y);
}

float hs_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * hs_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    // Heat source line (bottom)
    float heatY = 0.25;
    float yAboveHeat = uv.y - heatY;

    // Displacement intensity decreases with height
    float distIntensity = smoothstep(0.7, 0.0, yAboveHeat) * 0.02;
    if (yAboveHeat < 0.0) distIntensity = 0.0;

    // FBM displacement — vertical shimmer
    vec2 noiseP = vec2(uv.x * 8.0, uv.y * 4.0 + x_Time * 0.5);
    vec2 displacement = vec2(
        (hs_fbm(noiseP) - 0.5) * distIntensity,
        (hs_fbm(noiseP + vec2(5.2, 0.0)) - 0.5) * distIntensity * 0.3
    );

    // Sample terminal at displaced position (chromatic for air-density variation)
    vec3 terminalDisplaced;
    terminalDisplaced.r = x_Texture(uv + displacement + vec2(0.002, 0.0) * distIntensity).r;
    terminalDisplaced.g = x_Texture(uv + displacement).g;
    terminalDisplaced.b = x_Texture(uv + displacement - vec2(0.002, 0.0) * distIntensity).b;

    vec3 col = terminalDisplaced;

    // Heat source glow (below heatY)
    if (yAboveHeat < 0.0) {
        float heatIntensity = smoothstep(heatY, 0.0, uv.y);
        vec3 heatColor = hs_pal(fract(x_Time * 0.02 + heatIntensity * 0.3));
        col = mix(col, heatColor, heatIntensity * 0.7);
        // Cracked heat texture
        float crack = hs_fbm(vec2(uv.x * 20.0, uv.y * 30.0) + x_Time * 0.1);
        col += vec3(1.0, 0.5, 0.1) * crack * heatIntensity * 0.2;
    }

    // Rising heat wavey bands above source
    if (yAboveHeat > 0.0 && yAboveHeat < 0.7) {
        float riseT = yAboveHeat;
        vec2 bandP = uv * vec2(20.0, 10.0) + vec2(0.0, -x_Time * 2.0);
        float bands = hs_fbm(bandP) * smoothstep(0.0, 0.1, yAboveHeat) * smoothstep(0.7, 0.4, yAboveHeat);
        col += hs_pal(fract(x_Time * 0.04)) * bands * 0.15;
    }

    // Mirage inversion at very top — flip image partially
    if (yAboveHeat > 0.5) {
        float mirageMask = smoothstep(0.5, 0.7, yAboveHeat);
        vec2 mirageUV = vec2(uv.x, 2.0 * heatY + 0.5 - uv.y);
        mirageUV = clamp(mirageUV, 0.0, 1.0);
        vec3 mirage = x_Texture(mirageUV).rgb;
        col = mix(col, mirage * 0.7, mirageMask * 0.15);
    }

    // Composite (terminal already present in col via displaced sample)
    _wShaderOut = vec4(col, 1.0);
}
