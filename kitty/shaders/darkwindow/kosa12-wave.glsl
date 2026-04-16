// kosa12/CRTty wave — MIT License
// Ported from https://github.com/kosa12/CRTty/blob/main/examples/wave.glsl

void windowShader(inout vec4 _wShaderOut) {
    float wave = sin(v_texcoord.y * 40.0 + x_Time * 3.0) * 0.003;
    vec3 c;
    c.r = x_Texture(v_texcoord + vec2(wave, 0.0)).r;
    c.g = x_Texture(v_texcoord).g;
    c.b = x_Texture(v_texcoord - vec2(wave, 0.0)).b;

    float scan = 0.95 + 0.05 * sin(v_texcoord.y * x_WindowSize.y * 3.14159 + x_Time * 2.0);
    c *= scan;

    _wShaderOut = vec4(c, 1.0);
}
