// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Ghost in the Shell — thermoptic-camo dissolve: humanoid silhouette materializes from digital noise with pixel particles flying inward, steady phase with ripple edge, then dissolves outward. Background digital glitch pattern.

const int   PARTICLES = 110;
const float INTENSITY = 0.55;
const float CYCLE = 10.0;

vec3 gis_pal(float t) {
    vec3 cyan   = vec3(0.20, 0.85, 1.00);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 green  = vec3(0.30, 0.95, 0.55);
    vec3 white  = vec3(0.95, 0.98, 1.00);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, green, s - 2.0);
    else if (s < 4.0) return mix(green, white, s - 3.0);
    else              return mix(white, cyan, s - 4.0);
}

float gis_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float gis_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

// Humanoid silhouette function. Returns a soft mask [0, 1].
// The silhouette is centered at (0, 0).
float silhouetteMask(vec2 q) {
    // Head: circle at y~0.30
    vec2 headC = vec2(0.0, 0.35);
    float headR = 0.08;
    float head = smoothstep(headR + 0.008, headR - 0.008, length(q - headC));

    // Torso: rounded rectangle
    vec2 torsoC = vec2(0.0, 0.08);
    vec2 torsoR = vec2(0.14, 0.18);
    vec2 td = abs(q - torsoC) - torsoR;
    float torsoD = length(max(td, 0.0)) + min(max(td.x, td.y), 0.0);
    float torso = smoothstep(0.02, 0.0, torsoD);

    // Arms: angled rectangles on each side
    float lArm = 0.0, rArm = 0.0;
    {
        vec2 relL = q - vec2(-0.15, 0.03);
        // rotate slightly
        float c = cos(0.15), s = sin(0.15);
        vec2 rotL = vec2(relL.x * c - relL.y * s, relL.x * s + relL.y * c);
        vec2 aR = vec2(0.035, 0.12);
        vec2 ad = abs(rotL) - aR;
        float armD = length(max(ad, 0.0)) + min(max(ad.x, ad.y), 0.0);
        lArm = smoothstep(0.02, 0.0, armD);
    }
    {
        vec2 relR = q - vec2(0.15, 0.03);
        float c = cos(-0.15), s = sin(-0.15);
        vec2 rotR = vec2(relR.x * c - relR.y * s, relR.x * s + relR.y * c);
        vec2 aR = vec2(0.035, 0.12);
        vec2 ad = abs(rotR) - aR;
        float armD = length(max(ad, 0.0)) + min(max(ad.x, ad.y), 0.0);
        rArm = smoothstep(0.02, 0.0, armD);
    }

    // Legs: two vertical rectangles
    float lLeg = 0.0, rLeg = 0.0;
    {
        vec2 rel = q - vec2(-0.05, -0.22);
        vec2 lR = vec2(0.04, 0.15);
        vec2 ld = abs(rel) - lR;
        float legD = length(max(ld, 0.0)) + min(max(ld.x, ld.y), 0.0);
        lLeg = smoothstep(0.02, 0.0, legD);
    }
    {
        vec2 rel = q - vec2(0.05, -0.22);
        vec2 lR = vec2(0.04, 0.15);
        vec2 ld = abs(rel) - lR;
        float legD = length(max(ld, 0.0)) + min(max(ld.x, ld.y), 0.0);
        rLeg = smoothstep(0.02, 0.0, legD);
    }

    return clamp(head + torso + lArm + rArm + lLeg + rLeg, 0.0, 1.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.004, 0.008, 0.025);

    // === Background digital glitch pattern ===
    vec2 gridCell = floor(p * 80.0);
    float cellHash = gis_hash2(gridCell + floor(x_Time * 8.0));
    if (cellHash > 0.985) {
        col += gis_pal(cellHash) * 0.4;
    }
    // Horizontal scanline flicker bands
    float scanFlick = step(0.997, gis_hash(floor(p.y * 40.0) + floor(x_Time * 3.0)));
    col += vec3(0.15, 0.25, 0.35) * scanFlick * 0.3;

    // === Cycle phase ===
    float cycT = mod(x_Time, CYCLE) / CYCLE;
    // Phases: 0.0-0.3 materialize in, 0.3-0.7 stable with shimmer, 0.7-1.0 dissolve out
    float revealAmount;
    bool materializing = cycT < 0.3;
    if (cycT < 0.3) {
        revealAmount = cycT / 0.3;
    } else if (cycT < 0.7) {
        revealAmount = 1.0;
    } else {
        revealAmount = 1.0 - (cycT - 0.7) / 0.3;
    }

    // === Silhouette rendering ===
    float silm = silhouetteMask(p);
    if (silm > 0.0) {
        // Dissolve threshold based on per-pixel hash vs revealAmount
        float dissHash = gis_hash2(floor(p * 120.0));
        float dissolved = step(dissHash, revealAmount);

        // Body glow when fully revealed
        vec3 bodyCol = gis_pal(fract(x_Time * 0.1));
        float shimmer = 0.7 + 0.3 * sin(x_Time * 3.0 + p.y * 30.0);
        col += bodyCol * silm * dissolved * shimmer * 0.85;

        // Edge shimmer (regardless of dissolve)
        // Estimate edge by sampling silhouette at slightly-offset coords
        float silEdge = silm * (1.0 - silm * 2.0);
        silEdge = max(silEdge, 0.0);
        // Actually use gradient magnitude proxy: silm transitions fast at edges
        // Cheap proxy: the more silm deviates from 0 or 1, the more edge
        float edgeStrength = 4.0 * silm * (1.0 - silm);
        col += gis_pal(fract(0.4 + x_Time * 0.2)) * edgeStrength * revealAmount * 0.9;

        // Dark internal lines (cyberware circuitry suggestion)
        vec2 circuit = fract(p * 12.0);
        float circuitLine = step(0.92, max(abs(circuit.x - 0.5), abs(circuit.y - 0.5)));
        col -= vec3(0.1, 0.05, 0.15) * circuitLine * silm * revealAmount * 0.4;
    }

    // === Pixel particles flying toward / from silhouette ===
    for (int i = 0; i < PARTICLES; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        // Each particle has a random destination inside the silhouette
        vec2 dest = vec2(gis_hash(seed) * 0.3 - 0.15, gis_hash(seed * 3.7) * 0.6 - 0.2);
        // Start from random far position
        float ang = gis_hash(seed * 5.1) * 6.28;
        float farR = 1.2;
        vec2 far = vec2(cos(ang), sin(ang)) * farR;

        float pPhase;
        if (materializing) {
            // Flying from far to dest
            pPhase = cycT / 0.3;
        } else if (cycT < 0.7) {
            // Stable: particles stay at dest (no render)
            continue;
        } else {
            // Dissolving: flying outward from dest
            pPhase = 1.0 - (cycT - 0.7) / 0.3;
        }
        // Offset per particle (staggered arrival)
        float offset = gis_hash(seed * 11.0) * 0.4;
        float indivPhase = clamp((pPhase - offset) / (1.0 - offset), 0.0, 1.0);
        if (indivPhase <= 0.0 || indivPhase >= 1.0) continue;
        // Position
        vec2 partPos = mix(far, dest, materializing ? indivPhase : 1.0 - indivPhase);
        if (!materializing) partPos = mix(dest, far, 1.0 - indivPhase);

        float pd = length(p - partPos);
        float pSize = 0.003 + gis_hash(seed * 13.0) * 0.004;
        float pCore = exp(-pd * pd / (pSize * pSize) * 1.5);
        // Bright particle
        col += gis_pal(fract(seed * 0.1 + x_Time * 0.1)) * pCore * 1.1;
        // Trail
        vec2 dir = normalize(dest - far);
        if (!materializing) dir = -dir;
        float along = dot(p - partPos, -dir);
        float perp = length(p - partPos - (-dir) * along);
        if (along > 0.0 && along < 0.04) {
            float trailMask = exp(-perp * perp * 100000.0) * (1.0 - along / 0.04);
            col += gis_pal(fract(seed * 0.1)) * trailMask * 0.4;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
