precision highp float;
// Pixelize — animated block mosaic that oscillates between sharp and pixelated
// Category: Post-FX | Cost: MED | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;

    float t = 0.5 + 0.5 * sin(iTime * 1.57);
    float block = mix(40.0, 1.0, t);

    vec2 blockUv = (floor(fragCoord / block) * block + block * 0.5) / iResolution.xy;
    blockUv = clamp(blockUv, vec2(0.0), vec2(1.0));

    fragColor = texture(iChannel0, blockUv);
}
