#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

#define COLOR_BACK vec3(0.10, 0.10, 0.10)
#define COLOR_TRACE vec3(0.10, 1.10, 0.50)

// Function to apply lens distortion
vec2 applyLensDistortion(vec2 uv, float distortionAmount) {
    vec2 centeredUV = uv - 0.5;
    float dist = dot(centeredUV, centeredUV);
    vec2 distortedUV = uv + centeredUV * dist * distortionAmount;
    return distortedUV;
}

// Function to add film reel noise
float filmReelNoise(vec2 uv, float time) {
    float noise = sin(dot(uv + time, vec2(12.9898, 78.233))) * 43758.5453;
    return fract(noise);
}



void main()
{
    vec2 uv = gl_FragCoord.xy.xy / u_resolution;

    // Offset for chromatic aberration with a slight glitch
    float aberrationOffset = 0.0035;
    float glitchAmount = 0.015;

    float glitchX = mix(1.0, 1.0 + glitchAmount, fract(sin(dot(uv, vec2(12.9898, 78.233))) * 43758.5453));
    float glitchY = mix(1.0, 1.0 + glitchAmount, fract(sin(dot(uv, vec2(39.7424, 76.3456))) * 43758.5453));

    // Define the frequency and amplitude for palpitation
    float palpitationFrequency = 5.5; // Adjust this value for the frequency of palpitation
    float palpitationAmplitude = 0.05; // Adjust this value for the amplitude of palpitation

    // Calculate the scale based on a sine function to make it palpitate
    float time = u_time * palpitationFrequency;
    float palpitateScale = 0.10 + palpitationAmplitude * sin(time);

    // Draw u_input multiple times bigger but behind
    float scale = 1.0 * palpitateScale; // Adjust this value to control the overall scale of u_input
    float r_background = texture(u_input, uv + vec2(aberrationOffset * glitchX, 0.001) * scale).r;
    float g_background = texture(u_input, uv).g;
    float b_background = texture(u_input, uv - vec2(aberrationOffset * glitchY, 0.0) * scale).b;

    // Blend the background color with the original color from u_input
    float alpha = 0.05; // Adjust this value for the desired alpha
    vec3 finalColor = mix(vec3(r_background, g_background, b_background), COLOR_BACK, alpha);

    // Apply lens distortion
    float distortionAmount = 0.06; // Adjust this value for the desired lens distortion
    uv = applyLensDistortion(uv, distortionAmount);

    // Add moving old TV scan lines
    float scanLineIntensity = 0.02; // Adjust this value for the intensity of scan lines
    float scanLineSpacing = 10.0; // Adjust this value for the spacing of scan lines
    time = u_time * 0.5; // Adjust the multiplier for scan line movement speed
    float scanLine = mod(floor((uv.y + u_time * 0.2) * u_resolution.y / scanLineSpacing), 2.0);
    finalColor *= 1.0 - scanLineIntensity * scanLine;

    // Add film reel noise
    time = u_time * 0.3; // Adjust the multiplier for speed
    float noise = filmReelNoise(uv, time) * 0.055; // Adjust the multiplier for intensity

    // Add the film reel noise to the final color
    finalColor += vec3(noise);

    o_color = vec4(finalColor, 1.0);
}
