
vec2 norm(vec2 value, float isPosition) {
    return (value * 2.0 - (x_WindowSize * isPosition)) / x_WindowSize.y;
}

void windowShader(inout vec4 color)
{
    vec2 uv = x_PixelPos / x_WindowSize;

    float shakeDuration = 0.5; // seconds
    float timeSinceShake = x_Time - x_Time;

    vec2 shakeOffset = vec2(0.0);

    if (timeSinceShake >= 0.0 && timeSinceShake < shakeDuration) {
        float intensity = 0.0008; // Adjust shake intensity here

        float decay = 1.0 - (timeSinceShake / shakeDuration);

        shakeOffset.x = sin(x_Time * 40.0) * intensity * decay;
        shakeOffset.y = cos(x_Time * 35.0) * intensity * decay;
    }

    uv += shakeOffset;

    vec4 color = x_Texture(uv);

    color = color;
}
