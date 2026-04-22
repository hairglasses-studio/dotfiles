// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Pulsar with rotating lighthouse beams + magnetic cone jets

const float PULSE_RATE = 3.2;       // spins per second (fast neutron star)
const float JET_WIDTH  = 0.12;
const int   BEAM_SAMPS = 40;
const float INTENSITY  = 0.55;

vec3 pj_pal(float t) {
    vec3 hot   = vec3(1.00, 0.98, 0.90);
    vec3 cyan  = vec3(0.10, 0.82, 0.92);
    vec3 mag   = vec3(0.90, 0.30, 0.70);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(hot, cyan, s);
    else if (s < 2.0) return mix(cyan, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, vio, s - 2.0);
    else              return mix(vio, hot, s - 3.0);
}

float pj_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Neutron star core — tiny bright point
    float r = length(p);
    float core = exp(-r * r * 800.0) * 1.8;
    vec3 col = vec3(1.0, 0.98, 0.9) * core;

    // Core halo
    col += pj_pal(0.0) * exp(-r * r * 80.0) * 0.5;

    // Rotating beam axis — spinning pulsar
    float beamAngle = x_Time * PULSE_RATE * 6.28318;
    vec2 beamDir = vec2(cos(beamAngle), sin(beamAngle));
    vec2 beamPerp = vec2(-beamDir.y, beamDir.x);

    // Two beams (opposite directions)
    for (int sign = 0; sign < 2; sign++) {
        float signMul = sign == 0 ? 1.0 : -1.0;
        vec2 dir = beamDir * signMul;
        // Project p onto beam axis
        float along = dot(p, dir);
        float perp  = dot(p, beamPerp);
        if (along < 0.0) continue;

        // Beam width narrows with distance (or keep constant — here constant cone)
        float beamW = JET_WIDTH * (1.0 + along * 0.6);
        float widthMask = exp(-perp * perp / (beamW * beamW) * 2.0);

        // Turbulent structure along jet
        float jetTurb = 0.0;
        for (int k = 0; k < BEAM_SAMPS; k++) {
            float fk = float(k);
            float t = fk / float(BEAM_SAMPS);
            // Sample along beam
            float noise = pj_hash(vec2(along * 15.0 + fk * 3.0, x_Time * 2.0));
            jetTurb += noise * exp(-fk * 0.1);
        }
        jetTurb /= float(BEAM_SAMPS);

        // Beam brightness fades with distance
        float beamBright = exp(-along * 0.8);
        vec3 beamCol = pj_pal(fract(along * 0.5 + x_Time * 0.1));
        col += beamCol * widthMask * beamBright * (0.5 + jetTurb * 1.2) * 1.1;
    }

    // Magnetic field lines — spiral outward slowly
    float fieldAngle = atan(p.y, p.x);
    float fieldSpiral = fieldAngle + log(r + 0.05) * 0.5 + x_Time * 0.1;
    float fieldLines = pow(0.5 + 0.5 * cos(fieldSpiral * 8.0), 16.0);
    float fieldMask = smoothstep(0.05, 0.5, r) * smoothstep(1.0, 0.3, r);
    col += pj_pal(0.2) * fieldLines * fieldMask * 0.2;

    // Pulse flash — quick strobe visible everywhere at pulse rate
    float strobePhase = fract(x_Time * PULSE_RATE * 0.5);
    float strobe = pow(1.0 - strobePhase, 8.0) * 0.3;
    col += pj_pal(0.0) * strobe * exp(-r * r * 5.0);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
