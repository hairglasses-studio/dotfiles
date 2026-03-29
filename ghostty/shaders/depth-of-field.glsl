// Tilt-shift depth-of-field — sharp center band, blurred edges
precision highp float;

const float BAND_CENTER = 0.5;    // vertical center of the sharp band (0-1)
const float BAND_WIDTH = 0.15;    // half-width of the sharp band
const float MAX_BLUR = 9.0;       // maximum blur radius in pixels
const int NUM_SAMPLES = 8;

// Fixed 8-point radial offsets (unit circle, evenly spaced)
const vec2 OFFSETS[8] = vec2[8](
    vec2(1.0, 0.0), vec2(0.7071, 0.7071),
    vec2(0.0, 1.0), vec2(-0.7071, 0.7071),
    vec2(-1.0, 0.0), vec2(-0.7071, -0.7071),
    vec2(0.0, -1.0), vec2(0.7071, -0.7071)
);

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec2 texel = 1.0 / iResolution.xy;

    // Distance from the sharp band, normalized
    float dist = abs(uv.y - BAND_CENTER) - BAND_WIDTH;
    float blur = clamp(dist / (0.5 - BAND_WIDTH), 0.0, 1.0);
    float radius = blur * blur * MAX_BLUR; // quadratic falloff

    vec4 col = texture(iChannel0, uv);

    // Always sample — blend factor handles the transition smoothly
    vec4 sum = col;
    for (int i = 0; i < NUM_SAMPLES; i++) {
        vec2 offset = OFFSETS[i] * radius * texel;
        sum += texture(iChannel0, uv + offset);
    }
    vec4 blurred = sum / float(NUM_SAMPLES + 1);

    // Smooth blend from sharp to blurred (no hard seam)
    fragColor = mix(col, blurred, blur);
}
