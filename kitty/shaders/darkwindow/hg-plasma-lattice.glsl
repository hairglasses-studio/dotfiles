// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — 3D plasma lattice — raymarched grid of pulsing energy nodes with connecting arcs

const int   STEPS         = 48;
const float MAX_DIST      = 6.0;
const float LATTICE_SPACE = 0.85;
const float NODE_RADIUS   = 0.10;
const float INTENSITY     = 0.55;

vec3 pl_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.90, 0.30, 0.70); // pink
    vec3 c = vec3(0.55, 0.30, 0.98); // violet
    vec3 d = vec3(0.96, 0.70, 0.25); // amber
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// 3D repeated sphere SDF — tiled lattice of energy nodes
float sdSphere(vec3 p, float r) {
    return length(p) - r;
}

// Lattice distance: repeat 3D space on LATTICE_SPACE cells
float latticeDist(vec3 p, float t) {
    // Animated radius per node: each node pulses based on its grid position
    vec3 cellId = floor((p + LATTICE_SPACE * 0.5) / LATTICE_SPACE);
    float nodeHash = fract(sin(dot(cellId, vec3(12.9898, 78.233, 37.719))) * 43758.5);
    float pulseR = NODE_RADIUS * (0.7 + 0.5 * sin(t * 2.0 + nodeHash * 6.28));

    vec3 rep = mod(p + LATTICE_SPACE * 0.5, LATTICE_SPACE) - LATTICE_SPACE * 0.5;
    return sdSphere(rep, pulseR);
}

// Emission per position — colors based on nearest lattice center
vec3 latticeColor(vec3 p, float t) {
    vec3 cellId = floor((p + LATTICE_SPACE * 0.5) / LATTICE_SPACE);
    float nodeHash = fract(sin(dot(cellId, vec3(12.9898, 78.233, 37.719))) * 43758.5);
    return pl_pal(fract(nodeHash + t * 0.04));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Camera flying along +Z through the lattice, slight sway
    float t = x_Time;
    vec3 ro = vec3(0.25 * sin(t * 0.2), 0.25 * cos(t * 0.17), t * 0.5);
    vec3 rd = normalize(vec3(p, 1.3));

    // Slow roll
    float roll = x_Time * 0.07;
    float cr = cos(roll), sr = sin(roll);
    rd.xy = mat2(cr, -sr, sr, cr) * rd.xy;

    // Raymarch — accumulate emission and density
    float dist = 0.0;
    vec3 col = vec3(0.0);
    float transmittance = 1.0;

    for (int i = 0; i < STEPS; i++) {
        vec3 pos = ro + rd * dist;
        float d = latticeDist(pos, x_Time);

        // Emission glow around each node
        float glow = exp(-d * 6.0) * 0.08;
        if (glow > 0.001) {
            vec3 nc = latticeColor(pos, x_Time);
            col += nc * glow * transmittance;
            transmittance *= 1.0 - min(glow * 0.5, 0.3);
        }

        // Raymarch step — use SDF, but clamp minimum step so we don't stall at nodes
        dist += max(d * 0.9, 0.04);
        if (dist > MAX_DIST) break;
        if (transmittance < 0.02) break;
    }

    // Fog fades far nodes
    col *= exp(-dist * 0.08);

    // Connecting "arc" shimmer — adds thin interference pattern
    float arc = sin((p.x * 8.0 + p.y * 6.0) + x_Time) * 0.5 + 0.5;
    col += pl_pal(fract(x_Time * 0.05)) * pow(arc, 20.0) * 0.15;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
