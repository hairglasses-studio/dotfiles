// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Rotating ice crystal — hexagonal refractive prism with rainbow dispersion

const int   STEPS = 48;
const float MAX_DIST = 5.0;
const float EPS = 0.002;
const float INTENSITY = 0.55;

vec3 ic_pal(float t) {
    vec3 a = vec3(0.15, 0.85, 1.00);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.95, 0.30, 0.75);
    vec3 d = vec3(0.96, 0.80, 0.35);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Hexagonal prism SDF: hex cross-section, bounded in z
float sdHexPrism(vec3 p, vec2 h) {
    const vec3 k = vec3(-0.8660254, 0.5, 0.57735);
    p = abs(p);
    p.xy -= 2.0 * min(dot(k.xy, p.xy), 0.0) * k.xy;
    vec2 d = vec2(length(p.xy - vec2(clamp(p.x, -k.z * h.x, k.z * h.x), h.x)) * sign(p.y - h.x),
                  p.z - h.y);
    return min(max(d.x, d.y), 0.0) + length(max(d, 0.0));
}

vec3 calcNormal(vec3 p, vec2 h) {
    vec2 e = vec2(0.003, 0.0);
    return normalize(vec3(
        sdHexPrism(p + e.xyy, h) - sdHexPrism(p - e.xyy, h),
        sdHexPrism(p + e.yxy, h) - sdHexPrism(p - e.yxy, h),
        sdHexPrism(p + e.yyx, h) - sdHexPrism(p - e.yyx, h)
    ));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Camera orbit around crystal
    float t = x_Time * 0.3;
    vec3 ro = vec3(2.0 * cos(t), 0.8 * sin(t * 0.6), 2.0 * sin(t));
    vec3 target = vec3(0.0);
    vec3 forward = normalize(target - ro);
    vec3 right = normalize(cross(vec3(0.0, 1.0, 0.0), forward));
    vec3 up = cross(forward, right);
    vec3 rd = normalize(forward + right * p.x * 1.3 + up * p.y * 1.3);

    // Slow crystal rotation independent of camera
    float cRot = x_Time * 0.2;
    float cr = cos(cRot), sr = sin(cRot);
    mat3 crystalRot = mat3(cr, 0.0, -sr, 0.0, 1.0, 0.0, sr, 0.0, cr);

    vec2 h = vec2(0.3, 0.6);  // hex radius, length

    float dist = 0.0;
    vec3 col = vec3(0.0);
    float hit = 0.0;
    vec3 hitPos;
    for (int i = 0; i < STEPS; i++) {
        vec3 worldP = ro + rd * dist;
        vec3 localP = crystalRot * worldP;
        float d = sdHexPrism(localP, h);
        if (d < EPS) { hit = 1.0; hitPos = localP; break; }
        if (dist > MAX_DIST) break;
        dist += d * 0.9;
    }

    if (hit > 0.5) {
        vec3 n = calcNormal(hitPos, h);
        vec3 worldN = normalize(transpose(crystalRot) * n);

        vec3 lightDir = normalize(vec3(0.4, 0.7, -0.3));
        float diff = max(0.0, dot(worldN, lightDir));
        float fres = pow(1.0 - abs(dot(worldN, -rd)), 3.0);

        // Refractive dispersion — sample palette at 3 offset hues per channel
        float hue = fract(dot(worldN, vec3(1.0)) * 0.3 + x_Time * 0.05);
        vec3 rCol = ic_pal(fract(hue));
        vec3 gCol = ic_pal(fract(hue + 0.1));
        vec3 bCol = ic_pal(fract(hue + 0.2));

        col = rCol * 0.5 + gCol * 0.3 + bCol * 0.2;
        col *= 0.3 + diff * 0.7;
        // Bright fresnel rim — rainbow tint
        col += ic_pal(fract(hue + 0.5)) * fres * 0.9;
        // Inner sparkle — specular highlight
        col += vec3(1.0) * pow(diff, 48.0) * 0.6;

        col *= exp(-dist * 0.06);
    }

    // Soft ambient glow
    float r = length(p);
    col += ic_pal(fract(x_Time * 0.04)) * exp(-r * r * 5.0) * 0.08;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
