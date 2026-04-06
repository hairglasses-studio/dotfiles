// Matrix phosphor glow - clean and readable
// Adds subtle green glow around text without CRT distortion

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord.xy / iResolution.xy;
    vec4 color = texture(iChannel0, uv);

    // Glow radius in pixels
    float glowSize = 2.0;
    vec2 pixelSize = glowSize / iResolution.xy;

    // Sample surrounding pixels for glow
    vec4 glow = vec4(0.0);
    float samples = 0.0;

    for (float x = -2.0; x <= 2.0; x += 1.0) {
        for (float y = -2.0; y <= 2.0; y += 1.0) {
            if (x == 0.0 && y == 0.0) continue;
            float dist = length(vec2(x, y));
            float weight = 1.0 / (1.0 + dist);
            glow += texture(iChannel0, uv + vec2(x, y) * pixelSize) * weight;
            samples += weight;
        }
    }
    glow /= samples;

    // Calculate luminance of glow
    float glowLum = dot(glow.rgb, vec3(0.299, 0.587, 0.114));

    // Add subtle glow (Matrix green tinted)
    vec3 glowColor = vec3(0.0, 1.0, 0.3) * glowLum * 0.15;

    // Slight brightness boost to greens
    color.g *= 1.05;

    fragColor = vec4(color.rgb + glowColor, color.a);
}
