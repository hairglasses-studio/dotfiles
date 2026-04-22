// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Rotating crystalline shard — raymarched SDF with fresnel rim + internal refraction

const int   STEPS        = 64;
const float MAX_DIST     = 4.0;
const float EPS          = 0.0015;
const float INTENSITY    = 0.55;

vec3 cs_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.90, 0.30, 0.70); // magenta
    vec3 d = vec3(0.96, 0.85, 0.40); // gold
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Octahedron SDF
float sdOctahedron(vec3 p, float r) {
    p = abs(p);
    float m = p.x + p.y + p.z - r;
    vec3 q;
    if (3.0 * p.x < m)      q = p.xyz;
    else if (3.0 * p.y < m) q = p.yzx;
    else if (3.0 * p.z < m) q = p.zxy;
    else return m * 0.57735;
    float k = clamp(0.5 * (q.z - q.y + r), 0.0, r);
    return length(vec3(q.x, q.y - r + k, q.z - k));
}

// Crystal scene: two intersecting octahedra + internal slice
float sdCrystal(vec3 p) {
    float d1 = sdOctahedron(p, 0.5);
    // Secondary slightly rotated
    vec3 pr = p;
    float a = 0.5;
    pr.xz = mat2(cos(a), -sin(a), sin(a), cos(a)) * pr.xz;
    float d2 = sdOctahedron(pr, 0.42);
    return min(d1, d2);
}

vec3 calcNormal(vec3 p) {
    vec2 e = vec2(0.0015, 0.0);
    return normalize(vec3(
        sdCrystal(p + e.xyy) - sdCrystal(p - e.xyy),
        sdCrystal(p + e.yxy) - sdCrystal(p - e.yxy),
        sdCrystal(p + e.yyx) - sdCrystal(p - e.yyx)
    ));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Rotating camera around origin
    float t = x_Time * 0.3;
    vec3 ro = vec3(0.0, 0.0, -1.8);
    vec3 rd = normalize(vec3(p, 1.4));

    // Rotate the scene around y-axis
    float cr = cos(t), sr = sin(t);
    mat3 rotY = mat3(cr, 0.0, -sr, 0.0, 1.0, 0.0, sr, 0.0, cr);
    // Rotate around x too
    float cp = cos(t * 0.7), sp = sin(t * 0.7);
    mat3 rotX = mat3(1.0, 0.0, 0.0, 0.0, cp, -sp, 0.0, sp, cp);
    mat3 rot = rotX * rotY;

    // Raymarch to crystal surface
    float dist = 0.0;
    vec3 col = vec3(0.0);
    vec3 hitPos = ro;
    float hit = 0.0;
    for (int i = 0; i < STEPS; i++) {
        vec3 pos = rot * (ro + rd * dist);
        float d = sdCrystal(pos);
        if (d < EPS) { hit = 1.0; hitPos = pos; break; }
        if (dist > MAX_DIST) break;
        dist += d * 0.9;
    }

    if (hit > 0.5) {
        vec3 n = calcNormal(hitPos);
        vec3 lightDir = normalize(vec3(0.5, 0.7, -0.8));
        float diffuse = max(0.0, dot(n, lightDir));

        // Fresnel rim
        vec3 viewDir = rot * rd;
        float fresnel = pow(1.0 - abs(dot(n, -viewDir)), 3.0);

        // Face-colored — hue depends on normal direction
        float hue = fract((n.x + n.y + n.z) * 0.3 + x_Time * 0.05);
        vec3 faceCol = cs_pal(hue);

        // Bright internal highlight — fake refraction via secondary sample offset
        vec3 refractedCol = cs_pal(fract(hue + 0.3));
        float sparkle = pow(max(0.0, dot(reflect(-lightDir, n), -viewDir)), 32.0);

        col = faceCol * (0.25 + diffuse * 0.6);
        col += refractedCol * fresnel * 0.9;
        col += vec3(1.0, 0.95, 0.9) * sparkle;

        // Depth fade
        col *= exp(-dist * 0.08);
    }

    // Atmospheric glow around crystal
    float glow = 1.0 - smoothstep(0.0, 0.6, length(p));
    col += cs_pal(fract(x_Time * 0.04)) * glow * 0.08;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
