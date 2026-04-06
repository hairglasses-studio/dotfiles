float text(vec2 x_PixelPos)
{
    vec2 uv = mod(x_PixelPos.xy, 16.)*.0625;
    vec2 block = x_PixelPos*.0625 - uv;
    uv = uv*.8+.1; // scale the letters up a bit
    uv += floor(texture(iChannel1, block/iChannelResolution[1].xy + x_Time*.002).xy * 16.); // randomize letters
    uv *= .0625; // bring back into 0-1 range
    uv.x = -uv.x; // flip letters horizontally
    return x_Texture(uv).r;
}

vec3 rain(vec2 x_PixelPos)
{
	x_PixelPos.x -= mod(x_PixelPos.x, 16.);
    //x_PixelPos.y -= mod(x_PixelPos.y, 16.);
    
    float offset=sin(x_PixelPos.x*15.);
    float speed=cos(x_PixelPos.x*3.)*.3+.7;
   
    float y = fract(x_PixelPos.y/x_WindowSize.y + x_Time*speed + offset);
    return vec3(.1,1,.35) / (y*20.);
}

void windowShader(inout vec4 color)
{
    color = vec4(text(x_PixelPos)*rain(x_PixelPos),1.0);
}
