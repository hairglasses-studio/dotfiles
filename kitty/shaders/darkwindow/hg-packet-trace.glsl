// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Packet trace — 12 router nodes connected by a sparse edge graph, 10 packets in flight each following a deterministic 3-hop random-walk path, rendered with head + fading trail, node blink on hop arrival

const int   NODES = 12;
const int   PACKETS = 10;
const float PATH_SECONDS = 4.5;   // total time for a 3-hop path
const float INTENSITY = 0.55;

vec3 pkt_pal(float t) {
    vec3 cyan   = vec3(0.25, 0.90, 1.00);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 amber  = vec3(1.00, 0.70, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, amber, s - 2.0);
    else if (s < 4.0) return mix(amber, mint, s - 3.0);
    else              return mix(mint, cyan, s - 4.0);
}

float pkt_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Static node positions
vec2 nodePos(int i) {
    float fi = float(i);
    float seed = fi * 7.31;
    return vec2(pkt_hash(seed) * 1.8 - 0.9, pkt_hash(seed * 3.7) * 1.4 - 0.7);
}

// Sparse fixed graph: each node connects to i+1, i+3, i+5 mod N
int neighbor(int i, int k) {
    int off = (k == 0) ? 1 : (k == 1 ? 3 : 5);
    return (i + off) % NODES;
}

// Get the sequence of 4 nodes for packet i's 3-hop path
void packetPath(int pIdx, float t, out int n0, out int n1, out int n2, out int n3) {
    float fi = float(pIdx);
    float seed = fi * 11.1;
    // Path changes every PATH_SECONDS
    float pathSeed = floor((t + pkt_hash(seed * 3.1) * PATH_SECONDS) / PATH_SECONDS) + seed;
    n0 = int(pkt_hash(pathSeed) * float(NODES)) % NODES;
    int kA = int(pkt_hash(pathSeed * 1.7) * 3.0) % 3;
    n1 = neighbor(n0, kA);
    int kB = int(pkt_hash(pathSeed * 2.3) * 3.0) % 3;
    n2 = neighbor(n1, kB);
    int kC = int(pkt_hash(pathSeed * 3.5) * 3.0) % 3;
    n3 = neighbor(n2, kC);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.004, 0.008, 0.022);

    // Faint radial background
    float rdist = length(p);
    col += vec3(0.02, 0.03, 0.06) * (1.0 - smoothstep(0.0, 1.2, rdist)) * 0.4;

    // === Faint edges of the full graph (always shown dimly) ===
    float minEdgeD = 1e9;
    for (int i = 0; i < NODES; i++) {
        vec2 a = nodePos(i);
        for (int k = 0; k < 3; k++) {
            int j = neighbor(i, k);
            vec2 b = nodePos(j);
            vec2 ab = b - a;
            vec2 pa = p - a;
            float lenSq = dot(ab, ab);
            if (lenSq > 1e-6) {
                float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
                float d = length(pa - ab * h);
                if (d < minEdgeD) minEdgeD = d;
            }
        }
    }
    float edgeMask = exp(-minEdgeD * minEdgeD * 25000.0);
    col += vec3(0.15, 0.20, 0.30) * edgeMask * 0.7;

    // === Packets in flight ===
    float minPacketD = 1e9;
    vec2 closestPacketPos = vec2(0.0);
    int closestPacketIdx = 0;
    float closestPacketFade = 0.0;

    for (int pI = 0; pI < PACKETS; pI++) {
        int n0, n1, n2, n3;
        packetPath(pI, x_Time, n0, n1, n2, n3);

        float tInPath = mod(x_Time + pkt_hash(float(pI) * 1.7) * PATH_SECONDS, PATH_SECONDS) / PATH_SECONDS;
        float edgeIdx = floor(tInPath * 3.0);
        float edgePhase = fract(tInPath * 3.0);
        int nA, nB;
        if (edgeIdx < 0.5) { nA = n0; nB = n1; }
        else if (edgeIdx < 1.5) { nA = n1; nB = n2; }
        else { nA = n2; nB = n3; }

        vec2 a = nodePos(nA);
        vec2 b = nodePos(nB);
        vec2 packetHead = mix(a, b, edgePhase);
        // Trail length along edge
        float trailStart = max(0.0, edgePhase - 0.35);
        vec2 trailTail = mix(a, b, trailStart);
        // Distance to the head
        float dHead = length(p - packetHead);
        // Distance to segment from trailTail → packetHead
        vec2 ab = packetHead - trailTail;
        vec2 pa = p - trailTail;
        float lenSq = dot(ab, ab);
        if (lenSq > 1e-6) {
            float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
            float d = length(pa - ab * h);
            float packetMask = exp(-d * d * 80000.0);
            float fade = h;  // trail brighter near head
            vec3 pktCol = pkt_pal(fract(float(pI) * 0.13 + x_Time * 0.05));
            col += pktCol * packetMask * (0.3 + fade * 0.8) * 1.1;
        }
        // Bright head
        col += vec3(1.0, 0.98, 0.85) * exp(-dHead * dHead * 30000.0) * 1.3;
    }

    // === Nodes (routers) ===
    for (int i = 0; i < NODES; i++) {
        vec2 n = nodePos(i);
        float nd = length(p - n);
        // Node: small bright disc with rim
        float nodeSize = 0.014;
        float core = exp(-nd * nd / (nodeSize * nodeSize) * 1.3);
        // Rim
        float rim = exp(-pow(nd - nodeSize * 0.9, 2.0) * 50000.0);

        // Activity boost: brighten when any packet is near this node (edge endpoint)
        float activity = 0.0;
        for (int pI = 0; pI < PACKETS; pI++) {
            int n0, n1, n2, n3;
            packetPath(pI, x_Time, n0, n1, n2, n3);
            if (n0 == i || n1 == i || n2 == i || n3 == i) {
                float tInPath = mod(x_Time + pkt_hash(float(pI) * 1.7) * PATH_SECONDS, PATH_SECONDS) / PATH_SECONDS;
                // Check if current packet head is near this node
                float edgePhase = fract(tInPath * 3.0);
                float edgeIdx = floor(tInPath * 3.0);
                bool approaching = false;
                if (edgeIdx < 0.5 && n1 == i && edgePhase > 0.8) approaching = true;
                else if (edgeIdx < 1.5 && n2 == i && edgePhase > 0.8) approaching = true;
                else if (edgeIdx > 1.5 && n3 == i && edgePhase > 0.8) approaching = true;
                if (approaching) activity = max(activity, (edgePhase - 0.8) / 0.2);
            }
        }

        vec3 nodeCol = pkt_pal(fract(float(i) * 0.08));
        col += nodeCol * core * (1.0 + activity * 1.5);
        col += nodeCol * rim * 0.5;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
