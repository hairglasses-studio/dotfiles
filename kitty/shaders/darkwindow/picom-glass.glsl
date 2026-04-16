// Glass Shatter — angular shard decomposition with chromatic aberration
// Category: Post-FX | Cost: MED | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

#define PI 3.14159265
#define TWO_PI 6.28318530

float random(vec2 st) {
    return fract(sin(dot(st, vec2(12.9898, 78.233))) * 43758.5453123);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec2 center = vec2(0.5);

    const float num_shards = 20.0;
    float animation_progress = 0.5 + 0.5 * cos(x_Time * 1.047);

    vec2 toCenter = uv - center;
    float angle = atan(toCenter.y, toCenter.x);
    if (angle < 0.0) angle += TWO_PI;
    float shard_id = floor(angle / (TWO_PI / num_shards));

    float shard_delay = random(vec2(shard_id, shard_id * 0.31));
    float individual_duration = 0.7;
    float ripple_spread = 1.0 - individual_duration;
    float stagger_start = shard_delay * ripple_spread;
    float stagger_end = stagger_start + individual_duration;
    float shard_progress = smoothstep(stagger_start, stagger_end, animation_progress);

    if (shard_progress < 0.001) {
        _wShaderOut = vec4(0.0);
        return;
    }

    float displacement = 1.0 - shard_progress;
    float max_translation = 0.15;
    float max_rotation = (PI / 180.0) * 25.0 * random(vec2(shard_id * 0.7, shard_id));

    float shard_angle = (shard_id + 0.5) * (TWO_PI / num_shards);
    vec2 radial_dir = vec2(cos(shard_angle), sin(shard_angle));

    vec2 translation = radial_dir * max_translation * displacement;
    float rotation = max_rotation * displacement;

    vec2 p = uv - translation;
    vec2 rel = p - center;
    float c = cos(rotation);
    float s = sin(rotation);
    vec2 rotated = vec2(c * rel.x - s * rel.y, s * rel.x + c * rel.y);
    vec2 sampleUv = rotated + center;

    if (sampleUv.x < 0.0 || sampleUv.x > 1.0 || sampleUv.y < 0.0 || sampleUv.y > 1.0) {
        _wShaderOut = vec4(0.0);
        return;
    }

    float ca_strength = 0.008 * displacement;
    vec2 ca_offset = radial_dir * ca_strength;

    vec4 out_color;
    out_color.r = x_Texture(sampleUv + ca_offset).r;
    out_color.g = x_Texture(sampleUv).g;
    out_color.b = x_Texture(sampleUv - ca_offset).b;
    out_color.a = x_Texture(sampleUv).a;

    out_color.a *= shard_progress;
    _wShaderOut = out_color;
}
