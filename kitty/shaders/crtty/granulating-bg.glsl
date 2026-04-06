#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

// Granulating Watercolor Wash Background
// Pigment settles into paper texture, creating a speckled, grainy look.
// Dense pigment clusters in paper valleys, bare paper on the peaks.

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

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec4 orig = texture(u_input, uv);

    float distToBg = distance(orig.rgb, vec4(0.0, 0.0, 0.0, 1.0));
    float isBg = 1.0 - smoothstep(0.0, 0.15, distToBg);

    if (isBg < 0.3) {
        o_color = orig;
        return;
    }

    // --- Organic edge shape (pixel-based so it works at any window size) ---
    float dTop    = u_resolution.y - gl_FragCoord.xy.y;
    float dBottom = gl_FragCoord.xy.y;
    float dLeft   = gl_FragCoord.xy.x;
    float dRight  = u_resolution.x - gl_FragCoord.xy.x;

    float nTop    = fbm(vec2(gl_FragCoord.xy.x * 0.008, 0.0));
    float nBottom = fbm(vec2(gl_FragCoord.xy.x * 0.008, 100.0));
    float nLeft   = fbm(vec2(0.0, gl_FragCoord.xy.y * 0.008));
    float nRight  = fbm(vec2(100.0, gl_FragCoord.xy.y * 0.008));

    float edgePx = 32.0;
    float roughPx = 20.0;

    float paintTop    = step(edgePx + nTop * roughPx, dTop);
    float paintBottom = step(edgePx + nBottom * roughPx, dBottom);
    float paintLeft   = step(edgePx + nLeft * roughPx, dLeft);
    float paintRight  = step(edgePx + nRight * roughPx, dRight);

    float inPaint = paintTop * paintBottom * paintLeft * paintRight;

    // --- Granulating wash ---
    // WASH_HUE is replaced by randomize-shader.sh, default 0.6
    float hue = WASH_HUE;
    vec3 pigment = 0.3 + 0.2 * cos(6.28318 * (hue + vec3(0.0, 0.33, 0.67)));

    // Paper texture at multiple scales — pigment settles in the valleys
    // Coarse paper grain (cold-pressed texture)
    float coarseGrain = vnoise(gl_FragCoord.xy * 0.06);
    // Medium texture
    float medGrain = vnoise(gl_FragCoord.xy * 0.12 + vec2(50.0));
    // Fine tooth
    float fineGrain = hash21(floor(gl_FragCoord.xy * 0.3));

    // Combined paper surface: 0 = valley (catches pigment), 1 = peak (bare)
    float paperSurface = coarseGrain * 0.5 + medGrain * 0.3 + fineGrain * 0.2;

    // Pigment density: more pigment in valleys, less on peaks
    // This creates the characteristic granulation speckle
    float pigmentDensity = 1.0 - smoothstep(0.25, 0.65, paperSurface);

    // Large-scale wash variation (where more/less paint was applied)
    vec2 p = gl_FragCoord.xy * 0.001 + vec2(hue * 100.0, hue * 73.0);
    float washAmount = fbm(p * 1.5 + vec2(5.0, 3.0));
    pigmentDensity *= smoothstep(0.2, 0.6, washAmount);

    // Color: dense areas are saturated pigment, sparse areas show paper
    vec3 washColor = mix(vec4(0.0, 0.0, 0.0, 1.0), pigment, pigmentDensity * 0.7);

    // Slight color separation — granulating pigments often split into
    // warm and cool components as they settle
    float separation = vnoise(gl_FragCoord.xy * 0.08 + vec2(30.0));
    vec3 warmPigment = pigment * vec3(1.15, 1.0, 0.85);
    vec3 coolPigment = pigment * vec3(0.85, 1.0, 1.15);
    vec3 splitColor = mix(warmPigment, coolPigment, separation);
    washColor = mix(washColor, mix(vec4(0.0, 0.0, 0.0, 1.0), splitColor, pigmentDensity * 0.7), 0.3);

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

    o_color = vec4(clamp(result, 0.0, 1.0), alpha);
}
