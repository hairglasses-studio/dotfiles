// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Electric sphere — dense lightning ball with surface arcs + occasional chaos bursts

const int   ARCS = 12;
const int   ARC_SEG = 8;
const float SPHERE_R = 0.3;
const float INTENSITY = 0.55;

vec3 es_pal(float t) {
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 cyan = vec3(0.20, 0.85, 0.98);
    vec3 white = vec3(0.95, 0.98, 1.00);
    vec3 mag  = vec3(0.95, 0.30, 0.65);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(vio, cyan, s);
    else if (s < 2.0) return mix(cyan, white, s - 1.0);
    else if (s < 3.0) return mix(white, mag, s - 2.0);
    else              return mix(mag, vio, s - 3.0);
}

float es_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);

    vec3 col = vec3(0.01, 0.02, 0.05);

    // Sphere shell background
    if (r < SPHERE_R) {
        float sphereHeat = 1.0 - r / SPHERE_R;
        col += es_pal(0.0) * pow(sphereHeat, 4.0) * 0.25;
    }

    // Central core
    float coreMask = exp(-r * r / (SPHERE_R * 0.15 * SPHERE_R * 0.15) * 2.0);
    col += vec3(1.0, 0.98, 0.95) * coreMask * 2.0;
    col += es_pal(0.3) * exp(-r * r * 30.0) * 0.5;

    // Surface arcs — wrapping the sphere
    for (int a = 0; a < ARCS; a++) {
        float fa = float(a);
        float startAng = fa / float(ARCS) * 6.28 + x_Time * 0.3 + es_hash(fa + floor(x_Time * 5.0)) * 0.5;
        float endAng = startAng + 1.5 + es_hash(fa * 3.7) * 1.5;
        vec2 start = vec2(cos(startAng), sin(startAng)) * SPHERE_R;
        vec2 end = vec2(cos(endAng), sin(endAng)) * SPHERE_R;

        // Arc is a jagged line along sphere surface (approximated by zigzag)
        vec2 perp = normalize(vec2(-(end - start).y, (end - start).x));
        float minD = 1e9;
        vec2 prev = start;
        for (int seg = 1; seg <= ARC_SEG; seg++) {
            float tS = float(seg) / float(ARC_SEG);
            vec2 base = mix(start, end, tS);
            // Follow sphere surface (pull toward r=SPHERE_R)
            base *= SPHERE_R / length(base);
            // Jitter perpendicular to sphere normal
            float env = sin(tS * 3.14);
            float jit = (es_hash(fa + float(seg) + floor(x_Time * 8.0)) - 0.5) * 0.04 * env;
            vec2 pt = base + perp * jit;
            vec2 pa = p - prev;
            vec2 ba = pt - prev;
            float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
            minD = min(minD, length(pa - ba * h));
            prev = pt;
        }
        float arcCore = 1.0 - smoothstep(0.003, 0.005, minD);
        float arcGlow = exp(-minD * minD * 800.0) * 0.3;
        vec3 arcCol = es_pal(fract(fa * 0.05 + x_Time * 0.1));
        col += (vec3(1.0) * arcCore + arcCol * arcGlow) * 1.2;
    }

    // Outer halo
    if (r > SPHERE_R) {
        float haloD = r - SPHERE_R;
        float halo = exp(-haloD * 4.0) * 0.35;
        col += es_pal(fract(x_Time * 0.06)) * halo;
    }

    // Occasional chaos burst — brief full-brightness flash
    float burstHash = es_hash(floor(x_Time * 2.0));
    if (burstHash > 0.9) {
        float burstPhase = fract(x_Time * 2.0);
        if (burstPhase < 0.1) {
            float burstFade = 1.0 - burstPhase / 0.1;
            col += es_pal(0.0) * exp(-r * r * 5.0) * burstFade * 0.6;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
