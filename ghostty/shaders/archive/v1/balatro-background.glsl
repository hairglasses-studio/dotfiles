// Liquid swirl background — "Balatro Background" effect
// Ported from https://www.shadertoy.com/view/XXtBRr
precision highp float;

const float SPIN_SPEED   = 7.0;
const float SPIN_AMOUNT  = 0.8;
const float OFFSET       = 5.0;
const float COLOUR_1_HUE = 2.0;
const float COLOUR_2_HUE = 3.5;
const float CONTRAST      = 3.0;

vec3 hsl2rgb(float h, float s, float l) {
    vec3 rgb = clamp(abs(mod(h * 6.0 + vec3(0.0, 4.0, 2.0), 6.0) - 3.0) - 1.0, 0.0, 1.0);
    return l + s * (rgb - 0.5) * (1.0 - abs(2.0 * l - 1.0));
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = (fragCoord - 0.5 * iResolution.xy) / iResolution.y;
    float uv_len = length(uv);

    float speed = (SPIN_SPEED * 0.2);
    float new_pixel_angle = atan(uv.y, uv.x) + speed - SPIN_AMOUNT * uv_len;
    vec2 mid = (iResolution.xy * 0.5) / iResolution.y;
    uv = vec2(uv_len * cos(new_pixel_angle) + mid.x, uv_len * sin(new_pixel_angle) + mid.y) - mid;

    float c1p = (uv.x + OFFSET) * (uv.y + OFFSET);
    float c2p = (uv.x - OFFSET) * (uv.y - OFFSET);
    float c3p = (uv.x + OFFSET) * (uv.y - OFFSET);

    float t = iTime * speed;
    float paint1 = sin(c1p * 0.3 + t * 1.3) + sin(c1p * 0.1 - t * 0.7);
    float paint2 = sin(c2p * 0.3 - t * 1.1) + sin(c2p * 0.1 + t * 0.9);
    float paint3 = sin(c3p * 0.3 + t * 0.8) + sin(c3p * 0.1 - t * 1.2);

    float res = (paint1 + paint2 + paint3) / 3.0;
    res = pow(clamp(res * CONTRAST * 0.5 + 0.5, 0.0, 1.0), 1.5);

    vec3 col1 = hsl2rgb(COLOUR_1_HUE / 6.28318, 0.8, 0.35);
    vec3 col2 = hsl2rgb(COLOUR_2_HUE / 6.28318, 0.8, 0.55);
    vec3 col = mix(col1, col2, res);

    col = pow(col, vec3(0.8));

    // Blend with terminal text
    vec2 termUV = fragCoord / iResolution.xy;
    vec4 terminal = texture(iChannel0, termUV);
    float termLuma = dot(terminal.rgb, vec3(0.2126, 0.7152, 0.0722));
    float mask = 1.0 - smoothstep(0.05, 0.25, termLuma);
    fragColor = vec4(mix(terminal.rgb, col * 0.6, mask), terminal.a);
}
