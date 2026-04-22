// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Quasar core — brilliant accretion disk seen edge-on with relativistic jet perpendicular

const int   DISK_SAMPS = 32;
const float DISK_RAD   = 0.3;
const float JET_LEN    = 0.75;
const float INTENSITY  = 0.6;

vec3 qu_col(float heat) {
    vec3 white = vec3(1.00, 0.98, 0.85);
    vec3 hot   = vec3(1.00, 0.55, 0.15);
    vec3 mag   = vec3(0.95, 0.25, 0.55);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    if (heat > 0.75)      return mix(hot, white, (heat - 0.75) * 4.0);
    else if (heat > 0.5)  return mix(mag, hot, (heat - 0.5) * 4.0);
    else if (heat > 0.25) return mix(vio, mag, (heat - 0.25) * 4.0);
    else                  return vio * (heat * 4.0);
}

float qu_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.04);

    float r = length(p);
    if (r < 0.03) {
        col = vec3(0.0);
    } else {
        float photonR = 0.045;
        float photonD = abs(r - photonR);
        col += vec3(1.0, 0.95, 0.8) * exp(-photonD * photonD * 5000.0) * 0.9;
    }

    // Accretion disk — edge-on (flat horizontally)
    if (abs(p.y) < DISK_RAD * 0.15 && abs(p.x) < DISK_RAD && r > 0.05) {
        float diskThickness = exp(-pow(p.y / (DISK_RAD * 0.05), 2.0));
        float doppler = 0.6 + 0.4 * (p.x > 0.0 ? -1.0 : 1.0);
        float heat = 1.0 - abs(p.x) / DISK_RAD;
        float turb = qu_hash(vec2(abs(p.x) * 30.0, floor(x_Time * 8.0)));
        heat *= 0.7 + turb * 0.3;
        vec3 diskCol = qu_col(heat) * doppler;
        col += diskCol * diskThickness * 0.8;
    }

    // Relativistic jets — perpendicular to disk (vertical)
    for (int side = 0; side < 2; side++) {
        float signM = (side == 0) ? 1.0 : -1.0;
        float alongJet = p.y * signM;
        if (alongJet > 0.04 && alongJet < JET_LEN) {
            float jetR = 0.015 + alongJet * 0.06;
            float perpD = abs(p.x);
            if (perpD < jetR) {
                float jetMask = 1.0 - (perpD / jetR);
                float brightFade = exp(-alongJet * 2.5);
                float knots = qu_hash(vec2(floor(alongJet * 20.0), floor(x_Time * 3.0)));
                float knotPulse = pow(knots, 2.0);
                float jetIntensity = brightFade * jetMask * (0.7 + knotPulse * 0.6);
                col += qu_col(0.7 + knotPulse * 0.3) * jetIntensity * 1.3;
            }
        }
    }

    col += qu_col(0.2) * exp(-r * r * 3.0) * 0.3;

    vec2 sg = floor(p * 130.0);
    float sh = qu_hash(sg);
    if (sh > 0.996) {
        col += vec3(0.9, 0.9, 1.0) * (sh - 0.996) * 200.0 * 0.5;
    }

    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
