#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

void main() {
    vec2 uv = gl_FragCoord.xy/u_resolution;
    
    // Base color from terminal
    vec3 color = texture(u_input, uv).rgb;
    
    // Add bloom/glow
    float bloom = 0.05;
    vec3 glow = vec3(0.0);
    for(float i = 0.0; i < 4.0; i++) {
        vec2 offset = vec2(i) / u_resolution;
        glow += texture(u_input, uv + offset).rgb;
        glow += texture(u_input, uv - offset).rgb;
    }
    
    // Combine glow with original color
    color += glow * bloom;
    
    o_color = vec4(color, 1.0);
}
