// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Solar prominence — giant arching coronal loop connecting two magnetic footpoints, helical plasma flow along the loop, granulated limb surface, episodic flare kicks

const int   LOOP_SAMPS = 100;
const int   SPARKS = 20;
const float STAR_CY = -0.55;    // center of the star (below the frame)
const float STAR_R  = 0.55;     // star radius (so limb sits near bottom-middle)
const float INTENSITY = 0.55;

vec3 prom_pal(float t) {
    vec3 maroon = vec3(0.25, 0.04, 0.08);
    vec3 crimson= vec3(0.85, 0.18, 0.15);
    vec3 orange = vec3(1.00, 0.48, 0.18);
    vec3 amber  = vec3(1.00, 0.78, 0.30);
    vec3 cream  = vec3(1.00, 0.95, 0.70);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(maroon, crimson, s);
    else if (s < 2.0) return mix(crimson, orange, s - 1.0);
    else if (s < 3.0) return mix(orange, amber, s - 2.0);
    else              return mix(amber, cream, s - 3.0);
}

float prom_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }
float prom_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }

float prom_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(prom_hash(i), prom_hash(i + vec2(1, 0)), u.x),
               mix(prom_hash(i + vec2(0, 1)), prom_hash(i + vec2(1, 1)), u.x), u.y);
}

