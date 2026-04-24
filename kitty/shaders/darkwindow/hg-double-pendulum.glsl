// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Double pendulum — parametrically-approximated chaotic double-pendulum motion with two arms/masses, long fading phase-space trace of the end mass, and subtle background grid

const int   TRACE_SAMPS = 80;
const float L1 = 0.25;
const float L2 = 0.20;
const float INTENSITY = 0.55;

vec3 dp_pal(float t) {
    vec3 deep   = vec3(0.05, 0.08, 0.20);
    vec3 cyan   = vec3(0.30, 0.90, 1.00);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 amber  = vec3(1.00, 0.70, 0.30);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(deep, cyan, s);
    else if (s < 2.0) return mix(cyan, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else              return mix(mag, amber, s - 3.0);
}

// Pivot at top-center
vec2 pivot() { return vec2(0.0, 0.40); }

// Parametric approximation of chaotic double-pendulum angles.
// Real DP needs ODE integration; this mixes incommensurate frequencies.
vec2 pendulumAngles(float t) {
    float th1 = 1.1 * sin(t * 0.70)
              + 0.35 * sin(t * 1.83 + 1.7)
              + 0.18 * sin(t * 3.11 + 0.5);
    float th2 = 2.6 * cos(t * 0.89 + 0.3)
              + 0.55 * sin(t * 2.10 + 1.1)
              + 0.25 * cos(t * 4.17 + 2.2);
    return vec2(th1, th2);
}

// Returns (massA, massB) positions for pendulum at time t
void pendulumState(float t, out vec2 mA, out vec2 mB) {
    vec2 th = pendulumAngles(t);
    vec2 pv = pivot();
    mA = pv + vec2(sin(th.x), -cos(th.x)) * L1;
    mB = mA + vec2(sin(th.y), -cos(th.y)) * L2;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.004, 0.008, 0.022);

    // Subtle background grid
    vec2 gridP = fract(p * 8.0);
    float gridX = smoothstep(0.015, 0.0, abs(gridP.x - 0.5) - 0.48);
    float gridY = smoothstep(0.015, 0.0, abs(gridP.y - 0.5) - 0.48);
    col += vec3(0.12, 0.15, 0.25) * max(gridX, gridY) * 0.18;

    vec2 pv = pivot();

    // === Phase-space trace — connect past positions of the end mass ===
    float minTraceD = 1e9;
    float closestTraceS = 0.0;
    for (int i = 0; i < TRACE_SAMPS - 1; i++) {
        float age1 = float(i) * 0.05;
        float age2 = float(i + 1) * 0.05;
        vec2 a1, b1, a2, b2;
        pendulumState(x_Time - age1, a1, b1);
        pendulumState(x_Time - age2, a2, b2);
        // Segment from b1 to b2 in screen space
        vec2 seg = b2 - b1;
        vec2 pa = p - b1;
        float h = clamp(dot(pa, seg) / dot(seg, seg), 0.0, 1.0);
        float d = length(pa - seg * h);
        if (d < minTraceD) {
            minTraceD = d;
            closestTraceS = mix(age1, age2, h);
        }
    }
    // Trace rendering: thin bright line, fading with age
    float traceThick = 0.0025;
    float traceMask = exp(-minTraceD * minTraceD / (traceThick * traceThick) * 1.5);
    float ageFade = exp(-closestTraceS * 0.5);
    col += dp_pal(fract(closestTraceS * 0.3 + x_Time * 0.03)) * traceMask * ageFade * 1.1;
    // Soft halo
    col += dp_pal(0.4) * exp(-minTraceD * minTraceD * 800.0) * ageFade * 0.25;

    // === Current pendulum ===
    vec2 mA, mB;
    pendulumState(x_Time, mA, mB);

    // Arm 1: pivot → mA
    {
        vec2 ab = mA - pv;
        vec2 pa = p - pv;
        float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
        float d = length(pa - ab * h);
        float armMask = exp(-d * d * 20000.0);
        col += vec3(0.85, 0.90, 1.00) * armMask * 0.85;
    }
    // Arm 2: mA → mB
    {
        vec2 ab = mB - mA;
        vec2 pa = p - mA;
        float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
        float d = length(pa - ab * h);
        float armMask = exp(-d * d * 20000.0);
        col += vec3(0.85, 0.90, 1.00) * armMask * 0.85;
    }

    // Pivot mark
    float pvD = length(p - pv);
    col += vec3(0.95, 0.55, 0.25) * exp(-pvD * pvD * 8000.0) * 1.1;

    // Mass A (medium disc)
    float mAD = length(p - mA);
    col += vec3(0.30, 0.90, 1.00) * exp(-mAD * mAD * 5000.0) * 1.2;
    col += vec3(0.30, 0.90, 1.00) * exp(-mAD * mAD * 200.0) * 0.3;

    // Mass B (larger, brighter)
    float mBD = length(p - mB);
    col += vec3(1.00, 0.85, 0.35) * exp(-mBD * mBD * 3500.0) * 1.5;
    col += vec3(1.00, 0.85, 0.35) * exp(-mBD * mBD * 100.0) * 0.4;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
