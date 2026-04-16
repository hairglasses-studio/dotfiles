// Grayscale — CIELAB luminance-based desaturation filter
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 c = x_Texture(uv);

    // CIELAB luma based on human tristimulus response
    float g = 0.2126 * c.r + 0.7152 * c.g + 0.0722 * c.b;

    _wShaderOut = vec4(vec3(g), c.a);
}
