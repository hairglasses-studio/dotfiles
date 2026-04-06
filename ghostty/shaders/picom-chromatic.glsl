precision highp float;
// Chromatic Aberration — RGB channel offset scaled by distance from center
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec2 center = vec2(0.5);

    // Pixel offsets for each color channel (in UV space)
    const float offset = 3.0;
    vec2 px = 1.0 / iResolution.xy;
    vec2 uvr = vec2(offset, 0.0) * px;
    vec2 uvg = vec2(0.0, offset) * px;
    vec2 uvb = vec2(-offset, 0.0) * px;

    // Scale effect by distance from center
    const float scaling_factor = 1.0;
    const float base_strength = 0.0;
    vec2 scale = base_strength + scaling_factor * (uv - center);

    uvr *= scale;
    uvg *= scale;
    uvb *= scale;

    // Fetch each channel at its offset position
    float r = texture(iChannel0, uv + uvr).r;
    float g = texture(iChannel0, uv + uvg).g;
    float b = texture(iChannel0, uv + uvb).b;
    float a = texture(iChannel0, uv).a;

    fragColor = vec4(r, g, b, a);
}
