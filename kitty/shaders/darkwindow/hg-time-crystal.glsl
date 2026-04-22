// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Time crystal — phase-shifted hexagonal lattice evolving through Floquet states

const float LATTICE_SIZE = 0.09;
const float INTENSITY = 0.55;

vec3 tcr_pal(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(0.96, 0.85, 0.40);
    vec3 d = vec3(0.90, 0.25, 0.65);
    vec3 e = vec3(0.20, 0.95, 0.60);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

float tcr_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// Hex cell center for pixel position
vec2 hexCell(vec2 p) {
    vec2 q = vec2(p.x * 0.5774 - p.y * 0.3333, p.y * 0.6667) / LATTICE_SIZE;
    vec2 a = floor(q);
    vec2 b = fract(q);
    // Odd row shift
    bool evenRow = mod(a.y, 2.0) < 1.0;
    return (evenRow ? a : a + vec2(0.5, 0.0)) * LATTICE_SIZE;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.03);

    // Grid
    float aspect = x_WindowSize.x / x_WindowSize.y;
    vec2 scaleP = p / LATTICE_SIZE;
    vec2 cellId = floor(scaleP);
    vec2 cellF = fract(scaleP);

    // Each cell has its own phase
    float cellSeed = tcr_hash(cellId);

    // Time-crystal period: 2T (double the drive period)
    // Phase-shifted by position: wave traveling in discrete steps across lattice
    float drivePhase = fract(x_Time * 0.4);
    float cellPhase = fract(drivePhase + cellId.x * 0.05 + cellId.y * 0.03 + cellSeed * 0.2);

    // Subharmonic oscillation — period is 2T
    float oscTime = x_Time * 0.2;  // T period
    float twoTPhase = mod(oscTime + cellSeed * 2.0, 2.0);
    float subharmonic = (twoTPhase < 1.0) ? twoTPhase : (2.0 - twoTPhase);

    // Cell core visibility
    vec2 cellCenter = cellId + 0.5;
    float dFromCenter = length((cellF - 0.5) * LATTICE_SIZE);
    float cellSize = LATTICE_SIZE * 0.35 * (0.5 + subharmonic * 0.7);
    float cellMask = smoothstep(cellSize * 1.1, cellSize * 0.8, dFromCenter);
    float cellGlow = exp(-dFromCenter * dFromCenter * 300.0) * 0.25;

    vec3 cellCol = tcr_pal(fract(cellSeed + cellPhase + x_Time * 0.03));

    col += cellCol * cellMask * (0.4 + subharmonic * 0.8);
    col += cellCol * cellGlow * subharmonic;

    // Cell-to-cell connections (only between neighbors with opposite phase — period-doubling signature)
    // Visualize by adding bright edges between cells
    vec2 gridDist = abs(cellF - 0.5);
    float edgeMask = smoothstep(0.48, 0.5, max(gridDist.x, gridDist.y));
    if (edgeMask > 0.01) {
        float phaseDiff = abs(subharmonic - 0.5);
        col += cellCol * edgeMask * phaseDiff * 0.2;
    }

    // Wave front indicator — shows the phase shift propagating
    float wavePos = cellId.x * 0.05 + cellId.y * 0.03 + cellSeed * 0.2 + drivePhase;
    float waveFront = smoothstep(0.02, 0.0, abs(fract(wavePos) - 0.5));
    col += vec3(1.0) * waveFront * cellMask * 0.3;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
