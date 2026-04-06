#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

float text(vec2 gl_FragCoord.xy)
{
    vec2 uv = mod(gl_FragCoord.xy.xy, 16.)*.0625;
    vec2 block = gl_FragCoord.xy*.0625 - uv;
    uv = uv*.8+.1; // scale the letters up a bit
    uv += floor(texture(iChannel1, block/iChannelResolution[1].xy + u_time*.002).xy * 16.); // randomize letters
    uv *= .0625; // bring back into 0-1 range
    uv.x = -uv.x; // flip letters horizontally
    return texture(u_input, uv).r;
}

vec3 rain(vec2 gl_FragCoord.xy)
{
	gl_FragCoord.xy.x -= mod(gl_FragCoord.xy.x, 16.);
    //gl_FragCoord.xy.y -= mod(gl_FragCoord.xy.y, 16.);
    
    float offset=sin(gl_FragCoord.xy.x*15.);
    float speed=cos(gl_FragCoord.xy.x*3.)*.3+.7;
   
    float y = fract(gl_FragCoord.xy.y/u_resolution.y + u_time*speed + offset);
    return vec3(.1,1,.35) / (y*20.);
}

void main()
{
    o_color = vec4(text(gl_FragCoord.xy)*rain(gl_FragCoord.xy),1.0);
}
