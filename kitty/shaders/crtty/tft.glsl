#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

/** Size of TFT "pixels" */
float resolution = 4.0;

/** Strength of effect */
float strength = 0.5;

void _scanline(inout vec3 color, vec2 uv)
{
    float scanline = step(1.2, mod(uv.y * u_resolution.y, resolution));
    float grille   = step(1.2, mod(uv.x * u_resolution.x, resolution));
    color *= max(1.0 - strength, scanline * grille);
}

void main()
{
    vec2 uv = gl_FragCoord.xy.xy / u_resolution;
    vec3 color = texture(u_input, uv).rgb;

    _scanline(color, uv);

    o_color.xyz = color;
    o_color.w   = 1.0;
}
