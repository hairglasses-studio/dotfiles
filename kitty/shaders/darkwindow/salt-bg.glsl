#define WASH_HUE 0.6
// Shader attribution: JRMeyer
// (Watercolor) — Salt texture on watercolor wash

// Salt Texture Watercolor Wash Background
// Salt sprinkled on wet paint disrupts the wash into a fine speckled texture.
// Paper peeks through where crystals absorbed pigment; pigment concentrates between.

float hash21(vec2 p) {
    p = fract(p * vec2(234.34, 435.345));
    p += dot(p, p + 34.23);
    return fract(p.x * p.y);
}

float vnoise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    f = f * f * (3.0 - 2.0 * f);
    float a = hash21(i);
    float b = hash21(i + vec2(1.0, 0.0));
    float c = hash21(i + vec2(0.0, 1.0));
    float d = hash21(i + vec2(1.0, 1.0));
    return mix(mix(a, b, f.x), mix(c, d, f.x), f.y);
}

float fbm(vec2 p) {
    float s = 0.0, a = 0.5;
    for (int i = 0; i < 5; i++) { s += vnoise(p) * a; p *= 2.0; a *= 0.5; }
    return s;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 orig = x_Texture(uv);

    float distToBg = distance(orig.rgb, vec3(0.0, 0.0, 0.0));
    float isBg = 1.0 - smoothstep(0.0, 0.15, distToBg);

    if (isBg < 0.3) {
        _wShaderOut = orig;
        return;
    }

    // --- Organic edge shape (pixel-based so it works at any window size) ---
    float dTop    = x_WindowSize.y - x_PixelPos.y;
    float dBottom = x_PixelPos.y;
    float dLeft   = x_PixelPos.x;
    float dRight  = x_WindowSize.x - x_PixelPos.x;

    float nTop    = fbm(vec2(x_PixelPos.x * 0.008, 0.0));
    float nBottom = fbm(vec2(x_PixelPos.x * 0.008, 100.0));
    float nLeft   = fbm(vec2(0.0, x_PixelPos.y * 0.008));
    float nRight  = fbm(vec2(100.0, x_PixelPos.y * 0.008));

    float edgePx = 32.0;
    float roughPx = 20.0;

    float paintTop    = step(edgePx + nTop * roughPx, dTop);
    float paintBottom = step(edgePx + nBottom * roughPx, dBottom);
    float paintLeft   = step(edgePx + nLeft * roughPx, dLeft);
    float paintRight  = step(edgePx + nRight * roughPx, dRight);

    float inPaint = paintTop * paintBottom * paintLeft * paintRight;

    // --- Salt texture: fine speckled disruption of the wash ---
    // WASH_HUE is replaced by randomize-shader.sh, default 0.6
    float hue = WASH_HUE;
    vec3 pigment = 0.3 + 0.2 * cos(6.28318 * (hue + vec3(0.0, 0.33, 0.67)));

    // fbm at fine scale: 5 octaves create natural, organic texture
    // that doesn't look blobby or pixelated
    float saltTexture = fbm(x_PixelPos * 0.015);

    // Large-scale variation: where more/less salt was sprinkled
    vec2 p = x_PixelPos * 0.001 + vec2(hue * 100.0, hue * 73.0);
    float saltDensity = fbm(p * 2.5 + vec2(8.0, 3.0));

    // Combined: fine texture modulated by area variation
    float pigmentPresence = saltTexture * 0.6 + saltDensity * 0.4;

    // Soft mapping to color — wide smoothstep avoids hard camo edges
    vec3 washColor = mix(vec3(0.0, 0.0, 0.0), pigment,
                         smoothstep(0.2, 0.65, pigmentPresence) * 0.75);

    // Minimal paper grain
    washColor *= 0.97 + 0.06 * vnoise(x_PixelPos * 0.04);

    // --- Composite ---
    vec3 result = orig.rgb;
    float alpha = orig.a;

    if (isBg > 0.5) {
        if (inPaint > 0.5) {
            result = washColor;
            alpha = 0.9;
        } else {
            alpha = 0.0;
        }
    }

    _wShaderOut = vec4(clamp(result, 0.0, 1.0), alpha);
}
