// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Newton's fractal / magnetic pendulum basin of attraction — 3 poles, iterated attractor

const int   MAX_ITER = 24;
const float DAMPING  = 0.02;
const float G_MAG    = 0.6;
const float INTENSITY = 0.55;

const vec3 POLE_COL_A = vec3(0.10, 0.82, 0.92);
const vec3 POLE_COL_B = vec3(0.90, 0.30, 0.70);
const vec3 POLE_COL_C = vec3(0.20, 0.95, 0.60);

// Simulate a pendulum under gravity + 3 magnetic pull points, return index 0..2 of dominant pole
int magneticPendulum(vec2 p0, float t) {
    vec2 pole1 = vec2(0.0, 0.35);
    vec2 pole2 = vec2(-0.3, -0.2);
    vec2 pole3 = vec2(0.3, -0.2);
    // Slow rotation
    float ang = t * 0.1;
    float cr = cos(ang), sr = sin(ang);
    mat2 rot = mat2(cr, -sr, sr, cr);
    pole1 = rot * pole1;
    pole2 = rot * pole2;
    pole3 = rot * pole3;

    vec2 pos = p0;
    vec2 vel = vec2(0.0);
    for (int i = 0; i < MAX_ITER; i++) {
        // Force from each pole (inverse-square attraction)
        vec2 d1 = pole1 - pos; float r1 = dot(d1, d1) + 0.01; vec2 f1 = d1 / (r1 * sqrt(r1)) * G_MAG;
        vec2 d2 = pole2 - pos; float r2 = dot(d2, d2) + 0.01; vec2 f2 = d2 / (r2 * sqrt(r2)) * G_MAG;
        vec2 d3 = pole3 - pos; float r3 = dot(d3, d3) + 0.01; vec2 f3 = d3 / (r3 * sqrt(r3)) * G_MAG;
        // Restoring pull toward origin + damping
        vec2 fRest = -pos * 0.05;
        vec2 fDamp = -vel * DAMPING * 10.0;
        vec2 totalF = f1 + f2 + f3 + fRest + fDamp;
        vel += totalF * 0.04;
        pos += vel * 0.04;
    }
    // Dominant pole = nearest at end
    float d1f = distance(pos, pole1);
    float d2f = distance(pos, pole2);
    float d3f = distance(pos, pole3);
    if (d1f < d2f && d1f < d3f) return 0;
    if (d2f < d3f) return 1;
    return 2;
}

// Like above but returns the full final distance vector (for convergence shading)
vec3 pendulumSample(vec2 p0, float t) {
    vec2 pole1 = vec2(0.0, 0.35);
    vec2 pole2 = vec2(-0.3, -0.2);
    vec2 pole3 = vec2(0.3, -0.2);
    float ang = t * 0.1;
    float cr = cos(ang), sr = sin(ang);
    mat2 rot = mat2(cr, -sr, sr, cr);
    pole1 = rot * pole1;
    pole2 = rot * pole2;
    pole3 = rot * pole3;

    vec2 pos = p0;
    vec2 vel = vec2(0.0);
    float totalDist = 0.0;
    for (int i = 0; i < MAX_ITER; i++) {
        vec2 d1 = pole1 - pos; float r1 = dot(d1, d1) + 0.01; vec2 f1 = d1 / (r1 * sqrt(r1)) * G_MAG;
        vec2 d2 = pole2 - pos; float r2 = dot(d2, d2) + 0.01; vec2 f2 = d2 / (r2 * sqrt(r2)) * G_MAG;
        vec2 d3 = pole3 - pos; float r3 = dot(d3, d3) + 0.01; vec2 f3 = d3 / (r3 * sqrt(r3)) * G_MAG;
        vec2 fRest = -pos * 0.05;
        vec2 fDamp = -vel * DAMPING * 10.0;
        vec2 totalF = f1 + f2 + f3 + fRest + fDamp;
        vel += totalF * 0.04;
        vec2 newPos = pos + vel * 0.04;
        totalDist += distance(pos, newPos);
        pos = newPos;
    }
    // Return: .xy = final pos, .z = total path length (convergence metric)
    return vec3(pos, totalDist);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Determine pole and convergence speed
    vec3 result = pendulumSample(p, x_Time);
    vec2 finalPos = result.xy;
    float pathLen = result.z;

    // Basin color from which pole the ball ended at
    vec2 pole1 = vec2(0.0, 0.35);
    vec2 pole2 = vec2(-0.3, -0.2);
    vec2 pole3 = vec2(0.3, -0.2);
    float ang = x_Time * 0.1;
    float cr = cos(ang), sr = sin(ang);
    mat2 rot = mat2(cr, -sr, sr, cr);
    pole1 = rot * pole1; pole2 = rot * pole2; pole3 = rot * pole3;

    float d1 = distance(finalPos, pole1);
    float d2 = distance(finalPos, pole2);
    float d3 = distance(finalPos, pole3);

    vec3 basinCol;
    if (d1 < d2 && d1 < d3) basinCol = POLE_COL_A;
    else if (d2 < d3)       basinCol = POLE_COL_B;
    else                    basinCol = POLE_COL_C;

    // Brightness from convergence speed (shorter path = faster = brighter)
    float brightness = 1.0 / (1.0 + pathLen * 0.2);
    vec3 col = basinCol * (0.3 + brightness * 0.9);

    // Pole markers
    float d1P = length(p - pole1);
    float d2P = length(p - pole2);
    float d3P = length(p - pole3);
    col += POLE_COL_A * exp(-d1P * d1P * 2500.0) * 1.4;
    col += POLE_COL_B * exp(-d2P * d2P * 2500.0) * 1.4;
    col += POLE_COL_C * exp(-d3P * d3P * 2500.0) * 1.4;
    // Pole halos
    col += POLE_COL_A * exp(-d1P * d1P * 120.0) * 0.15;
    col += POLE_COL_B * exp(-d2P * d2P * 120.0) * 0.15;
    col += POLE_COL_C * exp(-d3P * d3P * 120.0) * 0.15;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 finalCol = mix(terminal.rgb, col, visibility * 0.85);

    _wShaderOut = vec4(finalCol, 1.0);
}
