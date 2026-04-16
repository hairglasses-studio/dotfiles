// Shader attribution: 0xhckr
// (CRT) — Retro green phosphor terminal

// Original shader collected from: https://www.shadertoy.com/view/WsVSzV
// Licensed under Shadertoy's default since the original creator didn't provide any license. (CC BY NC SA 3.0)
// Slight modifications were made to give a green-ish effect.

float warp = 0.25; // simulate curvature of CRT monitor
float scan = 0.50; // simulate darkness between scanlines

void windowShader(inout vec4 _wShaderOut)
{
    // squared distance from center
    vec2 uv = x_PixelPos / x_WindowSize;
    vec2 dc = abs(0.5 - uv);
    dc *= dc;
    
    // warp the fragment coordinates
    uv.x -= 0.5; uv.x *= 1.0 + (dc.y * (0.3 * warp)); uv.x += 0.5;
    uv.y -= 0.5; uv.y *= 1.0 + (dc.x * (0.4 * warp)); uv.y += 0.5;

    // sample inside boundaries, otherwise set to black
    if (uv.y > 1.0 || uv.x < 0.0 || uv.x > 1.0 || uv.y < 0.0)
        _wShaderOut = vec4(0.0, 0.0, 0.0, 1.0);
    else
    {
        // determine if we are drawing in a scanline
        float apply = abs(sin(x_PixelPos.y) * 0.5 * scan);
        
        // sample the texture and apply a teal tint
        vec3 color = x_Texture(uv).rgb;
        vec3 tealTint = vec3(0.0, 0.8, 0.6); // teal color (slightly more green than blue)

        // mix the sampled color with the teal tint based on scanline intensity
        _wShaderOut = vec4(mix(color * tealTint, vec3(0.0), apply), 1.0);
    }
}
