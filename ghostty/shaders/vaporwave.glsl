// Vaporwave Filter — Ghostty terminal overlay
// Applies a retro pink/cyan/purple color grade with scan artifacts

float rand(vec2 co) {
    return fract(sin(dot(co, vec2(12.9898, 78.233))) * 43758.5453);
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 term = texture(iChannel0, uv);

    // Extract luminance
    float luma = dot(term.rgb, vec3(0.299, 0.587, 0.114));

    // Vaporwave color palette: pink, cyan, purple gradient based on screen position
    vec3 pink   = vec3(1.0, 0.4, 0.7);
    vec3 cyan   = vec3(0.3, 0.9, 0.95);
    vec3 purple = vec3(0.6, 0.2, 0.8);

    // Blend palette by vertical position with slow animation
    float t = uv.y + sin(iTime * 0.3) * 0.1;
    vec3 tint;
    if (t < 0.5) {
        tint = mix(cyan, pink, t * 2.0);
    } else {
        tint = mix(pink, purple, (t - 0.5) * 2.0);
    }

    // Apply color grade: tint the terminal colors
    vec3 color = mix(term.rgb, term.rgb * tint, 0.4);

    // Boost saturation slightly
    float gray = dot(color, vec3(0.333));
    color = mix(vec3(gray), color, 1.3);

    // Subtle scan lines
    float scan = sin(uv.y * iResolution.y * 0.8) * 0.03;
    color -= scan;

    // Occasional horizontal noise line
    float noiseLine = step(0.998, rand(vec2(iTime * 0.1, floor(uv.y * 50.0))));
    color += noiseLine * 0.15;

    fragColor = vec4(color, term.a);
}
