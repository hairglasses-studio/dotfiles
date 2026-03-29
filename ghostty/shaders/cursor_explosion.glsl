precision highp float;
// -----------------------------------------------------------------
// 辅助函数 (Correct)
// -----------------------------------------------------------------
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

// -----------------------------------------------------------------
// 效果参数 (最终调整版)
// -----------------------------------------------------------------
const float EFFECT_DURATION = 0.5;      // 稍微延长总时长以容纳随机性

// --- 粒子参数 ---
const int NUM_PARTICLES = 45;           // 粒子数量
const float PARTICLE_SPEED = 4.0;       // 粒子基础速度
const float PARTICLE_SIZE = 0.008;      // 粒子大小
// --- (新增) 随机化参数 ---
const float PARTICLE_EMISSION_DELAY = 0.15; // 最大发射延迟(秒), 0.0 = 同时发射
const float PARTICLE_LIFESPAN_VARIATION = 0.6; // 生命周期随机范围, 0.0 = 同样长, 1.0 = 差异极大

// --- 冲击波参数 ---
const bool ENABLE_SHOCKWAVE = true;
const float SHOCKWAVE_SPEED = 5.0;
const float SHOCKWAVE_THICKNESS = 0.005;
const vec3 SHOCKWAVE_COLOR = vec3(1.0, 0.8, 0.5);

// --- 光标参数 ---
const vec3 CURSOR_COLOR = vec3(1.0, 1.0, 1.0);

// -----------------------------------------------------------------
// 主函数
// -----------------------------------------------------------------
void mainImage(out vec4 fragColor, in vec2 fragCoord)
{
    fragColor = texture(iChannel0, fragCoord.xy / iResolution.xy);
    vec2 vu = normalize(fragCoord, 1.);
    vec4 currentCursor = vec4(normalize(iCurrentCursor.xy, 1.), normalize(iCurrentCursor.zw, 0.));
    vec2 cursorCenter = currentCursor.xy + vec2(currentCursor.z * 0.5, -currentCursor.w * 0.5);

    float sdfCursor = sdBox(vu, cursorCenter, currentCursor.zw * 0.5);
    fragColor = mix(vec4(CURSOR_COLOR, 1.0), fragColor, smoothstep(0.0, 0.005, sdfCursor));

    float timeSinceMove = iTime - iTimeCursorChange;

    if (timeSinceMove < EFFECT_DURATION) {
        float t_global = timeSinceMove / EFFECT_DURATION;
        float fade_global = pow(1.0 - t_global, 2.0);

        if (ENABLE_SHOCKWAVE) {
            float shockwaveRadius = t_global * SHOCKWAVE_SPEED * currentCursor.w;
            float shockwaveThickness = SHOCKWAVE_THICKNESS * fade_global;
            float sdfShockwave = abs(length(vu - cursorCenter) - shockwaveRadius) - shockwaveThickness;
            vec4 shockwaveColor = vec4(SHOCKWAVE_COLOR, 1.0) * fade_global * 1.5;
            fragColor = mix(fragColor, shockwaveColor, 1.0 - smoothstep(0.0, 0.003, sdfShockwave));
        }

        for (int i = 0; i < NUM_PARTICLES; i++) {
            float id = float(i);

            // ================== 新增随机化逻辑 ==================
            // 1. 随机发射延迟
            float emissionDelay = hash(id + 2.0) * PARTICLE_EMISSION_DELAY;
            float particleTime = timeSinceMove - emissionDelay;

            // 如果还没到这个粒子的发射时间，就跳过它
            if (particleTime < 0.0) continue;

            // 2. 随机生命周期
            // lifespanFactor范围: [1.0 - VARIATION, 1.0]
            float lifespanFactor = 1.0 - hash(id + 3.0) * PARTICLE_LIFESPAN_VARIATION;
            float particleDuration = EFFECT_DURATION * lifespanFactor;

            // 计算这个粒子自己的动画进度 t_particle
            float t_particle = particleTime / particleDuration;

            // 如果这个粒子已经结束了生命，就跳过它
            if (t_particle > 1.0) continue;
            // ====================================================

            // 使用 t_particle 来计算淡出和位置
            float fade_particle = pow(1.0 - t_particle, 2.0);

            float angle = hash(id) * 6.2831;
            vec2 dir = vec2(cos(angle), sin(angle));
            float speed = mix(0.5, 1.5, hash(id + 1.0)) * PARTICLE_SPEED;

            // 使用 t_particle 来计算位置
            vec2 p = cursorCenter + dir * t_particle * speed * currentCursor.w;

            vec3 particleColor = 0.5 + 0.5 * cos(id + vec3(0.0, 2.0, 4.0));
            float sdfParticle = length(vu - p) - PARTICLE_SIZE * fade_particle;
            vec4 pColor = vec4(particleColor, 1.0) * fade_particle;

            fragColor = mix(fragColor, pColor, 1.0 - smoothstep(0.0, 0.002, sdfParticle));
        }
    }
}
