// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Pinwheel galaxy (M101-style) — 4-arm spiral face-on with stars + HII regions

const int   ARMS        = 4;
const float ARM_PITCH   = 0.35;
const int   HII_REGIONS = 20;
const int   STAR_CLUSTER = 48;
const float INTENSITY   = 0.55;

vec3 pg_pal(float t) {
    vec3 core    = vec3(1.00, 0.90, 0.55);
    vec3 warm    = vec3(0.95, 0.55, 0.30);
    vec3 mag     = vec3(0.85, 0.25, 0.60);
    vec3 cyan    = vec3(0.20, 0.85, 0.95);
    vec3 vio     = vec3(0.50, 0.25, 0.95);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(core, warm, s);
    else if (s < 2.0) return mix(warm, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, cyan, s - 2.0);
    else if (s < 4.0) return mix(cyan, vio, s - 3.0);
    else              return mix(vio, core, s - 4.0);
}

float pg_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  pg_hash2(float n) { return vec2(pg_hash(n), pg_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.01, 0.03);

    // Spiral rotation
    float t = x_Time * 0.06;
    float cr = cos(t), sr = sin(t);
    vec2 rp = mat2(cr, -sr, sr, cr) * p;

    float r = length(rp);
    float a = atan(rp.y, rp.x);

    // Bright core
    float core = exp(-r * r * 400.0) * 1.6;
    col += pg_pal(0.0) * core;
    col += pg_pal(0.05) * exp(-r * r * 25.0) * 0.3;

    // Spiral arms — logarithmic spiral at ARMS positions
    float armStrength = 0.0;
    for (int arm = 0; arm < ARMS; arm++) {
        float armOffset = float(arm) / float(ARMS) * 6.28318;
        float armAngle = a - log(r + 0.02) / ARM_PITCH + armOffset;
        // Normalize to [0,1]
        armAngle = mod(armAngle, 6.28318);
        // Arm density: Gaussian around arm midline
        float armDist = min(armAngle, 6.28318 - armAngle);
        float armMask = exp(-armDist * armDist * 15.0);
        // Arms extend from ~0.08 to 0.6
        float rMask = smoothstep(0.06, 0.15, r) * smoothstep(0.55, 0.35, r);
        armStrength = max(armStrength, armMask * rMask);
    }
    col += pg_pal(0.3) * armStrength * 0.5;

    // Dust lanes — darker bands between arms
    for (int arm = 0; arm < ARMS; arm++) {
        float armOffset = float(arm) / float(ARMS) * 6.28318 + 0.3;  // offset from arm
        float dustAngle = mod(a - log(r + 0.02) / ARM_PITCH + armOffset, 6.28318);
        float dustDist = min(dustAngle, 6.28318 - dustAngle);
        float dust = exp(-dustDist * dustDist * 30.0) * smoothstep(0.08, 0.3, r) * smoothstep(0.6, 0.3, r);
        col *= 1.0 - dust * 0.5;
    }

    // HII regions — bright pink star-forming spots along arms
    for (int h = 0; h < HII_REGIONS; h++) {
        float fh = float(h);
        // Place along arms
        float hArm = mod(fh, float(ARMS));
        float hArmOffset = hArm / float(ARMS) * 6.28318;
        float hAlongArm = 0.15 + pg_hash(fh * 3.7) * 0.4;
        float hAng = hAlongArm / ARM_PITCH * (-1.0) + hArmOffset;
        vec2 hPos = vec2(cos(hAng), sin(hAng)) * hAlongArm;
        // Add perpendicular jitter
        hPos += (pg_hash2(fh * 5.1) - 0.5) * 0.02;

        float hd = length(rp - hPos);
        float core2 = exp(-hd * hd * 6000.0);
        float halo = exp(-hd * hd * 800.0) * 0.4;
        col += pg_pal(0.4) * (core2 * 1.4 + halo);
    }

    // Individual stars
    for (int s = 0; s < STAR_CLUSTER; s++) {
        float fs = float(s);
        vec2 sp = pg_hash2(fs * 7.3) * 2.0 - 1.0;
        sp *= 0.7;
        // Concentrate around disk center with some scatter
        float rs = length(sp);
        if (rs > 0.7) continue;
        float sd = length(rp - sp);
        float core2 = exp(-sd * sd * 30000.0);
        col += vec3(0.9, 0.9, 1.0) * core2 * 0.9;
    }

    // Outer halo
    col += pg_pal(0.1) * exp(-r * r * 3.0) * 0.1;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
