precision highp float;
// Brightness — proportional brightness adjustment
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 c = texture(iChannel0, uv);

    const float brightness_level = 0.7; // 0.0 = black, 1.0 = unchanged

    c.rgb *= brightness_level;

    fragColor = c;
}
