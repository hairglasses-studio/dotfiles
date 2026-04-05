precision highp float;
// source: https://gist.github.com/qwerasd205/c3da6c610c8ffe17d6d2d3cc7068f17f
// credits: https://github.com/qwerasd205
// Golden spiral samples, [x, y, weight] weight is inverse of distance.
const vec3[12] samples = {
  vec3(-0.5819, 0.7289, 0.8536),
  vec3(0.3539, -1.3850, 0.5387),
  vec3(0.2009, 1.7814, 0.4277),
  vec3(-0.9350, -1.8733, 0.3658),
  vec3(1.7057, 1.6261, 0.3248),
  vec3(-2.3704, -1.0508, 0.2951),
  vec3(2.8016, 0.2081, 0.2723),
  vec3(-2.9026, 0.7979, 0.2541),
  vec3(2.6214, -1.8330, 0.2391),
  vec3(-1.9590, 2.7501, 0.2265),
  vec3(0.9715, -3.4098, 0.2157),
  vec3(0.2358, 3.6992, 0.2063)
  };

float lum(vec4 c) {
  return 0.299 * c.r + 0.587 * c.g + 0.114 * c.b;
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
  vec2 uv = fragCoord.xy / iResolution.xy;

  vec4 color = texture(iChannel0, uv);

  vec2 step = vec2(1.414) / iResolution.xy;

  for (int i = 0; i < 12; i++) {
    vec3 s = samples[i];
    vec4 c = texture(iChannel0, uv + s.xy * step);
    float l = lum(c);
    if (l > 0.2) {
      color += l * s.z * c * 0.4;
    }
  }

  fragColor = color;
}