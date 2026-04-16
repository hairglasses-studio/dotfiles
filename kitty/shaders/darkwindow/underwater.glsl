// Shader attribution: alex-sherwin
// (Background) — Underwater caustics and light rays

// adapted by Alex Sherwin for Ghostty from https://www.shadertoy.com/view/lljGDt

#define BLACK_BLEND_THRESHOLD .4

float hash21(vec2 p) {
    p = fract(p * vec2(233.34, 851.73));
    p += dot(p, p + 23.45);
    return fract(p.x * p.y);
}

float rayStrength(vec2 raySource, vec2 rayRefDirection, vec2 coord, float seedA, float seedB, float speed)
{
    vec2 sourceToCoord = coord - raySource;
    float cosAngle = dot(normalize(sourceToCoord), rayRefDirection);
    
    // Add subtle dithering based on screen coordinates
    float dither = hash21(coord) * 0.015 - 0.0075;
    
    float ray = clamp(
        (0.45 + 0.15 * sin(cosAngle * seedA + x_Time * speed)) +
        (0.3 + 0.2 * cos(-cosAngle * seedB + x_Time * speed)) + dither,
        0.0, 1.0);
        
    // Smoothstep the distance falloff
    float distFade = smoothstep(0.0, x_WindowSize.x, x_WindowSize.x - length(sourceToCoord));
    return ray * mix(0.5, 1.0, distFade);
}

void windowShader(inout vec4 _wShaderOut)
{
	vec2 uv = x_PixelPos.xy / x_WindowSize;

	uv.y = 1.0 - uv.y;
	vec2 coord = vec2(x_PixelPos.x, x_WindowSize.y - x_PixelPos.y);
	
	// Set the parameters of the sun rays
	vec2 rayPos1 = vec2(x_WindowSize.x * 0.7, x_WindowSize.y * 1.1);
	vec2 rayRefDir1 = normalize(vec2(1.0, 0.116));
	float raySeedA1 = 36.2214;
	float raySeedB1 = 21.11349;
	float raySpeed1 = 1.1;
	
	vec2 rayPos2 = vec2(x_WindowSize.x * 0.8, x_WindowSize.y * 1.2);
	vec2 rayRefDir2 = normalize(vec2(1.0, -0.241));
	const float raySeedA2 = 22.39910;
	const float raySeedB2 = 18.0234;
	const float raySpeed2 = 0.9;
	
	// Calculate the colour of the sun rays on the current fragment
	vec4 rays1 =
		vec4(1.0, 1.0, 1.0, 0.0) *
		rayStrength(rayPos1, rayRefDir1, coord, raySeedA1, raySeedB1, raySpeed1);
	 
	vec4 rays2 =
		vec4(1.0, 1.0, 1.0, 0.0) *
		rayStrength(rayPos2, rayRefDir2, coord, raySeedA2, raySeedB2, raySpeed2);
	
	vec4 col = rays1 * 0.5 + rays2 * 0.4;
	
	// Attenuate brightness towards the bottom, simulating light-loss due to depth.
	// Give the whole thing a blue-green tinge as well.
	float brightness = 1.0 - (coord.y / x_WindowSize.y);
	col.r *= 0.05 + (brightness * 0.8);
	col.g *= 0.15 + (brightness * 0.6);
	col.b *= 0.3 + (brightness * 0.5);

  vec2 termUV = x_PixelPos.xy / x_WindowSize;
  vec4 terminalColor = x_Texture(termUV);

  float alpha = step(length(terminalColor.rgb), BLACK_BLEND_THRESHOLD);
  vec3 blendedColor = mix(terminalColor.rgb * 1.0, col.rgb * 0.3, alpha);
  
  _wShaderOut = vec4(blendedColor, terminalColor.a);
}
