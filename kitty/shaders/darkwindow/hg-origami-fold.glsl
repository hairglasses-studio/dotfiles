// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Origami fold — tessellating triangular paper creases with lit edges, animated folding

const float TILE_SIZE = 0.12;
const float INTENSITY = 0.55;

vec3 of_pal(float t) {
    vec3 a = vec3(0.95, 0.55, 0.35);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.20, 0.85, 0.95);
    vec3 d = vec3(0.96, 0.80, 0.35);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float of_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Triangular tiling: 6 triangles per hexagonal cell
    // Easier approach: square tiling with diagonal split
    vec2 scaleP = p / TILE_SIZE;
    vec2 cellId = floor(scaleP);
    vec2 cellF = fract(scaleP);

    // Determine triangle within cell: top-left or bottom-right
    int triIdx = cellF.x + cellF.y > 1.0 ? 1 : 0;
    vec2 triF;
    if (triIdx == 0) triF = cellF;
    else             triF = vec2(1.0) - cellF;

    // Per-triangle hash for animation
    float triSeed = of_hash(cellId + vec2(triIdx, 0.0));

    // Fold intensity — animated
    float foldPhase = fract(x_Time * 0.2 + triSeed * 4.0);
    float fold = sin(foldPhase * 3.14);  // 0 → 1 → 0

    // Base triangle color — shaded by simulated normal
    // Simulate light direction
    vec2 lightDir = normalize(vec2(0.4, 0.7));
    // Normal direction varies with fold
    vec2 normal = vec2(
        sin(triSeed * 6.28 + foldPhase * 6.28) * fold,
        cos(triSeed * 6.28 + foldPhase * 6.28) * fold
    );
    float diffuse = max(0.2, dot(normal, lightDir) * 0.5 + 0.5);

    vec3 triCol = of_pal(fract(triSeed + x_Time * 0.03)) * diffuse;

    // Edge highlights — crease lines
    // Compute distance to triangle edges
    float edgeD = min(
        min(triF.x, triF.y),
        abs(triF.x + triF.y - 1.0) * 0.7071  // hypotenuse
    );
    float edge = smoothstep(0.04, 0.01, edgeD);

    // Brighter edge on fold crease
    vec3 edgeCol = of_pal(fract(triSeed + 0.3 + x_Time * 0.04));
    // The fold-crease edge animates brighter as fold increases
    float creaseGlow = edge * (0.5 + fold * 0.8);

    vec3 col = triCol * 0.5 + edgeCol * creaseGlow;

    // Global subtle shimmer
    col *= 0.9 + 0.1 * sin(x_Time * 0.5 + triSeed * 3.0);

    // Soft vignette
    col *= 1.0 - length(p) * 0.15;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility * 0.9);

    _wShaderOut = vec4(result, 1.0);
}
