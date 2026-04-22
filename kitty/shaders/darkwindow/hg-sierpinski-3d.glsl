// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — 3D Sierpinski tetrahedron — iterated folding, raymarched with neon shading

const int   MAX_ITER  = 10;
const int   STEPS     = 64;
const float MAX_DIST  = 6.0;
const float EPS       = 0.0015;
const float INTENSITY = 0.55;

vec3 sp_pal(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(0.90, 0.30, 0.70);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// 4 tetrahedron vertices (unit tetra)
const vec3 V0 = vec3( 1.0,  1.0,  1.0);
const vec3 V1 = vec3(-1.0, -1.0,  1.0);
const vec3 V2 = vec3(-1.0,  1.0, -1.0);
const vec3 V3 = vec3( 1.0, -1.0, -1.0);

// Iterated Sierpinski folding DE
float sdSierpinski(vec3 p) {
    float scale = 2.0;
    for (int i = 0; i < MAX_ITER; i++) {
        // Find nearest vertex and reflect
        float d0 = length(p - V0);
        float d1 = length(p - V1);
        float d2 = length(p - V2);
        float d3 = length(p - V3);
        float minD = min(min(d0, d1), min(d2, d3));
        vec3 nearest;
        if (minD == d0) nearest = V0;
        else if (minD == d1) nearest = V1;
        else if (minD == d2) nearest = V2;
        else nearest = V3;
        p = p * scale - nearest * (scale - 1.0);
    }
    return length(p) * pow(scale, -float(MAX_ITER));
}

vec3 calcNormal(vec3 p) {
    vec2 e = vec2(0.002, 0.0);
    return normalize(vec3(
        sdSierpinski(p + e.xyy) - sdSierpinski(p - e.xyy),
        sdSierpinski(p + e.yxy) - sdSierpinski(p - e.yxy),
        sdSierpinski(p + e.yyx) - sdSierpinski(p - e.yyx)
    ));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Camera orbit
    float t = x_Time * 0.2;
    vec3 ro = vec3(3.0 * cos(t), 1.2 + 0.5 * sin(t * 0.7), 3.0 * sin(t));
    vec3 target = vec3(0.0);
    vec3 forward = normalize(target - ro);
    vec3 right = normalize(cross(vec3(0.0, 1.0, 0.0), forward));
    vec3 up = cross(forward, right);
    vec3 rd = normalize(forward + right * p.x * 1.3 + up * p.y * 1.3);

    // Raymarch
    float dist = 0.0;
    vec3 col = vec3(0.0);
    float hit = 0.0;
    int hitStep = STEPS;
    for (int i = 0; i < STEPS; i++) {
        vec3 pos = ro + rd * dist;
        float d = sdSierpinski(pos);
        if (d < EPS) { hit = 1.0; hitStep = i; break; }
        if (dist > MAX_DIST) break;
        dist += d * 0.85;
    }

    if (hit > 0.5) {
        vec3 pos = ro + rd * dist;
        vec3 n = calcNormal(pos);
        vec3 lightDir = normalize(vec3(0.5, 0.8, -0.3));
        float diffuse = max(0.0, dot(n, lightDir));
        float fresnel = pow(1.0 - abs(dot(n, -rd)), 2.5);

        // Color based on hit position + normal
        float hue = fract(length(pos) * 0.3 + x_Time * 0.05);
        vec3 base = sp_pal(hue);

        // AO from step count (fewer steps = solid hit, more = grazing)
        float ao = 1.0 - float(hitStep) / float(STEPS);

        col = base * (0.3 + diffuse * 0.6) * ao;
        col += sp_pal(fract(hue + 0.3)) * fresnel * 0.7;
        col += vec3(1.0) * pow(diffuse, 32.0) * 0.4;

        // Distance fog
        col *= exp(-dist * 0.1);
    }

    // Ambient glow
    float glow = smoothstep(1.5, 0.4, length(p));
    col += sp_pal(fract(x_Time * 0.04)) * glow * 0.06;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
