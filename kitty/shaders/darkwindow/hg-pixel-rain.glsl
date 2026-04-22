// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Pixel rain — blocky 8-bit-style falling raindrops with chromatic sub-channels

const float PIXEL_SIZE = 80.0;
const float INTENSITY = 0.5;

vec3 pr_pal(float t) {
    vec3 cyan = vec3(0.20, 0.90, 0.98);
    vec3 mint = vec3(0.25, 0.95, 0.55);
    vec3 mag  = vec3(0.95, 0.30, 0.65);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, mint, s);
    else if (s < 2.0) return mix(mint, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, vio, s - 2.0);
    else              return mix(vio, cyan, s - 3.0);
}

float pr_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    // Pixelate coordinates
    vec2 pixGrid = floor(uv * PIXEL_SIZE) / PIXEL_SIZE;

    // Per-column drop seed — determines color and speed
    float colX = floor(pixGrid.x * PIXEL_SIZE);
    float colSeed = pr_hash(vec2(colX, 0.0));

    // Drop head y-position in this column — moves down over time
    float dropSpeed = 0.3 + colSeed * 0.5;
    float headPhase = fract(colX * 0.137 + x_Time * dropSpeed);
    float pixelsPerScreen = PIXEL_SIZE;
    float headY = headPhase;

    // Current pixel's y in pixel units
    float pixY = pixGrid.y;

    // Distance from head (in normalized y)
    float distFromHead = headY - pixY;
    if (distFromHead < 0.0) distFromHead += 1.0;

    // Trail length
    float trailLen = 0.15 + colSeed * 0.25;
    float trail = smoothstep(trailLen, 0.0, distFromHead);

    // Probability of drop at this pixel — density threshold
    float dropMask = 0.0;
    if (distFromHead < trailLen) {
        // Random sparkle within the trail
        float sparkle = pr_hash(vec2(colX, floor(pixY * PIXEL_SIZE)));
        dropMask = step(0.3, sparkle) * trail;
    }

    vec3 col = vec3(0.0);

    // Base color for this column
    vec3 baseCol = pr_pal(fract(colSeed + x_Time * 0.03));

    // Chromatic sub-channels — draw R at colX-1/8 pixel, B at colX+1/8 pixel
    // This is a single-pixel shader, so fake it by treating the rainbow as per-frame modulation
    float headBright = 1.0 - distFromHead / trailLen;
    float headIntense = pow(headBright, 3.0);

    col = baseCol * dropMask;
    // Head bright white
    col += vec3(1.0, 1.0, 0.95) * headIntense * dropMask;

    // Ambient dim backdrop with very subtle static
    vec3 ambient = vec3(0.02, 0.03, 0.05);
    col += ambient;
    float staticNoise = pr_hash(pixGrid + vec2(0.0, floor(x_Time * 40.0)));
    if (staticNoise > 0.97) col += baseCol * 0.2;

    // Scanlines
    col *= 0.85 + 0.15 * sin(x_PixelPos.y * 2.0);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
