// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Fireworks display — 8 simultaneous bursts with radial particle trails + star-core + smoke

const int   BURSTS      = 8;
const int   PARTICLES_PER = 32;
const float INTENSITY = 0.55;

vec3 fw_pal(float t) {
    vec3 a = vec3(1.00, 0.80, 0.30);
    vec3 b = vec3(0.95, 0.30, 0.55);
    vec3 c = vec3(0.55, 0.30, 0.98);
    vec3 d = vec3(0.20, 0.85, 0.95);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float fw_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  fw_hash2(float n) { return vec2(fw_hash(n), fw_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = mix(vec3(0.01, 0.01, 0.04), vec3(0.03, 0.02, 0.08), 1.0 - uv.y);

    for (int b = 0; b < BURSTS; b++) {
        float fb = float(b);
        float cycle = 2.0 + fw_hash(fb * 2.0) * 1.5;
        float phase = fract((x_Time + fw_hash(fb) * cycle) / cycle);
        float cycleID = floor((x_Time + fw_hash(fb) * cycle) / cycle);
        float seed = fb * 11.3 + cycleID * 13.7;

        // Position per cycle
        vec2 center = fw_hash2(seed) * 1.4 - 0.7;
        center.x *= x_WindowSize.x / x_WindowSize.y;

        // Phase: 0-0.05 = rising trail; 0.05-0.07 = explode; 0.07-1.0 = particle expansion
        vec2 burstCenter = center;
        burstCenter.y += (phase < 0.05) ? (0.05 - phase) * 8.0 : 0.0;  // rising trail

        if (phase < 0.05) {
            // Rising rocket trail
            float trailY = (0.05 - phase) * 8.0;
            float trailD = abs(p.x - center.x);
            float alongTrail = (p.y - (center.y - trailY));
            if (alongTrail > 0.0 && alongTrail < trailY) {
                float trailMask = exp(-trailD * trailD * 10000.0);
                col += vec3(1.0, 0.85, 0.4) * trailMask * 0.8;
            }
            continue;
        }

        if (phase < 0.07) {
            // Explosion flash
            float flashR = (phase - 0.05) * 10.0;
            float flashD = length(p - center);
            float flash = exp(-flashD * flashD / (flashR * flashR) * 2.0);
            col += vec3(1.0) * flash * 2.0;
            continue;
        }

        // Particle expansion
        float expandPhase = (phase - 0.07) / 0.93;
        float expandR = pow(expandPhase, 0.7) * 0.4;
        float expandFade = pow(1.0 - expandPhase, 1.5);

        vec3 burstCol = fw_pal(fract(seed * 0.03));

        // Particles radiate from center
        for (int pi = 0; pi < PARTICLES_PER; pi++) {
            float fpi = float(pi);
            float particleAng = fpi / float(PARTICLES_PER) * 6.28 + seed;
            // Slight angular jitter
            particleAng += (fw_hash(seed + fpi) - 0.5) * 0.1;
            float particleSpeed = 0.8 + fw_hash(seed + fpi * 3.1) * 0.4;
            vec2 particlePos = center + vec2(cos(particleAng), sin(particleAng)) * expandR * particleSpeed;
            // Gravity pulls down
            particlePos.y -= expandPhase * expandPhase * 0.15;
            float pd = length(p - particlePos);
            // Core + trail
            float core = exp(-pd * pd * 20000.0);
            // Trail pointing toward center (motion blur)
            vec2 toCenterFromPart = normalize(center - particlePos);
            vec2 trailCheck = p - particlePos;
            float trailAlong = dot(trailCheck, toCenterFromPart);
            float trailPerp = abs(dot(trailCheck, vec2(-toCenterFromPart.y, toCenterFromPart.x)));
            float trail = 0.0;
            if (trailAlong > 0.0 && trailAlong < 0.02) {
                trail = exp(-trailPerp * trailPerp * 20000.0) * (1.0 - trailAlong / 0.02) * 0.5;
            }

            col += burstCol * (core * 1.4 + trail) * expandFade;
        }

        // Residual smoke
        if (expandPhase > 0.5) {
            float smokeD = length(p - center);
            float smoke = exp(-smokeD * smokeD * 15.0) * (expandPhase - 0.5) * 0.4 * expandFade;
            col += burstCol * smoke * 0.3;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
