// Shader attribution: 12jihan
// (Post-FX) — Simple bloom/glow effect

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos/x_WindowSize;
    
    // Base color from terminal
    vec3 color = x_Texture(uv).rgb;
    
    // Add bloom/glow
    float bloom = 0.05;
    vec3 glow = vec3(0.0);
    for(float i = 0.0; i < 4.0; i++) {
        vec2 offset = vec2(i) / x_WindowSize;
        glow += x_Texture(uv + offset).rgb;
        glow += x_Texture(uv - offset).rgb;
    }
    
    // Combine glow with original color
    color += glow * bloom;
    
    _wShaderOut = vec4(color, 1.0);
}
