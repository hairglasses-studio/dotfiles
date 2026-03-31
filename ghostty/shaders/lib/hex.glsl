// lib/hex.glsl — Hexagonal grid tiling utilities
// Axial hex coordinates, edge distance, neighbor lookup

// Returns the center (in world space) of the hexagon containing point p
// scale controls hex size (larger = bigger hexagons)
vec2 hexCenter(vec2 p, float scale) {
    p /= scale;
    vec2 a = mod(p, vec2(1.5, sqrt(3.0))) - vec2(0.75, sqrt(3.0) * 0.5);
    vec2 b = mod(p - vec2(0.75, sqrt(3.0) * 0.5), vec2(1.5, sqrt(3.0))) - vec2(0.75, sqrt(3.0) * 0.5);
    vec2 c = dot(a, a) < dot(b, b) ? a : b;
    return (p - c) * scale;
}

// Returns the distance from point p to the nearest hexagon edge
float hexEdgeDist(vec2 p, float scale) {
    vec2 center = hexCenter(p, scale);
    vec2 d = abs(p - center) / scale;
    // Hexagon SDF (flat-top orientation)
    float hex = max(d.x * 0.5 + d.y * (sqrt(3.0) / 2.0), d.x) - 0.5;
    return -hex * scale;
}

// Returns axial hex coordinates (integer grid ID) for the hex containing p
ivec2 hexAxial(vec2 p, float scale) {
    vec2 center = hexCenter(p, scale);
    return ivec2(floor(center / scale + 0.5));
}

// Manhattan distance between two hex cells in axial coordinates
float hexDist(ivec2 a, ivec2 b) {
    ivec2 d = abs(a - b);
    return float(max(max(d.x, d.y), abs(d.x + d.y)));
}
