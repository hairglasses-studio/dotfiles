// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Lightning network — 20 nodes connected by branching lightning bolts + firing cascades

const int   NODES = 20;
const float INTENSITY = 0.55;

vec3 ln_pal(float t) {
    vec3 white = vec3(0.90, 0.92, 1.00);
    vec3 cyan  = vec3(0.20, 0.85, 0.98);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 mag   = vec3(0.95, 0.30, 0.65);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(white, cyan, s);
    else if (s < 2.0) return mix(cyan, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else              return mix(mag, white, s - 3.0);
}

float ln_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  ln_hash2(float n) { return vec2(ln_hash(n), ln_hash(n * 1.37 + 11.0)); }

// Jagged bolt distance
float sdJaggedBolt(vec2 p, vec2 a, vec2 b, float seed) {
    vec2 dir = normalize(b - a);
    vec2 perp = vec2(-dir.y, dir.x);
    float len = length(b - a);
    float minD = 1e9;
    vec2 prev = a;
    for (int i = 1; i <= 6; i++) {
        float t = float(i) / 6.0;
        vec2 base = mix(a, b, t);
        float env = sin(t * 3.14);
        float jit = (ln_hash(seed + float(i)) - 0.5) * 0.03 * env;
        vec2 pt = base + perp * jit;
        vec2 pa = p - prev;
        vec2 ba = pt - prev;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        minD = min(minD, length(pa - ba * h));
        prev = pt;
    }
    return minD;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.02, 0.02, 0.05);

    // Node positions
    vec2 nodes[20];
    for (int i = 0; i < NODES; i++) {
        float fi = float(i);
        vec2 pos = ln_hash2(fi * 3.71) * 1.6 - 0.8;
        pos.x *= x_WindowSize.x / x_WindowSize.y;
        pos += 0.03 * vec2(sin(x_Time * 0.3 + fi), cos(x_Time * 0.25 + fi * 1.3));
        nodes[i] = pos;
    }

    // Connections — for each node, link to nearest neighbor
    for (int a = 0; a < NODES; a++) {
        // Find nearest neighbor
        float minDist = 1e9;
        int b = 0;
        for (int j = 0; j < NODES; j++) {
            if (j == a) continue;
            float d = distance(nodes[a], nodes[j]);
            if (d < minDist) { minDist = d; b = j; }
        }
        if (minDist > 0.3) continue;

        // Only fire periodically
        float fireSeed = float(a) * 17.0 + float(b) * 23.0;
        float firePhase = fract(x_Time * 0.8 + ln_hash(fireSeed));
        if (firePhase > 0.25) continue;
        float fireFade = 1.0 - firePhase / 0.25;

        float boltD = sdJaggedBolt(p, nodes[a], nodes[b], fireSeed + floor(x_Time * 10.0));
        float core = 1.0 - smoothstep(0.002, 0.004, boltD);
        float glow = exp(-boltD * boltD * 1000.0) * 0.3;
        vec3 boltCol = ln_pal(fract(fireSeed * 0.02 + x_Time * 0.05));
        col += (vec3(1.0) * core * 1.4 + boltCol * glow) * fireFade;
    }

    // Nodes — bright dots
    for (int i = 0; i < NODES; i++) {
        float fi = float(i);
        float pulseFreq = 0.8 + ln_hash(fi * 3.7) * 1.0;
        float pulse = pow(fract(x_Time * pulseFreq + ln_hash(fi)), 0.3);
        float pulseAmp = 1.0 - pulse;
        float nd = length(p - nodes[i]);
        float core = exp(-nd * nd * 15000.0);
        float halo = exp(-nd * nd * 800.0) * 0.3;
        vec3 nc = ln_pal(fract(fi * 0.08 + x_Time * 0.04));
        col += nc * (core * (0.8 + pulseAmp * 1.5) + halo);
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
