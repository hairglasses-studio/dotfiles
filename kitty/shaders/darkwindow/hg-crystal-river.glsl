// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Flowing crystal river — tumbling angular shards carried by animated current

const int   SHARDS   = 48;
const float INTENSITY = 0.5;

vec3 cr_pal(float t) {
    vec3 a = vec3(0.20, 0.85, 0.98);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.25, 0.60);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float cr_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  cr_hash2(float n) { return vec2(cr_hash(n), cr_hash(n * 1.37 + 11.0)); }

// Triangle SDF (equilateral)
float sdTriangle(vec2 p, float size) {
    const float k = sqrt(3.0);
    p.x = abs(p.x) - size;
    p.y = p.y + size / k;
    if (p.x + k * p.y > 0.0) p = vec2(p.x - k * p.y, -k * p.x - p.y) / 2.0;
    p.x -= clamp(p.x, -2.0 * size, 0.0);
    return -length(p) * sign(p.y);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // River flow direction (left to right), with wavy current
    float currentY = 0.03 * sin(p.x * 4.0 + x_Time);

    // Base river color
    vec3 bg = mix(vec3(0.02, 0.06, 0.15), vec3(0.01, 0.03, 0.10), uv.y);
    vec3 col = bg;

    // Shards flow left to right
    for (int i = 0; i < SHARDS; i++) {
        float fi = float(i);
        float seed = fi * 3.71;
        float speed = 0.2 + cr_hash(seed) * 0.2;
        float phase = fract(x_Time * speed + cr_hash(seed * 5.1));

        // Starting X (left edge) → ending X (right edge)
        float flowRange = 2.5 * x_WindowSize.x / x_WindowSize.y;
        float shardX = -flowRange * 0.5 + phase * flowRange;
        // Vertical position
        float baseY = (cr_hash(seed * 7.3) - 0.5) * 1.0;
        baseY += currentY * 0.5;
        vec2 shardPos = vec2(shardX, baseY);

        // Shard rotation
        float rotSpeed = (cr_hash(seed * 11.3) - 0.5) * 4.0;
        float angle = seed + x_Time * rotSpeed;

        // Shard size
        float size = 0.01 + cr_hash(seed * 13.7) * 0.02;

        // Distance in rotated frame
        vec2 localP = p - shardPos;
        float ca = cos(angle), sa = sin(angle);
        vec2 rotP = mat2(ca, -sa, sa, ca) * localP;
        float d = sdTriangle(rotP, size);

        vec3 shardCol = cr_pal(fract(seed * 0.08 + x_Time * 0.05));

        // Fill shard
        if (d < 0.0) {
            // Inside shard — with subtle translucent blue base
            float fill = 0.6;
            col = mix(col, shardCol * 0.5, fill);
        }
        // Edge
        float edge = smoothstep(0.002, 0.0, abs(d));
        col += shardCol * edge * 0.8;
        // Outer glow
        col += shardCol * exp(-d * d * 1500.0) * 0.15;

        // Sparkle — random bright specular point
        if (d < -0.001 && cr_hash(seed + floor(x_Time * 3.0)) > 0.7) {
            vec2 sparkOff = cr_hash2(seed + floor(x_Time * 3.0)) * size - size * 0.5;
            float sparkD = length(rotP - sparkOff);
            col += vec3(1.0) * exp(-sparkD * sparkD * 50000.0) * 0.8;
        }
    }

    // Surface ripples (horizontal shine bands)
    float ripple = sin(p.x * 15.0 + x_Time * 2.0) * 0.05 + sin(p.y * 20.0 + x_Time * 3.0) * 0.03;
    col += cr_pal(fract(x_Time * 0.04)) * max(0.0, ripple) * 0.3;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
