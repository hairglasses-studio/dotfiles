// Gamma — power-law gamma correction
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 c = x_Texture(uv);

    const float gamma = 0.7;
    const float inv_gamma = 1.0 / gamma;

    c.r = pow(c.r, inv_gamma);
    c.g = pow(c.g, inv_gamma);
    c.b = pow(c.b, inv_gamma);

    _wShaderOut = c;
}
