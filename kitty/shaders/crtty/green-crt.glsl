#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;

    // Sample the text texture (assuming white text on black background)
    float textAlpha = texture(u_input, uv).r;

    // Green phosphor color (slightly yellowish-green for realism)
    vec3 greenPhosphor = vec3(0.5, 1.0, 0.4);

    // Sharper inner glow effect (higher influence near text)
    float innerGlow = textAlpha * 1.2;  
    innerGlow += texture(u_input, uv + vec2(0.001, 0.0)).r * 0.6;
    innerGlow += texture(u_input, uv - vec2(0.001, 0.0)).r * 0.6;
    innerGlow += texture(u_input, uv + vec2(0.0, 0.001)).r * 0.6;
    innerGlow += texture(u_input, uv - vec2(0.0, 0.001)).r * 0.6;

    // Softer outer glow effect (minimal spread to avoid blurring)
    float outerGlow = textAlpha * 0.4;
    outerGlow += texture(u_input, uv + vec2(0.0025, 0.0)).r * 0.3;
    outerGlow += texture(u_input, uv - vec2(0.0025, 0.0)).r * 0.3;
    outerGlow += texture(u_input, uv + vec2(0.0, 0.0025)).r * 0.3;
    outerGlow += texture(u_input, uv - vec2(0.0, 0.0025)).r * 0.3;

    // Sharpened glow combination
    float glow = mix(innerGlow, outerGlow, 0.2);

    // Subtle green background glow
    float bgGlowIntensity = 0.4;  // Adjust for stronger/weaker effect
    float bgGlow = bgGlowIntensity * smoothstep(1.2, 0.2, length(uv - 0.5));
    vec3 backgroundColor = greenPhosphor * bgGlow * 0.3;  // Soft green glow

    // Subtle scanline effect
    float scanlineStrength = 0.98 + 0.02 * sin(uv.y * u_resolution.y * 3.1415 * 1.5);

    // Subtle flicker effect
    float flicker = 0.99 + 0.01 * sin(u_time * 100.0);

    // Final color composition
    vec3 color = backgroundColor + (greenPhosphor * glow * scanlineStrength * flicker);

    // Ensure sharp, readable text by making alpha stronger
    o_color = vec4(color, clamp(textAlpha + glow * 0.5, 0.0, 1.0));
}
