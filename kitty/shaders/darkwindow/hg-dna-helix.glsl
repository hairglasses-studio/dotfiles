// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — DNA double helix — rotating base pairs, backbone strands, fluorescent glow

const int   BASE_PAIRS  = 28;
const float HELIX_RAD   = 0.18;
const float HELIX_LEN   = 1.8;
const float PITCH       = 0.35;    // height per base pair
const float ROT_SPEED   = 0.25;
const float INTENSITY   = 0.55;

vec3 dna_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);  // cyan — strand 1
    vec3 b = vec3(0.90, 0.25, 0.65);  // magenta — strand 2
    vec3 c = vec3(0.20, 0.95, 0.60);  // mint — base pair A
    vec3 d = vec3(0.96, 0.75, 0.30);  // amber — base pair T
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float dna_hash(float n) { return fract(sin(n * 127.1) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);
    float t = x_Time * ROT_SPEED;

    // Center strand axis — vertical, slight camera drift
    float centerX = 0.05 * sin(x_Time * 0.15);

    // Each base pair is at a specific y; two strands spiral around y-axis
    for (int i = 0; i < BASE_PAIRS; i++) {
        float fi = float(i);
        // Y position along helix (slowly scrolls)
        float y = mod(fi / float(BASE_PAIRS) * HELIX_LEN + x_Time * 0.07, HELIX_LEN) - HELIX_LEN * 0.5;
        // Rotation angle at this y
        float theta = y / PITCH * 6.28318 + t * 4.0;

        // Strand 1 position
        float x1 = centerX + HELIX_RAD * cos(theta);
        float z1 = HELIX_RAD * sin(theta);          // for depth-fade only
        // Strand 2 (opposite side)
        float x2 = centerX - HELIX_RAD * cos(theta);
        float z2 = -HELIX_RAD * sin(theta);

        // Depth → brightness (simulating 3D perspective)
        float depth1 = (z1 + HELIX_RAD) / (2.0 * HELIX_RAD);   // 0 far, 1 near
        float depth2 = (z2 + HELIX_RAD) / (2.0 * HELIX_RAD);

        vec2 s1 = vec2(x1, y);
        vec2 s2 = vec2(x2, y);

        float d1 = length(p - s1);
        float d2 = length(p - s2);

        // Strand backbone: small bright dot
        vec3 c1 = dna_pal(0.0);
        vec3 c2 = dna_pal(fract(0.5 + x_Time * 0.04));
        float size = 0.01 * (0.6 + depth1 * 0.7);
        float core1 = exp(-d1 * d1 / (size * size) * 3.0) * (0.5 + depth1 * 1.2);
        size = 0.01 * (0.6 + depth2 * 0.7);
        float core2 = exp(-d2 * d2 / (size * size) * 3.0) * (0.5 + depth2 * 1.2);

        col += c1 * core1;
        col += c2 * core2;

        // Base pair "rung": line segment between s1 and s2
        vec2 midRung = (s1 + s2) * 0.5;
        vec2 rungDir = s2 - s1;
        float rungLen = length(rungDir);
        vec2 rungUnit = rungDir / rungLen;
        vec2 rungPerp = vec2(-rungUnit.y, rungUnit.x);
        // Distance perpendicular to rung axis
        vec2 toP = p - s1;
        float along = clamp(dot(toP, rungUnit), 0.0, rungLen);
        vec2 onSeg = s1 + rungUnit * along;
        float rungD = length(p - onSeg);
        float rungMask = exp(-rungD * rungD * 1800.0);
        // Base pair color — alternating A/T (mint/amber) based on index parity
        float parity = mod(fi, 2.0);
        vec3 rungCol = mix(dna_pal(fract(0.25 + x_Time * 0.02)),
                           dna_pal(fract(0.75 + x_Time * 0.02)), parity);
        col += rungCol * rungMask * 0.5;
    }

    // Ambient halo around strand axis
    float axisDist = abs(p.x - centerX);
    float halo = exp(-axisDist * axisDist * 10.0) * 0.08;
    col += dna_pal(fract(x_Time * 0.03)) * halo;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
