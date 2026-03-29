// lib/hash.glsl — Fast integer hash functions for GPU
// Uses uvec2 multiplication (5x faster than fract(sin(dot())) on Apple Silicon)

// Single float hash from vec2 (most common variant)
float hash(vec2 p) {
    uvec2 q = uvec2(p * 256.0) * uvec2(1597334673u, 3812015801u);
    uint n = (q.x ^ q.y) * 1597334673u;
    return float(n) / float(0xffffffffu);
}

// Single float hash from float
float hash(float p) {
    uint n = uint(p * 256.0) * 1597334673u;
    n = n ^ (n >> 16u);
    n *= 2246822519u;
    return float(n) / float(0xffffffffu);
}

// vec2 hash from vec2 (for 2D noise, Voronoi, etc.)
vec2 hash2(vec2 v) {
    uvec2 q = uvec2(v * mat2(127.1, 311.7, 269.5, 183.3) * 256.0);
    q *= uvec2(1597334673u, 3812015801u);
    q = q ^ (q >> 16u);
    return vec2(q) / float(0xffffffffu);
}

// vec4 hash from vec2 (for complex noise patterns)
vec4 hash4(vec2 v) {
    vec4 p = vec4(v * mat4x2(127.1, 311.7, 269.5, 183.3, 113.5, 271.9, 246.1, 124.6));
    uvec4 q = uvec4(p * 256.0) * uvec4(1597334673u, 3812015801u, 2798796415u, 1370589171u);
    q = q ^ (q >> 16u);
    return vec4(q) / float(0xffffffffu);
}

// vec4 hash from vec3 (for 3D volumetric patterns)
vec4 hash4(vec3 v) {
    vec4 p = vec4(dot(v, vec3(127.1, 311.7, 74.7)),
                  dot(v, vec3(269.5, 183.3, 246.1)),
                  dot(v, vec3(113.5, 271.9, 124.6)),
                  dot(v, vec3(271.9, 269.5, 311.7)));
    uvec4 q = uvec4(p * 256.0) * uvec4(1597334673u, 3812015801u, 2798796415u, 1370589171u);
    q = q ^ (q >> 16u);
    return vec4(q) / float(0xffffffffu);
}
