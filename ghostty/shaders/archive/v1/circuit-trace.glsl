// circuit-trace.glsl — Procedural circuit board with animated data flow
// Category: Cyberpunk | Cost: MED | Source: original
precision highp float;

float _hash(vec2 p) {
    uvec2 q = uvec2(p * 256.0) * uvec2(1597334673u, 3812015801u);
    uint n = (q.x ^ q.y) * 1597334673u;
    return float(n) / float(0xffffffffu);
}

float termLuminance(vec3 rgb) {
    return dot(rgb, vec3(0.2126, 0.7152, 0.0722));
}
float termMask(vec3 termColor) {
    return 1.0 - smoothstep(0.05, 0.25, termLuminance(termColor));
}

// Edge-based hashing: ensures adjacent cells agree on shared edges
// Returns 0.0 or 1.0 for whether an edge is "on"
float edgeOn(vec2 cellA, vec2 cellB) {
    vec2 edgeID = cellA + cellB; // symmetric: same from either side
    return step(0.45, _hash(edgeID * 0.37));
}

// Line segment SDF
float lineSDF(vec2 p, vec2 a, vec2 b, float w) {
    vec2 pa = p - a, ba = b - a;
    float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
    return length(pa - ba * h) - w;
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 term = texture(iChannel0, uv);
    float t = iTime;

    // --- Configuration ---
    float gridSize     = 20.0;     // cell size in pixels
    float traceWidth   = 1.2;      // trace line width
    float viaRadius    = 2.5;      // via/pad circle radius
    float dataSpeed    = 2.0;      // data pulse travel speed
    float dataLength   = 0.15;     // pulse trail length (0-1 of cell)

    // --- Colors ---
    vec3 substrate  = vec3(0.01, 0.04, 0.02);   // dark green PCB
    vec3 copper     = vec3(0.72, 0.55, 0.20);    // copper trace
    vec3 copperDark = vec3(0.35, 0.28, 0.10);    // darker copper
    vec3 dataCyan   = vec3(0.1, 1.0, 0.9);       // data pulse color
    vec3 viaMetal   = vec3(0.6, 0.6, 0.65);      // via pad silver

    // --- Grid ---
    vec2 pos = fragCoord;
    vec2 cellF = pos / gridSize;
    vec2 cellID = floor(cellF);
    vec2 cellUV = fract(cellF); // 0-1 within cell

    vec2 cellCenter = vec2(0.5);

    // --- Determine trace directions using edge hashing ---
    // Each cell can connect to 4 neighbors: right, up, left, down
    float right = edgeOn(cellID, cellID + vec2(1.0, 0.0));
    float up    = edgeOn(cellID, cellID + vec2(0.0, 1.0));
    float left  = edgeOn(cellID, cellID + vec2(-1.0, 0.0));
    float down  = edgeOn(cellID, cellID + vec2(0.0, -1.0));

    // At least one connection (avoid orphan cells)
    float totalConn = right + up + left + down;
    if (totalConn < 0.5) {
        right = 1.0;
        down = 1.0;
    }

    // --- Draw traces as line segments ---
    float traceDist = 999.0;
    float dataGlow = 0.0;
    float w = traceWidth / gridSize;

    // Cell-local coordinates for SDF
    vec2 lp = cellUV;

    // Right trace
    if (right > 0.5) {
        traceDist = min(traceDist, lineSDF(lp, cellCenter, vec2(1.0, 0.5), w));
        float dataPhase = fract(_hash(cellID * 1.1) + t * dataSpeed * 0.2);
        float dataPos = dataPhase;
        float dataDist = abs(lp.x - mix(0.5, 1.0, dataPos));
        float dataMask = step(abs(lp.y - 0.5), w * 2.0);
        dataGlow += exp(-dataDist * dataDist / (dataLength * dataLength)) * dataMask * step(0.5, lp.x);
    }

    // Up trace
    if (up > 0.5) {
        traceDist = min(traceDist, lineSDF(lp, cellCenter, vec2(0.5, 1.0), w));
        float dataPhase = fract(_hash(cellID * 2.3) + t * dataSpeed * 0.18);
        float dataPos = dataPhase;
        float dataDist = abs(lp.y - mix(0.5, 1.0, dataPos));
        float dataMask = step(abs(lp.x - 0.5), w * 2.0);
        dataGlow += exp(-dataDist * dataDist / (dataLength * dataLength)) * dataMask * step(0.5, lp.y);
    }

    // Left trace
    if (left > 0.5) {
        traceDist = min(traceDist, lineSDF(lp, cellCenter, vec2(0.0, 0.5), w));
        float dataPhase = fract(_hash(cellID * 3.7) + t * dataSpeed * 0.22);
        float dataPos = dataPhase;
        float dataDist = abs(lp.x - mix(0.5, 0.0, dataPos));
        float dataMask = step(abs(lp.y - 0.5), w * 2.0);
        dataGlow += exp(-dataDist * dataDist / (dataLength * dataLength)) * dataMask * step(lp.x, 0.5);
    }

    // Down trace
    if (down > 0.5) {
        traceDist = min(traceDist, lineSDF(lp, cellCenter, vec2(0.5, 0.0), w));
        float dataPhase = fract(_hash(cellID * 4.1) + t * dataSpeed * 0.15);
        float dataPos = dataPhase;
        float dataDist = abs(lp.y - mix(0.5, 0.0, dataPos));
        float dataMask = step(abs(lp.x - 0.5), w * 2.0);
        dataGlow += exp(-dataDist * dataDist / (dataLength * dataLength)) * dataMask * step(lp.y, 0.5);
    }

    // --- Via pads at junctions ---
    float viaDist = length(lp - cellCenter);
    float viaR = viaRadius / gridSize;
    float viaMask = smoothstep(viaR, viaR - 0.02, viaDist);
    float viaHole = smoothstep(viaR * 0.4, viaR * 0.35, viaDist);

    // --- IC/component placement (sparse) ---
    float componentMask = 0.0;
    vec3 componentColor = vec3(0.08, 0.08, 0.1);
    float compHash = _hash(cellID * 0.73);
    if (compHash > 0.92 && totalConn >= 2.0) {
        // Small rectangle (IC chip)
        vec2 compSize = vec2(0.35, 0.25);
        vec2 compDist = abs(lp - cellCenter) - compSize;
        float comp = step(max(compDist.x, compDist.y), 0.0);
        componentMask = comp;
        // Pin marks on edges
        float pins = step(0.9, sin(lp.x * 40.0)) * step(abs(lp.y - 0.5), 0.28);
        componentColor = mix(componentColor, vec3(0.5, 0.45, 0.2), pins * 0.5);
    }

    // --- Composite circuit ---
    float traceAA = smoothstep(0.005, -0.005, traceDist);

    // Trace color with subtle variation
    float copperVar = 0.85 + 0.15 * _hash(cellID);
    vec3 traceColor = mix(copperDark, copper, copperVar) * traceAA;

    // Data pulse glow
    dataGlow = clamp(dataGlow, 0.0, 1.0);
    vec3 dataColor = dataCyan * dataGlow * 1.5;

    // Via rendering
    vec3 viaColor = viaMetal * viaMask * (1.0 - viaHole * 0.7);

    // Build scene
    vec3 scene = substrate;
    scene = mix(scene, componentColor, componentMask);
    scene = mix(scene, traceColor, traceAA);
    scene += viaColor * step(0.5, totalConn);
    scene += dataColor;

    // Subtle substrate texture
    float subTex = _hash(floor(pos * 0.5)) * 0.02;
    scene += subTex * vec3(0.0, 0.3, 0.1);

    // --- Blend with terminal ---
    float mask = termMask(term.rgb);
    vec3 finalColor = mix(term.rgb, scene, mask * 0.9);

    fragColor = vec4(finalColor, term.a);
}
