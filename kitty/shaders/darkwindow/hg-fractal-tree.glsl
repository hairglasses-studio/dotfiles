// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Fractal tree — recursive L-system branching with swaying + glowing leaves

const int   BRANCH_DEPTH = 6;
const float INTENSITY = 0.55;

vec3 ft_pal(float t) {
    vec3 leaf_cyan = vec3(0.20, 0.90, 0.95);
    vec3 branch_vio = vec3(0.55, 0.30, 0.98);
    vec3 leaf_mag   = vec3(0.90, 0.30, 0.70);
    vec3 leaf_gold  = vec3(0.96, 0.80, 0.35);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(leaf_cyan, branch_vio, s);
    else if (s < 2.0) return mix(branch_vio, leaf_mag, s - 1.0);
    else if (s < 3.0) return mix(leaf_mag, leaf_gold, s - 2.0);
    else              return mix(leaf_gold, leaf_cyan, s - 3.0);
}

float ft_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// SDF of a capsule (line segment) with radius
float sdCapsule(vec2 p, vec2 a, vec2 b, float r) {
    vec2 pa = p - a, ba = b - a;
    float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
    return length(pa - ba * h) - r;
}

// Recursive branch evaluation. Since GLSL has no true recursion, we manually unroll.
// Returns minimum distance to any branch in tree
float treeDist(vec2 p, float t, out float leafGlow) {
    float minD = 1e9;
    leafGlow = 0.0;

    // Sway factor
    float sway = sin(t * 0.5) * 0.15;

    // Level 0 — trunk
    vec2 trunkStart = vec2(0.0, -0.5);
    vec2 trunkEnd = vec2(sway * 0.05, -0.2);
    minD = min(minD, sdCapsule(p, trunkStart, trunkEnd, 0.012));

    // Level 1 — main branches
    for (int b1 = 0; b1 < 3; b1++) {
        float fb1 = float(b1);
        float ang1 = (fb1 - 1.0) * 0.5;
        vec2 dir1 = vec2(sin(ang1 + sway * 0.5), cos(ang1 + sway * 0.5));
        vec2 b1Start = trunkEnd;
        vec2 b1End = b1Start + dir1 * 0.2;
        minD = min(minD, sdCapsule(p, b1Start, b1End, 0.008));

        // Level 2 — sub-branches
        for (int b2 = 0; b2 < 3; b2++) {
            float fb2 = float(b2);
            float ang2 = ang1 + (fb2 - 1.0) * 0.6 + sway * 0.3;
            vec2 dir2 = vec2(sin(ang2), cos(ang2));
            vec2 b2Start = b1End;
            vec2 b2End = b2Start + dir2 * 0.12;
            minD = min(minD, sdCapsule(p, b2Start, b2End, 0.005));

            // Level 3 — twigs
            for (int b3 = 0; b3 < 2; b3++) {
                float fb3 = float(b3);
                float ang3 = ang2 + (fb3 - 0.5) * 0.8 + sway * 0.5;
                vec2 dir3 = vec2(sin(ang3), cos(ang3));
                vec2 b3Start = b2End;
                vec2 b3End = b3Start + dir3 * 0.06;
                minD = min(minD, sdCapsule(p, b3Start, b3End, 0.003));

                // Leaves at twig tips
                float leafSeed = fb1 * 19.3 + fb2 * 7.3 + fb3 * 3.1;
                float leafPulse = 0.5 + 0.5 * sin(t * 2.0 + leafSeed);
                float leafD = length(p - b3End);
                leafGlow += exp(-leafD * leafD * 4000.0) * (0.4 + leafPulse * 0.5);
            }
        }
    }

    return minD;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.03);

    float leafGlow = 0.0;
    float d = treeDist(p, x_Time, leafGlow);

    // Branch body
    float core = smoothstep(0.002, 0.0, d);
    float glow = exp(-d * d * 800.0) * 0.3;
    vec3 branchCol = ft_pal(0.25);
    col += branchCol * (core * 1.0 + glow);

    // Leaves (bright glows at twig tips)
    col += ft_pal(fract(x_Time * 0.05)) * leafGlow * 0.5;

    // Ground line
    float groundMask = smoothstep(0.002, 0.0, abs(p.y + 0.5));
    col += branchCol * groundMask * 0.4;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
