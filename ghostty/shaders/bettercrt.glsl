// Original shader collected from: https://www.shadertoy.com/view/WsVSzV
// Licensed under Shadertoy's default since the original creator didn't provide any license. (CC BY NC SA 3.0)
// Slight modifications were made to give a green-ish effect.

// This shader was modified by April Hall (arithefirst)
// Sourced from https://github.com/m-ahdal/ghostty-shaders/blob/main/retro-terminal.glsl
// Changes made:
// - Removed tint
// - Made the boundaries match ghostty's background color

// This shader was then modified by Constantine Chukhlomin (chuhlomin)
// Sourced from https://github.com/hackr-sh/ghostty-shaders/blob/main/bettercrt.glsl
// Changes made:
// - Made scanlines follow the curvature of the screen
// - Added rounded corners

float warp = 0.25; // simulate curvature of CRT monitor
float scan = 0.50; // simulate darkness between scanlines
float cornerRadius = 0.0125; // controls how rounded the corners appear

void mainImage(out vec4 fragColor, in vec2 fragCoord)
{
    // squared distance from center
    vec2 uv = fragCoord / iResolution.xy;
    vec2 dc = abs(0.5 - uv);
    dc *= dc;

    vec2 warpedUV = uv;
    warpedUV.x -= 0.5; warpedUV.x *= 1.0 + (dc.y * (0.3 * warp)); warpedUV.x += 0.5;
    warpedUV.y -= 0.5; warpedUV.y *= 1.0 + (dc.x * (0.4 * warp)); warpedUV.y += 0.5;

    // calculate the warped Y position for scanlines
    float warpedY = warpedUV.y * iResolution.y;

    // determine if we are drawing in a scanline using the warped position
    float apply = abs(sin(warpedY) * 0.25 * scan);

    // sample the texture
    vec3 color = texture(iChannel0, warpedUV).rgb;

    // calculate distance from corner for rounded effect
    vec2 fromCenter = abs(warpedUV - 0.5) * 2.0;
    float cornerDistance = length(max(fromCenter - vec2(1.0 - cornerRadius), 0.0)) / cornerRadius;

    // for areas outside the curved screen or beyond corner radius, make them black
    if(warpedUV.x < 0.0 || warpedUV.x > 1.0 || warpedUV.y < 0.0 || warpedUV.y > 1.0 || cornerDistance > 1.0) {
        color = vec3(0.0);
        apply = 0.0;
    }

    // mix the sampled color with the scanline intensity
    fragColor = vec4(mix(color, vec3(0.0), apply), 1.0);
}
