// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Oort cloud comet — icy body leaving an outer-solar-system shell, plunging toward a distant sun with twin (ion + dust) tails pointing away from the star

const int   OORT_BODIES = 110;
const int   BG_STARS = 60;
const float INTENSITY = 0.55;
const float CYCLE = 16.0;

vec3 ort_pal(float t) {
    vec3 deep    = vec3(0.03, 0.05, 0.18);
    vec3 blue    = vec3(0.20, 0.55, 0.95);
    vec3 cyan    = vec3(0.30, 0.90, 1.00);
    vec3 white   = vec3(1.00, 0.98, 0.95);
    vec3 amber   = vec3(1.00, 0.78, 0.38);
    vec3 orange  = vec3(1.00, 0.48, 0.18);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(deep, blue, s);
    else if (s < 2.0) return mix(blue, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, white, s - 2.0);
    else if (s < 4.0) return mix(white, amber, s - 3.0);
    else              return mix(amber, orange, s - 4.0);
}

float ort_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Comet trajectory: enters from upper-left (oort shell), arcs toward sun, exits
// Parameterize by cyT in [0,1]. Sun is at SUN_POS.
vec2 cometPath(float cyT) {
    // Start far at (-0.9, 0.55) — exit at (0.95, -0.40) through perihelion
    vec2 start = vec2(-0.90, 0.55);
    vec2 end   = vec2( 0.95, -0.40);
    vec2 mid   = vec2( 0.18,  0.06);   // bends around the sun
    // Quadratic Bezier
    float t = cyT;
    float u = 1.0 - t;
    return u * u * start + 2.0 * u * t * mid + t * t * end;
}

vec2 sunPos() { return vec2(0.22, 0.04); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.003, 0.007, 0.020);

    // === Background stars ===
    for (int i = 0; i < BG_STARS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        vec2 sp = vec2(ort_hash(seed) * 2.0 - 1.0, ort_hash(seed * 3.7) * 1.6 - 0.8);
        float sd = length(p - sp);
        float mag = 0.3 + ort_hash(seed * 5.1) * 0.5;
        col += vec3(0.85, 0.9, 1.0) * exp(-sd * sd * 40000.0) * mag * 0.3;
    }

    // === Distant sun (small bright point with soft halo) ===
    vec2 sun = sunPos();
    float sunD = length(p - sun);
    col += vec3(1.0, 0.95, 0.75) * exp(-sunD * sunD * 15000.0) * 1.4;
    col += vec3(1.0, 0.75, 0.40) * exp(-sunD * sunD * 80.0) * 0.35;

    // === Oort cloud shell — scattered icy bodies at the edges of the frame ===
    for (int i = 0; i < OORT_BODIES; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        // Angular + radial scattering for shell appearance (big radius)
        float sh_ang = ort_hash(seed) * 6.28;
        float sh_r = 0.80 + ort_hash(seed * 3.7) * 0.25;
        // Slow tangential drift
        sh_ang += x_Time * (0.005 + ort_hash(seed * 5.1) * 0.01);
        vec2 bodyPos = sun + vec2(cos(sh_ang), sin(sh_ang)) * sh_r;
        float bd = length(p - bodyPos);
        float bSize = 0.002 + ort_hash(seed * 7.3) * 0.003;
        float bCore = exp(-bd * bd / (bSize * bSize) * 1.2);
        // Subtle twinkle
        float twink = 0.5 + 0.5 * sin(x_Time * (0.5 + ort_hash(seed * 11.0)) + seed);
        col += vec3(0.7, 0.8, 1.0) * bCore * (0.5 + twink * 0.5) * 0.4;
    }

    // === Active comet (currently plunging inward) ===
    float cyT = mod(x_Time, CYCLE) / CYCLE;
    vec2 cometP = cometPath(cyT);
    float cd = length(p - cometP);

    // Comet brightness ramps up near perihelion
    vec2 cToSun = sun - cometP;
    float distToSun = length(cToSun);
    float heat = 1.0 / (1.0 + distToSun * distToSun * 4.0);

    // Nucleus
    float nucCore = exp(-cd * cd * 20000.0);
    col += vec3(1.0, 0.95, 0.80) * nucCore * 1.5 * (0.6 + heat);
    // Coma
    float coma = exp(-cd * cd * 400.0);
    col += vec3(0.85, 0.92, 1.0) * coma * 0.55 * (0.5 + heat);

    // Direction vectors for tails: ion tail straight away from sun, dust tail
    // slightly curved (lagging behind)
    vec2 awayFromSun = normalize(cometP - sun);
    // Ion tail (blue, straight, long)
    {
        vec2 pc = p - cometP;
        float along = dot(pc, awayFromSun);
        float perp = length(pc - awayFromSun * along);
        if (along > 0.0 && along < 0.45) {
            float width = 0.006 + along * 0.08;
            float tailMask = exp(-perp * perp / (width * width) * 1.8);
            float fade = (1.0 - along / 0.45);
            // Stripes along the tail to show plasma bunches
            float stripes = 0.7 + 0.3 * sin(along * 80.0 - x_Time * 5.0);
            col += vec3(0.35, 0.75, 1.0) * tailMask * fade * stripes * heat * 1.2;
        }
    }
    // Dust tail (yellow-white, curved)
    {
        // Curve direction: rotate awayFromSun by a small angle
        float curveAng = 0.25;
        mat2 rotC = mat2(cos(curveAng), -sin(curveAng), sin(curveAng), cos(curveAng));
        vec2 dustDir = rotC * awayFromSun;
        vec2 pc = p - cometP;
        float along = dot(pc, dustDir);
        // Add curvature: as we go along, bend further
        // (approximate by sampling a rotated dir at each along)
        float perp = length(pc - dustDir * along);
        if (along > 0.0 && along < 0.35) {
            float width = 0.008 + along * 0.10;
            float tailMask = exp(-perp * perp / (width * width) * 1.5);
            float fade = (1.0 - along / 0.35);
            col += vec3(1.0, 0.88, 0.55) * tailMask * fade * heat * 0.9;
        }
    }

    // Bright flash when very close to perihelion
    if (distToSun < 0.25) {
        float flash = exp(-distToSun * 15.0) * heat;
        col += vec3(1.0, 0.92, 0.70) * flash * 0.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
