// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Stained glass — hex-cell colored panels with lead seams + backlight illumination + dust rays

const int   SEEDS = 30;
const float INTENSITY = 0.55;

vec3 sg_pal(float t) {
    vec3 a = vec3(0.90, 0.20, 0.30);
    vec3 b = vec3(0.20, 0.55, 0.95);
    vec3 c = vec3(0.35, 0.85, 0.40);
    vec3 d = vec3(0.95, 0.75, 0.25);
    vec3 e = vec3(0.60, 0.30, 0.85);
    vec3 f = vec3(0.15, 0.80, 0.85);
    float s = mod(t * 6.0, 6.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else if (s < 5.0) return mix(e, f, s - 4.0);
    else              return mix(f, a, s - 5.0);
}

float sg_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  sg_hash2(float n) { return vec2(sg_hash(n), sg_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Voronoi for panels
    float d1 = 1e9, d2 = 1e9;
    int i1 = 0;
    for (int i = 0; i < SEEDS; i++) {
        vec2 sp = sg_hash2(float(i) * 3.71) * 2.0 - 1.0;
        sp.x *= x_WindowSize.x / x_WindowSize.y;
        float d = length(p - sp);
        if (d < d1) { d2 = d1; d1 = d; i1 = i; }
        else if (d < d2) d2 = d;
    }
    float edgeDist = d2 - d1;

    // Panel color — slowly drifting
    float panelSeed = float(i1) * 0.3;
    vec3 panelCol = sg_pal(fract(panelSeed + x_Time * 0.02));

    // Backlight illumination
    vec2 lightSource = vec2(0.2, 0.4);
    float lightDist = length(p - lightSource);
    float backlight = exp(-lightDist * 0.6) * 0.7 + 0.3;

    // Modulated panel brightness + translucency flicker
    float panelBrightness = backlight * (0.5 + 0.5 * sg_hash(float(i1) * 7.3));
    vec3 col = panelCol * panelBrightness;

    // Lead seam — dark line between panels
    float leadMask = smoothstep(0.025, 0.005, edgeDist);
    col = mix(col, vec3(0.03, 0.02, 0.03), leadMask * 0.9);

    // Seam highlight (sharp bright line)
    float seamShine = smoothstep(0.005, 0.002, edgeDist);
    col += vec3(0.4, 0.35, 0.25) * seamShine * 0.5;

    // Sub-panel pattern — diamond lattice
    vec2 seed2D = sg_hash2(float(i1) * 3.71) * 2.0 - 1.0;
    seed2D.x *= x_WindowSize.x / x_WindowSize.y;
    vec2 localP = (p - seed2D) * 25.0;
    float pattern = step(0.5, fract((localP.x + localP.y) * 0.5));
    col *= 0.85 + pattern * 0.25;

    // Dust rays in the "light" from backlight
    for (int k = 0; k < 5; k++) {
        float fk = float(k);
        float rayAng = 0.3 + fk * 0.12 + x_Time * 0.03;
        vec2 rayDir = vec2(cos(rayAng), sin(rayAng));
        vec2 toP = p - lightSource;
        float along = dot(toP, rayDir);
        if (along < 0.0) continue;
        float perpD = abs(toP.x * rayDir.y - toP.y * rayDir.x);
        float rayMask = exp(-perpD * perpD * 700.0) * exp(-along * 0.9);
        col += sg_pal(fract(fk * 0.12)) * rayMask * 0.2;
    }

    // Grime/patina in dark panel corners
    float patina = sg_hash(floor(p * 50.0).x + floor(p * 50.0).y);
    col *= 0.95 + patina * 0.1;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility * 0.9);

    _wShaderOut = vec4(result, 1.0);
}
