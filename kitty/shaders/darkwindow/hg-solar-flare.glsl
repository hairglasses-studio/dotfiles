// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Solar flare — bright star surface with coronal mass ejection arcs + granulation

const int   CME_ARCS  = 4;
const int   GRANULES = 40;
const float SUN_RAD   = 0.28;
const float INTENSITY = 0.55;

vec3 sf_col(float heat) {
    vec3 dark   = vec3(0.6, 0.0, 0.05);
    vec3 orange = vec3(1.00, 0.35, 0.05);
    vec3 yellow = vec3(1.00, 0.80, 0.25);
    vec3 white  = vec3(1.00, 0.98, 0.90);
    if (heat < 0.33)      return mix(dark, orange, heat * 3.0);
    else if (heat < 0.66) return mix(orange, yellow, (heat - 0.33) * 3.0);
    else                  return mix(yellow, white, (heat - 0.66) * 3.0);
}

float sf_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float sf_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }

float sf_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(sf_hash(i), sf_hash(i + vec2(1,0)), u.x),
               mix(sf_hash(i + vec2(0,1)), sf_hash(i + vec2(1,1)), u.x), u.y);
}

float sf_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 4; i++) {
        v += a * sf_noise(p);
        p = p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);

    vec3 col = vec3(0.02, 0.01, 0.03);

    // Sun surface — turbulent granulation
    if (r < SUN_RAD) {
        float ang = atan(p.y, p.x);
        // Convective granulation
        float granMask = sf_fbm(p * 20.0 + x_Time * 0.2);
        float heat = 0.7 + granMask * 0.3;
        // Surface activity — brighter spots
        heat *= 0.9 + sf_noise(p * 50.0 + x_Time) * 0.2;
        // Darkening toward limb (limb darkening)
        float limbFactor = sqrt(1.0 - (r / SUN_RAD) * (r / SUN_RAD));
        heat *= 0.5 + limbFactor * 0.5;
        col = sf_col(heat);
    }

    // Bright corona / atmosphere
    float coronaD = max(0.0, r - SUN_RAD);
    float corona = exp(-coronaD * 3.0) * 0.5;
    col += sf_col(0.55) * corona;

    // Coronal Mass Ejection arcs — bright loops erupting from limb
    for (int i = 0; i < CME_ARCS; i++) {
        float fi = float(i);
        float cyclePhase = fract(x_Time * 0.15 + fi * 0.25);
        // Each CME lasts for full cycle
        float baseAng = sf_hash1(fi * 3.7 + floor(x_Time * 0.15 + fi * 0.25)) * 6.28;
        float cmeSize = 0.15 + sf_hash1(fi * 5.1) * 0.12;
        vec2 cmeCenter = vec2(cos(baseAng), sin(baseAng)) * (SUN_RAD + cmeSize * 0.4);
        // Arc from limb anchor point to cme center and back
        vec2 anchor = vec2(cos(baseAng), sin(baseAng)) * SUN_RAD;
        // Distance to arc (approximated as circle passing through anchor + cmeCenter)
        vec2 arcCenter = (anchor + cmeCenter) * 0.5;
        float arcR = length(anchor - cmeCenter) * 0.5;
        float distToArc = abs(length(p - arcCenter) - arcR);
        // Width
        float arcWidth = 0.003 + cyclePhase * 0.02;
        float arc = exp(-distToArc * distToArc / (arcWidth * arcWidth) * 2.0);
        // Only above sun limb
        float limbMask = smoothstep(SUN_RAD * 0.95, SUN_RAD * 1.05, length(p));
        float cycleFade = pow(1.0 - cyclePhase, 1.5);
        col += sf_col(0.8) * arc * limbMask * cycleFade * 1.2;
    }

    // Prominence spikes — brief radial jets
    for (int g = 0; g < GRANULES; g++) {
        float fg = float(g);
        float gAng = fg / float(GRANULES) * 6.28;
        vec2 gDir = vec2(cos(gAng), sin(gAng));
        float gLen = 0.01 + sf_hash1(fg) * 0.03 * (0.8 + 0.2 * sin(x_Time * 5.0 + fg));
        vec2 gTip = gDir * (SUN_RAD + gLen);
        // Distance to radial line from anchor to tip
        float along = dot(p, gDir) - SUN_RAD;
        float perpD = abs(dot(p, vec2(-gDir.y, gDir.x)));
        if (along > 0.0 && along < gLen && perpD < 0.005) {
            float spikeI = 1.0 - along / gLen;
            col += sf_col(0.85) * spikeI * exp(-perpD * perpD * 30000.0) * 0.8;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.55);
    float mask = clamp(length(col) * 0.85, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
