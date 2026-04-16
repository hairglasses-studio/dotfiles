// Brightness — proportional brightness adjustment
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 c = x_Texture(uv);

    const float brightness_level = 0.7; // 0.0 = black, 1.0 = unchanged

    c.rgb *= brightness_level;

    _wShaderOut = c;
}
