// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Volcanic eruption — spewing molten rock particles + smoke plume + lava glow

const int   PARTICLES = 100;
const float INTENSITY = 0.55;

vec3 ve_col(float heat) {
    vec3 dark = vec3(0.08, 0.01, 0.02);
    vec3 red  = vec3(0.95, 0.20, 0.05);
    vec3 orange = vec3(1.00, 0.55, 0.15);
    vec3 yellow = vec3(1.00, 0.90, 0.45);
    vec3 white  = vec3(1.00, 0.98, 0.90);
    if (heat < 0.25)      return mix(dark, red, heat * 4.0);
    else if (heat < 0.5)  return mix(red, orange, (heat - 0.25) * 4.0);
    else if (heat < 0.75) return mix(orange, yellow, (heat - 0.5) * 4.0);
    else                  return mix(yellow, white, (heat - 0.75) * 4.0);
}

float ve_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Dark backdrop with ember glow from below
    vec3 col = mix(vec3(0.02, 0.01, 0.04), vec3(0.20, 0.05, 0.02), smoothstep(0.2, -0.5, p.y));

    // Crater/peak silhouette — triangle at bottom
    vec2 crater = vec2(0.0, -0.3);
    float peakSlope = 0.5;
    float fromCrater = abs(p.x - crater.x);
    float peakY = crater.y - fromCrater * peakSlope;
    if (p.y < peakY && fromCrater < 0.5) {
        col = vec3(0.02, 0.01, 0.02);
        // Lava flows down peak — sparse glowing veins
        float veinPhase = atan(p.x - crater.x, crater.y - p.y);
        float veinPattern = 0.5 + 0.5 * cos(veinPhase * 20.0 + x_Time * 0.5);
        float veinMask = smoothstep(0.0, 0.1, peakY - p.y) * smoothstep(0.4, 0.0, fromCrater);
        col += ve_col(0.7) * pow(veinPattern, 8.0) * veinMask * 0.5;
    }

    // Particles erupting up and outward
    for (int i = 0; i < PARTICLES; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float life = fract(x_Time * 0.3 + ve_hash(seed));
        // Initial velocity — upward + random horizontal
        float launchAng = -1.2 + (ve_hash(seed * 3.1) - 0.5) * 1.6;   // near vertical with spread
        float speed = 1.5 + ve_hash(seed * 5.1) * 0.7;
        vec2 vel = vec2(sin(launchAng), cos(launchAng)) * speed;
        // Position: projectile motion from crater
        vec2 pos = crater + vel * life;
        pos.y -= 2.0 * life * life;  // gravity

        float pd = length(p - pos);
        float partSize = 0.004 * (1.0 - life * 0.6);
        float core = exp(-pd * pd / (partSize * partSize) * 2.0);
        // Heat cools as particle arcs (higher at start)
        float heat = 0.9 - life * 0.6;
        float halo = exp(-pd * pd * 1500.0) * 0.25 * heat;
        col += ve_col(heat) * (core * 1.4 + halo);
    }

    // Rising smoke plume — darker FBM
    if (p.y > peakY && fromCrater < 0.25) {
        float plumeHeight = p.y - peakY;
        float plumeWidth = 0.08 + plumeHeight * 0.2;
        if (fromCrater < plumeWidth) {
            float plumeMask = exp(-fromCrater * fromCrater / (plumeWidth * plumeWidth) * 2.0);
            // Turbulent
            float plumePhase = p.y * 10.0 - x_Time * 2.0;
            float turb = sin(plumePhase + p.x * 20.0) * 0.5 + 0.5;
            float fadeUp = smoothstep(1.0, 0.0, plumeHeight);
            col = mix(col, vec3(0.15, 0.12, 0.18), plumeMask * turb * fadeUp * 0.7);
        }
    }

    // Bottom lava pool glow
    float poolGlow = smoothstep(-0.3, -0.5, p.y) * exp(-fromCrater * fromCrater * 8.0);
    col += ve_col(0.85) * poolGlow * 0.7;

    // Crater mouth glow (bright orange spot)
    float craterD = length(p - crater);
    col += ve_col(0.9) * exp(-craterD * craterD * 200.0) * 1.2;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
