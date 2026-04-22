// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Binary neutron star merger — inspiral + kilonova flash + gravitational wave ripples

const int   WAVE_RINGS = 6;
const float CYCLE      = 6.0;
const float INTENSITY  = 0.6;

vec3 bc_col(float heat) {
    vec3 deep   = vec3(0.20, 0.04, 0.12);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.25, 0.60);
    vec3 gold   = vec3(1.00, 0.85, 0.45);
    vec3 white  = vec3(1.00, 0.98, 0.90);
    if (heat < 0.25)      return mix(deep, vio, heat * 4.0);
    else if (heat < 0.5)  return mix(vio, mag, (heat - 0.25) * 4.0);
    else if (heat < 0.75) return mix(mag, gold, (heat - 0.5) * 4.0);
    else                  return mix(gold, white, (heat - 0.75) * 4.0);
}

float bc_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.04);

    // Cycle phase: 0-0.7 = inspiral, 0.7-0.8 = merger flash, 0.8-1.0 = kilonova
    float phase = mod(x_Time, CYCLE) / CYCLE;

    if (phase < 0.7) {
        // Inspiral: two stars orbiting, separation shrinks
        float sep = 0.25 * (1.0 - phase / 0.7);
        sep = pow(sep, 0.5);                        // slower shrink
        float orbitSpeed = 0.5 + phase * 30.0;       // faster as they merge
        float theta = orbitSpeed * x_Time;
        vec2 star1 = vec2(cos(theta), sin(theta)) * sep;
        vec2 star2 = -star1;

        // Star bodies (increasingly bright as merger approaches)
        float d1 = length(p - star1);
        float d2 = length(p - star2);
        float starSize = 0.02;
        float bright = 1.0 + phase * 3.0;
        col += bc_col(0.8) * exp(-d1 * d1 / (starSize * starSize) * 2.0) * bright;
        col += bc_col(0.8) * exp(-d2 * d2 / (starSize * starSize) * 2.0) * bright;
        // Halos
        col += bc_col(0.4) * exp(-d1 * d1 * 300.0) * 0.4;
        col += bc_col(0.4) * exp(-d2 * d2 * 300.0) * 0.4;
        // Gas stream between them (Lagrangian)
        vec2 mid = (star1 + star2) * 0.5;
        float midD = length(p - mid);
        vec2 axis = normalize(star2 - star1);
        float along = abs(dot(p - mid, axis));
        float perp = abs(dot(p - mid, vec2(-axis.y, axis.x)));
        if (along < sep * 0.8 && perp < 0.01) {
            col += bc_col(0.7) * exp(-perp * perp * 5000.0) * 0.4;
        }
    } else if (phase < 0.78) {
        // Merger flash
        float flashPhase = (phase - 0.7) / 0.08;
        float r = length(p);
        float flash = exp(-r * r * 20.0) * (1.0 - flashPhase);
        col += bc_col(0.95) * flash * 3.0;
        col += vec3(1.0) * exp(-r * r * 200.0) * (1.0 - flashPhase) * 1.5;
    } else {
        // Kilonova — expanding debris cloud
        float kPhase = (phase - 0.78) / 0.22;
        float r = length(p);
        float debrisR = kPhase * 0.8;
        float shellDist = abs(r - debrisR);
        float shellWidth = 0.01 + kPhase * 0.15;
        float shell = exp(-shellDist * shellDist / (shellWidth * shellWidth) * 2.0);
        shell *= (1.0 - kPhase);
        float heat = 0.8 - kPhase * 0.6;
        col += bc_col(heat) * shell * 1.4;
        // Remnant at center
        col += bc_col(0.9) * exp(-r * r * 500.0) * (1.0 - kPhase * 0.8);
    }

    // Gravitational waves — expanding ripples during inspiral/merger
    if (phase > 0.4 && phase < 0.9) {
        for (int w = 0; w < WAVE_RINGS; w++) {
            float fw = float(w);
            float waveAge = fract(x_Time * 0.8 + fw * 0.15);
            float waveR = waveAge * 1.2;
            float r = length(p);
            float wd = abs(r - waveR);
            float wave = exp(-wd * wd * 400.0);
            float waveFade = 1.0 - waveAge;
            col += bc_col(0.5) * wave * waveFade * 0.15;
        }
    }

    // Background stars
    vec2 sg = floor(p * 120.0);
    float sh = bc_hash(sg);
    if (sh > 0.996) {
        col += vec3(0.9, 0.9, 1.0) * (sh - 0.996) * 200.0 * 0.5;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
