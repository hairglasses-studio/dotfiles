// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Spirit flame — ethereal flickering flame-spirits rising, with translucent body + face hint

const int   SPIRITS = 6;
const int   OCTAVES = 5;
const float INTENSITY = 0.55;

vec3 sp_pal(float t) {
    vec3 cyan = vec3(0.20, 0.85, 0.95);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 mint = vec3(0.25, 0.95, 0.55);
    vec3 gold = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mint, s - 1.0);
    else if (s < 3.0) return mix(mint, gold, s - 2.0);
    else              return mix(gold, cyan, s - 3.0);
}

float sp_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float sp_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }

float sp_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(sp_hash(i), sp_hash(i + vec2(1,0)), u.x),
               mix(sp_hash(i + vec2(0,1)), sp_hash(i + vec2(1,1)), u.x), u.y);
}

float sp_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * sp_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.02, 0.05);

    for (int i = 0; i < SPIRITS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float cycle = 5.0 + sp_hash1(seed) * 3.0;
        float phase = fract((x_Time + sp_hash1(seed * 3.7) * cycle) / cycle);

        // Starting x — random
        float baseX = (sp_hash1(seed * 5.1) - 0.5) * 1.6 * x_WindowSize.x / x_WindowSize.y;
        // Wafting
        float waft = 0.1 * sin(x_Time * 0.5 + seed) * phase;
        float spiritX = baseX + waft;
        // Rising
        float spiritY = -0.8 + phase * 2.0;

        vec2 spiritCenter = vec2(spiritX, spiritY);
        vec2 toSpirit = p - spiritCenter;

        // Body shape — tall tapered flame, wider at bottom, narrow at top
        float bodyH = 0.25;
        float bodyW = 0.05 * (1.0 - (toSpirit.y + 0.1) / bodyH);
        bodyW = max(0.0, bodyW);

        if (toSpirit.y < bodyH && toSpirit.y > -0.05 && abs(toSpirit.x) < bodyW) {
            // Flickering wisp shape
            float flicker = sp_fbm(toSpirit * vec2(20.0, 10.0) + vec2(x_Time * 1.5 + fi, 0.0));
            float bodyMask = exp(-toSpirit.x * toSpirit.x / (bodyW * bodyW + 0.001) * 2.0);
            bodyMask *= 1.0 - smoothstep(0.0, bodyH, toSpirit.y);
            bodyMask *= 0.7 + flicker * 0.6;

            // Fade out over lifetime
            float lifeFade = 1.0 - smoothstep(0.7, 1.0, phase);

            vec3 spColor = sp_pal(fract(seed * 0.03 + x_Time * 0.05));
            col += spColor * bodyMask * lifeFade * 0.9;

            // Core bright
            col += vec3(1.0) * bodyMask * lifeFade * 0.2;
        }

        // Face hint — two dark eye spots near top of body
        vec2 facePos = spiritCenter + vec2(0.0, bodyH * 0.65);
        vec2 toFace = p - facePos;
        if (abs(toFace.x) < bodyW * 0.6 && abs(toFace.y) < 0.02) {
            vec2 eye1 = facePos + vec2(-bodyW * 0.2, 0.0);
            vec2 eye2 = facePos + vec2(bodyW * 0.2, 0.0);
            float e1d = length(p - eye1);
            float e2d = length(p - eye2);
            float eyeMask = min(e1d, e2d);
            if (eyeMask < 0.003) {
                col *= 1.0 - smoothstep(0.003, 0.001, eyeMask) * 0.8;
            }
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
