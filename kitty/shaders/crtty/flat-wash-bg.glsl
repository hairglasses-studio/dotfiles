#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

// Flat Watercolor Wash Background
// Uniform color with organic, irregular edges like real watercolor on paper.

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
    // Distance from each window edge in pixels
    float dTop    = u_resolution.y - gl_FragCoord.xy.y;
    float dBottom = gl_FragCoord.xy.y;
    float dLeft   = gl_FragCoord.xy.x;
    float dRight  = u_resolution.x - gl_FragCoord.xy.x;

    // Noise along each edge for irregular boundary
    float nTop    = fbm(vec2(gl_FragCoord.xy.x * 0.008, 0.0));
    float nBottom = fbm(vec2(gl_FragCoord.xy.x * 0.008, 100.0));
    float nLeft   = fbm(vec2(0.0, gl_FragCoord.xy.y * 0.008));
    float nRight  = fbm(vec2(100.0, gl_FragCoord.xy.y * 0.008));

    // Each edge: paint stops at a noisy pixel boundary
    float edgePx = 32.0;    // pixels from window edge where paint border sits
    float roughPx = 20.0;   // pixels of wobble

    float paintTop    = step(edgePx + nTop * roughPx, dTop);
    float paintBottom = step(edgePx + nBottom * roughPx, dBottom);
    float paintLeft   = step(edgePx + nLeft * roughPx, dLeft);
    float paintRight  = step(edgePx + nRight * roughPx, dRight);

    float inPaint = paintTop * paintBottom * paintLeft * paintRight;

    // --- Flat wash color ---
    // WASH_HUE is replaced by randomize-shader.sh, default 0.6
    float hue = WASH_HUE;
    vec3 washColor = 0.3 + 0.2 * cos(6.28318 * (hue + vec3(0.0, 0.33, 0.67)));

    // Very subtle pigment settling (keeps it flat but alive)
    float settle = fbm(gl_FragCoord.xy * 0.008);
    washColor *= 0.95 + 0.1 * settle;

    // Minimal paper grain
    washColor *= 0.97 + 0.06 * vnoise(gl_FragCoord.xy * 0.04);

    // --- Composite ---
    vec3 result = orig.rgb;
    float alpha = orig.a;

    if (isBg > 0.5) {
        if (inPaint > 0.5) {
            // Inside the wash — opaque with wash color
            result = mix(vec4(0.0, 0.0, 0.0, 1.0), washColor, 0.6);
            alpha = 0.9;
        } else {
            // Outside the wash — fully transparent
            alpha = 0.0;
        }
    }

    o_color = vec4(clamp(result, 0.0, 1.0), alpha);
}
