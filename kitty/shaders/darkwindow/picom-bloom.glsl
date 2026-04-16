// Bloom — Gaussian blur + screen blending on bright pixels
// Category: Post-FX | Cost: MED | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec2 px = 1.0 / x_WindowSize;

    const float cutoff = 0.55;
    const float light_brightness = 1.0;
    const float base_brightness = 1.2;

    // 7x7 Gaussian kernel
    float kernel[49];
    kernel[ 0]=0.0051; kernel[ 1]=0.0094; kernel[ 2]=0.0135; kernel[ 3]=0.0153; kernel[ 4]=0.0135; kernel[ 5]=0.0094; kernel[ 6]=0.0051;
    kernel[ 7]=0.0094; kernel[ 8]=0.0173; kernel[ 9]=0.0250; kernel[10]=0.0282; kernel[11]=0.0250; kernel[12]=0.0173; kernel[13]=0.0094;
    kernel[14]=0.0135; kernel[15]=0.0250; kernel[16]=0.0361; kernel[17]=0.0407; kernel[18]=0.0361; kernel[19]=0.0250; kernel[20]=0.0135;
    kernel[21]=0.0153; kernel[22]=0.0282; kernel[23]=0.0407; kernel[24]=0.0461; kernel[25]=0.0407; kernel[26]=0.0282; kernel[27]=0.0153;
    kernel[28]=0.0135; kernel[29]=0.0250; kernel[30]=0.0361; kernel[31]=0.0407; kernel[32]=0.0361; kernel[33]=0.0250; kernel[34]=0.0135;
    kernel[35]=0.0094; kernel[36]=0.0173; kernel[37]=0.0250; kernel[38]=0.0282; kernel[39]=0.0250; kernel[40]=0.0173; kernel[41]=0.0094;
    kernel[42]=0.0051; kernel[43]=0.0094; kernel[44]=0.0135; kernel[45]=0.0153; kernel[46]=0.0135; kernel[47]=0.0094; kernel[48]=0.0051;

    int radius = 3;

    vec4 total = vec4(0.0);
    for (int y = -radius; y <= radius; y++) {
        for (int x = -radius; x <= radius; x++) {
            vec4 s = x_Texture(uv + vec2(float(x), float(y)) * px);
            float bright = (s.r + s.g + s.b) / 3.0;
            if (bright < cutoff) s = vec4(0.0);
            int idx = (x + radius) + (y + radius) * 7;
            s *= kernel[idx];
            s.rgb *= light_brightness;
            total += s;
        }
    }

    vec4 c = x_Texture(uv);
    float bright = (c.r + c.g + c.b) / 3.0;
    if (bright >= cutoff) {
        c.rgb = min(c.rgb * base_brightness, vec3(1.0));
    }

    // Screen blending mode
    c = 1.0 - (1.0 - total) * (1.0 - c);

    _wShaderOut = c;
}
