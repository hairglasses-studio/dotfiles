// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Aurora veil — stationary hanging aurora with detailed filament structure + vertical rays

const int   VEIL_LAYERS = 6;
const int   OCTAVES = 5;
const float INTENSITY = 0.55;

vec3 av_pal(float t) {
    vec3 green = vec3(0.12, 0.98, 0.55);
    vec3 cyan  = vec3(0.20, 0.78, 0.98);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 mag   = vec3(0.95, 0.35, 0.70);
    vec3 gold  = vec3(0.98, 0.80, 0.30);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(green, cyan, s);
    else if (s < 2.0) return mix(cyan, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else if (s < 4.0) return mix(mag, gold, s - 3.0);
    else              return mix(gold, green, s - 4.0);
}

float av_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float av_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(av_hash(i), av_hash(i + vec2(1,0)), u.x),
               mix(av_hash(i + vec2(0,1)), av_hash(i + vec2(1,1)), u.x), u.y);
}

float av_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * av_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Night sky
    vec3 bg = mix(vec3(0.02, 0.05, 0.15), vec3(0.005, 0.02, 0.08), 1.0 - uv.y);
    vec3 col = bg;

    // Stars
    vec2 sg = floor(p * 100.0);
    float sh = av_hash(sg);
    if (sh > 0.996 && uv.y > 0.3) {
        float tw = 0.4 + 0.6 * sin(x_Time * (2.0 + sh * 3.0));
        col += vec3(0.9, 0.9, 1.0) * (sh - 0.996) * 200.0 * tw;
    }

    // Aurora veils — hanging vertical curtains
    for (int i = 0; i < VEIL_LAYERS; i++) {
        float fi = float(i);
        float yCenter = 0.25 + fi * 0.06;  // horizontal "hang" band
        float yThickness = 0.12 + fi * 0.01;

        // Horizontal position: FBM gives sideways displacement
        vec2 fbmP = vec2(p.x * 4.0 + x_Time * 0.1, fi * 1.0 + x_Time * 0.05);
        float xDisp = av_fbm(fbmP) * 0.3;

        // Column position: check if fragment is in veil vertically
        float vDistFromCenter = abs(uv.y - yCenter);
        if (vDistFromCenter > yThickness * 1.5) continue;

        // Vertical fade
        float vMask = exp(-vDistFromCenter * vDistFromCenter / (yThickness * yThickness) * 1.2);

        // Detailed filament structure — vertical rays
        float rayPhase = p.x * 30.0 + xDisp * 20.0 + x_Time * 0.3 + fi * 5.0;
        float rayStrength = av_fbm(vec2(rayPhase, uv.y * 4.0)) * 0.5 + 0.5;
        rayStrength = pow(rayStrength, 2.0);

        // Vertical extension below center — rays hang downward
        float below = smoothstep(0.0, -yThickness * 2.0, uv.y - yCenter);
        float rayExtend = below * (0.5 + rayStrength * 0.5) * 0.3;

        float total = vMask * rayStrength + rayExtend;
        vec3 vc = av_pal(fract(fi * 0.15 + p.x * 0.2 + x_Time * 0.03));
        col += vc * total * 0.35;
    }

    // Ground reflection — subtle bottom glow
    if (uv.y < 0.1) {
        float groundMask = smoothstep(0.1, 0.0, uv.y);
        col += av_pal(0.0) * groundMask * 0.1;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
