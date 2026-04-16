// rain-on-glass.glsl — Raindrops sliding on glass with neon bokeh refraction
// Category: Cyberpunk | Cost: HIGH | Source: original

// --- Hash functions ---
float _hash(vec2 p) {
    uvec2 q = uvec2(p * 256.0) * uvec2(1597334673u, 3812015801u);
    uint n = (q.x ^ q.y) * 1597334673u;
    return float(n) / float(0xffffffffu);
}
vec2 _hash2(vec2 v) {
    uvec2 q = uvec2(v * mat2(127.1, 311.7, 269.5, 183.3) * 256.0);
    q *= uvec2(1597334673u, 3812015801u);
    q = q ^ (q >> 16u);
    return vec2(q) / float(0xffffffffu);
}

float termLuminance(vec3 rgb) {
    return dot(rgb, vec3(0.2126, 0.7152, 0.0722));
}
float termMask(vec3 termColor) {
    return 1.0 - smoothstep(0.05, 0.25, termLuminance(termColor));
}

// Value noise for mist
float vnoise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(_hash(i), _hash(i + vec2(1.0, 0.0)), u.x),
               mix(_hash(i + vec2(0.0, 1.0)), _hash(i + vec2(1.0, 1.0)), u.x), u.y);
}

float fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < 5; i++) {
        v += a * vnoise(p);
        p = rot * p * 2.0 + 0.1;
        a *= 0.5;
    }
    return v;
}

// Single raindrop layer
// Returns: xy = refraction normal, z = drop alpha
vec3 rainLayer(vec2 uv, float t, float scale, float speed) {
    vec2 aspect = vec2(x_WindowSize.x / x_WindowSize.y, 1.0);
    vec2 p = uv * aspect * scale;

    // Grid for drop placement
    float gridY = 6.0;
    float gridX = 4.0;
    p.y += t * speed;

    vec2 cellID = floor(p * vec2(gridX, gridY));
    vec2 cellUV = fract(p * vec2(gridX, gridY));

    // Randomize drop position within cell
    vec2 dropCenter = _hash2(cellID) * 0.6 + 0.2;

    // Drop only in some cells
    float dropExists = step(0.55, _hash(cellID * 1.7));

    // Wobble path as drop slides
    float wobble = sin(cellID.x * 3.0 + t * 2.0 + p.y * 4.0) * 0.08;
    dropCenter.x += wobble;

    vec2 d = cellUV - dropCenter;
    d.y *= 1.6; // elongate vertically

    // Drop head (circle)
    float headRadius = 0.12 + _hash(cellID * 2.1) * 0.06;
    float headDist = length(d);
    float head = smoothstep(headRadius, headRadius - 0.04, headDist);

    // Trailing smear above the drop
    float trail = 0.0;
    if (d.y < 0.0) {
        float trailWidth = headRadius * 0.5 * smoothstep(-0.5, 0.0, d.y);
        trail = smoothstep(trailWidth, trailWidth - 0.02, abs(d.x)) *
                smoothstep(-0.5, -0.02, d.y);
    }

    float dropAlpha = max(head, trail * 0.4) * dropExists;

    // Refraction normal from drop shape (central difference on SDF-like field)
    float eps = 0.01;
    float dx = length(d + vec2(eps, 0.0)) - length(d - vec2(eps, 0.0));
    float dy = length(d + vec2(0.0, eps)) - length(d - vec2(0.0, eps));
    vec2 normal = vec2(dx, dy);

    // Stronger refraction near edges, less at center
    float edgeFactor = smoothstep(0.0, headRadius, headDist) * head;
    normal *= (0.5 + edgeFactor * 2.0);

    return vec3(normal * dropAlpha, dropAlpha);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 term = x_Texture(uv);
    float t = x_Time;
    float mask = termMask(term.rgb);

    // --- Configuration ---
    float refractStrength = 0.02;   // UV distortion strength
    float mistOpacity     = 0.12;   // condensation mist
    int   bokehCount      = 12;     // neon bokeh light count
    float bokehSize       = 0.06;   // bokeh circle softness

    // --- Raindrop layers (parallax) ---
    vec3 layer1 = rainLayer(uv, t, 1.0, 0.3);
    vec3 layer2 = rainLayer(uv + 0.37, t, 1.5, 0.45);
    vec3 layer3 = rainLayer(uv + 0.71, t, 2.2, 0.25);

    // Combine refraction normals
    vec2 totalRefract = layer1.xy * 1.0 + layer2.xy * 0.6 + layer3.xy * 0.3;
    float totalAlpha = max(max(layer1.z, layer2.z * 0.7), layer3.z * 0.4);

    // Refracted terminal sample
    vec2 refractedUV = uv + totalRefract * refractStrength;
    refractedUV = clamp(refractedUV, 0.0, 1.0);
    vec4 refracted = x_Texture(refractedUV);

    // --- Neon bokeh lights (visible in dark areas) ---
    vec3 bokeh = vec3(0.0);
    vec3 bokehColors[4];
    bokehColors[0] = vec3(1.0, 0.15, 0.6);  // hot pink
    bokehColors[1] = vec3(0.1, 0.8, 1.0);   // cyan
    bokehColors[2] = vec3(0.7, 0.2, 1.0);   // purple
    bokehColors[3] = vec3(0.1, 1.0, 0.5);   // green

    for (int i = 0; i < 12; i++) {
        if (i >= bokehCount) break;
        vec2 bPos = _hash2(vec2(float(i) * 7.3, 1.5));
        // Gentle drift
        bPos.x += sin(t * 0.3 + float(i)) * 0.05;
        bPos.y += cos(t * 0.2 + float(i) * 1.3) * 0.03;

        float bDist = length((uv - bPos) * vec2(x_WindowSize.x / x_WindowSize.y, 1.0));
        float bGlow = bokehSize / (bDist * bDist + bokehSize);
        bGlow *= 0.015;

        int colorIdx = i - (i / 4) * 4; // i % 4 without mod
        vec3 bColor = bokehColors[colorIdx];
        // Slight color cycling
        bColor = mix(bColor, bColor.gbr, sin(t * 0.5 + float(i)) * 0.2 + 0.2);

        bokeh += bColor * bGlow;
    }

    // Bokeh also refracts through drops
    vec3 bokehRefracted = bokeh;
    if (totalAlpha > 0.01) {
        // Brighter bokeh through drops (lens magnification)
        bokehRefracted = bokeh * (1.0 + totalAlpha * 2.0);
    }

    // --- Condensation mist ---
    float mist = fbm(uv * 8.0 + t * 0.1) * mistOpacity;
    vec3 mistColor = vec3(0.5, 0.6, 0.7) * mist;

    // --- Composite ---
    // Base: refracted terminal where drops are, normal elsewhere
    vec3 baseColor = mix(term.rgb, refracted.rgb, totalAlpha);

    // Add effects in dark areas
    vec3 effectColor = bokehRefracted + mistColor;

    // Drop highlight (specular-like bright edge on drops)
    float specular = pow(max(totalAlpha, 0.0), 3.0) * 0.3;
    vec3 specColor = vec3(0.8, 0.9, 1.0) * specular;

    vec3 finalColor = baseColor + effectColor * mask + specColor;

    _wShaderOut = vec4(finalColor, term.a);
}
