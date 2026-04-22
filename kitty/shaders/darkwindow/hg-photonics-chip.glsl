// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Silicon photonics chip — waveguides with light pulses, ring resonators, IC pads

const float TRACE_WIDTH = 0.003;
const int   PULSES = 10;
const float INTENSITY = 0.55;

vec3 pc_pal(float t) {
    vec3 cyan = vec3(0.20, 0.90, 0.95);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 mag  = vec3(0.95, 0.30, 0.70);
    vec3 white = vec3(0.95, 0.98, 1.00);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, white, s - 2.0);
    else              return mix(white, cyan, s - 3.0);
}

float pc_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// Predefined waveguide paths (points on grid)
// 5 waveguide segments for illustration
float sdSegment(vec2 p, vec2 a, vec2 b) {
    vec2 pa = p - a;
    vec2 ba = b - a;
    float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
    return length(pa - ba * h);
}

float sdRing(vec2 p, vec2 c, float r) {
    return abs(length(p - c) - r);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.02, 0.04, 0.05);

    // Substrate grid (very subtle)
    vec2 gridDist = abs(fract(p * 20.0) - 0.5);
    float grid = 1.0 - smoothstep(0.005, 0.0, min(gridDist.x, gridDist.y));
    col += vec3(0.02, 0.05, 0.08) * grid * 0.2;

    // Waveguide network (straight and bent segments)
    float minD = 1e9;
    float bestT = 0.0;
    vec2 bestDir = vec2(1.0, 0.0);

    // Horizontal trunk
    float d1 = sdSegment(p, vec2(-0.5, 0.0), vec2(0.5, 0.0));
    if (d1 < minD) { minD = d1; bestT = (p.x + 0.5) / 1.0; bestDir = vec2(1.0, 0.0); }

    // Top branches (vertical)
    float d2 = sdSegment(p, vec2(-0.2, 0.0), vec2(-0.2, 0.3));
    if (d2 < minD) { minD = d2; bestT = 1.2 + p.y / 0.3; bestDir = vec2(0.0, 1.0); }
    float d3 = sdSegment(p, vec2(0.2, 0.0), vec2(0.2, 0.3));
    if (d3 < minD) { minD = d3; bestT = 1.5 + p.y / 0.3; bestDir = vec2(0.0, 1.0); }
    float d4 = sdSegment(p, vec2(-0.2, 0.3), vec2(0.2, 0.3));
    if (d4 < minD) { minD = d4; bestT = 1.8 + (p.x + 0.2) / 0.4; bestDir = vec2(1.0, 0.0); }

    // Bottom branches
    float d5 = sdSegment(p, vec2(-0.1, 0.0), vec2(-0.1, -0.2));
    if (d5 < minD) { minD = d5; bestT = 2.2 + -p.y / 0.2; bestDir = vec2(0.0, -1.0); }
    float d6 = sdSegment(p, vec2(0.1, 0.0), vec2(0.1, -0.2));
    if (d6 < minD) { minD = d6; bestT = 2.5 + -p.y / 0.2; bestDir = vec2(0.0, -1.0); }

    // Ring resonators (coupled to trunks)
    float dRing1 = sdRing(p, vec2(-0.35, 0.05), 0.045);
    if (dRing1 < minD) { minD = dRing1; bestT = 3.0 + atan(p.y - 0.05, p.x + 0.35) / 6.28; bestDir = vec2(1.0, 0.0); }
    float dRing2 = sdRing(p, vec2(0.35, 0.05), 0.045);
    if (dRing2 < minD) { minD = dRing2; bestT = 3.5 + atan(p.y - 0.05, p.x - 0.35) / 6.28; bestDir = vec2(1.0, 0.0); }

    // Trace rendering
    float core = smoothstep(TRACE_WIDTH, 0.0, minD);
    float glow = exp(-minD * minD * 2000.0) * 0.3;
    col += pc_pal(0.1) * (core * 0.7 + glow * 0.5);

    // Light pulses traveling along waveguides
    for (int i = 0; i < PULSES; i++) {
        float fi = float(i);
        float pulsePhase = fract(x_Time * 0.4 + fi * 0.15);
        // Is this pixel on the path right now at this phase?
        float phaseDiff = abs(fract(bestT) - pulsePhase);
        phaseDiff = min(phaseDiff, 1.0 - phaseDiff);
        float pulse = exp(-phaseDiff * phaseDiff * 200.0) * core;
        col += pc_pal(fract(fi * 0.1 + x_Time * 0.05)) * pulse * 1.2;
    }

    // IC pads — small bright squares at ends of waveguides
    for (int pad = 0; pad < 4; pad++) {
        float fpad = float(pad);
        vec2 padPos = vec2((fpad - 1.5) * 0.25, fpad < 2.0 ? 0.3 : -0.2);
        vec2 padP = p - padPos;
        vec2 padD = abs(padP) - vec2(0.015);
        float pd = length(max(padD, 0.0));
        float padMask = smoothstep(0.001, 0.0, pd - 0.005);
        col += pc_pal(fract(fpad * 0.2 + x_Time * 0.03)) * padMask * 0.8;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
