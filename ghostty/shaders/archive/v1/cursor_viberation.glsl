precision highp float;
// =================================================================
// 辅助函数 (整合自两个文件，已去重)
// =================================================================
vec2 normalize(vec2 value, float isPosition) {
    return (value * 2.0 - (iResolution.xy * isPosition)) / iResolution.y;
}

float hash(float n) {
    return fract(sin(n) * 43758.5453123);
}

float sdBox(in vec2 p, in vec2 xy, in vec2 b) {
    vec2 d = abs(p - xy) - b;
    return length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);
}

// 拖尾的 SDF 函数
float seg(in vec2 p, in vec2 a, in vec2 b, inout float s, float d) {
    vec2 e = b - a; vec2 w = p - a;
    vec2 proj = a + e * clamp(dot(w, e) / dot(e, e), 0.0, 1.0);
    float segd = dot(p - proj, p - proj);
    d = min(d, segd);
    float c0 = step(0.0, p.y - a.y); float c1 = 1.0 - step(0.0, p.y - b.y);
    float c2 = 1.0 - step(0.0, e.x * w.y - e.y * w.x);
    float allCond = c0 * c1 * c2; float noneCond = (1.0 - c0) * (1.0 - c1) * (1.0 - c2);
    float flip = mix(1.0, -1.0, step(0.5, allCond + noneCond));
    s *= flip;
    return d;
}

float getSdfParallelogram(in vec2 p, in vec2 v0, in vec2 v1, in vec2 v2, in vec2 v3) {
    float s = 1.0; float d = dot(p - v0, p - v0);
    d = seg(p, v0, v3, s, d); d = seg(p, v1, v0, s, d);
    d = seg(p, v2, v1, s, d); d = seg(p, v3, v2, s, d);
    return s * sqrt(d);
}

float determineStartVertexFactor(vec2 a, vec2 b) {
    float c1 = step(b.x, a.x) * step(a.y, b.y);
    float c2 = step(a.x, b.x) * step(b.y, a.y);
    return 1.0 - max(c1, c2);
}

vec2 getRectangleCenter(vec4 rectangle) {
    return vec2(rectangle.x + (rectangle.z / 2.), rectangle.y - (rectangle.w / 2.));
}

float antialising(float distance) {
    return 1. - smoothstep(0., normalize(vec2(2., 2.), 0.).x, distance);
}

// =================================================================
// 效果参数 (整合与调整)
// =================================================================

// --- 打击感参数 ---
const float IMPACT_DURATION = 0.1;
const float VIBRATION_FREQUENCY = 90.0;
const float VIBRATION_STRENGTH = 0.06;
const float GLOW_EXPANSION = 0.012;

// --- 拖尾参数 ---
const float TRAIL_DURATION = .8;
const float TRAIL_OPACITY = .4; // 稍微调高一点透明度以适应发光背景
const float WAVE_FREQUENCY = 0.0;
const float WAVE_AMPLITUDE = 0.0;
const bool HIDE_TRAILS_ON_THE_SAME_LINE = false;
const float DRAW_THRESHOLD = 1.5;

// --- 颜色预设 ---
const vec4 COLOR_A_START = vec4(1.0, 0.7, 0.2, 1.0); const vec4 COLOR_A_END = vec4(1.0, 0.9, 0.5, 1.0);
const vec4 COLOR_B_START = vec4(0.9, 0.2, 0.5, 1.0); const vec4 COLOR_B_END = vec4(0.2, 0.8, 0.9, 1.0);
const vec4 COLOR_C_START = vec4(0.1, 0.9, 0.6, 1.0); const vec4 COLOR_C_END = vec4(0.7, 1.0, 0.7, 1.0);
const vec4 COLOR_D_START = vec4(0.3, 0.5, 1.0, 1.0); const vec4 COLOR_D_END = vec4(0.7, 0.8, 1.0, 1.0);


