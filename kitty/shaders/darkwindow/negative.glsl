// Shader attribution: m-ahdal
// (Post-FX) — Color negative/inversion filter


void windowShader(inout vec4 _wShaderOut)
{
    vec2 uv = x_PixelPos/x_WindowSize;
    vec4 color = x_Texture(uv);
    _wShaderOut = vec4(1.0 - color.x, 1.0 - color.y, 1.0 - color.z, color.w);
}
