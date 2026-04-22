// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Ion engine plume — segmented stage-thrust with magnetic confinement rings + blue glow

const int   STAGES    = 5;
const int   PARTICLES = 60;
const float INTENSITY = 0.55;

vec3 ie_pal(float t) {
    vec3 white = vec3(0.95, 0.98, 1.00);
    vec3 cyan  = vec3(0.20, 0.85, 0.98);
    vec3 blue  = vec3(0.15, 0.45, 0.95);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    if (t > 0.75)      return mix(cyan, white, (t - 0.75) * 4.0);
    else if (t > 0.5)  return mix(blue, cyan, (t - 0.5) * 4.0);
    else if (t > 0.25) return mix(vio, blue, (t - 0.25) * 4.0);
    else               return vio * (t * 4.0);
}

float ie_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  ie_hash2(float n) { return vec2(ie_hash(n), ie_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Engine mounted at top of screen, plume extending downward
    vec2 enginePos = vec2(0.0, 0.4);
    // Plume direction = down
    vec2 plumeDir = vec2(0.0, -1.0);
    vec2 perpDir = vec2(1.0, 0.0);

    vec3 col = vec3(0.01, 0.01, 0.04);

    // Project p onto plume axis
    vec2 toP = p - enginePos;
    float along = dot(toP, plumeDir);   // positive = downstream
    float perpD = abs(dot(toP, perpDir));

    // Engine body — small thruster chamber above
    if (along < 0.0 && abs(along) < 0.08 && perpD < 0.05) {
        col = vec3(0.15, 0.15, 0.20);
        // Nozzle lip
        if (abs(along) > 0.06) col = vec3(0.4, 0.35, 0.45);
    }

    // Plume stages — segmented with gaps (multi-grid ion thruster)
    if (along > 0.0 && along < 1.5) {
        // Width of plume grows then tapers
        float plumeR = 0.03 + along * 0.08 - pow(along / 1.5, 2.0) * 0.05;

        if (perpD < plumeR) {
            float rNorm = perpD / plumeR;
            // Brighter core, dimmer edges
            float coreProfile = exp(-rNorm * rNorm * 3.0);

            // Stage segmentation — periodic brightness bumps
            float stagePhase = along * float(STAGES) - x_Time * 0.4;
            float stageBump = 0.5 + 0.5 * sin(stagePhase * 6.28);
            stageBump = pow(stageBump, 2.0);

            // Distance-based temperature fall
            float heat = (1.0 - along * 0.6) * (0.7 + stageBump * 0.6);

            vec3 plumeCol = ie_pal(heat);
            col += plumeCol * coreProfile * 1.3;

            // Confinement rings — bright thin rings where magnetic field constricts
            for (int k = 0; k < STAGES; k++) {
                float fk = float(k);
                float ringPos = (fk + 0.5) / float(STAGES) * 1.2;
                float ringD = abs(along - ringPos);
                if (ringD < 0.015 && perpD < plumeR * 0.9 && perpD > plumeR * 0.6) {
                    float ring = exp(-ringD * ringD / (0.008 * 0.008) * 2.0);
                    col += ie_pal(0.9) * ring * 0.9;
                }
            }
        }

        // Ion particles streaming downstream
        for (int i = 0; i < PARTICLES; i++) {
            float fi = float(i);
            float seed = fi * 3.71;
            float partSpeed = 0.4 + ie_hash(seed) * 0.4;
            float partPhase = fract(x_Time * partSpeed + ie_hash(seed * 3.7));
            float partAlong = partPhase * 1.4;
            float partPerp = (ie_hash(seed * 5.3) - 0.5) * 0.12 * (0.3 + partAlong * 0.4);
            vec2 partPos = enginePos + plumeDir * partAlong + perpDir * partPerp;
            float pd = length(p - partPos);
            float core2 = exp(-pd * pd * 30000.0);
            col += ie_pal(0.8 - partAlong * 0.3) * core2 * 1.2;
        }

        // Halo outside plume
        float outerHalo = exp(-(perpD - plumeR) * (perpD - plumeR) * 600.0) * 0.2;
        if (perpD > plumeR) {
            col += ie_pal(0.4) * outerHalo * smoothstep(0.0, 1.5, along) * smoothstep(2.0, 1.2, along);
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
