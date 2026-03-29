// Pixelation + Sobel edge overlay for an ASCII-art-like appearance
precision highp float;

const float GRID_SIZE = 6.0;       // pixel block size
const float EDGE_INTENSITY = 1.5;  // brightness of detected edges
const float EDGE_THRESHOLD = 0.08; // minimum edge strength to show
const float EDGE_MIX = 0.6;       // blend factor for edge overlay

// Sample luminance at offset (in grid-snapped space)
float lum(vec2 uv, vec2 offset) {
    vec2 texel = 1.0 / iResolution.xy;
    vec3 c = texture(iChannel0, uv + offset * texel * GRID_SIZE).rgb;
    return dot(c, vec3(0.299, 0.587, 0.114));
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;

    // Pixelate: snap to grid
    vec2 blockUV = floor(fragCoord / GRID_SIZE) * GRID_SIZE + GRID_SIZE * 0.5;
    vec2 pixUV = blockUV / iResolution.xy;
    vec4 block = texture(iChannel0, pixUV);

    // Sobel edge detection on the pixelated coordinates
    // Sample 3x3 neighborhood luminances
    float tl = lum(pixUV, vec2(-1, -1));
    float tc = lum(pixUV, vec2( 0, -1));
    float tr = lum(pixUV, vec2( 1, -1));
    float ml = lum(pixUV, vec2(-1,  0));
    float mr = lum(pixUV, vec2( 1,  0));
    float bl = lum(pixUV, vec2(-1,  1));
    float bc = lum(pixUV, vec2( 0,  1));
    float br = lum(pixUV, vec2( 1,  1));

    // Sobel operators
    float gx = -tl - 2.0*ml - bl + tr + 2.0*mr + br;
    float gy = -tl - 2.0*tc - tr + bl + 2.0*bc + br;
    float edge = sqrt(gx * gx + gy * gy);

    // Threshold and scale
    edge = smoothstep(EDGE_THRESHOLD, EDGE_THRESHOLD + 0.15, edge) * EDGE_INTENSITY;

    // Composite: pixelated base with bright edge overlay
    vec3 edgeColor = vec3(edge);
    vec3 result = mix(block.rgb, block.rgb + edgeColor, EDGE_MIX);

    fragColor = vec4(clamp(result, 0.0, 1.0), block.a);
}
