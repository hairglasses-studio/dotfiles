// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Iron filings around magnetic poles — field-line alignment + pole rotation

const float INTENSITY = 0.5;

vec3 mf_pal(float t) {
    vec3 a = vec3(0.90, 0.30, 0.50);  // red (N pole)
    vec3 b = vec3(0.15, 0.55, 0.98);  // blue (S pole)
    vec3 c = vec3(0.60, 0.65, 0.75);  // iron grey (filings)
    float s = mod(t * 3.0, 3.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else              return mix(c, a, s - 2.0);
}

float mf_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// 2D magnetic field at point p from two opposite poles
vec2 dipoleField(vec2 p, vec2 polePos, vec2 poleNeg) {
    vec2 dPos = p - polePos;
    vec2 dNeg = p - poleNeg;
    float rPos = max(length(dPos), 0.01);
    float rNeg = max(length(dNeg), 0.01);
    // Field from N pole points outward, field from S points inward
    vec2 bPos =  dPos / (rPos * rPos * rPos);
    vec2 bNeg = -dNeg / (rNeg * rNeg * rNeg);
    return bPos + bNeg;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Rotating dipole
    float ang = x_Time * 0.2;
    float cr = cos(ang), sr = sin(ang);
    mat2 rot = mat2(cr, -sr, sr, cr);
    vec2 polePos = rot * vec2(0.22, 0.0);
    vec2 poleNeg = -polePos;

    // Field at p
    vec2 field = dipoleField(p, polePos, poleNeg);
    float fieldMag = length(field);
    vec2 fieldDir = fieldMag > 0.0 ? field / fieldMag : vec2(1.0, 0.0);

    vec3 col = vec3(0.02, 0.025, 0.04);   // dark paper

    // "Filings" = anisotropic stripe pattern aligned with field direction
    // Sample perpendicular distance from fragment to closest field-line iso
    // Use perpendicular dot product projected and a high-freq cosine
    float along = dot(p, fieldDir) * 60.0;
    float perp  = dot(p, vec2(-fieldDir.y, fieldDir.x)) * 60.0;
    float filings = 0.5 + 0.5 * cos(along);   // parallel stripes
    filings = pow(filings, 4.0);

    // Filing density falls off with distance from field lines (perp)
    float stripeSharp = pow(filings, 1.5 + (1.0 - smoothstep(0.0, 2.0, fieldMag * 1.5)) * 2.0);

    // Field strength modulates visibility
    float strengthMask = smoothstep(0.5, 5.0, fieldMag);
    col += mf_pal(0.7) * stripeSharp * strengthMask * 0.7;

    // Grain noise per "filing"
    vec2 grainP = floor(p * 200.0);
    float grain = mf_hash(grainP + floor(x_Time * 20.0));
    col += mf_pal(0.7) * step(0.99, grain) * stripeSharp * 0.5;

    // Poles
    float dPos = length(p - polePos);
    float dNeg = length(p - poleNeg);
    // Pole bodies
    col = mix(col, mf_pal(0.0), smoothstep(0.04, 0.035, dPos));
    col = mix(col, mf_pal(0.35), smoothstep(0.04, 0.035, dNeg));
    // Pole label glow
    col += mf_pal(0.0) * exp(-dPos * dPos * 600.0) * 0.5;
    col += mf_pal(0.35) * exp(-dNeg * dNeg * 600.0) * 0.5;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
