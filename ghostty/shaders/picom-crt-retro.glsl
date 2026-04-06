precision highp float;
// CRT Retro — scanlines, spherical curvature, color distortion, edge shadow
// Category: CRT | Cost: MED | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

#define PI 3.1415926538

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec2 res = iResolution.xy;
    vec2 center = vec2(0.5);

    const float sc_freq = 0.2;
    const float sc_intensity = 0.6;
    const bool grid = false;
    const float distortion_offset = 2.0;
    const int downscale_factor = 2;
    const float sph_distance = 500.0;
    const float curvature = 1.5;
    const float shadow_cutoff = 1.0;
    const int shadow_intensity = 1;

    float radius = res.x / curvature;

    vec2 offset = uv - center;
    vec2 scaled = offset * res;

    float z = sqrt(max(0.0, (radius + sph_distance) * (radius + sph_distance) -
                   scaled.x * scaled.x - scaled.y * scaled.y));
    vec2 curved = scaled * ((radius + sph_distance) / z);
    vec2 curvedUv = curved / res + center;

    if (curvedUv.x >= 1.0 || curvedUv.x <= 0.0 ||
        curvedUv.y >= 1.0 || curvedUv.y <= 0.0) {
        fragColor = vec4(0.0, 0.0, 0.0, 1.0);
        return;
    }

    vec2 px = 1.0 / res;

    vec4 c = vec4(0.0);
    if (downscale_factor < 2) {
        c = texture(iChannel0, curvedUv);
    } else {
        ivec2 pixPos = ivec2(curvedUv * res);
        ivec2 blockRel = pixPos % downscale_factor;
        for (int j = 0; j < downscale_factor; j++) {
            for (int i = 0; i < downscale_factor; i++) {
                vec2 sUv = (vec2(pixPos - blockRel + ivec2(i, j)) + 0.5) / res;
                if (sUv.x >= 0.0 && sUv.x <= 1.0 && sUv.y >= 0.0 && sUv.y <= 1.0)
                    c += texture(iChannel0, sUv);
            }
        }
        c /= float(downscale_factor * downscale_factor);
    }

    vec4 c_right = texture(iChannel0, curvedUv + vec2(distortion_offset * px.x, 0.0));
    vec4 c_left  = texture(iChannel0, curvedUv - vec2(distortion_offset * px.x, 0.0));
    c = vec4(c_left.r, c.g, c_right.b, c.a);

    float scanY = sin(2.0 * PI * sc_freq * fragCoord.y) / (2.0 / sc_intensity) + 1.0 - sc_intensity / 2.0;
    c.rgb *= scanY;

    if (grid) {
        float scanX = sin(2.0 * PI * sc_freq * fragCoord.x) / (2.0 / sc_intensity) + 1.0 - sc_intensity / 2.0;
        c.rgb *= scanX;
    }

    if (shadow_intensity > 0) {
        vec2 distFromCenter = abs(curvedUv - center) / center;
        float brightness = 1.0;
        brightness *= -pow(distFromCenter.y * shadow_cutoff, float((5 / shadow_intensity) * 2)) + 1.0;
        brightness *= -pow(distFromCenter.x * shadow_cutoff, float((5 / shadow_intensity) * 2)) + 1.0;
        c.rgb *= max(brightness, 0.0);
    }

    fragColor = c;
}
