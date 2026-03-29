// Classic CMYK halftone / Ben-Day dots effect
precision highp float;

const float DOT_SIZE = 5.0;
const float DOT_SPACING = 10.0;
const float DOT_SMOOTHNESS = 0.4;

// Channel rotation angles in radians
const float ANGLE_C = 0.2618; // 15 degrees
const float ANGLE_M = 1.3090; // 75 degrees
const float ANGLE_Y = 0.0;    //  0 degrees
const float ANGLE_K = 0.7854; // 45 degrees

mat2 rot(float a) {
    float c = cos(a), s = sin(a);
    return mat2(c, -s, s, c);
}

float halftoneDot(vec2 fragCoord, float angle, float intensity) {
    vec2 p = rot(angle) * fragCoord;
    vec2 cell = mod(p, DOT_SPACING) - DOT_SPACING * 0.5;
    float d = length(cell);
    float radius = max(DOT_SIZE * sqrt(intensity) * 0.5, DOT_SMOOTHNESS);
    return 1.0 - smoothstep(radius - DOT_SMOOTHNESS, radius + DOT_SMOOTHNESS, d);
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 tex = texture(iChannel0, uv);
    vec3 rgb = tex.rgb;

    // Convert RGB to CMY + K
    float k = 1.0 - max(rgb.r, max(rgb.g, rgb.b));
    float invK = k < 1.0 ? 1.0 / (1.0 - k) : 0.0;
    float c = (1.0 - rgb.r - k) * invK;
    float m = (1.0 - rgb.g - k) * invK;
    float y = (1.0 - rgb.b - k) * invK;

    // Render each channel as a halftone dot grid
    float cDot = halftoneDot(fragCoord, ANGLE_C, c);
    float mDot = halftoneDot(fragCoord, ANGLE_M, m);
    float yDot = halftoneDot(fragCoord, ANGLE_Y, y);
    float kDot = halftoneDot(fragCoord, ANGLE_K, k);

    // Convert CMYK dots back to RGB
    vec3 result = vec3(1.0);
    result -= cDot * vec3(1.0, 0.0, 0.0); // cyan absorbs red
    result -= mDot * vec3(0.0, 1.0, 0.0); // magenta absorbs green
    result -= yDot * vec3(0.0, 0.0, 1.0); // yellow absorbs blue
    result -= kDot * vec3(1.0, 1.0, 1.0); // key absorbs all

    fragColor = vec4(clamp(result, 0.0, 1.0), tex.a);
}
