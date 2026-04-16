// Chromatic Aberration — RGB channel offset scaled by distance from center
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec2 center = vec2(0.5);

    // Pixel offsets for each color channel (in UV space)
    const float offset = 3.0;
    vec2 px = 1.0 / x_WindowSize;
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
    float r = x_Texture(uv + uvr).r;
    float g = x_Texture(uv + uvg).g;
    float b = x_Texture(uv + uvb).b;
    float a = x_Texture(uv).a;

    _wShaderOut = vec4(r, g, b, a);
}
