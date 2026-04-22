// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Neon noir — dramatic venetian blind shadows + smoky atmosphere + backlit silhouette

const float INTENSITY = 0.55;

vec3 nn_pal(float t) {
    vec3 a = vec3(0.95, 0.30, 0.60);
    vec3 b = vec3(0.15, 0.75, 0.95);
    vec3 c = vec3(1.00, 0.70, 0.30);
    vec3 d = vec3(0.55, 0.30, 0.98);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float nn_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float nn_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(nn_hash(i), nn_hash(i + vec2(1,0)), u.x),
               mix(nn_hash(i + vec2(0,1)), nn_hash(i + vec2(1,1)), u.x), u.y);
}

float nn_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 4; i++) {
        v += a * nn_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.02, 0.01, 0.03);

    // Light source — animated neon sign off-screen
    vec2 lightPos = vec2(0.8, 0.5) + 0.05 * vec2(sin(x_Time * 0.3), cos(x_Time * 0.27));

    // Venetian blinds — horizontal bands casting shadows
    float blindFreq = 40.0;
    float blindAngle = 0.2;  // slight tilt
    float blindCoord = (p.y - p.x * blindAngle) * blindFreq;
    float blindFrac = fract(blindCoord);
    float blindThickness = 0.6 + 0.1 * sin(x_Time * 0.3);  // blinds open/close
    float inShadow = smoothstep(blindThickness - 0.02, blindThickness, blindFrac);

    // Light coming from lightPos direction — attenuated by distance + shadow
    float distToLight = length(p - lightPos);
    float lightBrightness = exp(-distToLight * 0.5) * (1.0 - inShadow);
    vec3 lightCol = nn_pal(fract(x_Time * 0.05));

    col += lightCol * lightBrightness * 0.35;

    // Shadow side of room (opposite light)
    float roomSide = smoothstep(lightPos.x - 1.5, lightPos.x + 0.5, p.x);
    col *= 0.4 + roomSide * 0.8;

    // Atmospheric smoke — horizontal bands drifting
    float smokeY = uv.y;
    vec2 smokeP = vec2(uv.x * 4.0 + x_Time * 0.1, smokeY * 2.0);
    float smoke = nn_fbm(smokeP) * smoothstep(0.2, 0.8, smokeY);
    col += nn_pal(0.2) * smoke * lightBrightness * 0.4;

    // Light shafts visible in smoke
    for (int k = 0; k < 5; k++) {
        float fk = float(k);
        float shaftAngle = 2.8 + fk * 0.08 + x_Time * 0.02;
        vec2 shaftDir = vec2(cos(shaftAngle), sin(shaftAngle));
        // Distance from fragment to line through lightPos with direction shaftDir
        vec2 toP = p - lightPos;
        float shaftPerp = abs(toP.x * shaftDir.y - toP.y * shaftDir.x);
        float shaftMask = exp(-shaftPerp * shaftPerp * 800.0);
        // Only on far side from light
        float shaftAlong = dot(toP, shaftDir);
        if (shaftAlong < 0.0) shaftMask = 0.0;
        col += lightCol * shaftMask * smoke * 0.4;
    }

    // Neon sign — the actual light source
    float signD = length(p - lightPos);
    if (signD < 0.04) {
        col += lightCol * smoothstep(0.04, 0.02, signD) * 1.5;
    }

    // Film grain
    float grain = nn_hash(x_PixelPos + vec2(floor(x_Time * 30.0), 0.0));
    col *= 0.9 + grain * 0.15;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
