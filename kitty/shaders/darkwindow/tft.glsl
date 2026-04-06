/** Size of TFT "pixels" */
float resolution = 4.0;

/** Strength of effect */
float strength = 0.5;

void _scanline(inout vec3 color, vec2 uv)
{
    float scanline = step(1.2, mod(uv.y * x_WindowSize.y, resolution));
    float grille   = step(1.2, mod(uv.x * x_WindowSize.x, resolution));
    color *= max(1.0 - strength, scanline * grille);
}

void windowShader(inout vec4 color)
{
    vec2 uv = x_PixelPos.xy / x_WindowSize;
    vec3 color = x_Texture(uv).rgb;

    _scanline(color, uv);

    color.xyz = color;
    color.w   = 1.0;
}
