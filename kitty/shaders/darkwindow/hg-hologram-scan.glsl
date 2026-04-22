// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Hologram scan — horizontal scan bar reveals wireframe object + interference noise

const float INTENSITY = 0.5;

vec3 hs_pal(float t) {
    vec3 cyan = vec3(0.20, 0.95, 0.98);
    vec3 mag  = vec3(0.95, 0.30, 0.70);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 white = vec3(0.95, 0.98, 1.00);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, white, s - 2.0);
    else              return mix(white, cyan, s - 3.0);
}

float hs_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

// Simple "scanned object" — a 3D cube projected to 2D
vec2 project(vec3 v) {
    float zDist = 3.0;
    float z = 1.0 / (zDist - v.z);
    return v.xy * z;
}

// Distance to cube wireframe (projected)
float cubeWireframeDist(vec2 p, mat3 rot) {
    // 8 cube vertices
    vec3 verts[8];
    verts[0] = vec3(-1.0, -1.0, -1.0); verts[1] = vec3( 1.0, -1.0, -1.0);
    verts[2] = vec3(-1.0,  1.0, -1.0); verts[3] = vec3( 1.0,  1.0, -1.0);
    verts[4] = vec3(-1.0, -1.0,  1.0); verts[5] = vec3( 1.0, -1.0,  1.0);
    verts[6] = vec3(-1.0,  1.0,  1.0); verts[7] = vec3( 1.0,  1.0,  1.0);

    // Project rotated vertices
    vec2 projVerts[8];
    for (int i = 0; i < 8; i++) {
        projVerts[i] = project(rot * verts[i] * 0.5);
    }

    // 12 edges of cube
    int edges[24];
    edges[0]=0; edges[1]=1;  edges[2]=0; edges[3]=2;   edges[4]=0; edges[5]=4;
    edges[6]=1; edges[7]=3;  edges[8]=1; edges[9]=5;   edges[10]=2; edges[11]=3;
    edges[12]=2; edges[13]=6; edges[14]=3; edges[15]=7; edges[16]=4; edges[17]=5;
    edges[18]=4; edges[19]=6; edges[20]=5; edges[21]=7; edges[22]=6; edges[23]=7;

    float minD = 1e9;
    for (int e = 0; e < 12; e++) {
        vec2 a = projVerts[edges[e * 2]];
        vec2 b = projVerts[edges[e * 2 + 1]];
        vec2 pa = p - a;
        vec2 ba = b - a;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        float d = length(pa - ba * h);
        minD = min(minD, d);
    }
    return minD;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.02, 0.05);

    // Rotating cube
    float t = x_Time * 0.4;
    float cr = cos(t), sr = sin(t);
    mat3 rotY = mat3(cr, 0.0, -sr, 0.0, 1.0, 0.0, sr, 0.0, cr);
    float cp = cos(t * 0.7), sp = sin(t * 0.7);
    mat3 rotX = mat3(1.0, 0.0, 0.0, 0.0, cp, -sp, 0.0, sp, cp);
    mat3 rot = rotX * rotY;

    float edgeD = cubeWireframeDist(p, rot);

    // Scan bar — horizontal line moving down
    float scanY = mod(x_Time * 0.4, 1.4) - 0.7;
    float scanDist = abs(p.y - scanY);
    float scanBar = exp(-scanDist * scanDist * 5000.0);
    float scanBrightBand = smoothstep(0.02, 0.0, scanDist);

    // Edge brightness dependent on proximity to scan bar (wireframe only revealed where scanned)
    float edgeCore = smoothstep(0.003, 0.0, edgeD);
    float edgeGlow = exp(-edgeD * edgeD * 1200.0) * 0.25;
    // Scan intensity — bright near scan bar, fades as it passes
    float scanInfluence = exp(-pow((p.y - scanY) * 2.0, 2.0) * 1.5);

    vec3 edgeColor = hs_pal(fract(x_Time * 0.05));
    col += edgeColor * (edgeCore * 0.8 + edgeGlow) * (0.3 + scanInfluence * 1.2);

    // Scan bar itself
    col += hs_pal(0.0) * scanBar * 0.5;
    col += vec3(1.0) * scanBrightBand * 0.3;

    // Interference noise lines
    float noiseBand = sin(p.y * 120.0 + x_Time * 30.0);
    noiseBand = pow(max(0.0, noiseBand), 4.0);
    col += edgeColor * noiseBand * 0.04;

    // Random glitch artifacts in scan band
    if (scanDist < 0.05) {
        float artifactHash = hs_hash(vec2(floor(p.x * 60.0), floor(x_Time * 15.0)));
        if (artifactHash > 0.95) {
            col += hs_pal(fract(artifactHash + x_Time)) * 0.3;
        }
    }

    // Holographic flicker
    col *= 0.9 + 0.1 * sin(x_PixelPos.y * 0.5 + x_Time * 8.0);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
