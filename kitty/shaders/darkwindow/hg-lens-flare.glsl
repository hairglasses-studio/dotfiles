// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Anamorphic lens flare — hexagonal ghosts, chromatic streaks, halo

const int   GHOST_COUNT  = 7;
const float STREAK_WIDTH = 0.003;
const float CHROM_SPREAD = 0.008;
const float INTENSITY    = 0.55;

vec3 lf_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);  // cyan
    vec3 b = vec3(1.00, 0.70, 0.20);  // warm gold
    vec3 c = vec3(0.90, 0.20, 0.55);  // magenta
    vec3 d = vec3(0.55, 0.30, 0.98);  // violet
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Hexagonal aperture SDF
float hexMask(vec2 p, float size) {
    p = abs(p);
    float d = max(p.x * 0.866 + p.y * 0.5, p.y);  // hex inequality
    return smoothstep(size, size * 0.85, d);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Light source — moves in big circle around screen center
    vec2 lightPos = 0.4 * vec2(cos(x_Time * 0.15), sin(x_Time * 0.15) * 0.7);

    // Direction from center to light — ghosts arrange along this axis
    vec2 axis = normalize(lightPos);

    vec3 col = vec3(0.0);

    // Light-source core
    float lightDist = length(p - lightPos);
    float core = exp(-lightDist * lightDist * 800.0) * 1.5;
    float halo = exp(-lightDist * 3.5) * 0.35;
    vec3 lightCol = lf_pal(fract(x_Time * 0.03));
    col += lightCol * (core + halo);

    // Horizontal streak — anamorphic characteristic: stretched along x-axis
    vec2 streakVec = p - lightPos;
    float streakAxis = abs(streakVec.y);
    float streakLen  = abs(streakVec.x);
    float streak = exp(-streakAxis * streakAxis / (STREAK_WIDTH * STREAK_WIDTH))
                 * exp(-streakLen * 1.2);
    // Chromatic spread on streak
    col.r += lightCol.r * exp(-(streakAxis + CHROM_SPREAD) * (streakAxis + CHROM_SPREAD) / (STREAK_WIDTH * STREAK_WIDTH)) * exp(-streakLen * 1.2) * 0.7;
    col.b += lightCol.b * exp(-(streakAxis - CHROM_SPREAD) * (streakAxis - CHROM_SPREAD) / (STREAK_WIDTH * STREAK_WIDTH)) * exp(-streakLen * 1.2) * 0.7;
    col += lightCol * streak * 0.5;

    // Hexagonal ghosts along the anti-optical axis
    for (int i = 0; i < GHOST_COUNT; i++) {
        float fi = float(i);
        float offset = -0.4 + fi * 0.18;  // spread along axis, mostly on anti-side
        vec2 ghostPos = lightPos * offset;
        vec2 gp = p - ghostPos;

        float ghostSize = 0.04 + 0.025 * abs(sin(fi * 1.7));
        // Random-ish rotation per ghost
        float rot = fi * 0.7 + x_Time * 0.1;
        float cr = cos(rot), sr = sin(rot);
        vec2 rp = mat2(cr, -sr, sr, cr) * gp;

        // Hexagonal aperture mask
        float hex = hexMask(rp, ghostSize);

        // Chromatic ghosts: slightly different positions per channel
        vec2 gpr = gp + axis * (fi * 0.005);
        vec2 gpb = gp - axis * (fi * 0.005);
        float hexR = hexMask(mat2(cr, -sr, sr, cr) * gpr, ghostSize);
        float hexB = hexMask(mat2(cr, -sr, sr, cr) * gpb, ghostSize);

        vec3 ghostCol = lf_pal(fract(fi * 0.15 + x_Time * 0.02));
        col.r += hexR * ghostCol.r * 0.25;
        col.g += hex  * ghostCol.g * 0.25;
        col.b += hexB * ghostCol.b * 0.25;
    }

    // Radial halo from light — like lens bloom
    float radialHalo = exp(-lightDist * 1.5) * 0.15;
    col += lightCol * radialHalo;

    // Aperture ring — dim ring at distance from light
    float ringR = 0.3;
    float ringDist = abs(lightDist - ringR);
    float ring = exp(-ringDist * ringDist * 900.0) * 0.4;
    col += lf_pal(fract(x_Time * 0.04 + 0.5)) * ring;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
