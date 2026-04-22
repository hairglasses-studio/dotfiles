// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Bismuth crystal — hopper crystal SDF with rainbow iridescent oxide coating

const int   STEPS = 56;
const float MAX_DIST = 5.0;
const float EPS = 0.002;
const float INTENSITY = 0.55;

vec3 bc_pal(float t) {
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 blue = vec3(0.20, 0.50, 0.98);
    vec3 cyan = vec3(0.20, 0.85, 0.95);
    vec3 green = vec3(0.30, 0.95, 0.55);
    vec3 gold  = vec3(0.96, 0.85, 0.30);
    vec3 mag   = vec3(0.95, 0.30, 0.65);
    float s = mod(t * 6.0, 6.0);
    if (s < 1.0)      return mix(vio, blue, s);
    else if (s < 2.0) return mix(blue, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, green, s - 2.0);
    else if (s < 4.0) return mix(green, gold, s - 3.0);
    else if (s < 5.0) return mix(gold, mag, s - 4.0);
    else              return mix(mag, vio, s - 5.0);
}

// Stepped hopper crystal — series of nested boxes
float sdHopper(vec3 p) {
    float d = 1e9;
    float size = 0.35;
    for (int i = 0; i < 5; i++) {
        float fi = float(i);
        // Inside box at decreasing size, offset slightly
        vec3 boxSize = vec3(size - fi * 0.05, size - fi * 0.05, size - fi * 0.05);
        vec3 q = abs(p) - boxSize;
        float box = length(max(q, 0.0)) + min(max(q.x, max(q.y, q.z)), 0.0);
        // Subtract from main shape to create stepped hopper
        if (i == 0) d = box;
        else d = max(d, -box + 0.03);
    }
    return d;
}

vec3 bcNormal(vec3 p) {
    vec2 e = vec2(0.003, 0.0);
    return normalize(vec3(
        sdHopper(p + e.xyy) - sdHopper(p - e.xyy),
        sdHopper(p + e.yxy) - sdHopper(p - e.yxy),
        sdHopper(p + e.yyx) - sdHopper(p - e.yyx)
    ));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Camera orbit
    float t = x_Time * 0.2;
    vec3 ro = vec3(1.8 * cos(t), 0.4 + 0.3 * sin(t * 0.6), 1.8 * sin(t));
    vec3 target = vec3(0.0);
    vec3 forward = normalize(target - ro);
    vec3 right = normalize(cross(vec3(0.0, 1.0, 0.0), forward));
    vec3 up = cross(forward, right);
    vec3 rd = normalize(forward + right * p.x * 1.3 + up * p.y * 1.3);

    float dist = 0.0;
    vec3 col = vec3(0.0);
    float hit = 0.0;
    for (int i = 0; i < STEPS; i++) {
        vec3 pos = ro + rd * dist;
        float d = sdHopper(pos);
        if (d < EPS) { hit = 1.0; break; }
        if (dist > MAX_DIST) break;
        dist += max(d * 0.8, 0.02);
    }

    if (hit > 0.5) {
        vec3 pos = ro + rd * dist;
        vec3 n = bcNormal(pos);
        vec3 lightDir = normalize(vec3(0.5, 0.7, -0.3));
        float diff = max(0.0, dot(n, lightDir));
        // Iridescent color from normal direction (thin-film interference)
        float irid = dot(n, -rd) * 0.5 + 0.5;
        irid = fract(irid * 3.0 + x_Time * 0.05 + pos.x * 0.3 + pos.y * 0.3);
        vec3 iridCol = bc_pal(irid);

        // Specular
        float spec = pow(diff, 64.0);

        col = iridCol * (0.3 + diff * 0.6);
        col += vec3(1.0) * spec * 0.8;
        col *= exp(-dist * 0.08);
    }

    // Ambient glow
    float r = length(p);
    col += bc_pal(fract(x_Time * 0.04)) * exp(-r * r * 3.0) * 0.07;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
