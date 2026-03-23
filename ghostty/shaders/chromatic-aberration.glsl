// Chromatic Aberration — Ghostty terminal overlay
// Separates RGB channels with radial offset from screen center

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;

    // Distance from center for radial aberration
    vec2 center = uv - 0.5;
    float dist = length(center);

    // Aberration strength increases toward edges
    float strength = 0.004 * dist;

    // Offset direction (radial from center)
    vec2 dir = normalize(center + 0.0001);

    // Sample each channel at slightly different positions
    float r = texture(iChannel0, uv + dir * strength).r;
    float g = texture(iChannel0, uv).g;
    float b = texture(iChannel0, uv - dir * strength).b;
    float a = texture(iChannel0, uv).a;

    fragColor = vec4(r, g, b, a);
}
