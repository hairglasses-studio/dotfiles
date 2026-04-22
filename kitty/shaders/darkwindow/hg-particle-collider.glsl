// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Particle collider — 2 beams converging at center, collision creates jet sprays + reconstructed tracks

const int   BEAM_PARTICLES = 40;
const int   JET_TRACKS = 20;
const float INTENSITY = 0.55;

vec3 pc_col(float t) {
    vec3 blue = vec3(0.25, 0.55, 0.98);
    vec3 white = vec3(1.00, 0.98, 0.90);
    vec3 mag  = vec3(0.95, 0.30, 0.65);
    vec3 gold = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(blue, white, s);
    else if (s < 2.0) return mix(white, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, gold, s - 2.0);
    else              return mix(gold, blue, s - 3.0);
}

float pc_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.04);

    // Beam pipe — horizontal line at y=0
    float pipeD = abs(p.y);
    if (pipeD < 0.004) {
        col = vec3(0.15, 0.18, 0.25);
    }

    // Interaction point = origin
    vec2 IP = vec2(0.0);

    // Collision cycle
    float cycle = 3.0;
    float phase = mod(x_Time, cycle) / cycle;

    // Incoming beams — particles accelerating toward IP
    for (int i = 0; i < BEAM_PARTICLES; i++) {
        float fi = float(i);
        bool leftBeam = mod(fi, 2.0) < 1.0;
        float side = leftBeam ? 1.0 : -1.0;
        float particlePhase = fract(x_Time * 0.6 + fi * 0.025);

        if (phase < 0.4) {
            // Pre-collision: beam traveling
            float particleX = -side * 0.8 + side * particlePhase * 0.8;
            vec2 partPos = vec2(particleX, (pc_hash(fi * 3.7) - 0.5) * 0.006);
            float pd = length(p - partPos);
            col += pc_col(0.15) * exp(-pd * pd * 50000.0) * 0.9;
        }
    }

    // Collision flash at phase 0.4
    if (phase > 0.4 && phase < 0.5) {
        float flashPhase = (phase - 0.4) / 0.1;
        float flash = exp(-length(p - IP) * length(p - IP) * 200.0) * (1.0 - flashPhase);
        col += vec3(1.0) * flash * 2.5;
        col += pc_col(0.9) * exp(-length(p - IP) * length(p - IP) * 20.0) * (1.0 - flashPhase) * 0.8;
    }

    // Post-collision: reconstructed jet tracks
    if (phase > 0.5) {
        float jetPhase = (phase - 0.5) / 0.5;
        for (int t = 0; t < JET_TRACKS; t++) {
            float ft = float(t);
            float trackSeed = ft * 7.31;
            float ang = pc_hash(trackSeed) * 6.28;
            float len = 0.5 + pc_hash(trackSeed * 3.1) * 0.4;
            vec2 trackEnd = IP + vec2(cos(ang), sin(ang)) * len * jetPhase;

            // Distance from p to track line
            vec2 dir = normalize(trackEnd - IP);
            vec2 perp = vec2(-dir.y, dir.x);
            float along = dot(p - IP, dir);
            float perpD = abs(dot(p - IP, perp));
            if (along > 0.0 && along < length(trackEnd - IP)) {
                // Track curvature from magnetic field (slight spiral)
                float curve = sin(along * 10.0 + trackSeed) * 0.015;
                float effPerpD = abs(perpD - curve * along);
                float trackMask = exp(-effPerpD * effPerpD * 50000.0);
                float fadeAlong = 1.0 - along / len;
                col += pc_col(fract(trackSeed * 0.02 + jetPhase + x_Time * 0.04)) * trackMask * fadeAlong * 0.9;
            }
        }

        // Particle end-points (detector hits)
        for (int t = 0; t < JET_TRACKS; t++) {
            float ft = float(t);
            float trackSeed = ft * 7.31;
            float ang = pc_hash(trackSeed) * 6.28;
            float len = 0.5 + pc_hash(trackSeed * 3.1) * 0.4;
            vec2 endP = IP + vec2(cos(ang), sin(ang)) * len * jetPhase;
            float ed = length(p - endP);
            col += vec3(1.0, 0.9, 0.7) * exp(-ed * ed * 30000.0) * 1.2;
        }
    }

    // Detector ring in background
    float ringD = abs(length(p - IP) - 0.7);
    col += vec3(0.15, 0.2, 0.3) * smoothstep(0.01, 0.005, ringD) * 0.6;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
