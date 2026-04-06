precision highp float;
// Gamma — power-law gamma correction
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 c = texture(iChannel0, uv);

    const float gamma = 0.7;
    const float inv_gamma = 1.0 / gamma;

    c.r = pow(c.r, inv_gamma);
    c.g = pow(c.g, inv_gamma);
    c.b = pow(c.b, inv_gamma);

    fragColor = c;
}
