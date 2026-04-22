// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Rotating 4D tesseract — inner + outer cubes projected from 4D with neon edges

const int   EDGES        = 32;    // 32 edges in a tesseract
const float EDGE_WIDTH   = 0.005;
const float VERTEX_SIZE  = 0.012;
const float INTENSITY    = 0.55;

vec3 ts_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.90, 0.30, 0.70); // magenta
    vec3 c = vec3(0.55, 0.30, 0.98); // violet
    vec3 d = vec3(0.96, 0.85, 0.40); // gold
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// 16 tesseract vertices in 4D, each at (±1, ±1, ±1, ±1)
vec4 tessVertex(int i) {
    return vec4(
        (i & 1) == 0 ? -1.0 : 1.0,
        (i & 2) == 0 ? -1.0 : 1.0,
        (i & 4) == 0 ? -1.0 : 1.0,
        (i & 8) == 0 ? -1.0 : 1.0
    );
}

// Rotate in xy plane
vec4 rotXY(vec4 v, float a) {
    float cr = cos(a), sr = sin(a);
    return vec4(v.x * cr - v.y * sr, v.x * sr + v.y * cr, v.z, v.w);
}
vec4 rotXW(vec4 v, float a) {
    float cr = cos(a), sr = sin(a);
    return vec4(v.x * cr - v.w * sr, v.y, v.z, v.x * sr + v.w * cr);
}
vec4 rotYW(vec4 v, float a) {
    float cr = cos(a), sr = sin(a);
    return vec4(v.x, v.y * cr - v.w * sr, v.z, v.y * sr + v.w * cr);
}
vec4 rotZW(vec4 v, float a) {
    float cr = cos(a), sr = sin(a);
    return vec4(v.x, v.y, v.z * cr - v.w * sr, v.z * sr + v.w * cr);
}

// 4D → 3D projection (perspective with w-distance)
vec3 project4to3(vec4 v) {
    float wDist = 3.0;
    float w = 1.0 / (wDist - v.w);
    return v.xyz * w;
}

// 3D → 2D projection (simple perspective)
vec2 project3to2(vec3 v) {
    float zDist = 3.0;
    float z = 1.0 / (zDist - v.z);
    return v.xy * z;
}

// Distance from point p to segment [a,b] in 2D
float sdSegment2(vec2 p, vec2 a, vec2 b) {
    vec2 pa = p - a, ba = b - a;
    float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
    return length(pa - ba * h);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y * 2.0;

    // 4D rotations — combined for rich motion
    float t = x_Time;
    // Compute projected 2D vertices for all 16
    vec2 verts[16];
    for (int i = 0; i < 16; i++) {
        vec4 v = tessVertex(i);
        v = rotXY(v, t * 0.3);
        v = rotXW(v, t * 0.4);
        v = rotYW(v, t * 0.27);
        v = rotZW(v, t * 0.33);
        vec3 v3 = project4to3(v);
        verts[i] = project3to2(v3);
    }

    vec3 col = vec3(0.0);

    // Draw all edges — connect vertices whose IDs differ in exactly 1 bit
    for (int a = 0; a < 16; a++) {
        for (int b = a + 1; b < 16; b++) {
            int diff = a ^ b;
            // Bit count of diff should be 1
            if (diff != 1 && diff != 2 && diff != 4 && diff != 8) continue;
            float d = sdSegment2(p, verts[a], verts[b]);
            float edgeMask = smoothstep(EDGE_WIDTH, 0.0, d);
            // Color: inner cube vs outer, based on w-bit
            float outerLayer = float((a >> 3) & 1 + (b >> 3) & 1) * 0.5;
            vec3 ec = ts_pal(fract(outerLayer * 0.5 + x_Time * 0.05));
            col += ec * edgeMask * 0.8;
            // Glow
            col += ec * exp(-d * d * 10000.0) * 0.3;
        }
    }

    // Vertex dots
    for (int i = 0; i < 16; i++) {
        float d = length(p - verts[i]);
        float core = exp(-d * d / (VERTEX_SIZE * VERTEX_SIZE) * 3.0);
        float outerLayer = float((i >> 3) & 1);
        vec3 vc = ts_pal(fract(outerLayer * 0.5 + x_Time * 0.05));
        col += vc * core * 1.5;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
