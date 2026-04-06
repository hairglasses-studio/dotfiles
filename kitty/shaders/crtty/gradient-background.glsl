#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

// credits: https://github.com/unkn0wncode
void main()
{
    // Normalize pixel coordinates (range from 0 to 1)
    vec2 uv = gl_FragCoord.xy.xy / u_resolution;

    // Create a gradient from bottom right to top left as a function (x + y)/2
    float gradientFactor = (uv.x + uv.y) / 2.0;

    // Define gradient colors (adjust to your preference)
    vec3 gradientStartColor = vec3(0.1, 0.1, 0.5); // Start color (e.g., dark blue)
    vec3 gradientEndColor = vec3(0.5, 0.1, 0.1); //      End color (e.g., dark red)

    vec3 gradientColor = mix(gradientStartColor, gradientEndColor, gradientFactor);

    // Sample the terminal screen texture including alpha channel
    vec4 terminalColor = texture(u_input, uv);

    // Make a mask that is 1.0 where the terminal content is not black
    float mask = 1 - step(0.5, dot(terminalColor.rgb, vec3(1.0)));
    vec3 blendedColor = mix(terminalColor.rgb, gradientColor, mask);

    // Apply terminal's alpha to control overall opacity
    o_color = vec4(blendedColor, terminalColor.a);
}
