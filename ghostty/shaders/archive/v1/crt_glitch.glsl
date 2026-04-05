precision highp float;
#define COLOR_BACK vec3(0.10, 0.10, 0.10)
#define COLOR_TRACE vec3(0.10, 1.10, 0.50)

float _hash(vec2 p) {
    uvec2 q = uvec2(p * 256.0) * uvec2(1597334673u, 3812015801u);
    uint n = (q.x ^ q.y) * 1597334673u;
    return float(n) / float(0xffffffffu);
}

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



void mainImage(out vec4 fragColor, in vec2 fragCoord)
{
    vec2 uv = fragCoord.xy / iResolution.xy;

    // Offset for chromatic aberration with a slight glitch
    float aberrationOffset = 0.0035;
    float glitchAmount = 0.015;

    float glitchX = mix(1.0, 1.0 + glitchAmount, _hash(uv));
    float glitchY = mix(1.0, 1.0 + glitchAmount, _hash(uv * 1.3));

    // Define the frequency and amplitude for palpitation
    float palpitationFrequency = 5.5; // Adjust this value for the frequency of palpitation
    float palpitationAmplitude = 0.05; // Adjust this value for the amplitude of palpitation

    // Calculate the scale based on a sine function to make it palpitate
    float time = iTime * palpitationFrequency;
    float palpitateScale = 0.10 + palpitationAmplitude * sin(time);

    // Draw iChannel0 multiple times bigger but behind
    float scale = 1.0 * palpitateScale; // Adjust this value to control the overall scale of iChannel0
    float r_background = texture(iChannel0, uv + vec2(aberrationOffset * glitchX, 0.001) * scale).r;
    float g_background = texture(iChannel0, uv).g;
    float b_background = texture(iChannel0, uv - vec2(aberrationOffset * glitchY, 0.0) * scale).b;

    // Blend the background color with the original color from iChannel0
    float alpha = 0.05; // Adjust this value for the desired alpha
    vec3 finalColor = mix(vec3(r_background, g_background, b_background), COLOR_BACK, alpha);

    // Apply lens distortion
    float distortionAmount = 0.06; // Adjust this value for the desired lens distortion
    uv = applyLensDistortion(uv, distortionAmount);

    // Add moving old TV scan lines
    float scanLineIntensity = 0.02; // Adjust this value for the intensity of scan lines
    float scanLineSpacing = 10.0; // Adjust this value for the spacing of scan lines
    time = iTime * 0.5; // Adjust the multiplier for scan line movement speed
    float scanLine = mod(floor((uv.y + iTime * 0.2) * iResolution.y / scanLineSpacing), 2.0);
    finalColor *= 1.0 - scanLineIntensity * scanLine;

    // Add film reel noise
    time = iTime * 0.3; // Adjust the multiplier for speed
    float noise = filmReelNoise(uv, time) * 0.055; // Adjust the multiplier for intensity

    // Add the film reel noise to the final color
    finalColor += vec3(noise);

    fragColor = vec4(finalColor, 1.0);
}