#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

// Based on https://www.shadertoy.com/view/ms3cWn
float map(float value, float min1, float max1, float min2, float max2) {
  return min2 + (value - min1) * (max2 - min2) / (max1 - min1);
}

void main()
{
    vec2 uv = gl_FragCoord.xy / u_resolution;
    float d = length(uv - 0.5) * 2.0;
    float t = d * d * 25.0 - u_time * 2.0;
    vec3 col = 0.5 + 0.5 * cos(t / 20.0 + uv.xyx + vec3(0.0,2.0,4.0));

	vec2 center = u_resolution * 0.5;
    float distCentre = distance(gl_FragCoord.xy.xy, center);
    float dCSin = sin(distCentre * 0.05);

    vec2 anim = vec2(map(sin(u_time),-1.0,1.0,0.0,u_resolution.x),map(sin(u_time*1.25),-1.0,1.0,0.0,u_resolution.y));
    float distMouse = distance(gl_FragCoord.xy.xy, anim);
    float dMSin = sin(distMouse * 0.05);

    float greycol = (((dMSin * dCSin) + 1.0) * 0.5);
    greycol = greycol * map(d, 0.0, 1.4142135623730951, 0.5, 0.0);

    vec4 terminalColor = texture(u_input, uv);
    vec3 blendedColor = mix(terminalColor.rgb, vec3(greycol * col.x, greycol * col.y, greycol * col.z), 0.25);

    o_color = vec4(blendedColor, terminalColor.a);
}
