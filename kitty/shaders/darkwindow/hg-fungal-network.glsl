// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Fungal network — bioluminescent mycelium: 16 nutrient nodes connected by a sparse hypha graph, nutrient-transport pulses flowing along edges, ambient forest-floor FBM with dark leaf litter

const int   NODES = 16;
const int   PULSES = 10;
const int   FBM_OCT = 4;
const float INTENSITY = 0.55;

vec3 fn_pal(float t) {
    vec3 deep   = vec3(0.03, 0.06, 0.08);
    vec3 teal   = vec3(0.15, 0.65, 0.55);
    vec3 mint   = vec3(0.35, 0.95, 0.65);
    vec3 amber  = vec3(1.00, 0.80, 0.35);
    vec3 red    = vec3(0.95, 0.35, 0.20);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(deep, teal, s);
    else if (s < 2.0) return mix(teal, mint, s - 1.0);
    else if (s < 3.0) return mix(mint, amber, s - 2.0);
    else              return mix(amber, red, s - 3.0);
}

float fn_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float fn_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float fn_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(fn_hash2(i), fn_hash2(i + vec2(1, 0)), u.x),
               mix(fn_hash2(i + vec2(0, 1)), fn_hash2(i + vec2(1, 1)), u.x), u.y);
}

float fn_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < FBM_OCT; i++) {
        v += a * fn_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

vec2 nodePos(int i, float t) {
    float fi = float(i);
    float seed = fi * 7.31;
    vec2 base = vec2(fn_hash(seed) * 2.0 - 1.0, fn_hash(seed * 3.7) * 1.6 - 0.8);
    // Very slow breathing motion
    vec2 breathe = vec2(0.015 * sin(t * 0.3 + seed), 0.012 * cos(t * 0.25 + seed * 1.3));
    return base + breathe;
}

// Fixed sparse graph: each node connects to i+1, i+3, i+5 mod N
int neighbor(int i, int k) {
    int off = (k == 0) ? 1 : (k == 1 ? 3 : 5);
    return (i + off) % NODES;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // === Forest-floor / leaf-litter FBM background ===
    float litter = fn_fbm(p * 3.5);
    vec3 floorCol = mix(vec3(0.03, 0.04, 0.03), vec3(0.12, 0.08, 0.05), litter);
    // Darker flecks
    float flecks = smoothstep(0.8, 1.0, fn_fbm(p * 14.0 + 3.0));
    floorCol *= 1.0 - flecks * 0.3;
    vec3 col = floorCol;

    // === Hyphae — draw min-distance segments along the sparse graph ===
    float minD = 1e9;
    int closestEdgeA = 0, closestEdgeB = 0;
    float closestAlong = 0.0;
    for (int i = 0; i < NODES; i++) {
        vec2 a = nodePos(i, x_Time);
        for (int k = 0; k < 3; k++) {
            int j = neighbor(i, k);
            vec2 b = nodePos(j, x_Time);
            float segLen = length(b - a);
            if (segLen > 0.75) continue;  // only short connections
            vec2 ab = b - a;
            vec2 pa = p - a;
            float lenSq = dot(ab, ab);
            float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
            float d = length(pa - ab * h);
            if (d < minD) {
                minD = d;
                closestEdgeA = i;
                closestEdgeB = j;
                closestAlong = h;
            }
        }
    }

    // Render hypha as thin line
    float hyphaThick = 0.002;
    float hyphaMask = exp(-minD * minD / (hyphaThick * hyphaThick) * 1.5);
    vec3 hyphaCol = fn_pal(fract(float(closestEdgeA) * 0.11 + float(closestEdgeB) * 0.07));
    col += hyphaCol * hyphaMask * 0.85;
    // Soft halo
    col += hyphaCol * exp(-minD * minD * 1500.0) * 0.15;

    // === Nutrient pulses flowing along hyphae ===
    for (int i = 0; i < PULSES; i++) {
        float fi = float(i);
        float seed = fi * 11.1;
        // Pick an edge
        int eA = int(fn_hash(seed) * float(NODES));
        int kOff = int(fn_hash(seed * 3.1) * 3.0);
        int eB = neighbor(eA, kOff);
        vec2 a = nodePos(eA, x_Time);
        vec2 b = nodePos(eB, x_Time);
        float edgeLen = length(b - a);
        if (edgeLen > 0.75 || edgeLen < 1e-4) continue;
        // Pulse position along edge [0,1]
        float speed = 0.2 + fn_hash(seed * 5.1) * 0.4;
        float phase = fract((x_Time + fn_hash(seed * 7.3) * 3.0) * speed);
        vec2 pulsePos = mix(a, b, phase);
        float pd = length(p - pulsePos);
        float pulseCore = exp(-pd * pd * 25000.0);
        col += fn_pal(0.65) * pulseCore * 1.1;
        // Halo
        col += fn_pal(0.5) * exp(-pd * pd * 2000.0) * 0.2;
    }

    // === Nodes (bright bioluminescent spots) ===
    for (int i = 0; i < NODES; i++) {
        vec2 n = nodePos(i, x_Time);
        float nd = length(p - n);
        float fi = float(i);
        float seed = fi * 7.31;
        float nodeR = 0.005 + fn_hash(seed * 11.0) * 0.006;
        float core = exp(-nd * nd / (nodeR * nodeR) * 1.2);
        // Breathe pulse per node
        float breathe = 0.75 + 0.25 * sin(x_Time * (0.8 + fn_hash(seed * 13.0) * 0.7) + seed);
        vec3 nodeCol = fn_pal(fract(0.55 + fn_hash(seed * 17.0) * 0.3));
        col += nodeCol * core * breathe * 1.5;
        // Outer halo
        col += nodeCol * exp(-nd * nd * 400.0) * breathe * 0.18;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
