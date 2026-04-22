// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Refractive droplet — single large lens-like droplet with chromatic refraction + sparkle

const float INTENSITY = 0.55;

vec3 rd_pal(float t) {
    vec3 a = vec3(0.20, 0.85, 0.95);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.95, 0.30, 0.70);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float rd_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Droplet position — slowly drifts
    vec2 dropletPos = 0.15 * vec2(sin(x_Time * 0.2), cos(x_Time * 0.17));
    float dropletR = 0.32;

    vec2 toDrop = p - dropletPos;
    float r = length(toDrop);

    vec3 col = vec3(0.0);

    if (r < dropletR) {
        // Inside droplet — refract terminal sample
        float angFromCenter = r / dropletR;
        // Refraction strength grows toward edge
        vec2 normal = toDrop / max(r, 0.001);
        // Fake "lens" refraction — pull sample toward center by a factor dependent on position
        float refractStrength = pow(angFromCenter, 2.0) * 0.1;
        vec2 refractedUV = uv - normal * refractStrength;

        // Chromatic aberration
        float chrom = refractStrength * 0.3;
        vec3 refracted;
        refracted.r = x_Texture(refractedUV + normal * chrom).r;
        refracted.g = x_Texture(refractedUV).g;
        refracted.b = x_Texture(refractedUV - normal * chrom).b;

        // Invert image (because lens would)
        vec2 centeredUV = (refractedUV - 0.5) * -1.0 + 0.5;
        vec3 inverted = x_Texture(centeredUV).rgb;
        // Mix inverted into refracted based on distance from edge
        refracted = mix(refracted, inverted, smoothstep(0.3, 1.0, angFromCenter) * 0.7);

        col = refracted;

        // Fresnel rim
        float fresnel = pow(angFromCenter, 4.0);
        col += rd_pal(fract(x_Time * 0.04)) * fresnel * 0.4;

        // Sparkle on surface (specular)
        vec2 specCenter = dropletPos + vec2(-dropletR * 0.3, dropletR * 0.4);
        float specD = length(p - specCenter);
        col += vec3(1.0) * exp(-specD * specD * 500.0) * 0.9;

        // Edge highlight
        float edgeDist = dropletR - r;
        float edgeMask = smoothstep(0.02, 0.0, edgeDist);
        col += rd_pal(fract(0.3 + x_Time * 0.05)) * edgeMask * 0.5;

        // Internal swirl — subtle pattern
        float ang = atan(toDrop.y, toDrop.x);
        float swirl = sin(ang * 8.0 + r * 20.0 + x_Time) * 0.04;
        col += rd_pal(fract(ang / 6.28)) * swirl * (1.0 - angFromCenter);
    } else {
        // Outside: pass-through with subtle halo
        vec4 terminal = x_Texture(uv);
        col = terminal.rgb;
        // Outer halo glow
        float halo = exp(-(r - dropletR) * 10.0) * 0.15;
        col += rd_pal(fract(x_Time * 0.04)) * halo;
    }

    _wShaderOut = vec4(col, 1.0);
}
