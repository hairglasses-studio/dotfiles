// Film Grain Pro — Luminance-adaptive cinematic grain
// Category: Post-FX | Cost: LOW | Source: original (Shadertoy research)
// Applies organic grain that is stronger in midtones, weaker in blacks/whites.
// Perfect daily driver — subtle enough for extended coding sessions.

float hash12(vec2 p) {
    vec3 p3 = fract(vec3(p.xyx) * 0.1031);
    p3 += dot(p3, p3.yzx + 33.33);
    return fract((p3.x + p3.y) * p3.z);
}

void windowShader(inout vec4 color) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 term = x_Texture(uv);

    float lum = dot(term.rgb, vec3(0.299, 0.587, 0.114));

    // Luminance-adaptive intensity: strongest in midtones (0.3-0.7)
    float midtoneMask = smoothstep(0.0, 0.3, lum) * smoothstep(1.0, 0.7, lum);
    float intensity = 0.04 + 0.06 * midtoneMask;

    // Multi-octave grain for organic texture
    float grain1 = hash12(x_PixelPos + fract(x_Time * 43.17) * 1000.0) - 0.5;
    float grain2 = hash12(x_PixelPos * 2.0 + fract(x_Time * 71.23) * 1000.0) - 0.5;
    float grain = grain1 * 0.7 + grain2 * 0.3;

    // Slight warm/cool color variation in grain
    vec3 grainColor = vec3(
        grain + 0.003,   // slightly warm reds
        grain,
        grain - 0.003    // slightly cool blues
    );

    vec3 result = term.rgb + grainColor * intensity;

    color = vec4(result, term.a);
}
