precision highp float;
// Zoom — magnification / demagnification centered on terminal
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec2 center = vec2(0.5);

    const float scale = 0.5; // >1 = zoom in, <1 = zoom out

    // Displace UV around center
    vec2 newUv = (uv - center) * (1.0 / scale) + center;

    // Clamp to bounds — return black for out-of-range
    if (newUv.x > 1.0 || newUv.x < 0.0 || newUv.y > 1.0 || newUv.y < 0.0) {
        fragColor = vec4(0.0);
        return;
    }

    fragColor = texture(iChannel0, newUv);
}
