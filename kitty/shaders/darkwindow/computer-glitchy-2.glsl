void windowShader(inout vec4 color) {
    vec2 uv = x_PixelPos/x_WindowSize;
    
    // Brutal digital glitch
    float glitchTime = floor(x_Time * 10.0) / 10.0;
    
    // Reduced full-screen glitch block
    float glitchBlock = step(0.99, fract(sin(glitchTime * 57.0)));
    
    // Aggressive digital distortions
    vec2 glitchUV = uv;
    
    // Random horizontal splits
    if (fract(sin(glitchTime * 23.0)) > 0.9) {
        float splitHeight = fract(sin(glitchTime * 17.0));
        float splitIntensity = step(abs(uv.y - splitHeight), 0.01) * 
                                step(fract(sin(glitchTime * 11.0)), 0.5);
        glitchUV.x += splitIntensity * 0.1 * sin(x_Time * 50.0);
    }
    
    // Channel shifting
    vec3 rChannel = x_Texture(glitchUV + vec2(0.005, 0.0)).r * vec3(1.0, 0.0, 0.0);
    vec3 gChannel = x_Texture(glitchUV).g * vec3(0.0, 1.0, 0.0);
    vec3 bChannel = x_Texture(glitchUV - vec2(0.005, 0.0)).b * vec3(0.0, 0.0, 1.0);
    
    // Digital noise
    float digitalNoise = fract(sin(dot(uv, vec2(12.9898, 78.233))) * 43758.5453123);
    float noiseThreshold = step(0.95, digitalNoise);
    
    // Subtle local glitches instead of full-screen
    vec3 localGlitch = vec3(digitalNoise * 0.1);
    
    // Harsh color manipulation
    vec3 glitchColor = rChannel + gChannel + bChannel;
    glitchColor += localGlitch * 0.2;
    
    // Harsh digital lines
    float digitalLines = step(fract(uv.y * 100.0 + x_Time * 10.0), 0.5);
    glitchColor *= 1.0 - digitalLines * 0.2;
    
    // Color quantization (limited color palette)
    glitchColor = floor(glitchColor * 4.0) / 4.0;
    
    color = vec4(glitchColor, 1.0);
}
