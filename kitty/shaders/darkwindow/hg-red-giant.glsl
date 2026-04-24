// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Red giant — swollen evolved star with convection-cell granulation, limb darkening, solar wind rays, and episodic mass-loss puffs

const int   FBM_OCTAVES = 5;
const int   WIND_RAYS = 24;
const int   PUFFS = 6;
const float STAR_R = 0.28;
const float INTENSITY = 0.55;

vec3 rg_pal(float t) {
    vec3 deep    = vec3(0.25, 0.04, 0.06);
    vec3 crimson = vec3(0.75, 0.15, 0.12);
    vec3 orange  = vec3(1.00, 0.48, 0.15);
    vec3 amber   = vec3(1.00, 0.78, 0.35);
    vec3 cream   = vec3(1.00, 0.95, 0.75);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(deep, crimson, s);
    else if (s < 2.0) return mix(crimson, orange, s - 1.0);
    else if (s < 3.0) return mix(orange, amber, s - 2.0);
    else              return mix(amber, cream, s - 3.0);
}

float rg_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }
float rg_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }

float rg_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(rg_hash(i), rg_hash(i + vec2(1, 0)), u.x),
               mix(rg_hash(i + vec2(0, 1)), rg_hash(i + vec2(1, 1)), u.x), u.y);
}

float rg_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < FBM_OCTAVES; i++) {
        v += a * rg_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);
    float ang = atan(p.y, p.x);

    vec3 col = vec3(0.005, 0.01, 0.03);

    // Sparse background starfield
    vec2 sg = floor(p * 140.0);
    float sh = rg_hash(sg);
    if (sh > 0.996) col += vec3(0.9, 0.92, 1.0) * (sh - 0.996) * 200.0 * 0.25;

    // === Star body with convective granulation ===
    if (r < STAR_R) {
        // Limb factor: 1 at center, 0 at edge (for limb darkening)
        float limb = sqrt(max(0.0, 1.0 - (r / STAR_R) * (r / STAR_R)));

        // Fake slow rotation: pretend we're sampling in 3D
        // Project (p.x, p.y, z) onto sphere where z = STAR_R * limb
        vec3 sph = vec3(p.x, p.y, STAR_R * limb) / STAR_R;
        // Slow axial rotation
        float rotA = x_Time * 0.04;
        vec3 rsph = vec3(sph.x * cos(rotA) - sph.z * sin(rotA),
                         sph.y,
                         sph.x * sin(rotA) + sph.z * cos(rotA));

        // Granulation: FBM driven by 3D-ish coords + time
        vec2 granUV = rsph.xy * 6.0 + vec2(rsph.z * 3.0, x_Time * 0.08);
        float gran = rg_fbm(granUV);
        // Secondary, finer scale cells
        float gran2 = rg_fbm(granUV * 2.8 + 7.3);
        float cells = gran * 0.7 + gran2 * 0.3;

        // Hot spots (brighter cells erupting)
        float hotspot = smoothstep(0.68, 0.85, cells);

        // Base red-giant color driven by cell intensity
        vec3 base = rg_pal(0.2 + cells * 0.6);
        // Limb darkening: factor 0.35..1.0
        float darken = 0.35 + 0.65 * pow(limb, 0.6);
        col += base * darken * 1.15;
        // Add hot-spot boost
        col += rg_pal(0.85) * hotspot * darken * 0.8;

        // Subtle H-alpha ring near limb
        float limbRing = exp(-pow(r - STAR_R * 0.96, 2.0) * 4000.0);
        col += vec3(0.95, 0.35, 0.20) * limbRing * 0.5;
    }

    // === Corona glow just outside the photosphere ===
    float coronaD = r - STAR_R;
    if (coronaD > 0.0) {
        float corona = exp(-coronaD * coronaD * 180.0);
        col += vec3(0.85, 0.45, 0.22) * corona * 0.55;
        // Flickery outer halo
        float halo = exp(-coronaD * 6.0) * 0.2;
        float twinkle = 0.7 + 0.3 * sin(x_Time * 3.0 + ang * 5.0);
        col += rg_pal(fract(ang * 0.25 + x_Time * 0.05)) * halo * twinkle * 0.35;
    }

    // === Solar wind rays — radial streaks outside the star ===
    if (r > STAR_R) {
        // Sample angular noise for ray density
        float rayAng = ang * float(WIND_RAYS) / 6.28318;
        float rayBase = floor(rayAng);
        float rayFrac = fract(rayAng);
        // Per-ray hash controls intensity, phase, speed
        float rHash = rg_hash1(rayBase * 3.17);
        float raySpeed = 0.15 + rHash * 0.2;
        // Streak profile: thin, bright near star, fading outward
        float thickness = 0.25 + rHash * 0.5;
        float width = exp(-pow(rayFrac - 0.5, 2.0) * (40.0 + thickness * 60.0));
        // Length fades
        float along = (r - STAR_R) / 0.75;
        if (along >= 0.0 && along < 1.0) {
            // Flowing outward (advection pattern)
            float flow = fract(along * 2.0 - x_Time * raySpeed + rHash);
            float pulse = smoothstep(0.0, 0.15, flow) * smoothstep(1.0, 0.85, flow);
            float fade = 1.0 - along;
            float rayMask = width * pulse * fade;
            col += rg_pal(fract(0.4 + rHash * 0.4 + x_Time * 0.02)) * rayMask * 0.9;
        }
    }

    // === Episodic mass-loss puffs (coronal mass ejections) ===
    for (int i = 0; i < PUFFS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        // Each puff repeats over its own cycle
        float cycle = 6.0 + rg_hash1(seed * 1.7) * 4.0;
        float phase = fract((x_Time + rg_hash1(seed * 5.1) * cycle) / cycle);
        // Eject direction + travel
        float puffAng = rg_hash1(seed) * 6.28;
        float travel = phase * 0.55;
        vec2 puffOrigin = vec2(cos(puffAng), sin(puffAng)) * STAR_R;
        vec2 puffPos = puffOrigin + vec2(cos(puffAng), sin(puffAng)) * travel;
        float pd = length(p - puffPos);
        // Expanding clump: size grows with travel
        float puffSize = 0.025 + phase * 0.06;
        float puffMask = exp(-pd * pd / (puffSize * puffSize) * 1.2);
        float puffFade = 1.0 - phase;
        col += rg_pal(0.55 + rg_hash1(seed * 11.0) * 0.3) * puffMask * puffFade * 0.9;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
