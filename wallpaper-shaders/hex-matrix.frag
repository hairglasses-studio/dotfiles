// Hex Matrix — hexagonal grid with falling data streams
// Cyberpunk matrix aesthetic in green (#3dffb5)
// Shadertoy-compatible: mainImage(out vec4, in vec2)

float hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
}

// Hexagonal grid — returns cell center and cell ID
vec4 hexGrid(vec2 uv, float scale) {
    vec2 s = vec2(1.0, 1.732);
    vec2 a = mod(uv * scale, s) - s * 0.5;
    vec2 b = mod(uv * scale - s * 0.5, s) - s * 0.5;
    vec2 gv = dot(a, a) < dot(b, b) ? a : b;
    vec2 id = uv * scale - gv;
    return vec4(gv, id);
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    uv.x *= iResolution.x / iResolution.y;

    vec4 hex = hexGrid(uv, 12.0);
    vec2 gv = hex.xy;
    vec2 id = hex.zw;

    // Per-cell timing offset
    float cellHash = hash(floor(id));
    float t = iTime * 0.4 + cellHash * 6.28;

    // Falling stream effect
    float stream = fract(t + cellHash * 3.0);
    stream = smoothstep(0.0, 0.3, stream) * smoothstep(1.0, 0.6, stream);

    // Only some cells are activeCell
    float activeCell = step(0.65, cellHash);
    stream *= activeCell;

    // Hex border glow
    float d = length(gv);
    float border = smoothstep(0.5, 0.48, d) - smoothstep(0.48, 0.35, d);

    // Colors
    vec3 green = vec3(0.239, 1.0, 0.710);   // #3dffb5
    vec3 bg = vec3(0.020, 0.027, 0.051);     // #05070d

    vec3 col = bg;
    col += green * stream * 0.12;
    col += green * border * 0.03;

    // Occasional bright flash cell
    float flash = step(0.92, sin(t * 2.0) * 0.5 + 0.5) * activeCell;
    col += green * flash * 0.08;

    fragColor = vec4(col, 1.0);
}
