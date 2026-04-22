// Shader attribution: fielding / JAM (https://github.com/fielding/ghostty-shader-adventures)
// License: no upstream LICENSE — personal/non-commercial use, attribution preserved.
// Ported to DarkWindow by hairglasses.
//
// Semantics changed from the original:
//   - iPreviousCursor has no DarkWindow equivalent, so a slow orbital "phantom" tracks
//     the cursor. The effect becomes constant electric arcs dancing from cursor to phantom
//     rather than typing-reactive arcs to the previous cursor position.
//   - iTimeCursorChange stubbed to a gentle oscillation → arcs pulse and breathe instead
//     of flaring on keystroke. Heat is capped below the screen-shake threshold.
//
// (Cyberpunk) — Electric arcs pulse from cursor to an orbital phantom

const float HEAT_RAMP_FAST = 0.03;
const float HEAT_RAMP_SLOW = 0.4;
const float ARC_DURATION   = 1.4;
const float MAX_ARC_THICKNESS = 0.014;
const float MIN_ARC_THICKNESS = 0.003;
const float GLOW_MULTIPLIER = 5.0;
const float BRANCH_THRESHOLD = 0.15;
const float TAIL_EXTENSION = 1.5;
const float SPARK_COUNT = 12.0;

const vec3 COL_PINK   = vec3(0.906, 0.204, 0.612);
const vec3 COL_CYAN   = vec3(0.102, 0.816, 0.839);
const vec3 COL_PURPLE = vec3(0.596, 0.443, 0.996);
const vec3 COL_GOLD   = vec3(0.949, 0.651, 0.200);
const vec3 COL_BLUE   = vec3(0.271, 0.541, 0.886);
const vec3 COL_WHITE  = vec3(0.973, 0.965, 0.949);

float el_hash21(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453123);
}

float el_vnoise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(
        mix(el_hash21(i), el_hash21(i + vec2(1.0, 0.0)), u.x),
        mix(el_hash21(i + vec2(0.0, 1.0)), el_hash21(i + vec2(1.0, 1.0)), u.x),
        u.y
    );
}

float el_fbm4(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 6; i++) {
        v += a * el_vnoise(p);
        p *= 2.0;
        a *= 0.5;
    }
    return v;
}

vec2 el_px2norm(vec2 v, float isPos) {
    return (v * 2.0 - (x_WindowSize * isPos)) / x_WindowSize.y;
}

vec2 el_cursorCenter(vec4 rect) {
    return vec2(rect.x + rect.z * 0.5, rect.y - rect.w * 0.5);
}

float el_blend01(float t) {
    float s = t * t;
    return s / (2.0 * (s - t) + 1.0);
}

float el_arc(vec2 p, vec2 a, vec2 b, float tm, float intensity) {
    vec2 ab = b - a;
    float len = length(ab);
    if (len < 0.001) return 999.0;
    vec2 dir = ab / len;
    vec2 perp = vec2(-dir.y, dir.x);
    float t = clamp(dot(p - a, dir) / len, 0.0, 1.0);
    vec2 proj = a + dir * t * len;
    float envelope = sin(t * 3.14159);
    float disp = (el_fbm4(vec2(t * 15.0, tm * 10.0)) - 0.5) * 0.18 * intensity * envelope;
    disp += (el_vnoise(vec2(t * 40.0, tm * 25.0)) - 0.5) * 0.04 * intensity * envelope;
    return length(p - (proj + perp * disp));
}

float el_branch(vec2 p, vec2 a, vec2 b, float tm, float seed) {
    vec2 ab = b - a;
    float len = length(ab);
    if (len < 0.001) return 999.0;
    vec2 dir = ab / len;
    vec2 perp = vec2(-dir.y, dir.x);
    float t = clamp(dot(p - a, dir) / len, 0.0, 1.0);
    vec2 proj = a + dir * t * len;
    float disp = (el_fbm4(vec2(t * 20.0 + seed * 7.0, tm * 14.0 + seed * 3.0)) - 0.5)
                 * 0.10 * sin(t * 3.14159);
    disp += (el_vnoise(vec2(t * 50.0 + seed * 13.0, tm * 30.0)) - 0.5) * 0.03 * sin(t * 3.14159);
    return length(p - (proj + perp * disp));
}

