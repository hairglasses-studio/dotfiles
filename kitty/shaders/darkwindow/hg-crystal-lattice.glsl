// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — 3D crystal lattice — raymarched repeating octahedron array with depth fog + fresnel

const int   STEPS    = 56;
const float MAX_DIST = 8.0;
const float EPS      = 0.0015;
const float SPACING  = 0.7;
const float OCT_R    = 0.13;
const float INTENSITY = 0.55;

vec3 cl_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.30, 0.70);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float sdOct(vec3 p, float s) {
    p = abs(p);
    return (p.x + p.y + p.z - s) * 0.57735;
}

float sdLattice(vec3 p) {
    // Repeat space
    vec3 rep = mod(p + SPACING * 0.5, SPACING) - SPACING * 0.5;
    return sdOct(rep, OCT_R);
}

vec3 latticeNormal(vec3 p) {
    vec2 e = vec2(0.002, 0.0);
    return normalize(vec3(
        sdLattice(p + e.xyy) - sdLattice(p - e.xyy),
        sdLattice(p + e.yxy) - sdLattice(p - e.yxy),
        sdLattice(p + e.yyx) - sdLattice(p - e.yyx)
    ));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Flying through lattice
    float t = x_Time;
    vec3 ro = vec3(0.1 * sin(t * 0.2), 0.1 * cos(t * 0.15), t * 0.4);
    vec3 rd = normalize(vec3(p, 1.3));

    // Slow roll
    float roll = t * 0.08;
    float cr = cos(roll), sr = sin(roll);
    rd.xy = mat2(cr, -sr, sr, cr) * rd.xy;

    // Raymarch
    float dist = 0.0;
    vec3 col = vec3(0.0);
    float hit = 0.0;
    int hitStep = STEPS;
    for (int i = 0; i < STEPS; i++) {
        vec3 pos = ro + rd * dist;
        float d = sdLattice(pos);
        if (d < EPS) { hit = 1.0; hitStep = i; break; }
        if (dist > MAX_DIST) break;
        dist += d * 0.85;
    }

    if (hit > 0.5) {
        vec3 pos = ro + rd * dist;
        vec3 n = latticeNormal(pos);
        vec3 lightDir = normalize(vec3(0.4, 0.7, -0.5));
        float diff = max(0.0, dot(n, lightDir));
        float fres = pow(1.0 - abs(dot(n, -rd)), 2.5);

        // Per-cell color from cell id
        vec3 cellId = floor((pos + SPACING * 0.5) / SPACING);
        float cellHash = fract(sin(dot(cellId, vec3(12.9898, 78.233, 37.719))) * 43758.5);
        vec3 base = cl_pal(fract(cellHash + x_Time * 0.04));

        float ao = 1.0 - float(hitStep) / float(STEPS);
        col = base * (0.3 + diff * 0.6) * ao;
        col += cl_pal(fract(cellHash + 0.3)) * fres * 0.8;
        // Sparkle
        col += vec3(1.0) * pow(diff, 64.0) * 0.3;
        // Depth fog
        col *= exp(-dist * 0.15);
    }

    // Ambient depth glow
    col += cl_pal(fract(x_Time * 0.03)) * exp(-length(p) * length(p) * 1.5) * 0.08;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