void mainImage(out vec4 fragColor, in vec2 fragCoord)
{
    // --- 1. 初始化和颜色选择 ---
    fragColor = texture(iChannel0, fragCoord.xy / iResolution.xy);
    vec2 vu = normalize(fragCoord, 1.);

    vec4 currentCursor = vec4(normalize(iCurrentCursor.xy, 1.), normalize(iCurrentCursor.zw, 0.));
    vec4 previousCursor = vec4(normalize(iPreviousCursor.xy, 1.), normalize(iPreviousCursor.zw, 0.));
    vec2 cursorCenter = getRectangleCenter(currentCursor);
    vec2 cursorSize = currentCursor.zw * 0.5;

    // 根据时间戳随机选择一套颜色
    vec4 startColor, endColor, glowColor;
    float randomVal = hash(iTimeCursorChange);
    if (randomVal < 0.25)      { startColor = COLOR_A_START; endColor = COLOR_A_END; }
    else if (randomVal < 0.5)  { startColor = COLOR_B_START; endColor = COLOR_B_END; }
    else if (randomVal < 0.75) { startColor = COLOR_C_START; endColor = COLOR_C_END; }
    else                       { startColor = COLOR_D_START; endColor = COLOR_D_END; }
    glowColor = startColor; // 发光颜色使用拖尾的起始色，保持和谐

    float timeSinceMove = iTime - iTimeCursorChange;
    vec2 finalCursorCenter = cursorCenter; // 默认光标中心点

    // --- 2. 打击感效果 (发光与振动) ---
    if (timeSinceMove < IMPACT_DURATION) {
        float t = timeSinceMove / IMPACT_DURATION;
        float fade = pow(1.0 - t, 3.0);

        // 绘制发光 (图层2)
        vec2 glowSize = cursorSize + GLOW_EXPANSION * fade;
        float sdfGlow = sdBox(vu, cursorCenter, glowSize);
        vec4 finalGlowColor = vec4(glowColor.rgb, 1.0) * fade * 1.5;
        fragColor = mix(fragColor, finalGlowColor, 1.0 - smoothstep(0.0, 0.015, sdfGlow));

        // 计算振动位移，但不在此处绘制光标
        float vibration = sin(t * VIBRATION_FREQUENCY) * VIBRATION_STRENGTH * fade;
        vec2 shakeDirection = normalize(vec2(hash(1.37), hash(2.89)));
        finalCursorCenter += shakeDirection * vibration;
    }

    // --- 3. 拖尾效果 (曲线与渐变) ---
    vec2 prevCursorCenter = getRectangleCenter(previousCursor);
    float lineLength = distance(cursorCenter, prevCursorCenter);
    bool isFarEnough = lineLength > (DRAW_THRESHOLD * max(cursorSize.x, cursorSize.y));
    bool isOnSeparateLine = HIDE_TRAILS_ON_THE_SAME_LINE ? currentCursor.y != previousCursor.y : true;

    if (timeSinceMove < TRAIL_DURATION && isFarEnough && isOnSeparateLine) {
        float t = timeSinceMove / TRAIL_DURATION;
        float fade = pow(1.0 - t, 5.0); // 使用拖尾自己的衰减

        float vertexFactor = determineStartVertexFactor(currentCursor.xy, previousCursor.xy);
        vec2 v0 = vec2(currentCursor.x + currentCursor.z * vertexFactor, currentCursor.y - currentCursor.w);
        vec2 v1 = vec2(currentCursor.x + currentCursor.z * (1.0-vertexFactor), currentCursor.y);
        vec2 v2 = vec2(previousCursor.x + currentCursor.z * (1.0-vertexFactor), previousCursor.y);
        vec2 v3 = vec2(previousCursor.x + currentCursor.z * vertexFactor, previousCursor.y - previousCursor.w);

        vec2 trailVector = cursorCenter - prevCursorCenter;
        vec2 trailDir = normalize(trailVector);
        vec2 trailPerp = vec2(-trailDir.y, trailDir.x);
        vec2 pixelVector = vu - prevCursorCenter;
        float trailProgress = clamp(dot(pixelVector, trailDir) / lineLength, 0.0, 1.0);

        float waveOffset = sin(trailProgress * WAVE_FREQUENCY - iTime * 10.0) * WAVE_AMPLITUDE;
        waveOffset *= fade * (1.0 - trailProgress);
        vec2 warpedVu = vu - trailPerp * waveOffset;

        float sdfTrail = getSdfParallelogram(warpedVu, v0, v1, v2, v3);

        // 绘制拖尾 (图层3)
        vec4 finalTrailColor = mix(startColor, endColor, trailProgress);
        finalTrailColor.a *= fade * antialising(sdfTrail) * TRAIL_OPACITY;
        fragColor.rgb = mix(fragColor.rgb, finalTrailColor.rgb, finalTrailColor.a);
    }

    // --- 4. 最终绘制光标 (统一图层) ---
    // 无论如何，最后在所有效果之上绘制光标
    // 它会使用被振动修改过的 finalCursorCenter
    // 它的颜色是拖尾的结束色，保持一致
    float sdfCursor = sdBox(vu, finalCursorCenter, cursorSize);
    fragColor = mix(fragColor, vec4(endColor.rgb, 1.0), antialising(sdfCursor));
}
