// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Neural network mesh: nodes, pulsing edges, data packets

const int   NODE_COUNT  = 16;
const int   EDGE_LINKS  = 3;       // each node connects to 3 nearest neighbors
const float NODE_RADIUS = 0.01;
const float EDGE_WIDTH  = 0.0035;
const float PACKET_W    = 0.008;
const float INTENSITY   = 0.55;

vec3 nm_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.90, 0.20, 0.55); // magenta
    vec3 d = vec3(0.20, 0.95, 0.60); // mint
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float nm_hash(float n) { return fract(sin(n * 43.37) * 43758.5453); }
vec2 nm_hash2(float n) { return vec2(nm_hash(n), nm_hash(n * 1.37 + 11.0)); }

// Animated node position — slowly drifts on lissajous path
vec2 nodePos(int idx, float t) {
    float fi = float(idx);
    vec2 base = nm_hash2(fi * 3.71) * 2.0 - 1.0;   // random in [-1,1]^2
    // Aspect-corrected: keep within window
    base *= vec2(x_WindowSize.x / x_WindowSize.y, 1.0) * 0.6;
    // Slow drift
    float f1 = 0.5 + nm_hash(fi * 5.13) * 0.5;
    float f2 = 0.5 + nm_hash(fi * 7.37) * 0.5;
    base += 0.04 * vec2(sin(t * f1 + fi), cos(t * f2 + fi * 1.7));
    return base;
}

// Distance from point p to segment [a,b]
float sdSegment(vec2 p, vec2 a, vec2 b) {
    vec2 pa = p - a, ba = b - a;
    float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
    return length(pa - ba * h);
}

// Distance along segment (for packet positioning)
float projT(vec2 p, vec2 a, vec2 b) {
    vec2 pa = p - a, ba = b - a;
    return clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    // Pre-compute all nodes
    vec2 nodes[16];
    for (int i = 0; i < NODE_COUNT; i++) {
        nodes[i] = nodePos(i, x_Time);
    }

    // Edges — for each node, connect to EDGE_LINKS nearest
    // To keep it single-pass, draw edges as we find them
    for (int a = 0; a < NODE_COUNT; a++) {
        // Find EDGE_LINKS closest neighbors (simple selection, no sort — use threshold)
        // Instead, check all (a,b) pairs where b > a and distance < threshold
        for (int b = a + 1; b < NODE_COUNT; b++) {
            float dist = length(nodes[a] - nodes[b]);
            if (dist > 0.4 || dist < 0.02) continue;   // prune too-near, too-far

            // Only link if this pair is among the closest — approximate via distance threshold
            float linkStrength = smoothstep(0.4, 0.1, dist);

            // Edge line SDF
            float ed = sdSegment(p, nodes[a], nodes[b]);
            float edgeMask = smoothstep(EDGE_WIDTH, 0.0, ed);

            // Pulse traveling along edge — phase by pair seed
            float pairSeed = float(a) * 13.7 + float(b) * 29.3;
            float pulsePhase = fract(x_Time * (0.4 + nm_hash(pairSeed) * 0.6) + nm_hash(pairSeed));
            float t = projT(p, nodes[a], nodes[b]);
            float packetDist = abs(t - pulsePhase);
            float packetMask = exp(-packetDist * packetDist * 300.0) * step(ed, PACKET_W);

            vec3 edgeColor = nm_pal(fract(pairSeed * 0.01 + x_Time * 0.05));
            col += edgeColor * edgeMask * 0.5 * linkStrength;
            col += vec3(0.95, 0.95, 1.0) * packetMask * linkStrength * 2.0;
        }
    }

    // Nodes — bright, with halo
    for (int i = 0; i < NODE_COUNT; i++) {
        float nd = length(p - nodes[i]);
        float pulse = 0.7 + 0.3 * sin(x_Time * 2.0 + float(i) * 1.7);
        float core = exp(-nd * nd / (NODE_RADIUS * NODE_RADIUS) * 4.0);
        float halo = exp(-nd * nd / 0.01 * 1.5) * 0.15;
        vec3 nc = nm_pal(fract(float(i) * 0.065 + x_Time * 0.05));
        col += nc * (core * 1.2 + halo) * pulse;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
