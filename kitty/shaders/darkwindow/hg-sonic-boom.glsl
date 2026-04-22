// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Sonic boom — Mach cone with shockwave rings + compression vapor

const int   RINGS = 6;
const float INTENSITY = 0.55;

vec3 sb_pal(float t) {
    vec3 cyan = vec3(0.20, 0.80, 0.95);
    vec3 white = vec3(0.95, 0.98, 1.00);
    vec3 mag  = vec3(0.90, 0.30, 0.60);
    vec3 gold = vec3(1.00, 0.70, 0.30);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, white, s);
    else if (s < 2.0) return mix(white, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, gold, s - 2.0);
    else              return mix(gold, cyan, s - 3.0);
}

float sb_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.03, 0.04, 0.08);

    // Projectile position — moves right
    float cycle = 3.0;
    float phase = mod(x_Time, cycle) / cycle;
    vec2 projPos = vec2(-0.8 + phase * 2.0, 0.0);

    // Mach cone opens behind (Mach angle ~30°)
    float coneAngle = 0.52;
    vec2 toP = p - projPos;
    float along = -toP.x;  // positive behind projectile
    float perpD = abs(toP.y);

    if (along > 0.0) {
        float expectedPerp = along * tan(coneAngle);
        float coneDist = abs(perpD - expectedPerp);
        float cone = exp(-coneDist * coneDist * 300.0);
        float distFade = exp(-along * 1.5);
        col += sb_pal(0.0) * cone * distFade * 0.6;
    }

    // Shock rings — expanding outward from projectile position
    for (int r = 0; r < RINGS; r++) {
        float fr = float(r);
        // Each ring fired at a different time
        float ringAge = fract(x_Time * 2.0 + fr * 0.15);
        float ringR = ringAge * 0.6;
        float ringD = abs(length(toP) - ringR);
        float ringWidth = 0.003 + ringAge * 0.03;
        float ring = exp(-ringD * ringD / (ringWidth * ringWidth) * 2.0);
        float ringFade = 1.0 - ringAge;
        col += sb_pal(fract(fr * 0.07 + x_Time * 0.05)) * ring * ringFade * 0.5;
    }

    // Projectile body + exhaust
    float projD = length(p - projPos);
    col += vec3(1.0, 0.95, 0.8) * exp(-projD * projD * 5000.0);
    col += vec3(1.0, 0.7, 0.3) * exp(-projD * projD * 400.0) * 0.3;

    // Exhaust trail
    if (along > 0.0 && along < 0.4 && perpD < 0.008) {
        float trailAlpha = exp(-along * 5.0);
        col += vec3(1.0, 0.7, 0.3) * trailAlpha * 0.7;
    }

    // Compression vapor cloud around projectile (subsonic precursor ring)
    float vaporD = length(p - projPos) - 0.1;
    if (vaporD > -0.03 && vaporD < 0.01) {
        float vapor = smoothstep(0.01, -0.03, vaporD);
        float turb = sb_hash(vec2(floor((p.x + phase) * 30.0), floor(p.y * 30.0)));
        col += vec3(0.8, 0.85, 0.95) * vapor * (0.4 + turb * 0.4) * 0.3;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
