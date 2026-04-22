// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Supernova shockwave — expanding ring + hot bright core + radial ejecta

const int   EJECTA = 24;
const float CYCLE  = 5.0;      // supernova repeat period
const float INTENSITY = 0.55;

vec3 sn_pal(float heat) {
    vec3 vio   = vec3(0.45, 0.15, 0.95);
    vec3 mag   = vec3(0.95, 0.25, 0.60);
    vec3 orange = vec3(1.00, 0.55, 0.20);
    vec3 yellow = vec3(1.00, 0.95, 0.55);
    vec3 white  = vec3(1.00, 0.98, 0.90);
    if (heat < 0.25)      return mix(vio, mag, heat * 4.0);
    else if (heat < 0.5)  return mix(mag, orange, (heat - 0.25) * 4.0);
    else if (heat < 0.75) return mix(orange, yellow, (heat - 0.5) * 4.0);
    else                  return mix(yellow, white, (heat - 0.75) * 4.0);
}

float sn_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);

    // Cycle phase — explosion every CYCLE seconds
    float phase = fract(x_Time / CYCLE);
    float cycleID = floor(x_Time / CYCLE);

    vec3 col = vec3(0.01, 0.01, 0.03);

    // Phase 0-0.08: pre-flash (dim star)
    // Phase 0.08-0.15: bright flash
    // Phase 0.15-0.7: expanding shockwave
    // Phase 0.7-1.0: fade + cooling remnants

    // 1. Core star (always visible, brightest at pre-flash, dims after explosion)
    float starBright;
    if (phase < 0.08) {
        starBright = 0.3 + phase * 3.0;    // rising
    } else if (phase < 0.15) {
        starBright = 5.0 * (1.0 - (phase - 0.08) / 0.07);  // decay
    } else {
        starBright = 0.1 * (1.0 - smoothstep(0.15, 0.8, phase));
    }
    float coreMask = exp(-r * r * 400.0);
    col += sn_pal(0.95) * coreMask * starBright;

    // 2. Expanding shockwave ring
    float shockR = 0.0;
    float shockIntensity = 0.0;
    if (phase > 0.1 && phase < 0.9) {
        shockR = (phase - 0.1) * 1.1;
        // Ring thickness increases with expansion (thermal broadening)
        float ringW = 0.01 + (phase - 0.1) * 0.05;
        float ringD = abs(r - shockR);
        shockIntensity = exp(-ringD * ringD / (ringW * ringW) * 3.0);
        // Ring fades as it expands (conservation of energy)
        shockIntensity *= (1.0 - (phase - 0.1) / 0.8);
        // Heat of shock — hotter closer to explosion start, cools as expands
        float shockHeat = 1.0 - (phase - 0.1) / 0.8;
        col += sn_pal(shockHeat) * shockIntensity * 1.4;
    }

    // 3. Radial ejecta streaks — filamentary structure
    float angle = atan(p.y, p.x);
    for (int e = 0; e < EJECTA; e++) {
        float fe = float(e);
        float ejectaSeed = fe * 3.7 + cycleID * 13.7;
        float ejectaAngle = fe / float(EJECTA) * 6.28 + sn_hash(ejectaSeed) * 0.5;
        float ejectaLen = 0.2 + sn_hash(ejectaSeed * 1.7) * 0.4;
        // Angular distance to ejecta direction
        float angDist = abs(mod(angle - ejectaAngle + 3.14159, 6.28318) - 3.14159);
        float angMask = exp(-angDist * angDist * 400.0);
        // Radial reach: ejecta extends from core outward
        if (r < ejectaLen * phase * 2.0 && phase > 0.12) {
            float reachMask = smoothstep(0.0, 0.1, r) * smoothstep(ejectaLen * phase * 2.0, ejectaLen * phase * 1.5, r);
            float ejectaHeat = 0.8 - (r / ejectaLen) * 0.6;
            col += sn_pal(ejectaHeat) * angMask * reachMask * 0.6;
        }
    }

    // 4. Lingering nebula remnant (final phase)
    if (phase > 0.7) {
        float remPhase = (phase - 0.7) / 0.3;
        float nebR = 0.3 + remPhase * 0.2;
        float nebDist = abs(r - nebR);
        float nebMask = exp(-nebDist * nebDist * 60.0) * (1.0 - remPhase) * 0.3;
        col += sn_pal(0.25) * nebMask;
    }

    // Starfield backdrop
    vec2 sg = floor(p * 120.0);
    float sh = sn_hash(sg.x * 31.0 + sg.y);
    if (sh > 0.996) {
        col += vec3(0.9, 0.9, 1.0) * (sh - 0.996) * 200.0 * 0.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
