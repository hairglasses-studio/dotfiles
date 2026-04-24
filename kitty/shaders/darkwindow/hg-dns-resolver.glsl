// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — DNS resolver — 4-level hierarchical tree (root → 4 TLDs → 3 SLDs each → 2 records each), live query pulses descending a random path over time, cache-hit preloading of parallel paths, node blink on pulse arrival

const int   TLDS = 4;
const int   SLDS_PER_TLD = 3;
const int   RECS_PER_SLD = 2;
const int   QUERIES = 4;
const float QUERY_CYCLE = 4.0;
const float INTENSITY = 0.55;

vec3 dns_pal(int level) {
    if (level == 0) return vec3(1.00, 0.95, 0.70); // root = warm white
    if (level == 1) return vec3(0.30, 0.90, 1.00); // TLD = cyan
    if (level == 2) return vec3(0.95, 0.30, 0.70); // SLD = magenta
    return vec3(0.40, 0.95, 0.60);                  // record = mint
}

float dns_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Positions for each node: 4-level tree, y = 0.55 - level*0.3
vec2 rootPos() { return vec2(0.0, 0.55); }
vec2 tldPos(int t) {
    float ft = float(t);
    float x = (ft + 0.5) / float(TLDS) * 1.6 - 0.8;
    return vec2(x, 0.25);
}
vec2 sldPos(int t, int s) {
    float ft = float(t), fs = float(s);
    vec2 tp = tldPos(t);
    float x = tp.x + (fs - 1.0) * 0.09;
    return vec2(x, -0.05);
}
vec2 recPos(int t, int s, int r) {
    float fr = float(r);
    vec2 sp = sldPos(t, s);
    float x = sp.x + (fr - 0.5) * 0.04;
    return vec2(x, -0.35);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.008, 0.022);

    // Map-like gridlines
    vec2 gp = fract(p * 10.0);
    float gx = smoothstep(0.015, 0.0, abs(gp.x - 0.5) - 0.48);
    float gy = smoothstep(0.015, 0.0, abs(gp.y - 0.5) - 0.48);
    col += vec3(0.08, 0.10, 0.18) * max(gx, gy) * 0.15;

    vec2 root = rootPos();

    // === Draw tree edges ===
    float minEdgeD = 1e9;
    int edgeLevel = 0;
    for (int t = 0; t < TLDS; t++) {
        vec2 tp = tldPos(t);
        // Edge: root → TLD
        vec2 ab = tp - root;
        vec2 pa = p - root;
        float lenSq = dot(ab, ab);
        if (lenSq > 1e-6) {
            float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
            float d = length(pa - ab * h);
            if (d < minEdgeD) { minEdgeD = d; edgeLevel = 0; }
        }
        for (int s = 0; s < SLDS_PER_TLD; s++) {
            vec2 sp = sldPos(t, s);
            ab = sp - tp;
            pa = p - tp;
            lenSq = dot(ab, ab);
            if (lenSq > 1e-6) {
                float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
                float d = length(pa - ab * h);
                if (d < minEdgeD) { minEdgeD = d; edgeLevel = 1; }
            }
            for (int r = 0; r < RECS_PER_SLD; r++) {
                vec2 rp = recPos(t, s, r);
                ab = rp - sp;
                pa = p - sp;
                lenSq = dot(ab, ab);
                if (lenSq > 1e-6) {
                    float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
                    float d = length(pa - ab * h);
                    if (d < minEdgeD) { minEdgeD = d; edgeLevel = 2; }
                }
            }
        }
    }
    float edgeMask = exp(-minEdgeD * minEdgeD * 30000.0);
    col += vec3(0.20, 0.22, 0.30) * edgeMask * 0.5;

    // === Query pulses descending the tree ===
    for (int q = 0; q < QUERIES; q++) {
        float fq = float(q);
        float queryOffset = dns_hash(fq * 1.7);
        float tInCycle = mod(x_Time + queryOffset * QUERY_CYCLE, QUERY_CYCLE) / QUERY_CYCLE;
        // Choose path (rerolls every cycle)
        float pathSeed = floor(x_Time / QUERY_CYCLE + queryOffset) + fq;
        int tChoose = int(dns_hash(pathSeed) * float(TLDS)) % TLDS;
        int sChoose = int(dns_hash(pathSeed * 2.3) * float(SLDS_PER_TLD)) % SLDS_PER_TLD;
        int rChoose = int(dns_hash(pathSeed * 3.1) * float(RECS_PER_SLD)) % RECS_PER_SLD;

        // Pulse goes through 3 hops: root→TLD, TLD→SLD, SLD→rec over the cycle
        int hopIdx = int(floor(tInCycle * 3.0));
        float hopPhase = fract(tInCycle * 3.0);
        vec2 pulseStart, pulseEnd;
        int pulseLevel;
        if (hopIdx == 0) {
            pulseStart = rootPos();
            pulseEnd = tldPos(tChoose);
            pulseLevel = 0;
        } else if (hopIdx == 1) {
            pulseStart = tldPos(tChoose);
            pulseEnd = sldPos(tChoose, sChoose);
            pulseLevel = 1;
        } else {
            pulseStart = sldPos(tChoose, sChoose);
            pulseEnd = recPos(tChoose, sChoose, rChoose);
            pulseLevel = 2;
        }
        vec2 pulsePos = mix(pulseStart, pulseEnd, hopPhase);
        float pd = length(p - pulsePos);
        vec3 pulseCol = dns_pal(pulseLevel + 1);
        col += pulseCol * exp(-pd * pd * 20000.0) * 1.3;
        // Trail
        vec2 trailTail = mix(pulseStart, pulseEnd, max(0.0, hopPhase - 0.35));
        vec2 ab = pulsePos - trailTail;
        vec2 pa = p - trailTail;
        float lenSq = dot(ab, ab);
        if (lenSq > 1e-6) {
            float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
            float d = length(pa - ab * h);
            float trailMask = exp(-d * d * 40000.0);
            col += pulseCol * trailMask * h * 0.6;
        }
    }

    // === Nodes ===
    // Root
    float rd = length(p - root);
    col += dns_pal(0) * exp(-rd * rd * 6000.0) * 1.4;
    col += dns_pal(0) * exp(-rd * rd * 200.0) * 0.2;

    // TLDs
    for (int t = 0; t < TLDS; t++) {
        vec2 tp = tldPos(t);
        float d = length(p - tp);
        col += dns_pal(1) * exp(-d * d * 10000.0) * 1.0;
    }
    // SLDs
    for (int t = 0; t < TLDS; t++) {
        for (int s = 0; s < SLDS_PER_TLD; s++) {
            vec2 sp = sldPos(t, s);
            float d = length(p - sp);
            col += dns_pal(2) * exp(-d * d * 15000.0) * 0.75;
        }
    }
    // Records
    for (int t = 0; t < TLDS; t++) {
        for (int s = 0; s < SLDS_PER_TLD; s++) {
            for (int r = 0; r < RECS_PER_SLD; r++) {
                vec2 rp = recPos(t, s, r);
                float d = length(p - rp);
                col += dns_pal(3) * exp(-d * d * 25000.0) * 0.55;
            }
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
