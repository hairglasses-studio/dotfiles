// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Tachyon beam — faster-than-light particle beam with retrograde Cherenkov cone

const int   TACHYONS = 5;
const float INTENSITY = 0.55;

vec3 tb_pal(float t) {
    vec3 blue  = vec3(0.20, 0.55, 0.98);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 white = vec3(0.95, 0.98, 1.00);
    vec3 mag   = vec3(0.95, 0.30, 0.70);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(blue, vio, s);
    else if (s < 2.0) return mix(vio, white, s - 1.0);
    else if (s < 3.0) return mix(white, mag, s - 2.0);
    else              return mix(mag, blue, s - 3.0);
}

float tb_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  tb_hash2(float n) { return vec2(tb_hash(n), tb_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.02, 0.05);

    for (int i = 0; i < TACHYONS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float cycle = 2.0 + tb_hash(seed) * 1.5;
        float phase = fract((x_Time + tb_hash(seed * 3.7) * cycle) / cycle);
        float cycleID = floor((x_Time + tb_hash(seed * 3.7) * cycle) / cycle);

        vec2 startP = tb_hash2(seed + cycleID * 13.7) * 2.0 - 1.0;
        startP.x *= x_WindowSize.x / x_WindowSize.y;
        vec2 endP = tb_hash2(seed * 3.7 + cycleID * 17.3) * 2.0 - 1.0;
        endP.x *= x_WindowSize.x / x_WindowSize.y;

        vec2 partPos = mix(startP, endP, phase);
        vec2 partDir = normalize(endP - startP);
        vec2 partPerp = vec2(-partDir.y, partDir.x);

        vec2 toP = p - partPos;
        float along = dot(toP, partDir);
        float perp = abs(dot(toP, partPerp));

        // Tachyon cone points FORWARD (faster than light in vacuum)
        if (along > 0.0) {
            float coneAngle = 0.6;
            float expectedPerp = along * tan(coneAngle);
            float coneDist = abs(perp - expectedPerp);
            float cone = exp(-coneDist * coneDist * 400.0);
            float distFade = exp(-along * 1.5);
            col += tb_pal(0.4) * cone * distFade * 0.7;
        }

        // Particle itself — bright UV glow
        float pd = length(p - partPos);
        col += vec3(0.85, 0.95, 1.0) * exp(-pd * pd * 2000.0) * 0.8;
        col += tb_pal(0.6) * exp(-pd * pd * 200.0) * 0.3;

        // Causality distortion — ghost image behind particle (time-reversed)
        vec2 ghostPos = partPos - partDir * 0.1;
        float gd = length(p - ghostPos);
        col += tb_pal(fract(seed * 0.03 + x_Time * 0.1)) * exp(-gd * gd * 1000.0) * 0.4;

        // Track from start to current position
        vec2 pa = p - startP;
        vec2 ba = partPos - startP;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        float trackD = length(pa - ba * h);
        float trackMask = exp(-trackD * trackD * 5000.0);
        col += vec3(0.9, 0.95, 1.0) * trackMask * 0.15;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
