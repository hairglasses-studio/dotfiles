#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Cyber Grid — Tron-style infinite neon grid behind terminal text
// Category: Cyberpunk | Cost: LOW-MED | Source: original (Shadertoy research)
// Receding perspective grid with pulsing neon lines on black background.

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec4 term = texture(u_input, uv);
    float termLum = dot(term.rgb, vec3(0.299, 0.587, 0.114));

    // Grid coordinates — perspective projection
    vec2 p = (gl_FragCoord.xy - 0.5 * u_resolution) / u_resolution.y;
    float t = u_time * 0.4;

    // Ground plane: y maps to depth via perspective divide
    float horizon = -0.1;
    float depth = 0.3 / (p.y - horizon + 0.001);
    if (p.y < horizon) depth = 0.0;

    float gridX = p.x * depth;
    float gridZ = depth + t * 2.0;

    // Grid lines
    float lineX = smoothstep(0.02, 0.0, abs(fract(gridX) - 0.5) - 0.47);
    float lineZ = smoothstep(0.02, 0.0, abs(fract(gridZ * 0.5) - 0.5) - 0.47);
    float grid = max(lineX, lineZ);

    // Fade with distance
    float fade = exp(-depth * 0.08) * step(horizon, p.y);
    grid *= fade;

    // Pulse at grid intersections
    float pulse = sin(gridZ * 3.14159 + t * 4.0) * 0.5 + 0.5;
    float intersection = lineX * lineZ * pulse;

    // Snazzy palette
    vec3 cyan    = vec3(0.341, 0.780, 1.0);   // #57c7ff
    vec3 magenta = vec3(1.0, 0.416, 0.757);   // #ff6ac1

    vec3 gridColor = mix(cyan, magenta, intersection) * grid * 0.35;

    // Horizon glow
    float horizonGlow = exp(-abs(p.y - horizon) * 15.0) * 0.15;
    gridColor += cyan * horizonGlow;

    // Composite: grid behind text (masked by terminal luminance)
    vec3 result = term.rgb + gridColor * (1.0 - termLum * 0.9);

    o_color = vec4(result, term.a);
}
