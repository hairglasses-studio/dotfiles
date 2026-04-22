// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Hex grid with data packets flowing along edges, lit cell centers, pulsing gridlines

const float HEX_SIZE     = 0.065;
const float EDGE_WIDTH   = 0.005;
const float PACKET_LEN   = 0.3;      // along-edge length
const int   FLOW_STREAMS = 6;        // simultaneous packet streams
const float INTENSITY    = 0.5;

vec3 hd_pal(float t) {
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

float hd_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5);
}

// Hex axial coordinates → pixel center
vec2 hexToPixel(vec2 ax, float size) {
    float x = size * 1.732 * (ax.x + ax.y * 0.5);
    float y = size * 1.5 * ax.y;
    return vec2(x, y);
}

// Pixel → nearest hex axial coords
vec2 pixelToHex(vec2 p, float size) {
    float q = (p.x * 0.5774 - p.y * 0.3333) / size;
    float r = (p.y * 0.6667) / size;
    // Round to nearest hex
    float rx = floor(q + 0.5);
    float ry = floor(r + 0.5);
    float rz = floor(-q - r + 0.5);
    float x_diff = abs(rx - q);
    float y_diff = abs(ry - r);
    float z_diff = abs(rz - (-q - r));
    if (x_diff > y_diff && x_diff > z_diff)      rx = -ry - rz;
    else if (y_diff > z_diff)                    ry = -rx - rz;
    return vec2(rx, ry);
}

// Distance to nearest hex edge (within a given cell)
float hexEdgeDist(vec2 p, vec2 center, float size) {
    vec2 d = p - center;
    d.x = abs(d.x);
    d.y = abs(d.y);
    float ed = max(d.x * 0.866 + d.y * 0.5, d.y);  // distance to hex boundary
    return abs(ed - size * 0.866);
}

// Distance to nearest hex vertex (corner)
float hexVertexDist(vec2 p, vec2 center, float size) {
    vec2 d = p - center;
    float minD = 1e9;
    for (int i = 0; i < 6; i++) {
        float a = float(i) * 1.0472 + 0.5236;
        vec2 v = vec2(cos(a), sin(a)) * size * 0.866;
        minD = min(minD, length(d - v));
    }
    return minD;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Hex cell of this fragment
    vec2 ax = pixelToHex(p, HEX_SIZE);
    vec2 center = hexToPixel(ax, HEX_SIZE);

    // Cell seed (deterministic per cell)
    float cellSeed = hd_hash(ax);

    vec3 col = vec3(0.0);

    // Gridline base
    float ed = hexEdgeDist(p, center, HEX_SIZE);
    float edgeMask = smoothstep(EDGE_WIDTH, 0.0, ed);
    vec3 gridCol = hd_pal(fract(cellSeed + x_Time * 0.02)) * 0.35;
    col += gridCol * edgeMask;

    // Vertex dots (brighter corners)
    float vd = hexVertexDist(p, center, HEX_SIZE);
    float vertexMask = exp(-vd * vd * 4000.0);
    col += hd_pal(fract(cellSeed * 2.3 + x_Time * 0.03)) * vertexMask * 1.2;

    // Cell pulse — some cells "fire" periodically
    float cellPulse = fract(cellSeed * 7.1 + x_Time * 0.25);
    float pulseBright = pow(1.0 - cellPulse, 4.0);  // sharp peak near 0
    float cd = length(p - center);
    float pulseMask = exp(-cd * cd * 100.0) * pulseBright;
    col += hd_pal(fract(cellSeed * 1.7 + x_Time * 0.08)) * pulseMask * 1.6;

    // Data packets traveling along edges
    // Each stream picks a cell center and flows to a neighbor
    for (int s = 0; s < FLOW_STREAMS; s++) {
        float fs = float(s);
        float streamSeed = fs * 3.71 + floor(x_Time * 0.3);
        float streamHash = hd_hash(vec2(streamSeed, 0.0));
        // Pick a hex somewhere in a 4-ring radius of origin
        vec2 startAx = vec2(floor(streamHash * 6.0 - 3.0),
                            floor(hd_hash(vec2(streamSeed, 1.0)) * 6.0 - 3.0));
        int dirIdx = int(hd_hash(vec2(streamSeed, 2.0)) * 6.0);
        vec2 dirOffsets[6];
        dirOffsets[0] = vec2(1, 0); dirOffsets[1] = vec2(0, 1); dirOffsets[2] = vec2(-1, 1);
        dirOffsets[3] = vec2(-1, 0); dirOffsets[4] = vec2(0, -1); dirOffsets[5] = vec2(1, -1);
        vec2 endAx = startAx + dirOffsets[dirIdx];

        vec2 startP = hexToPixel(startAx, HEX_SIZE);
        vec2 endP   = hexToPixel(endAx, HEX_SIZE);

        // Packet phase — 0 to 1 over 0.8 second
        float packetPhase = mod(x_Time * 1.25 + streamHash * 3.0, 1.0);
        vec2 packetP = mix(startP, endP, packetPhase);

        // Distance along edge
        vec2 edgeVec = endP - startP;
        float edgeLen = length(edgeVec);
        vec2 toPkt = p - startP;
        float along = clamp(dot(toPkt, edgeVec) / (edgeLen * edgeLen), 0.0, 1.0);
        float tailPhase = packetPhase - PACKET_LEN;
        if (along < tailPhase || along > packetPhase) continue;

        float tailT = (along - tailPhase) / PACKET_LEN;  // [0,1] tail to head
        vec2 closest = startP + edgeVec * along;
        float d = length(p - closest);
        float kernel = exp(-d * d * 6000.0);
        vec3 packetCol = hd_pal(fract(fs * 0.2 + x_Time * 0.1));
        col += packetCol * kernel * tailT * 1.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
