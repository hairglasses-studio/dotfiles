// lib/sdf.glsl — Signed distance functions for cursor shaders
// Common SDF primitives used across cursor_smear, manga_slash, etc.

// Rectangle/Box SDF
float getSdfRectangle(in vec2 p, in vec2 xy, in vec2 b) {
    vec2 d = abs(p - xy) - b;
    return length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);
}

// Line segment distance helper (for polygon SDFs)
float seg(in vec2 p, in vec2 a, in vec2 b, inout float s, in float d) {
    vec2 pa = p - a, ba = b - a;
    float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
    vec2 dp = pa - ba * h;
    float nd = dot(dp, dp);
    if (nd < d) d = nd;
    bvec3 cond = bvec3(p.y >= a.y, p.y < b.y, ba.x * pa.y > ba.y * pa.x);
    if (all(cond) || all(not(cond))) s *= -1.0;
    return d;
}

// Parallelogram SDF (for smear trails, slash effects)
float getSdfParallelogram(in vec2 p, in vec2 v0, in vec2 v1, in vec2 v2, in vec2 v3) {
    float s = 1.0;
    float d = dot(p - v0, p - v0);
    d = seg(p, v0, v3, s, d);
    d = seg(p, v1, v0, s, d);
    d = seg(p, v2, v1, s, d);
    d = seg(p, v3, v2, s, d);
    return s * sqrt(d);
}

// Edge processing helper (for hexagon SDFs)
void processEdge(vec2 p, vec2 a, vec2 b, inout float minDist, inout float inside) {
    vec2 edge = b - a;
    vec2 pa = p - a;
    float lenSq = dot(edge, edge);
    float invLenSq = 1.0 / lenSq;
    float t = clamp(dot(pa, edge) * invLenSq, 0.0, 1.0);
    vec2 diff = pa - edge * t;
    minDist = min(minDist, dot(diff, diff));
    float cross_ = edge.x * pa.y - edge.y * pa.x;
    inside = min(inside, step(0.0, cross_));
}

// Hexagon SDF (for cursor_smear hex variant)
float sdHexagon(in vec2 p, in vec2 v0, in vec2 v1, in vec2 v2, in vec2 v3, in vec2 v4, in vec2 v5) {
    float minDist = 1e20;
    float inside = 1.0;
    processEdge(p, v0, v1, minDist, inside);
    processEdge(p, v1, v2, minDist, inside);
    processEdge(p, v2, v3, minDist, inside);
    processEdge(p, v3, v4, minDist, inside);
    processEdge(p, v4, v5, minDist, inside);
    processEdge(p, v5, v0, minDist, inside);
    float dist = sqrt(max(minDist, 0.0));
    return mix(dist, -dist, inside);
}
