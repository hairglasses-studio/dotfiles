// Shader attribution: 0xhckr
// (Background) — Static color gradient background

// credits: https://github.com/unkn0wncode
void windowShader(inout vec4 _wShaderOut)
{
    // Normalize pixel coordinates (range from 0 to 1)
    vec2 uv = x_PixelPos.xy / x_WindowSize;

    // Create a gradient from bottom right to top left as a function (x + y)/2
    float gradientFactor = (uv.x + uv.y) / 2.0;

    // Define gradient colors (adjust to your preference)
    vec3 gradientStartColor = vec3(0.1, 0.1, 0.5); // Start color (e.g., dark blue)
    vec3 gradientEndColor = vec3(0.5, 0.1, 0.1); //      End color (e.g., dark red)

    vec3 gradientColor = mix(gradientStartColor, gradientEndColor, gradientFactor);

    // Sample the terminal screen texture including alpha channel
    vec4 terminalColor = x_Texture(uv);

    // Make a mask that is 1.0 where the terminal content is not black
    float mask = 1 - step(0.5, dot(terminalColor.rgb, vec3(1.0)));
    vec3 blendedColor = mix(terminalColor.rgb, gradientColor, mask);

    // Apply terminal's alpha to control overall opacity
    _wShaderOut = vec4(blendedColor, terminalColor.a);
}
