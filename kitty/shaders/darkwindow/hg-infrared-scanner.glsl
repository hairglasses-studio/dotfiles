// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Infrared scanner — thermal camera view with heat sources + scan crosshair + temperature scale

const int   HEAT_SOURCES = 6;
const int   FBM_OCTAVES  = 5;
const float INTENSITY = 0.55;

vec3 ir_col(float heat) {
    // Standard thermal palette: purple → blue → green → yellow → red → white
    vec3 dark    = vec3(0.05, 0.02, 0.15);
    vec3 vio     = vec3(0.45, 0.10, 0.55);
    vec3 blue    = vec3(0.15, 0.55, 0.95);
    vec3 green   = vec3(0.25, 0.95, 0.45);
    vec3 yellow  = vec3(0.95, 0.85, 0.25);
    vec3 red     = vec3(0.95, 0.35, 0.15);
    vec3 white   = vec3(1.00, 0.98, 0.80);
    if (heat < 0.15)      return mix(dark, vio, heat / 0.15);
    else if (heat < 0.3)  return mix(vio, blue, (heat - 0.15) / 0.15);
    else if (heat < 0.5)  return mix(blue, green, (heat - 0.3) / 0.2);
    else if (heat < 0.7)  return mix(green, yellow, (heat - 0.5) / 0.2);
    else if (heat < 0.9)  return mix(yellow, red, (heat - 0.7) / 0.2);
    else                  return mix(red, white, (heat - 0.9) / 0.1);
}

float ir_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float ir_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  ir_hash2(float n) { return vec2(ir_hash1(n), ir_hash1(n * 1.37 + 11.0)); }

float ir_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(ir_hash(i), ir_hash(i + vec2(1,0)), u.x),
               mix(ir_hash(i + vec2(0,1)), ir_hash(i + vec2(1,1)), u.x), u.y);
}

float ir_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < FBM_OCTAVES; i++) {
        v += a * ir_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Heat field — FBM background + local sources
    float heat = 0.15 + ir_fbm(p * 3.0 + x_Time * 0.05) * 0.2;

    // Heat sources (people/objects) — animated, each with radial falloff
    for (int i = 0; i < HEAT_SOURCES; i++) {
        float fi = float(i);
        vec2 pos = ir_hash2(fi * 3.71) * 2.0 - 1.0;
        pos.x *= x_WindowSize.x / x_WindowSize.y;
        // Slow drift
        pos += 0.05 * vec2(sin(x_Time * 0.2 + fi), cos(x_Time * 0.15 + fi * 1.3));
        float d = length(p - pos);
        float sourceHeat = exp(-d * d * 80.0) * (0.6 + 0.4 * ir_hash1(fi));
        heat = max(heat, sourceHeat);
    }

    vec3 col = ir_col(heat);

    // Scan crosshair — slow sweep
    vec2 scanPos = 0.5 * vec2(sin(x_Time * 0.3), cos(x_Time * 0.25));
    vec2 toScan = p - scanPos;
    float hLine = exp(-toScan.y * toScan.y * 10000.0);
    float vLine = exp(-toScan.x * toScan.x * 10000.0);
    if (length(toScan) > 0.02 && length(toScan) < 0.08) {
        col += vec3(0.9, 0.95, 0.9) * (hLine + vLine) * 0.5;
    }
    // Crosshair center circle
    float crossD = abs(length(toScan) - 0.03);
    col += vec3(0.9, 0.95, 0.9) * exp(-crossD * crossD * 5000.0) * 0.6;

    // Temperature readout at crosshair
    float ch2 = length(toScan);
    float heatAtCross = heat;
    // Draw heat indicator as small gauge near crosshair
    if (abs(toScan.x - 0.1) < 0.003 && abs(toScan.y) < 0.04) {
        col += ir_col(heatAtCross) * 1.2;
    }

    // Scanline artifacts
    float scan = 0.9 + 0.1 * sin(x_PixelPos.y * 3.0 + x_Time * 2.0);
    col *= scan;

    // Edge vignette
    col *= smoothstep(1.5, 0.3, length(p));

    // Temperature scale bar at bottom-right corner
    vec2 scaleUV = uv - vec2(0.85, 0.1);
    if (scaleUV.x > 0.0 && scaleUV.x < 0.1 && scaleUV.y > 0.0 && scaleUV.y < 0.15) {
        col = ir_col(scaleUV.y / 0.15);
        // Tick lines
        if (abs(fract(scaleUV.y / 0.015 + 0.1) - 0.1) < 0.1) col *= 1.3;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.55);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
