precision highp float;
// Grayscale — CIELAB luminance-based desaturation filter
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 c = texture(iChannel0, uv);

    // CIELAB luma based on human tristimulus response
    float g = 0.2126 * c.r + 0.7152 * c.g + 0.0722 * c.b;

    fragColor = vec4(vec3(g), c.a);
}
