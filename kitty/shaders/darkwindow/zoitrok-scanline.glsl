// Shader attribution: zoitrok (https://github.com/zoitrok/ghostty-shaders)
// License: Unlicense (public domain)
// Ported to DarkWindow by hairglasses — thin scanline overlay, stacks cleanly on any base shader.
// (Post-FX) — Rolling scanline overlay with subtle horizontal jitter

#define SPEED       0.1
#define LINEHEIGHT  0.01

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    float s = step(mod(uv.y + x_Time * SPEED, 1.0), LINEHEIGHT) * 0.1;
    uv.x = uv.x + s * 0.02;
    vec4 col = x_Texture(uv);
    col = col * (1.0 + s * 3.0) + s * 0.5;
    _wShaderOut = col;
}
