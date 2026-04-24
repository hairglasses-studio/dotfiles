// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Lorenz attractor — forward-integrated Lorenz system (σ=10, ρ=28, β=8/3) projected to 2D as the butterfly shape, with a moving head pointer tracing the trajectory and an age-fading trail that peaks near the head

const int   STEPS = 180;
const float DT = 0.020;
const float SIGMA = 10.0;
const float RHO = 28.0;
const float BETA = 2.6667;
const float INTENSITY = 0.55;
const float CYCLE = 14.0;    // time to traverse full trajectory

vec3 lor_pal(float t) {
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

// Euler step of the Lorenz system
vec3 lorenzStep(vec3 s, float dt) {
    float dx = SIGMA * (s.y - s.x);
    float dy = s.x * (RHO - s.z) - s.y;
    float dz = s.x * s.y - BETA * s.z;
    return s + vec3(dx, dy, dz) * dt;
}

// Project 3D (x,y,z) to 2D screen space, centered and scaled to fit
vec2 projectLorenz(vec3 s) {
    // Attractor spans roughly x,y ∈ [-25,25], z ∈ [0,50]
    // Rotate slightly for more interesting view: look down from above
    float c = cos(0.35), sn = sin(0.35);
    // Treat y as "up-back", use x (horizontal) and (y*cos - z*sin) (vertical)
    vec2 v = vec2(s.x * 0.025, (s.y * c - (s.z - 25.0) * sn) * 0.025);
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.004, 0.007, 0.020);

    // Subtle fract-grid background
    vec2 gridP = fract(p * 10.0);
    float gridX = smoothstep(0.015, 0.0, abs(gridP.x - 0.5) - 0.48);
    float gridY = smoothstep(0.015, 0.0, abs(gridP.y - 0.5) - 0.48);
    col += vec3(0.10, 0.13, 0.22) * max(gridX, gridY) * 0.13;

    // Compute trajectory — integrate the Lorenz system STEPS times from (1,1,1)
    // and store projected 2D points.
    vec2 traj[STEPS];
    vec3 state = vec3(1.0, 1.0, 1.0);
    traj[0] = projectLorenz(state);
    for (int i = 1; i < STEPS; i++) {
        state = lorenzStep(state, DT);
        traj[i] = projectLorenz(state);
    }

    // Current head position cycles from 0 to STEPS
    float cycT = mod(x_Time, CYCLE) / CYCLE;
    float headF = cycT * float(STEPS);
    int headI = int(headF);

    // For each trajectory segment, compute fragment distance + age-based intensity
    float minD = 1e9;
    float closestAge = 0.0;
    int closestIdx = 0;
    for (int i = 0; i < STEPS - 1; i++) {
        vec2 a = traj[i];
        vec2 b = traj[i + 1];
        vec2 ab = b - a;
        vec2 pa = p - a;
        float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
        float d = length(pa - ab * h);
        // Age relative to head: older segments behind head fade more
        float fi = float(i);
        float age = headF - fi;
        // Wrap so segments ahead of head are essentially at maximum age
        if (age < 0.0) age += float(STEPS);
        // Only care about segments within a trail window
        float TRAIL = 100.0;
        if (age < TRAIL && d < minD) {
            minD = d;
            closestAge = age;
            closestIdx = i;
        }
    }

    // Trail rendering — brighter near head, dimmer older
    float traceThick = 0.004;
    float traceMask = exp(-minD * minD / (traceThick * traceThick) * 1.5);
    float ageFade = exp(-closestAge / 60.0);
    vec3 traceCol = lor_pal(fract(float(closestIdx) / float(STEPS) * 2.0 + x_Time * 0.03));
    col += traceCol * traceMask * ageFade * 1.1;
    // Halo
    col += traceCol * exp(-minD * minD * 800.0) * ageFade * 0.22;

    // Head pointer: bright point at traj[headI]
    if (headI >= 0 && headI < STEPS) {
        vec2 headPos = traj[headI];
        float hd = length(p - headPos);
        col += vec3(1.0, 0.97, 0.80) * exp(-hd * hd * 8000.0) * 1.5;
        col += vec3(1.0, 0.80, 0.40) * exp(-hd * hd * 200.0) * 0.25;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
