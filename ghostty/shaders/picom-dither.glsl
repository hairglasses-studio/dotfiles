precision highp float;
// Dither — ordered dithering with configurable pattern depth
// Category: Post-FX | Cost: MED | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec2 px = 1.0 / iResolution.xy;

    const bool monochrome = false;
    const int block_size = 2;
    const float bit_depth = 8.0;

    // 2x2 Bayer-style dither pattern (9 levels)
    float dither[18];
    dither[ 0]=0.0; dither[ 1]=0.0;
    dither[ 2]=0.5; dither[ 3]=0.0;
    dither[ 4]=0.5; dither[ 5]=0.0;
    dither[ 6]=0.5; dither[ 7]=0.5;
    dither[ 8]=0.5; dither[ 9]=0.5;
    dither[10]=1.0; dither[11]=0.5;
    dither[12]=1.0; dither[13]=0.5;
    dither[14]=1.0; dither[15]=1.0;
    dither[16]=1.0; dither[17]=1.0;

    float dither_row2[18];
    dither_row2[ 0]=0.0; dither_row2[ 1]=0.0;
    dither_row2[ 2]=0.0; dither_row2[ 3]=0.0;
    dither_row2[ 4]=0.0; dither_row2[ 5]=0.5;
    dither_row2[ 6]=0.0; dither_row2[ 7]=0.5;
    dither_row2[ 8]=0.5; dither_row2[ 9]=0.5;
    dither_row2[10]=0.5; dither_row2[11]=0.5;
    dither_row2[12]=0.5; dither_row2[13]=1.0;
    dither_row2[14]=0.5; dither_row2[15]=1.0;
    dither_row2[16]=1.0; dither_row2[17]=1.0;

    ivec2 pixCoord = ivec2(fragCoord);
    ivec2 blockPos = pixCoord % block_size;

    vec3 blockColor = vec3(0.0);
    float alpha = 0.0;
    for (int y = 0; y < block_size; y++) {
        for (int x = 0; x < block_size; x++) {
            vec2 sampleUv = (vec2(pixCoord - blockPos + ivec2(x, y)) + 0.5) / iResolution.xy;
            vec4 s = texture(iChannel0, sampleUv);
            if (x == 0 && y == 0) alpha = s.a;
            if (monochrome) {
                blockColor.r += (s.r + s.g + s.b) / 3.0;
            } else {
                blockColor += s.rgb;
            }
        }
    }

    blockColor /= float(block_size * block_size);
    vec3 quantized;
    for (int ch = 0; ch < 3; ch++) {
        float val = (ch == 0) ? blockColor.r : (ch == 1) ? blockColor.g : blockColor.b;
        if (monochrome && ch > 0) {
            quantized[ch] = quantized[0];
            continue;
        }
        int level = int(round(val * bit_depth));
        level = clamp(level, 0, int(bit_depth));
        int idx = level * 2 + blockPos.x;
        if (blockPos.y == 0)
            quantized[ch] = dither[idx];
        else
            quantized[ch] = dither_row2[idx];
    }

    fragColor = vec4(quantized, alpha);
}
