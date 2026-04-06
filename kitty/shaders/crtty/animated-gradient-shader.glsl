#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

// credits: https://github.com/unkn0wncode
void main()
{
    vec2 uv = gl_FragCoord.xy.xy / u_resolution;

    // Create seamless gradient animation
    float speed = 0.2;
    float gradientFactor = (uv.x + uv.y) / 2.0;

    // Use smoothstep and multiple sin waves for smoother transition
    float t = sin(u_time * speed) * 0.5 + 0.5;
    gradientFactor = smoothstep(0.0, 1.0, gradientFactor);

    // Create smooth circular animation
    float angle = u_time * speed;
    vec3 color1 = vec3(0.1, 0.1, 0.5);
    vec3 color2 = vec3(0.5, 0.1, 0.1);
    vec3 color3 = vec3(0.1, 0.5, 0.1);

    // Smooth interpolation between colors using multiple mix operations
    vec3 gradientStartColor = mix(
            mix(color1, color2, smoothstep(0.0, 1.0, sin(angle) * 0.5 + 0.5)),
            color3,
            smoothstep(0.0, 1.0, sin(angle + 2.0) * 0.5 + 0.5)
        );

    vec3 gradientEndColor = mix(
            mix(color2, color3, smoothstep(0.0, 1.0, sin(angle + 1.0) * 0.5 + 0.5)),
            color1,
            smoothstep(0.0, 1.0, sin(angle + 3.0) * 0.5 + 0.5)
        );

    vec3 gradientColor = mix(gradientStartColor, gradientEndColor, gradientFactor);

    vec4 terminalColor = texture(u_input, uv);
    float mask = 1.0 - step(0.5, dot(terminalColor.rgb, vec3(1.0)));
    vec3 blendedColor = mix(terminalColor.rgb, gradientColor, mask);

    o_color = vec4(blendedColor, terminalColor.a);
}