float el_spark(vec2 p, vec2 origin, float tm, float seed) {
    float life = fract(seed * 3.17 + tm * 0.8);
    float age = life;
    float angle = seed * 6.2831 + seed * seed * 4.0;
    float speed = 0.15 + 0.25 * fract(seed * 5.31);
    vec2 vel = vec2(cos(angle), sin(angle)) * speed;
    vel.y -= age * 0.1;
    vec2 pos = origin + vel * age;
    float d = length(p - pos);
    float brightness = (1.0 - age) * (1.0 - age);
    float size = 0.003 * (1.0 - age * 0.7);
    return brightness * smoothstep(size, size * 0.2, d);
}

vec3 el_arcColor(float tm) {
    float phase = fract(tm * 0.5);
    if (phase < 0.2)      return mix(COL_CYAN, COL_BLUE, phase * 5.0);
    else if (phase < 0.4) return mix(COL_BLUE, COL_PURPLE, (phase - 0.2) * 5.0);
    else if (phase < 0.6) return mix(COL_PURPLE, COL_PINK, (phase - 0.4) * 5.0);
    else if (phase < 0.8) return mix(COL_PINK, COL_GOLD, (phase - 0.6) * 5.0);
    else                  return mix(COL_GOLD, COL_CYAN, (phase - 0.8) * 5.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;

    // Stub iTimeCursorChange → gentle oscillation, cap heat below shake threshold (0.6).
    float timeSinceType = 0.25 + 0.12 * sin(x_Time * 0.7);
    float heat = smoothstep(HEAT_RAMP_SLOW, HEAT_RAMP_FAST, timeSinceType);

    _wShaderOut = x_Texture(uv);
    vec2 vu = el_px2norm(x_PixelPos, 1.0);

    // Current cursor rect (10x20 px — DarkWindow lacks size info)
    vec4 curRect = vec4(x_CursorPos.x, x_CursorPos.y, 10.0, 20.0);
    // Phantom cursor — slow orbital motion around current position
    vec2 phantomOffset = vec2(cos(x_Time * 1.3) * 80.0, sin(x_Time * 0.9) * 50.0);
    vec4 prevRect = vec4(x_CursorPos.x + phantomOffset.x,
                         x_CursorPos.y + phantomOffset.y, 10.0, 20.0);

    vec4 cc = vec4(el_px2norm(curRect.xy, 1.0),  el_px2norm(curRect.zw, 0.0));
    vec4 cp = vec4(el_px2norm(prevRect.xy, 1.0), el_px2norm(prevRect.zw, 0.0));
    vec2 curPos  = el_cursorCenter(cc);
    vec2 prevPos = el_cursorCenter(cp);
    vec2 tailEnd = prevPos + (prevPos - curPos) * TAIL_EXTENSION;

    float progress = el_blend01(clamp(timeSinceType / ARC_DURATION, 0.0, 1.0));
    float fade = pow(1.0 - progress, 3.0);
    float tailLen = length(tailEnd - curPos);

    float thickness = mix(MIN_ARC_THICKNESS, MAX_ARC_THICKNESS, heat);
    thickness *= 1.0 + 0.5 * sin(x_Time * 40.0) * heat;

    vec3 arcColor  = el_arcColor(x_Time);
    vec3 glowColor = mix(arcColor, COL_WHITE, 0.4);
    vec3 coreColor = mix(arcColor, COL_WHITE, 0.8);

    // Main arc
    float mainDist = el_arc(vu, curPos, tailEnd, x_Time, 0.5 + heat * 0.5);
    float mainAlpha = (1.0 - smoothstep(thickness * 0.2, thickness, mainDist)) * fade;
    float mainCore  = (1.0 - smoothstep(0.0, thickness * 0.3, mainDist)) * fade;
    float mainGlow  = (1.0 - smoothstep(thickness, thickness * GLOW_MULTIPLIER, mainDist)) * fade;

    _wShaderOut.rgb = mix(_wShaderOut.rgb, arcColor, mainAlpha * 0.9);
    _wShaderOut.rgb = mix(_wShaderOut.rgb, coreColor, mainCore * 0.8);
    _wShaderOut.rgb += glowColor * mainGlow * 0.2 * heat;

    // Second arc
    if (heat > 0.2 && tailLen > 0.001) {
        float arc2Dist = el_arc(vu, curPos, tailEnd, x_Time + 100.0, 0.4 + heat * 0.4);
        float arc2Thick = thickness * 0.6;
        float arc2Alpha = (1.0 - smoothstep(arc2Thick * 0.2, arc2Thick, arc2Dist)) * fade * 0.5;
        _wShaderOut.rgb = mix(_wShaderOut.rgb, glowColor, arc2Alpha);
    }

    // Branch arcs
    if (heat > BRANCH_THRESHOLD && tailLen > 0.001) {
        float bi = smoothstep(BRANCH_THRESHOLD, 0.7, heat);
        int nb = int(2.0 + bi * 6.0);
        for (int i = 0; i < 8; i++) {
            if (i >= nb) break;
            float sd = float(i) * 1.37;
            float bt = fract(sd * 0.618 + x_Time * 0.6);
            vec2 bStart = mix(curPos, tailEnd, bt);
            vec2 mdir = tailEnd - curPos;
            float mlen = length(mdir);
            if (mlen > 0.001) {
                mdir = mdir / mlen;
                vec2 bPerp = vec2(-mdir.y, mdir.x);
                float bSide = (fract(sd * 3.14) > 0.5) ? 1.0 : -1.0;
                float bReach = 0.04 + 0.10 * fract(sd * 2.7) * bi;
                vec2 bEnd = bStart + bPerp * bSide * bReach;
                float bDist = el_branch(vu, bStart, bEnd, x_Time, sd);
                float bThick = thickness * 0.45;
                float bAlpha = (1.0 - smoothstep(bThick * 0.2, bThick, bDist)) * fade * bi;
                float bCore  = (1.0 - smoothstep(0.0, bThick * 0.3, bDist)) * fade * bi;
                _wShaderOut.rgb = mix(_wShaderOut.rgb, arcColor, bAlpha * 0.7);
                _wShaderOut.rgb = mix(_wShaderOut.rgb, coreColor, bCore * 0.4);
            }
        }
    }

    // Particle sparks
    if (heat > 0.1) {
        float sparkIntensity = smoothstep(0.1, 0.8, heat);
        int numSparks = int(SPARK_COUNT * sparkIntensity);
        for (int i = 0; i < 12; i++) {
            if (i >= numSparks) break;
            float sd = float(i) * 2.13;
            float spark = el_spark(vu, curPos, x_Time, sd) * sparkIntensity * fade;
            vec3 sparkCol = mix(arcColor, COL_WHITE, 0.7);
            _wShaderOut.rgb += sparkCol * spark * 1.5;
        }
    }

    // Cursor glow
    float cursorDist = length(vu - curPos);
    float cGlowBase = exp(-cursorDist * cursorDist * 60.0);
    float cGlowPulse = 0.15 + 0.35 * heat + 0.1 * sin(x_Time * 15.0) * heat;
    _wShaderOut.rgb += arcColor * cGlowBase * cGlowPulse;
    float cCore = exp(-cursorDist * cursorDist * 300.0) * heat * 0.4;
    _wShaderOut.rgb += COL_WHITE * cCore;

    _wShaderOut.a = 1.0;
}
