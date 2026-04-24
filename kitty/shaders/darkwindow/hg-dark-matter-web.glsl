// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Dark matter web — cosmic filaments connecting galaxy-cluster nodes, drifting slowly in time, with diffuse FBM density field and dark voids

const int   NODES = 18;
const int   BG_STARS = 80;
const int   FBM_OCT = 4;
const float INTENSITY = 0.55;

vec3 dmw_pal(float t) {
    vec3 deep    = vec3(0.02, 0.02, 0.10);
    vec3 indigo  = vec3(0.10, 0.08, 0.35);
    vec3 violet  = vec3(0.45, 0.20, 0.80);
    vec3 magenta = vec3(0.90, 0.30, 0.70);
    vec3 gold    = vec3(1.00, 0.80, 0.45);
    vec3 cream   = vec3(1.00, 0.95, 0.85);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(deep, indigo, s);
    else if (s < 2.0) return mix(indigo, violet, s - 1.0);
    else if (s < 3.0) return mix(violet, magenta, s - 2.0);
    else if (s < 4.0) return mix(magenta, gold, s - 3.0);
    else              return mix(gold, cream, s - 4.0);
}

float dmw_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float dmw_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float dmw_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(dmw_hash2(i), dmw_hash2(i + vec2(1, 0)), u.x),
               mix(dmw_hash2(i + vec2(0, 1)), dmw_hash2(i + vec2(1, 1)), u.x), u.y);
}

float dmw_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < FBM_OCT; i++) {
        v += a * dmw_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

// Node positions: deterministic + slow drift
vec2 nodePos(int i, float t) {
    float fi = float(i);
    float seed = fi * 7.31;
    vec2 base = vec2(dmw_hash(seed) * 2.0 - 1.0,
                     dmw_hash(seed * 3.7) * 1.6 - 0.8);
    // Slow orbital drift
    float driftAng = t * (0.04 + dmw_hash(seed * 5.1) * 0.03) + seed;
    float driftR = 0.04 + dmw_hash(seed * 7.1) * 0.03;
    return base + vec2(cos(driftAng), sin(driftAng)) * driftR;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.002, 0.005, 0.015);

    // === Diffuse large-scale density field (cosmic web background) ===
    vec2 fbmP = p * 2.3 + vec2(x_Time * 0.015, x_Time * 0.02);
    float density = dmw_fbm(fbmP);
    // Boost contrast to get web-like filament ridges
    density = pow(density, 1.6);
    // Ridged component for web-like filaments
    float ridged = 1.0 - abs(dmw_fbm(fbmP * 1.5) * 2.0 - 1.0);
    ridged = pow(ridged, 3.0);
    col += dmw_pal(0.15 + density * 0.3) * density * 0.35;
    col += dmw_pal(0.45) * ridged * 0.25;

    // === Filaments between nodes — only draw pairs below distance threshold ===
    float FIL_MAX = 0.55;     // only connect nearby pairs
    float FIL_THICK = 0.0055;
    for (int i = 0; i < NODES; i++) {
        vec2 a = nodePos(i, x_Time);
        // Connect to node i+1, i+3, i+7 mod N (fixed sparse graph)
        for (int k = 0; k < 3; k++) {
            int jOff = (k == 0) ? 1 : (k == 1 ? 3 : 7);
            int j = (i + jOff) % NODES;
            vec2 b = nodePos(j, x_Time);
            float segLen = length(b - a);
            if (segLen > FIL_MAX) continue;

            vec2 pa = p - a;
            vec2 ba = b - a;
            float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
            float d = length(pa - ba * h);
            // Thickness slightly increases near nodes
            float widthMod = 0.7 + 0.3 * (1.0 - abs(h - 0.5) * 2.0);
            float filMask = exp(-d * d / (FIL_THICK * FIL_THICK * widthMod * widthMod) * 1.8);
            // Intensity depends on distance (farther = dimmer)
            float distFade = 1.0 - segLen / FIL_MAX;
            // Filament color varies along its length
            vec3 fcol = dmw_pal(fract(float(i) * 0.07 + h * 0.3 + x_Time * 0.03));
            col += fcol * filMask * distFade * 0.85;

            // Faint halo around filaments
            float halo = exp(-d * d * 1200.0) * 0.18;
            col += dmw_pal(0.35) * halo * distFade;
        }
    }

    // === Nodes (galaxy clusters) — bright cores + halos ===
    for (int i = 0; i < NODES; i++) {
        vec2 n = nodePos(i, x_Time);
        float nd = length(p - n);
        float fi = float(i);
        float seed = fi * 7.31;
        float nodeSize = 0.008 + dmw_hash(seed * 11.0) * 0.012;
        float core = exp(-nd * nd / (nodeSize * nodeSize) * 1.2);
        float halo = exp(-nd * nd * 250.0);
        // Pulse rate differs per node
        float pulse = 0.8 + 0.2 * sin(x_Time * (1.0 + dmw_hash(seed * 13.0) * 1.5) + seed);
        vec3 nodeCol = dmw_pal(fract(0.7 + dmw_hash(seed * 17.0) * 0.3));
        col += nodeCol * core * pulse * 1.4;
        col += nodeCol * halo * 0.25;
    }

    // === Sparse background "field galaxies" (tiny stars) ===
    for (int i = 0; i < BG_STARS; i++) {
        float fi = float(i);
        float seed = fi * 19.1;
        vec2 sp = vec2(dmw_hash(seed) * 2.0 - 1.0, dmw_hash(seed * 3.7) * 1.6 - 0.8);
        float sd = length(p - sp);
        float mag = 0.3 + dmw_hash(seed * 5.1) * 0.4;
        col += vec3(0.85, 0.90, 1.0) * exp(-sd * sd * 40000.0) * mag * 0.35;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
