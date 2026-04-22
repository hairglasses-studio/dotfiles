// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Clifford torus — 4D torus projected to 3D, raymarched with neon shading

const int   STEPS    = 48;
const float MAX_DIST = 6.0;
const float EPS      = 0.0015;
const float TORUS_R1 = 1.0;
const float TORUS_R2 = 0.4;
const float INTENSITY = 0.55;

vec3 ct_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.30, 0.70);
    vec3 d = vec3(0.96, 0.70, 0.25);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Clifford torus: parameterize as (cos(u)cos(v), cos(u)sin(v), sin(u)cos(w), sin(u)sin(w))
// Project 4D → 3D by stereographic projection from north pole
vec3 cliffordProject(float u, float v, float w) {
    // 4D point on Clifford torus at unit sphere
    vec4 p4 = vec4(cos(u) * cos(v), cos(u) * sin(v), sin(u) * cos(w), sin(u) * sin(w));
    // Stereographic projection from (1,0,0,0)
    float denom = 1.0 - p4.x;
    if (abs(denom) < 0.01) denom = sign(denom) * 0.01;
    return vec3(p4.y, p4.z, p4.w) / denom;
}

// Approximate SDF to Clifford torus surface via tessellated sample
float sdClifford(vec3 p, float t) {
    // Sample U,V,W grid and find nearest projected point
    float minD = 1e9;
    for (int i = 0; i < 12; i++) {
        float u = float(i) / 12.0 * 6.28318 + t * 0.2;
        for (int j = 0; j < 12; j++) {
            float v = float(j) / 12.0 * 6.28318 + t * 0.3;
            // Fixed w offset varies
            float w = t * 0.5 + float(i + j) * 0.1;
            vec3 pp = cliffordProject(u, v, w);
            float d = length(p - pp) - 0.08;
            minD = min(minD, d);
        }
    }
    return minD;
}

vec3 ctNormal(vec3 p, float t) {
    vec2 e = vec2(0.002, 0.0);
    return normalize(vec3(
        sdClifford(p + e.xyy, t) - sdClifford(p - e.xyy, t),
        sdClifford(p + e.yxy, t) - sdClifford(p - e.yxy, t),
        sdClifford(p + e.yyx, t) - sdClifford(p - e.yyx, t)
    ));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Camera
    float t = x_Time * 0.2;
    vec3 ro = vec3(2.5 * cos(t), 1.0 * sin(t * 0.7), 2.5 * sin(t));
    vec3 target = vec3(0.0);
    vec3 forward = normalize(target - ro);
    vec3 right = normalize(cross(vec3(0.0, 1.0, 0.0), forward));
    vec3 up = cross(forward, right);
    vec3 rd = normalize(forward + right * p.x * 1.2 + up * p.y * 1.2);

    // Raymarch
    float dist = 0.0;
    vec3 col = vec3(0.0);
    float hit = 0.0;
    for (int i = 0; i < STEPS; i++) {
        vec3 pos = ro + rd * dist;
        float d = sdClifford(pos, x_Time);
        if (d < EPS) { hit = 1.0; break; }
        if (dist > MAX_DIST) break;
        dist += max(d * 0.7, 0.02);
    }

    if (hit > 0.5) {
        vec3 pos = ro + rd * dist;
        vec3 n = ctNormal(pos, x_Time);
        vec3 lightDir = normalize(vec3(0.4, 0.7, -0.3));
        float diffuse = max(0.0, dot(n, lightDir));
        float fresnel = pow(1.0 - abs(dot(n, -rd)), 2.5);

        float hue = fract(length(pos) * 0.3 + x_Time * 0.05);
        vec3 base = ct_pal(hue);
        col = base * (0.35 + diffuse * 0.55);
        col += ct_pal(fract(hue + 0.3)) * fresnel * 0.8;
        col += vec3(1.0) * pow(diffuse, 32.0) * 0.4;
        col *= exp(-dist * 0.08);
    }

    // Outer glow
    float r = length(p);
    col += ct_pal(fract(x_Time * 0.04)) * exp(-r * r * 2.0) * 0.08;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
