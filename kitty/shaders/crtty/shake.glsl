#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;


vec2 norm(vec2 value, float isPosition) {
    return (value * 2.0 - (u_resolution * isPosition)) / u_resolution.y;
}

void main()
{
    vec2 uv = gl_FragCoord.xy / u_resolution;

    float shakeDuration = 0.5; // seconds
    float timeSinceShake = u_time - 0.0;

    vec2 shakeOffset = vec2(0.0);

    if (timeSinceShake >= 0.0 && timeSinceShake < shakeDuration) {
        float intensity = 0.0008; // Adjust shake intensity here

        float decay = 1.0 - (timeSinceShake / shakeDuration);

        shakeOffset.x = sin(u_time * 40.0) * intensity * decay;
        shakeOffset.y = cos(u_time * 35.0) * intensity * decay;
    }

    uv += shakeOffset;

    vec4 color = texture(u_input, uv);

    o_color = color;
}
