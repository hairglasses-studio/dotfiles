// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Rotating 3D cube array — tiled cubes raymarched with per-cube hash color

const int   STEPS = 48;
const float MAX_DIST = 6.0;
const float EPS = 0.002;
const float CELL = 0.7;
const float CUBE_R = 0.18;
const float INTENSITY = 0.55;

vec3 ca_pal(float t) {
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

float sdBox(vec3 p, vec3 b) {
    vec3 d = abs(p) - b;
    return length(max(d, 0.0)) + min(max(d.x, max(d.y, d.z)), 0.0);
}

float sdCubeArray(vec3 p, float t) {
    // Tile with rotation per cube
    vec3 cellId = floor((p + CELL * 0.5) / CELL);
    vec3 rep = mod(p + CELL * 0.5, CELL) - CELL * 0.5;
    // Rotate cube based on time + cell id
    float cellHash = fract(sin(dot(cellId, vec3(12.9898, 78.233, 37.719))) * 43758.5);
    float rotA = t * 0.5 + cellHash * 6.28;
    float cr = cos(rotA), sr = sin(rotA);
    rep.xz = mat2(cr, -sr, sr, cr) * rep.xz;
    float cr2 = cos(rotA * 0.7), sr2 = sin(rotA * 0.7);
    rep.xy = mat2(cr2, -sr2, sr2, cr2) * rep.xy;
    return sdBox(rep, vec3(CUBE_R));
}

vec3 caNormal(vec3 p, float t) {
    vec2 e = vec2(0.003, 0.0);
    return normalize(vec3(
        sdCubeArray(p + e.xyy, t) - sdCubeArray(p - e.xyy, t),
        sdCubeArray(p + e.yxy, t) - sdCubeArray(p - e.yxy, t),
        sdCubeArray(p + e.yyx, t) - sdCubeArray(p - e.yyx, t)
    ));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Camera flying through
    vec3 ro = vec3(0.15 * sin(x_Time * 0.15), 0.15 * cos(x_Time * 0.12), x_Time * 0.3);
    vec3 rd = normalize(vec3(p, 1.3));

    float dist = 0.0;
    vec3 col = vec3(0.0);
    float hit = 0.0;
    int hitStep = STEPS;
    for (int i = 0; i < STEPS; i++) {
        vec3 pos = ro + rd * dist;
        float d = sdCubeArray(pos, x_Time);
        if (d < EPS) { hit = 1.0; hitStep = i; break; }
        if (dist > MAX_DIST) break;
        dist += d * 0.9;
    }

    if (hit > 0.5) {
        vec3 pos = ro + rd * dist;
        vec3 n = caNormal(pos, x_Time);
        vec3 lightDir = normalize(vec3(0.4, 0.7, -0.3));
        float diff = max(0.0, dot(n, lightDir));
        float fres = pow(1.0 - abs(dot(n, -rd)), 2.5);

        // Per-cube color
        vec3 cellId = floor((pos + CELL * 0.5) / CELL);
        float cellHash = fract(sin(dot(cellId, vec3(12.9898, 78.233, 37.719))) * 43758.5);
        vec3 base = ca_pal(fract(cellHash + x_Time * 0.04));

        float ao = 1.0 - float(hitStep) / float(STEPS);
        col = base * (0.3 + diff * 0.6) * ao;
        col += ca_pal(fract(cellHash + 0.3)) * fres * 0.8;
        col *= exp(-dist * 0.12);
    }

    // Ambient glow
    col += ca_pal(fract(x_Time * 0.03)) * exp(-length(p) * length(p) * 2.5) * 0.06;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
