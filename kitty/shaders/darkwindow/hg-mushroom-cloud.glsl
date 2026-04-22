// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Mushroom cloud — atomic explosion with expanding mushroom cap + stem + shockwave ring

const int   OCTAVES = 6;
const float INTENSITY = 0.55;

vec3 mc_col(float heat) {
    vec3 dark   = vec3(0.08, 0.04, 0.08);
    vec3 red    = vec3(0.95, 0.20, 0.05);
    vec3 orange = vec3(1.00, 0.50, 0.10);
    vec3 yellow = vec3(1.00, 0.85, 0.40);
    vec3 white  = vec3(1.00, 0.98, 0.85);
    if (heat < 0.25)      return mix(dark, red, heat * 4.0);
    else if (heat < 0.5)  return mix(red, orange, (heat - 0.25) * 4.0);
    else if (heat < 0.75) return mix(orange, yellow, (heat - 0.5) * 4.0);
    else                  return mix(yellow, white, (heat - 0.75) * 4.0);
}

float mc_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float mc_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(mc_hash(i), mc_hash(i + vec2(1,0)), u.x),
               mix(mc_hash(i + vec2(0,1)), mc_hash(i + vec2(1,1)), u.x), u.y);
}

float mc_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * mc_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Explosion cycle
    float cycle = 8.0;
    float phase = mod(x_Time, cycle) / cycle;

    vec3 col = mix(vec3(0.01, 0.02, 0.05), vec3(0.10, 0.05, 0.08), smoothstep(0.3, -0.3, p.y));

    // Growth envelope
    float growth = pow(phase, 0.5);
    float fade = 1.0 - smoothstep(0.7, 1.0, phase);

    // Stem (rising column)
    float stemHeight = growth * 0.5 - 0.4;  // rises from -0.4 to 0.1
    float stemW = 0.04 + growth * 0.03;
    float stemHeatTop = stemHeight;
    if (p.y > -0.5 && p.y < stemHeight && abs(p.x) < stemW) {
        float stemMask = exp(-p.x * p.x / (stemW * stemW) * 2.0);
        float stemTurb = mc_fbm(vec2(p.x * 20.0, p.y * 10.0 - x_Time * 3.0));
        float stemHeat = 0.8 - (stemHeight - p.y) / 0.5 * 0.3;
        col = mix(col, mc_col(stemHeat) * (0.6 + stemTurb * 0.5), stemMask * fade);
    }

    // Mushroom cap — rising with growth, expanding
    float capY = stemHeight + 0.12;
    vec2 capPos = p - vec2(0.0, capY);
    float capR = 0.15 + growth * 0.3;
    float capH = 0.12 + growth * 0.08;

    // Cap density — elliptical
    vec2 capEllipse = capPos;
    capEllipse.y *= capR / capH;
    float capDist = length(capEllipse);

    if (capDist < capR) {
        // Turbulence
        float capTurb = mc_fbm(vec2(atan(capPos.y, capPos.x) * 4.0, capDist * 10.0 - x_Time * 0.5));
        float capDensity = (1.0 - capDist / capR) * (0.7 + capTurb * 0.5);
        float capHeat = capDensity * (0.9 - phase * 0.4);
        col = mix(col, mc_col(capHeat), capDensity * fade);
    }

    // Rolling outer cap edge — billowing clouds
    if (capDist > capR * 0.7 && capDist < capR * 1.2 && capPos.y > -0.1) {
        float rolloff = capDist - capR * 0.7;
        float billTurb = mc_fbm(capPos * 8.0 - x_Time * 0.5);
        float billow = exp(-rolloff * rolloff * 100.0) * billTurb;
        col += mc_col(0.5) * billow * 0.6 * fade;
    }

    // Shockwave ring — expanding outward
    if (phase < 0.6) {
        float shockR = phase * 1.2;
        float shockDist = abs(length(p) - shockR);
        float shockWidth = 0.01 + phase * 0.04;
        float shockMask = exp(-shockDist * shockDist / (shockWidth * shockWidth) * 2.0);
        col += vec3(0.9, 0.95, 1.0) * shockMask * (1.0 - phase) * 0.5;
    }

    // Dust cloud at base
    if (p.y < -0.3) {
        float dustTurb = mc_fbm(p * 4.0 + x_Time * 0.2);
        float dustMask = growth * dustTurb * smoothstep(-0.5, -0.3, p.y);
        col = mix(col, vec3(0.15, 0.10, 0.08), dustMask * 0.6);
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.55);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
