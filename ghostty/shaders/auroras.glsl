// Northern lights — "Auroras" by nimitz (Stanton)
// Ported from https://www.shadertoy.com/view/XtGGRt
precision highp float;

const mat2 M2 = mat2(0.95534, 0.29552, -0.29552, 0.95534);

float tri(float x) { return clamp(abs(fract(x) - 0.5), 0.01, 0.49); }
vec2 tri2(vec2 p) { return vec2(tri(p.x) + tri(p.y), tri(p.y + tri(p.x))); }

float triNoise2d(vec2 p, float spd) {
    float z = 1.8;
    float z2 = 2.5;
    float rz = 0.0;
    p *= M2;
    vec2 bp = p;
    for (float i = 0.0; i < 5.0; i++) {
        vec2 dg = tri2(bp * 1.85) * 0.75;
        dg *= M2;
        p -= dg / z2;
        bp *= 1.3;
        z2 *= 0.45;
        z *= 0.42;
        p *= 1.21 + (rz - 1.0) * 0.02;
        rz += tri(p.x + tri(p.y)) * z;
        p *= -M2;
    }
    return clamp(1.0 / pow(rz * 29.0, 1.3), 0.0, 0.55);
}

float hash21(vec2 n) {
    uvec2 q = uvec2(n * 256.0) * uvec2(1597334673u, 3812015801u);
    uint h = (q.x ^ q.y) * 1597334673u;
    return float(h) / float(0xffffffffu);
}

vec4 aurora(vec3 ro, vec3 rd) {
    vec4 col = vec4(0);
    vec4 avgCol = vec4(0);

    for (float i = 0.0; i < 50.0; i++) {
        float of = 0.006 * hash21(gl_FragCoord.xy) * smoothstep(0.0, 15.0, i);
        float pt = ((0.8 + pow(i, 1.4) * 0.002) - ro.y) / (rd.y * 2.0 + 0.4);
        pt -= of;
        vec3 bpos = ro + pt * rd;
        vec2 p = bpos.zx;
        float rzt = triNoise2d(p, 0.06);
        vec4 col2 = vec4(0, 0, 0, rzt);
        col2.rgb = (sin(1.0 - vec3(2.15, -0.5, 1.2) + i * 0.043) * 0.5 + 0.5) * rzt;
        avgCol = mix(avgCol, col2, 0.5);
        col += avgCol * exp2(-i * 0.065 - 2.5) * smoothstep(0.0, 5.0, i);
    }
    col *= clamp(rd.y * 15.0 + 0.4, 0.0, 1.0);
    return col * 1.8;
}

vec3 bg(vec3 rd) {
    float sd = dot(normalize(vec3(-0.5, -0.6, 0.9)), rd) * 0.5 + 0.5;
    sd = pow(sd, 5.0);
    vec3 col = mix(vec3(0.05, 0.1, 0.2), vec3(0.1, 0.05, 0.2), sd);
    return col * 0.63;
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 q = fragCoord / iResolution.xy;
    vec2 p = q - 0.5;
    p.x *= iResolution.x / iResolution.y;

    vec3 ro = vec3(0, 0, -6.7);
    vec3 rd = normalize(vec3(p, 1.3));

    // Slow auto-rotation (replaces iMouse)
    float a = sin(iTime * 0.05) * 0.15;
    mat2 rot = mat2(cos(a), sin(a), -sin(a), cos(a));
    rd.xz *= rot;

    vec2 mo = vec2(sin(iTime * 0.08) * 0.3, sin(iTime * 0.06) * 0.1 + 0.1);
    rd.yz *= mat2(cos(mo.y), sin(mo.y), -sin(mo.y), cos(mo.y));
    rd.xz *= mat2(cos(mo.x), sin(mo.x), -sin(mo.x), cos(mo.x));

    vec3 col = vec3(0);
    vec3 brd = rd;
    float fade = smoothstep(0.0, 0.01, abs(brd.y)) * 0.1 + 0.9;
    col = bg(rd) * fade;

    if (rd.y > 0.0) {
        vec4 aur = smoothstep(0.0, 1.5, aurora(ro, rd)) * fade;
        col += aur.rgb;
    }

    col = pow(clamp(col, 0.0, 1.0), vec3(0.5, 0.5, 0.6));

    // Blend with terminal text
    vec4 terminal = texture(iChannel0, q);
    float termLuma = dot(terminal.rgb, vec3(0.2126, 0.7152, 0.0722));
    float mask = 1.0 - smoothstep(0.05, 0.2, termLuma);
    fragColor = vec4(mix(terminal.rgb, col, mask), terminal.a);
}
