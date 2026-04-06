#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

// Splatter Watercolor Wash Background
// Random droplets scattered across a light wash, like flicking a loaded brush.
// Multiple sizes from large drops to fine spray.

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

// Compute splatter droplets at a given grid scale
// cellSize: pixel size of each grid cell
// density: 0-1, chance a cell has a drop
// maxRadius: max drop radius in cell-fraction units
float splatDrops(vec2 gl_FragCoord.xy, float cellSize, float density, float maxRadius) {
    vec2 grid = gl_FragCoord.xy / cellSize;
    vec2 cell = floor(grid);
    vec2 local = fract(grid);

    float drop = 0.0;
    for (int y = -1; y <= 1; y++) {
        for (int x = -1; x <= 1; x++) {
            vec2 neighbor = cell + vec2(float(x), float(y));
            float exists = step(1.0 - density, hash21(neighbor + vec2(31.0)));
            vec2 dropPos = vec2(hash21(neighbor), hash21(neighbor + vec2(7.0)));
            float dropRadius = hash21(neighbor + vec2(13.0)) * maxRadius + maxRadius * 0.3;
            // Slight per-drop shape variation
            float wobble = 0.9 + 0.2 * hash21(neighbor + vec2(53.0));
            vec2 diff = vec2(float(x), float(y)) + dropPos - local;
            float d = length(diff);
            drop += exists * (1.0 - smoothstep(dropRadius * 0.7 * wobble, dropRadius * wobble, d));
        }
    }
    return clamp(drop, 0.0, 1.0);
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

    // --- Splatter: droplets on a light wash ---
    // WASH_HUE is replaced by randomize-shader.sh, default 0.6
    float hue = WASH_HUE;
    vec3 pigment = 0.3 + 0.2 * cos(6.28318 * (hue + vec3(0.0, 0.33, 0.67)));

    // Light base wash underneath the splatter
    vec2 p = gl_FragCoord.xy * 0.001 + vec2(hue * 100.0, hue * 73.0);
    float baseWash = fbm(p * 1.5 + vec2(5.0, 3.0));
    vec3 washColor = mix(vec4(0.0, 0.0, 0.0, 1.0), pigment, smoothstep(0.3, 0.6, baseWash) * 0.25);

    // Splatter at three scales: large drops, medium drops, fine spray
    float largeDrop = splatDrops(gl_FragCoord.xy, 80.0, 0.25, 0.25);
    float medDrop   = splatDrops(gl_FragCoord.xy, 40.0, 0.2, 0.22);
    float fineSpray = splatDrops(gl_FragCoord.xy, 18.0, 0.15, 0.18);

    // Combine all drop layers
    float allDrops = clamp(largeDrop + medDrop * 0.8 + fineSpray * 0.5, 0.0, 1.0);

    // Drops are pigmented; larger drops are slightly darker (more paint)
    vec3 dropColor = pigment * (0.85 + 0.15 * largeDrop);

    // Drops darken slightly at edges (pigment settles to rim)
    float edgeDarken = largeDrop * (1.0 - smoothstep(0.0, 0.8, largeDrop)) * 0.15;
    dropColor *= 1.0 - edgeDarken;

    // Mix drops onto the base wash
    washColor = mix(washColor, dropColor, allDrops * 0.7);

    // Very subtle pigment settling
    float settle = fbm(gl_FragCoord.xy * 0.008);
    washColor *= 0.95 + 0.1 * settle;

    // Minimal paper grain
    washColor *= 0.97 + 0.06 * vnoise(gl_FragCoord.xy * 0.04);

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
