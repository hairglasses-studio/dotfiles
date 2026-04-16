// Shader attribution: thijskok
// (CRT) — Green phosphor CRT with scanlines and flicker

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;

    // Sample the text texture (assuming white text on black background)
    float textAlpha = x_Texture(uv).r;

    // Green phosphor color (slightly yellowish-green for realism)
    vec3 greenPhosphor = vec3(0.5, 1.0, 0.4);

    // Sharper inner glow effect (higher influence near text)
    float innerGlow = textAlpha * 1.2;  
    innerGlow += x_Texture(uv + vec2(0.001, 0.0)).r * 0.6;
    innerGlow += x_Texture(uv - vec2(0.001, 0.0)).r * 0.6;
    innerGlow += x_Texture(uv + vec2(0.0, 0.001)).r * 0.6;
    innerGlow += x_Texture(uv - vec2(0.0, 0.001)).r * 0.6;

    // Softer outer glow effect (minimal spread to avoid blurring)
    float outerGlow = textAlpha * 0.4;
    outerGlow += x_Texture(uv + vec2(0.0025, 0.0)).r * 0.3;
    outerGlow += x_Texture(uv - vec2(0.0025, 0.0)).r * 0.3;
    outerGlow += x_Texture(uv + vec2(0.0, 0.0025)).r * 0.3;
    outerGlow += x_Texture(uv - vec2(0.0, 0.0025)).r * 0.3;

    // Sharpened glow combination
    float glow = mix(innerGlow, outerGlow, 0.2);

    // Subtle green background glow
    float bgGlowIntensity = 0.4;  // Adjust for stronger/weaker effect
    float bgGlow = bgGlowIntensity * smoothstep(1.2, 0.2, length(uv - 0.5));
    vec3 backgroundColor = greenPhosphor * bgGlow * 0.3;  // Soft green glow

    // Subtle scanline effect
    float scanlineStrength = 0.98 + 0.02 * sin(uv.y * x_WindowSize.y * 3.1415 * 1.5);

    // Subtle flicker effect
    float flicker = 0.99 + 0.01 * sin(x_Time * 100.0);

    // Final color composition
    vec3 color = backgroundColor + (greenPhosphor * glow * scanlineStrength * flicker);

    // Ensure sharp, readable text by making alpha stronger
    _wShaderOut = vec4(color, clamp(textAlpha + glow * 0.5, 0.0, 1.0));
}
