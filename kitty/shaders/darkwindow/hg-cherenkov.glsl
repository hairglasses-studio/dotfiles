// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Cherenkov radiation — blue glow cone from particle traveling faster than light in medium

const int   PARTICLES = 6;
const float INTENSITY = 0.55;

vec3 ck_pal(float t) {
    vec3 deep = vec3(0.02, 0.08, 0.25);
    vec3 blue = vec3(0.20, 0.70, 1.00);
    vec3 cyan = vec3(0.40, 0.95, 0.98);
    vec3 white = vec3(0.80, 0.95, 1.00);
    if (t < 0.33)      return mix(deep, blue, t * 3.0);
    else if (t < 0.66) return mix(blue, cyan, (t - 0.33) * 3.0);
    else               return mix(cyan, white, (t - 0.66) * 3.0);
}

float ck_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  ck_hash2(float n) { return vec2(ck_hash(n), ck_hash(n * 1.37 + 11.0)); }

// Particle position at time t
vec4 particleState(int i, float t) {
    float fi = float(i);
    float seed = fi * 7.31;
    // Travel time — each particle cycles
    float cyclePeriod = 2.0 + ck_hash(seed * 2.0) * 1.5;
    float phase = mod(t + ck_hash(seed) * cyclePeriod, cyclePeriod) / cyclePeriod;
    // Start and end positions
    float cycleID = floor((t + ck_hash(seed) * cyclePeriod) / cyclePeriod);
    vec2 startP = ck_hash2(seed + cycleID * 13.7) * 2.0 - 1.0;
    startP.x *= x_WindowSize.x / x_WindowSize.y;
    vec2 endP = ck_hash2(seed * 3.7 + cycleID * 17.3) * 2.0 - 1.0;
    endP.x *= x_WindowSize.x / x_WindowSize.y;
    // Animated position
    vec2 pos = mix(startP, endP, phase);
    vec2 dir = normalize(endP - startP);
    return vec4(pos, dir);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Deep water-medium backdrop
    vec3 bg = mix(ck_pal(0.0), vec3(0.01, 0.03, 0.12), uv.y);
    vec3 col = bg;

    for (int i = 0; i < PARTICLES; i++) {
        vec4 state = particleState(i, x_Time);
        vec2 partPos = state.xy;
        vec2 partDir = state.zw;

        // Cone behind the particle — angle ~42° (Cherenkov cone)
        float coneAngle = 0.7;
        vec2 toP = p - partPos;
        float along = dot(toP, partDir);
        float perp = abs(dot(toP, vec2(-partDir.y, partDir.x)));

        // Only in the backward cone
        if (along < 0.0) {
            float expectedPerp = -along * tan(coneAngle);
            float coneDist = abs(perp - expectedPerp);
            // Falloff with distance from cone edge
            float cone = exp(-coneDist * coneDist * 300.0);
            // Fades behind as propagates
            float backFade = exp(along * 2.0);  // closer to particle = brighter
            float lengthMask = smoothstep(0.0, -0.4, along);
            // Intensity increases with angular distance from axis (edges bright, interior dim)
            float axisDist = along * tan(coneAngle) * 0.3;
            float edgeBright = (perp > axisDist) ? 1.0 : 0.3;
            col += ck_pal(0.7) * cone * backFade * lengthMask * edgeBright * 0.8;
        }

        // Particle head glow
        float d = length(p - partPos);
        col += ck_pal(0.8) * exp(-d * d * 2000.0) * 0.9;
        col += ck_pal(1.0) * exp(-d * d * 10000.0) * 1.5;

        // Particle trail (brief straight line behind it)
        if (along < 0.0 && along > -0.03) {
            float trailD = perp;
            col += ck_pal(0.9) * exp(-trailD * trailD * 8000.0) * exp(along * 30.0) * 1.2;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