float prom_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < 4; i++) {
        v += a * prom_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

// Parametric coronal loop: s in [0,1], starts at left footpoint, rises into
// an arch, comes down at right footpoint. Footpoints sit on the star limb.
vec2 loopPoint(float s, float t) {
    // Footpoint angular positions on the limb (measured from top of star, +x = right)
    float fp1Ang = 2.15;        // ~ upper-left on the limb
    float fp2Ang = 0.95;        // ~ upper-right on the limb
    // Slight footpoint wobble from magnetic shuffling
    fp1Ang += 0.04 * sin(t * 0.3);
    fp2Ang += 0.04 * cos(t * 0.35);

    vec2 fp1 = vec2(cos(fp1Ang), sin(fp1Ang)) * STAR_R + vec2(0.0, STAR_CY);
    vec2 fp2 = vec2(cos(fp2Ang), sin(fp2Ang)) * STAR_R + vec2(0.0, STAR_CY);

    // Interpolate along an arch (semicircle or skewed)
    vec2 mid = (fp1 + fp2) * 0.5;
    vec2 side = normalize(fp2 - fp1);
    vec2 up = vec2(-side.y, side.x); // perpendicular, pointing away from star center
    float archH = length(fp2 - fp1) * 0.55;

    // Parabola + slight asymmetric kink for drama
    float h = 4.0 * s * (1.0 - s);  // peaks at s=0.5
    float kink = 0.12 * sin(s * 6.28 + t * 0.2);

    return mix(fp1, fp2, s) + up * (archH * h + kink * archH);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.01, 0.03);

    vec2 starCenter = vec2(0.0, STAR_CY);
    float starD = length(p - starCenter);

    // === Star body (lower half) ===
    if (starD < STAR_R) {
        // Limb darkening
        float limb = sqrt(max(0.0, 1.0 - (starD / STAR_R) * (starD / STAR_R)));
        // Granulation (convection cells)
        vec2 localP = (p - starCenter) * 4.0;
        float gran = prom_fbm(localP + vec2(0.0, x_Time * 0.05));
        float gran2 = prom_fbm(localP * 2.5 + 4.0);
        float cells = gran * 0.7 + gran2 * 0.3;
        vec3 base = prom_pal(0.3 + cells * 0.5) * (0.5 + limb * 0.6);
        col += base * 1.0;

        // Bright spots near limb (hot plage regions)
        float spot = smoothstep(0.72, 0.88, cells) * limb;
        col += vec3(1.0, 0.85, 0.45) * spot * 0.6;
    }

    // === Chromosphere ring just above photosphere ===
    if (starD > STAR_R && starD < STAR_R * 1.05) {
        float chromo = exp(-pow(starD - STAR_R * 1.02, 2.0) * 20000.0);
        col += vec3(0.95, 0.35, 0.25) * chromo * 0.8;
    }

    // === Corona glow extending above the limb ===
    float aboveLimb = starD - STAR_R;
    if (aboveLimb > 0.0 && aboveLimb < 0.6) {
        float corona = exp(-aboveLimb * 6.0) * 0.5;
        // Radial streaks
        float radAng = atan(p.y - STAR_CY, p.x);
        float streaks = 0.7 + 0.3 * sin(radAng * 18.0);
        col += vec3(0.90, 0.45, 0.25) * corona * streaks * 0.45;
    }

    // === Coronal loop — sample along its arc ===
    float minLoopD = 1e9;
    float closestS = 0.0;
    for (int i = 0; i < LOOP_SAMPS - 1; i++) {
        float s1 = float(i) / float(LOOP_SAMPS - 1);
        float s2 = float(i + 1) / float(LOOP_SAMPS - 1);
        vec2 a = loopPoint(s1, x_Time);
        vec2 b = loopPoint(s2, x_Time);
        vec2 pa = p - a;
        vec2 ba = b - a;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        float d = length(pa - ba * h);
        if (d < minLoopD) {
            minLoopD = d;
            closestS = mix(s1, s2, h);
        }
    }

    // Loop thickness tapers near footpoints, thicker at apex
    float loopThick = 0.007 + 0.01 * sin(closestS * 3.14159);
    float loopCore = exp(-minLoopD * minLoopD / (loopThick * loopThick) * 2.0);
    float loopHalo = exp(-minLoopD * minLoopD * 3000.0) * 0.25;
    // Plasma flow — advect pattern along loop
    float flow = fract(closestS * 3.0 - x_Time * 0.4);
    float pulse = sin(flow * 6.28 * 3.0) * 0.5 + 0.5;
    // Twist: helical striation perpendicular to loop
    float twist = sin(closestS * 40.0 + x_Time * 2.0) * 0.5 + 0.5;
    float loopIntensity = loopCore * (0.85 + pulse * 0.3 + twist * 0.2);
    col += prom_pal(0.35 + pulse * 0.4) * loopIntensity * 1.3;
    col += prom_pal(0.6) * loopHalo;

    // Footpoint brightening at s≈0 and s≈1
    if (closestS < 0.08) {
        float fp = exp(-pow(closestS / 0.08, 2.0) * 2.0) * loopCore;
        col += vec3(1.0, 0.95, 0.70) * fp * 1.2;
    } else if (closestS > 0.92) {
        float fp = exp(-pow((1.0 - closestS) / 0.08, 2.0) * 2.0) * loopCore;
        col += vec3(1.0, 0.95, 0.70) * fp * 1.2;
    }

    // === Episodic flare kick near the apex ===
    float flareCycle = 9.0;
    float flareT = fract(x_Time / flareCycle);
    if (flareT < 0.12) {
        float fInt = sin(flareT / 0.12 * 3.14159);
        vec2 apex = loopPoint(0.5, x_Time);
        float aD = length(p - apex);
        float flareMask = exp(-aD * aD * 400.0);
        col += vec3(1.0, 0.92, 0.65) * flareMask * fInt * 1.3;
    }

    // === Sparks (plasma droplets falling along the loop) ===
    for (int i = 0; i < SPARKS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float sparkS = fract(prom_hash1(seed) + x_Time * (0.12 + prom_hash1(seed * 3.1) * 0.15));
        vec2 sparkPos = loopPoint(sparkS, x_Time);
        // Offset slightly perpendicular for a spread
        float sparkOff = (prom_hash1(seed * 5.1) - 0.5) * 0.015;
        sparkPos += vec2(sparkOff, sparkOff * 0.5);
        float sd = length(p - sparkPos);
        col += vec3(1.0, 0.7, 0.35) * exp(-sd * sd * 25000.0) * 0.8;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
