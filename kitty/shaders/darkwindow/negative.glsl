
void windowShader(inout vec4 color)
{
    vec2 uv = x_PixelPos/x_WindowSize;
    vec4 color = x_Texture(uv);
    color = vec4(1.0 - color.x, 1.0 - color.y, 1.0 - color.z, color.w);
}
