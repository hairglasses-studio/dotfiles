// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Aurora whirlpool — aurora bands pulled into a rotating vortex with chromatic drift

const int   BANDS    = 5;
const float SWIRL    = 3.0;
const float INTENSITY = 0.55;

vec3 aw_pal(float t) {
    vec3 green = vec3(0.10, 0.95, 0.55);
    vec3 cyan  = vec3(0.15, 0.70, 0.95);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 mag   = vec3(0.90, 0.30, 0.70);
    vec3 gold  = vec3(0.96, 0.80, 0.35);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(green, cyan, s);
    else if (s < 2.0) return mix(cyan, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else if (s < 4.0) return mix(mag, gold, s - 3.0);
    else              return mix(gold, green, s - 4.0);
}

float aw_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float aw_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(aw_hash(i), aw_hash(i + vec2(1,0)), u.x),
               mix(aw_hash(i + vec2(0,1)), aw_hash(i + vec2(1,1)), u.x), u.y);
}

float aw_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 5; i++) {
        v += a * aw_noise(p);
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

    // Swirl: angle shifts with 1/r (tighter spiral near center)
    float swirlAmount = SWIRL / max(r, 0.05);
    float swirledA = a + swirlAmount * sign(sin(x_Time * 0.1 + 0.5));
    // Rotation animation
    swirledA += x_Time * 0.2;

    // Transformed polar coord
    vec2 swirledP = vec2(cos(swirledA), sin(swirledA)) * r;

    vec3 col = vec3(0.0);

    // Aurora bands — each band is an FBM-ribbon at varying radius
    for (int b = 0; b < BANDS; b++) {
        float fb = float(b);
        float bandR = 0.1 + fb * 0.09;
        float bandWidth = 0.04 + fb * 0.015;
        float rDist = abs(r - bandR);
        if (rDist > bandWidth * 2.0) continue;

        // Horizontal FBM (in swirled polar) — acts like a curtain
        vec2 fbmP = vec2(swirledA * 2.0, r * 10.0) + vec2(0.0, fb * 2.0);
        fbmP += x_Time * 0.1;
        float fbmVal = aw_fbm(fbmP);

        float bandMask = exp(-rDist * rDist / (bandWidth * bandWidth) * 1.5);
        float curtain = smoothstep(0.3, 0.8, fbmVal);

        vec3 bc = aw_pal(fract(fb * 0.15 + r * 0.5 + x_Time * 0.03));
        col += bc * bandMask * curtain * 0.5;
    }

    // Central vortex glow
    float vortexGlow = exp(-r * r * 120.0) * 0.8;
    col += aw_pal(0.1) * vortexGlow;

    // Sparse bright radial rays from center
    float rays = pow(0.5 + 0.5 * cos(swirledA * 16.0 + x_Time), 8.0);
    float rayRadial = (1.0 - smoothstep(0.05, 0.4, r));
    col += aw_pal(fract(x_Time * 0.08)) * rays * rayRadial * 0.3;

    // Dark outer vignette
    col *= smoothstep(1.3, 0.1, r);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.85, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
