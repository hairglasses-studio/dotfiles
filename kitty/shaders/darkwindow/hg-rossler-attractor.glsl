// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Rössler attractor — integrated Rössler system (a=0.2, b=0.2, c=5.7) with its characteristic single-band spiral + periodic upstroke, tilted orthographic projection, head pointer + age-faded trail

const int   STEPS = 200;
const float DT = 0.06;
const float R_A = 0.2;
const float R_B = 0.2;
const float R_C = 5.7;
const float INTENSITY = 0.55;
const float CYCLE = 18.0;

vec3 ros_pal(float t) {
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 amber  = vec3(1.00, 0.70, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    vec3 cyan   = vec3(0.25, 0.90, 1.00);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(vio, mag, s);
    else if (s < 2.0) return mix(mag, amber, s - 1.0);
    else if (s < 3.0) return mix(amber, mint, s - 2.0);
    else if (s < 4.0) return mix(mint, cyan, s - 3.0);
    else              return mix(cyan, vio, s - 4.0);
}

// Euler step of the Rössler system
vec3 rosslerStep(vec3 s, float dt) {
    float dx = -s.y - s.z;
    float dy = s.x + R_A * s.y;
    float dz = R_B + s.z * (s.x - R_C);
    return s + vec3(dx, dy, dz) * dt;
}

// Tilted orthographic projection
vec2 projectRossler(vec3 s) {
    float c = cos(0.45), sn = sin(0.45);
    // Attractor sits roughly around x,y ∈ [-10,10], z ∈ [0, 22]
    vec2 v = vec2(s.x * 0.055,
                  (s.y * c - (s.z - 4.0) * sn) * 0.055);
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.004, 0.006, 0.018);

    // Background radial fade
    float rdist = length(p);
    col += vec3(0.03, 0.02, 0.05) * (1.0 - smoothstep(0.0, 1.2, rdist));

    // Integrate Rössler from initial state, store 2D projections
    vec2 traj[STEPS];
    vec3 state = vec3(1.0, 1.0, 1.0);
    traj[0] = projectRossler(state);
    for (int i = 1; i < STEPS; i++) {
        state = rosslerStep(state, DT);
        traj[i] = projectRossler(state);
    }

    // Head advances over CYCLE
    float cycT = mod(x_Time, CYCLE) / CYCLE;
    float headF = cycT * float(STEPS);
    int headI = int(headF);

    // Trail + distance lookup
    float minD = 1e9;
    float closestAge = 0.0;
    int closestIdx = 0;
    float TRAIL_LEN = 120.0;
    for (int i = 0; i < STEPS - 1; i++) {
        vec2 a = traj[i];
        vec2 b = traj[i + 1];
        vec2 ab = b - a;
        vec2 pa = p - a;
        float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
        float d = length(pa - ab * h);
        float fi = float(i);
        float age = headF - fi;
        if (age < 0.0) age += float(STEPS);
        if (age < TRAIL_LEN && d < minD) {
            minD = d;
            closestAge = age;
            closestIdx = i;
        }
    }

    float traceThick = 0.004;
    float traceMask = exp(-minD * minD / (traceThick * traceThick) * 1.5);
    float ageFade = exp(-closestAge / 70.0);
    vec3 traceCol = ros_pal(fract(float(closestIdx) / float(STEPS) * 1.5 + x_Time * 0.03));
    col += traceCol * traceMask * ageFade * 1.15;
    // Halo
    col += traceCol * exp(-minD * minD * 700.0) * ageFade * 0.2;

    // Head pointer
    if (headI >= 0 && headI < STEPS) {
        vec2 headPos = traj[headI];
        float hd = length(p - headPos);
        col += vec3(1.0, 0.97, 0.80) * exp(-hd * hd * 8000.0) * 1.5;
        col += ros_pal(0.2) * exp(-hd * hd * 200.0) * 0.3;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
