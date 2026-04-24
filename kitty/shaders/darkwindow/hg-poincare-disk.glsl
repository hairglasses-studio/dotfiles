// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Poincaré disk — Escher-inspired hyperbolic tiling inside the unit disk via hyperbolic distance and rotating Möbius translation. Pattern infinitely accumulates toward the disk boundary, with angular sectors, hyperbolic rings, and a bright disk edge.

const float INTENSITY = 0.55;
const float DISK_R = 0.78;
const int   SECTORS = 8;     // angular symmetry
const float RING_DENSITY = 2.8;  // number of hyperbolic rings per unit d_h

vec3 poi_pal(float t) {
    vec3 cyan   = vec3(0.25, 0.90, 1.00);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 amber  = vec3(1.00, 0.70, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, amber, s - 2.0);
    else if (s < 4.0) return mix(amber, mint, s - 3.0);
    else              return mix(mint, cyan, s - 4.0);
}

// Möbius translation in the Poincaré disk: z → (z - a) / (1 - conj(a) z)
// where a is a complex parameter. Keep |a| < 1.
// For complex multiplication in 2D, (a+bi)(c+di) = (ac-bd) + (ad+bc)i
vec2 mobiusTranslate(vec2 z, vec2 a) {
    // Numerator: z - a
    vec2 num = z - a;
    // Denominator: 1 - conj(a) * z.  conj(a) = (a.x, -a.y)
    vec2 conjA = vec2(a.x, -a.y);
    float denR = 1.0 - (conjA.x * z.x - conjA.y * z.y);
    float denI = -(conjA.x * z.y + conjA.y * z.x);
    // num / den = (numR + numI i) / (denR + denI i)
    float denMag2 = denR * denR + denI * denI;
    return vec2((num.x * denR + num.y * denI) / denMag2,
                (num.y * denR - num.x * denI) / denMag2);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.004, 0.007, 0.020);

    // Normalize into the disk: scale by DISK_R so that p within DISK_R maps to z within unit disk
    vec2 z = p / DISK_R;
    float r = length(z);

    if (r < 1.0) {
        // Apply a slowly-varying Möbius translation to simulate a hyperbolic flow
        vec2 a = vec2(0.4 * sin(x_Time * 0.1), 0.3 * cos(x_Time * 0.07));
        vec2 zShift = mobiusTranslate(z, a);
        float rShift = length(zShift);

        // Hyperbolic distance from center: d_h = 2 * atanh(r)
        // For numerical safety clamp r slightly away from 1
        float rSafe = min(rShift, 0.998);
        float dH = 2.0 * atanh(rSafe);

        // Angular coordinate in shifted space
        float angShift = atan(zShift.y, zShift.x);

        // === Hyperbolic concentric rings ===
        float ringPhase = dH * RING_DENSITY - x_Time * 0.6;
        float ringFrac = fract(ringPhase);
        float ringMask = smoothstep(0.04, 0.0, abs(ringFrac - 0.5) - 0.46);
        float ringIdx = floor(ringPhase);

        // === Angular sectors — pie slices ===
        float sectorPhase = (angShift / 6.28318 + 0.5) * float(SECTORS) + x_Time * 0.08;
        float sectorFrac = fract(sectorPhase);
        float sectorMask = smoothstep(0.04, 0.0, abs(sectorFrac - 0.5) - 0.46);
        float sectorIdx = floor(sectorPhase);

        // Primary pattern: combine ring + sector into cell lattice
        float cell = ringMask + sectorMask * 0.7;
        vec3 patternCol = poi_pal(fract(ringIdx * 0.13 + sectorIdx * 0.07 + x_Time * 0.03));
        col += patternCol * cell * 1.1;

        // Cell interior color (pastel gradient keyed by ring+sector hash)
        vec2 cellId = vec2(ringIdx, sectorIdx);
        float cellHash = fract(sin(dot(cellId, vec2(127.1, 311.7))) * 43758.5);
        vec3 interiorCol = poi_pal(fract(cellHash + x_Time * 0.015));
        col += interiorCol * (1.0 - cell) * 0.15;

        // Boundary approach glow (infinite detail accumulates toward r=1)
        float boundaryGlow = exp(-(1.0 - rShift) * 18.0);
        col += poi_pal(fract(angShift * 0.5 + x_Time * 0.1)) * boundaryGlow * 0.6;

        // Center brighter (origin marker)
        float centerMark = exp(-rShift * rShift * 120.0);
        col += vec3(1.0, 0.95, 0.8) * centerMark * 0.7;
    }

    // === Disk boundary — bright ring at r = DISK_R ===
    float diskEdge = abs(length(p) - DISK_R);
    col += vec3(1.0, 0.85, 0.50) * exp(-diskEdge * diskEdge * 4000.0) * 0.85;
    // Outer faint halo
    col += vec3(0.80, 0.45, 0.25) * exp(-diskEdge * diskEdge * 60.0) * 0.15;

    // Subtle radial gradient outside disk
    if (length(p) > DISK_R) {
        float outD = length(p) - DISK_R;
        col += vec3(0.05, 0.03, 0.10) * exp(-outD * 5.0) * 0.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
