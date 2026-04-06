#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

void main()
{
    // Normalize pixel coordinates
    vec2 uv = gl_FragCoord.xy / u_resolution;

    // Get the video frame as a texture (use u_input for the video texture)
    vec4 videoColor = texture(u_input, uv);

    // Animate the noise by adding time-based variation
    float time = u_time * 0.1;  // Scale time for slower animation
    float noise = fract(sin(uv.x * 12.9898 + uv.y * 78.233 + time) * 43758.5453123);

    // Darken the noise effect by scaling it down
    float darkenedNoise = noise * 0.1;

    // Set the alpha channel for transparency (adjust as needed)
    float alpha = 0.2;  // Adjust this value for more or less transparency

    // Mix the video frame with the noise, adjusting the alpha for transparency
    o_color = mix(videoColor, vec4(vec3(darkenedNoise), alpha), 0.5);
}
