// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Cave petroglyphs — primitive glyph figures lit by flickering torch on dark stone wall

const int   GLYPHS = 5;
const float INTENSITY = 0.55;

vec3 pg_pal(float t) {
    vec3 warm_orange = vec3(0.95, 0.55, 0.25);
    vec3 red = vec3(0.85, 0.30, 0.20);
    vec3 gold = vec3(0.96, 0.85, 0.40);
    vec3 dim = vec3(0.55, 0.30, 0.15);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(warm_orange, red, s);
    else if (s < 2.0) return mix(red, gold, s - 1.0);
    else if (s < 3.0) return mix(gold, dim, s - 2.0);
    else              return mix(dim, warm_orange, s - 3.0);
}

float pg_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float pg_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(pg_hash(i), pg_hash(i + vec2(1,0)), u.x),
               mix(pg_hash(i + vec2(0,1)), pg_hash(i + vec2(1,1)), u.x), u.y);
}

float pg_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 4; i++) {
        v += a * pg_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

// SDF for a simple stick figure
float sdStickFigure(vec2 p) {
    // Head (circle)
    vec2 headCenter = vec2(0.0, 0.04);
    float headD = length(p - headCenter) - 0.012;
    // Body (vertical line)
    vec2 bodyStart = vec2(0.0, 0.025);
    vec2 bodyEnd = vec2(0.0, -0.015);
    vec2 pa = p - bodyStart;
    vec2 ba = bodyEnd - bodyStart;
    float bh = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
    float bodyD = length(pa - ba * bh) - 0.003;
    // Arms (horizontal spread)
    float armD = abs(p.y - 0.010) - 0.003;
    if (armD > 0.0) armD = length(vec2(armD, max(0.0, abs(p.x) - 0.018)));
    // Legs (diagonal)
    vec2 legsStart = vec2(0.0, -0.015);
    vec2 leg1End = vec2(-0.01, -0.03);
    vec2 leg2End = vec2(0.01, -0.03);
    vec2 pa1 = p - legsStart;
    vec2 ba1 = leg1End - legsStart;
    float h1 = clamp(dot(pa1, ba1) / dot(ba1, ba1), 0.0, 1.0);
    float leg1D = length(pa1 - ba1 * h1) - 0.003;
    vec2 ba2 = leg2End - legsStart;
    float h2 = clamp(dot(pa1, ba2) / dot(ba2, ba2), 0.0, 1.0);
    float leg2D = length(pa1 - ba2 * h2) - 0.003;
    return min(headD, min(bodyD, min(armD, min(leg1D, leg2D))));
}

// SDF for a simple sun
float sdSun(vec2 p) {
    float r = length(p);
    float disc = r - 0.02;
    // Rays — 8 spokes
    float ang = atan(p.y, p.x);
    float rayAngFrac = mod(ang, 6.28318 / 8.0) - 6.28318 / 16.0;
    float rayRay = r * 0.05 * cos(rayAngFrac) - 0.002;
    float rayMask = smoothstep(0.03, 0.04, r) * smoothstep(0.05, 0.04, r);
    float raysD = mix(0.005, rayRay, rayMask);
    return min(disc, raysD);
}

// SDF for spiral
float sdSpiral(vec2 p) {
    float r = length(p);
    float ang = atan(p.y, p.x);
    float spiralR = 0.002 * ang + 0.005;
    float distToSpiral = abs(r - spiralR);
    if (r > 0.025) return 1.0;
    return distToSpiral - 0.002;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Rough stone texture backdrop
    float stoneTexture = pg_fbm(p * 10.0) * 0.3 + 0.1;
    stoneTexture += pg_fbm(p * 40.0) * 0.1;
    vec3 col = vec3(0.15, 0.08, 0.05) * stoneTexture;

    // Torch light — flickering orange-red, wanders
    vec2 torchPos = vec2(0.25 * sin(x_Time * 0.1), -0.1);
    float torchFlicker = 0.7 + 0.3 * pg_fbm(vec2(x_Time * 3.0, 0.0));
    float torchD = length(p - torchPos);
    vec3 torchCol = pg_pal(0.0);
    col += torchCol * exp(-torchD * torchD * 5.0) * torchFlicker * 0.8;

    // Glyphs carved into wall (dark indentations visible in torchlight)
    vec2 glyphPositions[5];
    glyphPositions[0] = vec2(-0.3, 0.15);
    glyphPositions[1] = vec2(-0.1, 0.1);
    glyphPositions[2] = vec2(0.1, 0.15);
    glyphPositions[3] = vec2(0.3, 0.05);
    glyphPositions[4] = vec2(-0.15, -0.2);

    float minGlyphD = 1e9;
    for (int i = 0; i < GLYPHS; i++) {
        vec2 localP = p - glyphPositions[i];
        float d;
        if (i == 0 || i == 2) d = sdStickFigure(localP);
        else if (i == 1) d = sdSun(localP);
        else if (i == 3) d = sdSpiral(localP);
        else d = sdStickFigure(localP * 0.8);
        minGlyphD = min(minGlyphD, d);
    }

    // Dark indentation in stone
    float glyphMask = smoothstep(0.003, 0.0, minGlyphD);
    col *= 1.0 - glyphMask * 0.6;

    // Torch illuminates glyphs — brighter if near torch
    float glyphTorchLit = exp(-torchD * torchD * 10.0) * 0.5;
    // Pigment hint — faint red ochre on glyph edges
    float edgeGlow = exp(-minGlyphD * minGlyphD * 15000.0) * glyphTorchLit * torchFlicker;
    col += vec3(0.6, 0.2, 0.1) * edgeGlow;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.5);
    vec3 result = mix(terminal.rgb, col, visibility * 0.8);

    _wShaderOut = vec4(result, 1.0);
}
