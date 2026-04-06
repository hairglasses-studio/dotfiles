precision highp float;
// Pixelize — animated block mosaic that oscillates between sharp and pixelated
// Category: Post-FX | Cost: MED | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders
// Original used picom opacity transitions; adapted to iTime-driven loop

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;

    // Animate: oscillate between pixelated (block=40) and sharp (block=1)
    // using a smooth sine wave over 4 seconds
    float t = 0.5 + 0.5 * sin(iTime * 1.57); // 0→1→0 over ~4s
    float block = mix(40.0, 1.0, t);

    // Snap to block grid center
    vec2 blockUv = (floor(fragCoord / block) * block + block * 0.5) / iResolution.xy;

    // Clamp to valid UV range
    blockUv = clamp(blockUv, vec2(0.0), vec2(1.0));

    fragColor = texture(iChannel0, blockUv);
}
