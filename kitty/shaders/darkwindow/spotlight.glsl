// Shader attribution: m-ahdal
// (Post-FX) — Spotlight/vignette following cursor

// Created by Paul Robello


// Smooth oscillating function that varies over time
float smoothOscillation(float t, float frequency, float phase) {
    return sin(t * frequency + phase);
}

void windowShader(inout vec4 _wShaderOut) {
    // Resolution and UV coordinates
	vec2 uv = x_PixelPos.xy / x_WindowSize;

    // Used to fix distortion when calculating distance to circle center
    vec2 ratio = vec2(x_WindowSize.x / x_WindowSize.y, 1.0);

    // Get the texture from x_Texture
    vec4 texColor = x_Texture(uv);

    // Spotlight center moving based on a smooth random pattern
    float time = x_Time * 1.0; // Control speed of motion
    vec2 spotlightCenter = vec2(
        0.5 + 0.4 * smoothOscillation(time, 1.0, 0.0),  // Smooth X motion
        0.5 + 0.4 * smoothOscillation(time, 1.3, 3.14159) // Smooth Y motion with different frequency and phase
    );

    // Distance from the spotlight center
    float distanceToCenter = distance(uv * ratio, spotlightCenter);

    // Spotlight intensity based on distance
    float spotlightRadius = 0.25; // Spotlight radius
    float softness = 20.0;       // Spotlight edge softness. Higher values have sharper edge
    float spotlightIntensity = smoothstep(spotlightRadius, spotlightRadius - (1.0 / softness), distanceToCenter);

    // Ambient light level
    float ambientLight = 0.5; // Controls the minimum brightness across the texture

    // Combine the spotlight effect with the texture
    vec3 spotlightEffect = texColor.rgb * mix(vec3(ambientLight), vec3(1.0), spotlightIntensity);

    // Final color output
    _wShaderOut = vec4(spotlightEffect, texColor.a);
}
