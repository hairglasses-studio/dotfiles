// Shader attribution: m-ahdal
// (Background) — Animated lava/magma shader

// INFO: This shader is a port of https://www.shadertoy.com/view/3sySRK

// INFO: Change these variables to create some variation in the animation
#define BLACK_BLEND_THRESHOLD .4 // This is controls the dim of the screen
#define COLOR_SPEED 0.1          // This controls the speed at which the colors change
#define MOVEMENT_SPEED 0.1       // This controls the speed at which the balls move

float opSmoothUnion( float d1, float d2, float k )
{
    float h = clamp( 0.5 + 0.5*(d2-d1)/k, 0.0, 1.0 );
    return mix( d2, d1, h ) - k*h*(1.0-h);
}

float sdSphere( vec3 p, float s )
{
  return length(p)-s;
} 

float map(vec3 p)
{
	float d = 2.0;
	for (int i = 0; i < 16; i++) {
		float fi = float(i);
		float time = x_Time * (fract(fi * 412.531 + 0.513) - 0.5) * 2.0;
		d = opSmoothUnion(
            sdSphere(p + sin(time*MOVEMENT_SPEED + fi * vec3(52.5126, 64.62744, 632.25)) * vec3(2.0, 2.0, 0.8), mix(0.5, 1.0, fract(fi * 412.531 + 0.5124))),
			d,
			0.4
		);
	}
	return d;
}

vec3 calcNormal( in vec3 p )
{
    const float h = 1e-5; // or some other value
    const vec2 k = vec2(1,-1);
    return normalize( k.xyy*map( p + k.xyy*h ) + 
                      k.yyx*map( p + k.yyx*h ) + 
                      k.yxy*map( p + k.yxy*h ) + 
                      k.xxx*map( p + k.xxx*h ) );
}

void windowShader(inout vec4 _wShaderOut)
{
    vec2 uv = x_PixelPos/x_WindowSize;
    
	vec3 rayOri = vec3((uv - 0.5) * vec2(x_WindowSize.x/x_WindowSize.y, 1.0) * 6.0, 3.0);
	vec3 rayDir = vec3(0.0, 0.0, -1.0);
	
	float depth = 0.0;
	vec3 p;
	
	for(int i = 0; i < 64; i++) {
		p = rayOri + rayDir * depth;
		float dist = map(p);
        depth += dist;
		if (dist < 1e-6) {
			break;
		}
	}
	
    depth = min(6.0, depth);
	vec3 n = calcNormal(p);
    float b = max(0.0, dot(n, vec3(0.577)));
    vec3 col = (0.5 + 0.5 * cos((b + x_Time*COLOR_SPEED * 3.0) + uv.xyx * 2.0 + vec3(0,2,4))) * (0.85 + b * 0.35);
    col *= exp( -depth * 0.15 );
	

    vec2 termUV = x_PixelPos.xy / x_WindowSize;
    vec4 terminalColor = x_Texture(termUV);

    float alpha = step(length(terminalColor.rgb), BLACK_BLEND_THRESHOLD);
    vec3 blendedColor = mix(terminalColor.rgb * 1.0, col.rgb * 0.3, alpha);

    _wShaderOut = vec4(blendedColor, terminalColor.a);

}
