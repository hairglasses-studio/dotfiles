// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Vector field — arrow glyphs on grid showing curl-noise flow with traveling pulses

const float ARROW_GRID = 15.0;
const int   OCTAVES = 4;
const float INTENSITY = 0.55;

vec3 vf_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.25, 0.60);
    vec3 d = vec3(0.20, 0.95, 0.60);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float vf_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float vf_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(vf_hash(i), vf_hash(i + vec2(1,0)), u.x),
               mix(vf_hash(i + vec2(0,1)), vf_hash(i + vec2(1,1)), u.x), u.y);
}

float vf_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * vf_noise(p);
        p = p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

vec2 curlField(vec2 p, float t) {
    float eps = 0.01;
    float n1 = vf_fbm(p + vec2(0.0, eps) + vec2(t * 0.2, 0.0));
    float n2 = vf_fbm(p - vec2(0.0, eps) + vec2(t * 0.2, 0.0));
    float n3 = vf_fbm(p + vec2(eps, 0.0) + vec2(t * 0.2, 0.0));
    float n4 = vf_fbm(p - vec2(eps, 0.0) + vec2(t * 0.2, 0.0));
    return vec2((n1 - n2) / (2.0 * eps), -(n3 - n4) / (2.0 * eps));
}

// Draw an arrow at origin pointing in direction dir, length len
float arrow(vec2 p, vec2 dir, float len, float thick) {
    float dirLen = length(dir);
    if (dirLen < 0.01) return 1e9;
    dir /= dirLen;
    vec2 perp = vec2(-dir.y, dir.x);
    // Project onto along and perp axes
    float along = dot(p, dir);
    float perpD = abs(dot(p, perp));
    // Shaft: from 0 to len
    float shaft = 1e9;
    if (along > 0.0 && along < len * 0.7) {
        shaft = perpD;
    }
    // Arrowhead: from 0.7*len to len, wider then narrows
    float head = 1e9;
    if (along > len * 0.7 && along < len) {
        float headProgress = (along - len * 0.7) / (len * 0.3);
        float headW = thick * 3.0 * (1.0 - headProgress);
        if (perpD < headW) head = 0.0;
    }
    return min(shaft, head);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    // Grid cell
    float aspect = x_WindowSize.x / x_WindowSize.y;
    vec2 cellSize = vec2(1.0 / ARROW_GRID);
    vec2 gridCoord = p / cellSize.x;
    vec2 cellId = floor(gridCoord);
    vec2 cellCenter = (cellId + 0.5) * cellSize.x;

    // Compute flow at cell center
    vec2 flow = curlField(cellCenter, x_Time);
    float flowMag = length(flow);
    vec2 flowDir = flowMag > 0.001 ? flow / flowMag : vec2(1.0, 0.0);

    // Arrow within cell (relative to cell center)
    vec2 localP = p - cellCenter;
    float arrowLen = cellSize.x * 0.7 * clamp(flowMag * 3.0, 0.1, 1.0);
    float arrowD = arrow(localP, flowDir, arrowLen, cellSize.x * 0.03);
    float arrowMask = smoothstep(cellSize.x * 0.04, 0.0, arrowD);

    // Color based on direction
    float dirHue = atan(flowDir.y, flowDir.x) / 6.28318 + 0.5;
    vec3 arrowCol = vf_pal(fract(dirHue + x_Time * 0.03));
    col += arrowCol * arrowMask * 0.9;
    col += arrowCol * exp(-arrowD * arrowD * 1200.0) * 0.15;

    // Traveling pulse along flow — advect a pulse position
    float pulseSpeed = 0.15;
    float pulsePhase = fract(dot(p, flowDir) * 3.0 + x_Time * pulseSpeed);
    float pulseD = abs(pulsePhase - 0.5);
    if (arrowD < cellSize.x * 0.06) {
        col += vec3(1.0, 0.95, 0.9) * exp(-pulseD * pulseD * 80.0) * arrowMask * 0.6;
    }

    // Magnitude-based background tint
    col += arrowCol * flowMag * 0.05;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
