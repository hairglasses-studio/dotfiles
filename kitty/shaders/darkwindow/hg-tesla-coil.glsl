// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Tesla coil — high-voltage arcs from top electrode to random ground points, corona + plasma

const int   ARC_COUNT   = 8;
const int   ARC_SEGS    = 14;
const float INTENSITY   = 0.55;

vec3 tc_pal(float t) {
    vec3 a = vec3(0.70, 0.85, 1.00);  // white-blue arc core
    vec3 b = vec3(0.35, 0.60, 0.95);  // electric blue
    vec3 c = vec3(0.70, 0.30, 0.95);  // violet plasma
    vec3 d = vec3(0.95, 0.35, 0.60);  // magenta
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float tc_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Jagged bolt from start to end — polyline with random perpendicular displacement
float arcDist(vec2 p, vec2 start, vec2 end, float bendScale, float seed) {
    float minD = 1e9;
    vec2 dir = normalize(end - start);
    vec2 perp = vec2(-dir.y, dir.x);
    float totalLen = length(end - start);
    vec2 prev = start;
    for (int i = 1; i <= ARC_SEGS; i++) {
        float t = float(i) / float(ARC_SEGS);
        vec2 base = mix(start, end, t);
        // Random perpendicular displacement, zero at endpoints
        float envelope = sin(t * 3.14159);
        float disp = (tc_hash(seed + float(i) * 3.7) - 0.5) * bendScale * envelope;
        vec2 point = base + perp * disp;
        vec2 ab = point - prev;
        vec2 pa = p - prev;
        float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
        float d = length(pa - ab * h);
        minD = min(minD, d);
        prev = point;
    }
    return minD;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Dark backdrop (with subtle ambient)
    vec3 bg = mix(vec3(0.02, 0.01, 0.05), vec3(0.005, 0.005, 0.02), 1.0 - uv.y);
    vec3 col = bg;

    // Top electrode position (torus ring)
    vec2 elecPos = vec2(0.0, 0.4);
    // Ground baseline at bottom
    float groundY = -0.4;

    // Arcs
    float seedBase = floor(x_Time * 6.0);
    for (int a = 0; a < ARC_COUNT; a++) {
        float fa = float(a);
        float arcSeed = fa * 7.3 + seedBase;
        // Only ~3-4 arcs "activeFlag" at a time — on/off pattern
        float activeFlag = step(0.45, tc_hash(arcSeed));
        if (activeFlag < 0.5) continue;

        // Ground point for this arc
        float groundX = (tc_hash(arcSeed * 3.7) - 0.5) * 1.5;
        groundX *= x_WindowSize.x / x_WindowSize.y;
        vec2 groundP = vec2(groundX, groundY);

        // Brief lifetime — fade over time
        float lifePhase = fract(x_Time * 6.0 + arcSeed);
        float lifeFade = 1.0 - lifePhase;
        if (lifePhase > 0.7) continue;

        // Arc main bolt
        float d = arcDist(p, elecPos, groundP, 0.08, arcSeed);
        float core = 1.0 - smoothstep(0.002, 0.004, d);
        float glow = exp(-d * d * 500.0) * 0.6;
        float outerHalo = exp(-d * d * 50.0) * 0.15;

        vec3 arcCol = tc_pal(fract(fa * 0.07 + x_Time * 0.1));
        col += (vec3(1.0) * core * 1.5 + arcCol * glow + arcCol * outerHalo) * lifeFade;

        // Side branches from mid-path
        for (int b = 0; b < 3; b++) {
            float fb = float(b);
            float branchT = 0.3 + fb * 0.2 + tc_hash(arcSeed + fb) * 0.1;
            vec2 branchStart = mix(elecPos, groundP, branchT);
            float branchAngle = (tc_hash(arcSeed + fb * 2.1) - 0.5) * 2.0;
            float branchLen = 0.08 + tc_hash(arcSeed + fb * 3.7) * 0.06;
            vec2 dirToEnd = normalize(groundP - elecPos);
            vec2 perp = vec2(-dirToEnd.y, dirToEnd.x);
            vec2 branchEnd = branchStart + perp * branchAngle * branchLen + dirToEnd * branchLen * 0.3;
            float bd = arcDist(p, branchStart, branchEnd, 0.02, arcSeed + fb);
            float bCore = 1.0 - smoothstep(0.001, 0.003, bd);
            col += arcCol * bCore * lifeFade * 0.5;
        }
    }

    // Electrode corona — glowing ring at top
    float elecD = length(p - elecPos);
    float elecRingD = abs(elecD - 0.04);
    col += tc_pal(0.0) * exp(-elecRingD * elecRingD * 2000.0) * 0.6;
    col += tc_pal(0.3) * exp(-elecD * elecD * 40.0) * 0.2;

    // Ground line with sparks
    if (abs(p.y - groundY) < 0.005) {
        col += vec3(0.5, 0.5, 0.6) * smoothstep(0.005, 0.0, abs(p.y - groundY));
    }

    // Ambient electrical haze — ionized air blue tint
    col += vec3(0.03, 0.04, 0.08) * smoothstep(groundY + 0.1, 0.4, p.y);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
