// Shader attribution: erniee
// (Background) — Matrix digital rain

// Hash-based 2D pseudo-noise (replaces Shadertoy iChannel1 noise texture)
vec2 hash22(vec2 p) {
    p = vec2(dot(p, vec2(127.1, 311.7)),
             dot(p, vec2(269.5, 183.3)));
    return fract(sin(p) * 43758.5453);
}

float text(vec2 _fc)
{
    vec2 uv = mod(x_PixelPos.xy, 16.)*.0625;
    vec2 block = x_PixelPos*.0625 - uv;
    uv = uv*.8+.1; // scale the letters up a bit
    uv += floor(hash22(block * 0.00390625 + x_Time*.002) * 16.); // randomize letters (procedural)
    uv *= .0625; // bring back into 0-1 range
    uv.x = -uv.x; // flip letters horizontally
    return x_Texture(uv).r;
}

vec3 rain(vec2 _fc)
{
    vec2 pos = x_PixelPos;
    pos.x -= mod(pos.x, 16.);
    //pos.y -= mod(pos.y, 16.);

    float offset=sin(pos.x*15.);
    float speed=cos(pos.x*3.)*.3+.7;

    float y = fract(pos.y/x_WindowSize.y + x_Time*speed + offset);
    return vec3(.1,1,.35) / (y*20.);
}

void windowShader(inout vec4 _wShaderOut)
{
    _wShaderOut = vec4(text(x_PixelPos)*rain(x_PixelPos),1.0);
}
