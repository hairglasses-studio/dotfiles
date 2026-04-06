// Retro CRT - Configurable phosphor color for Terminal Eras
// Switchable green/amber/white phosphor or color (no tint)
//
// Based on [CRTS] PUBLIC DOMAIN CRT-STYLED SCALAR by Timothy Lottes
// source: https://gist.github.com/qwerasd205/c3da6c610c8ffe17d6d2d3cc7068f17f
// credits: https://github.com/qwerasd205
//
//==============================================================
//      LICENSE = UNLICENSE (aka PUBLIC DOMAIN)
//--------------------------------------------------------------
// This is free and unencumbered software released into the
// public domain.
//==============================================================

// --- Phosphor Type (uncomment ONE, or none for color displays) ---
#define PHOSPHOR_GREEN 1
// #define PHOSPHOR_AMBER 1
// #define PHOSPHOR_WHITE 1

// --- Phosphor glow colors ---
#if defined(PHOSPHOR_GREEN)
  #define GLOW_COLOR vec3(0.1, 1.0, 0.3)
#elif defined(PHOSPHOR_AMBER)
  #define GLOW_COLOR vec3(1.0, 0.7, 0.1)
#elif defined(PHOSPHOR_WHITE)
  #define GLOW_COLOR vec3(0.9, 0.9, 1.0)
#endif

// Phosphor glow intensity (only active when a phosphor type is set)
#if defined(PHOSPHOR_GREEN) || defined(PHOSPHOR_AMBER) || defined(PHOSPHOR_WHITE)
  #define PHOSPHOR_GLOW 0.10
#else
  #define PHOSPHOR_GLOW 0.0
#endif

// --- Core CRT Parameters ---

// "Scanlines" per real screen pixel.
//  High DPI: 0.33333333 | Low DPI: 0.66666666
#define SCALE 0.50

// "Tube" curvature - barrel distortion
#define CRTS_WARP 1

// Vignette darkness in corners (0.0=black, 1.0=none)
#define MIN_VIN 0.68

// Shadow mask - simulates RGB phosphor triads
// #define CRTS_MASK_GRILLE 1
// #define CRTS_MASK_GRILLE_LITE 1
// #define CRTS_MASK_NONE 1
#define CRTS_MASK_SHADOW 1

// Scanline thinness (0.50=fused, 0.70=default, 1.00=very thin)
#define INPUT_THIN 0.70

// Horizontal scan blur (-3.0=pixely, -2.5=default, -2.0=smooth)
#define INPUT_BLUR -2.6

// Shadow mask intensity (0.25=heavy, 0.50=default, 1.00=none)
#define INPUT_MASK 0.52

// --- Bloom & Effects ---

// Phosphor bloom intensity (0.0=none, 1.0=heavy)
#define BLOOM_AMOUNT 0.10

// Bloom spread in texels
#define BLOOM_SPREAD 1.5

// Chromatic aberration strength (RGB gun convergence error)
#define CHROMA_SHIFT 0.15

// Screen flicker intensity (60Hz refresh shimmer)
#define FLICKER_AMOUNT 0.012

// Phosphor persistence/afterglow brightness boost
#define PHOSPHOR_BOOST 1.05

// Edge shadow - darkens the very edge of the "tube glass"
#define EDGE_SHADOW 0.93

// --- Enhanced CRT Effects (opt-in, uncomment to enable) ---

// Static noise - film grain / signal noise
// #define ENABLE_NOISE 1
#define NOISE_AMOUNT 0.03

// Horizontal jitter - per-scanline signal instability
// #define ENABLE_JITTER 1
#define JITTER_AMOUNT 0.001

// Interlacing - alternating field darkening
// #define ENABLE_INTERLACE 1
#define INTERLACE_AMOUNT 0.06

// Halation - wide-spread bloom from internal glass reflections
// #define ENABLE_HALATION 1
#define HALATION_AMOUNT 0.08

// --- Core Functions ---

float FromSrgb1(float c) {
  return (c <= 0.04045) ? c * (1.0 / 12.92) :
  pow(c * (1.0 / 1.055) + (0.055 / 1.055), 2.4);
}
vec3 FromSrgb(vec3 c) {
  return vec3(
    FromSrgb1(c.r), FromSrgb1(c.g), FromSrgb1(c.b));
}

vec3 CrtsFetch(vec2 uv) {
  return FromSrgb(texture(iChannel0, uv.xy).rgb);
}

#define CrtsRcpF1(x) (1.0/(x))
#define CrtsSatF1(x) clamp((x),0.0,1.0)

float CrtsMax3F1(float a, float b, float c) {
  return max(a, max(b, c));
}

vec2 CrtsTone(
  float thin,
  float mask) {
  #ifdef CRTS_MASK_NONE
  mask = 1.0;
  #endif

  #ifdef CRTS_MASK_GRILLE_LITE
  mask = 0.5 + mask * 0.5;
  #endif

  vec2 ret;
  float midOut = 0.18 / ((1.5 - thin) * (0.5 * mask + 0.5));
  float pMidIn = 0.18;
  ret.x = ((-pMidIn) + midOut) / ((1.0 - pMidIn) * midOut);
  ret.y = ((-pMidIn) * midOut + pMidIn) / (midOut * (-pMidIn) + midOut);

  return ret;
}

vec3 CrtsMask(vec2 pos, float dark) {
  #ifdef CRTS_MASK_GRILLE
  vec3 m = vec3(dark, dark, dark);
  float x = fract(pos.x * (1.0 / 3.0));
  if (x < (1.0 / 3.0)) m.r = 1.0;
  else if (x < (2.0 / 3.0)) m.g = 1.0;
  else m.b = 1.0;
  return m;
  #endif

  #ifdef CRTS_MASK_GRILLE_LITE
  vec3 m = vec3(1.0, 1.0, 1.0);
  float x = fract(pos.x * (1.0 / 3.0));
  if (x < (1.0 / 3.0)) m.r = dark;
  else if (x < (2.0 / 3.0)) m.g = dark;
  else m.b = dark;
  return m;
  #endif

  #ifdef CRTS_MASK_NONE
  return vec3(1.0, 1.0, 1.0);
  #endif

  #ifdef CRTS_MASK_SHADOW
  pos.x += pos.y * 3.0;
  vec3 m = vec3(dark, dark, dark);
  float x = fract(pos.x * (1.0 / 6.0));
  if (x < (1.0 / 3.0)) m.r = 1.0;
  else if (x < (2.0 / 3.0)) m.g = 1.0;
  else m.b = 1.0;
  return m;
  #endif
}

vec3 CrtsFilter(
  vec2 ipos,
  vec2 inputSizeDivOutputSize,
  vec2 halfInputSize,
  vec2 rcpInputSize,
  vec2 rcpOutputSize,
  vec2 twoDivOutputSize,
  float inputHeight,
  vec2 warp,
  float thin,
  float blur,
  float mask,
  vec2 tone
) {
  vec2 pos;
  #ifdef CRTS_WARP
  pos = ipos * twoDivOutputSize - vec2(1.0, 1.0);
  pos *= vec2(
      1.0 + (pos.y * pos.y) * warp.x,
      1.0 + (pos.x * pos.x) * warp.y);
  float vin = 1.0 - (
      (1.0 - CrtsSatF1(pos.x * pos.x)) * (1.0 - CrtsSatF1(pos.y * pos.y)));
  vin = CrtsSatF1((-vin) * inputHeight + inputHeight);
  pos = pos * halfInputSize + halfInputSize;
  #else
  pos = ipos * inputSizeDivOutputSize;
  #endif

  float y0 = floor(pos.y - 0.5) + 0.5;
  float x0 = floor(pos.x - 1.5) + 0.5;

  vec2 p = vec2(x0 * rcpInputSize.x, y0 * rcpInputSize.y);
  vec3 colA0 = CrtsFetch(p);
  p.x += rcpInputSize.x;
  vec3 colA1 = CrtsFetch(p);
  p.x += rcpInputSize.x;
  vec3 colA2 = CrtsFetch(p);
  p.x += rcpInputSize.x;
  vec3 colA3 = CrtsFetch(p);
  p.y += rcpInputSize.y;
  vec3 colB3 = CrtsFetch(p);
  p.x -= rcpInputSize.x;
  vec3 colB2 = CrtsFetch(p);
  p.x -= rcpInputSize.x;
  vec3 colB1 = CrtsFetch(p);
  p.x -= rcpInputSize.x;
  vec3 colB0 = CrtsFetch(p);

  float off = pos.y - y0;
  float pi2 = 6.28318530717958;
  float hlf = 0.5;
  float scanA = cos(min(0.5, off * thin) * pi2) * hlf + hlf;
  float scanB = cos(min(0.5, (-off) * thin + thin) * pi2) * hlf + hlf;

  float off0 = pos.x - x0;
  float off1 = off0 - 1.0;
  float off2 = off0 - 2.0;
  float off3 = off0 - 3.0;
  float pix0 = exp2(blur * off0 * off0);
  float pix1 = exp2(blur * off1 * off1);
  float pix2 = exp2(blur * off2 * off2);
  float pix3 = exp2(blur * off3 * off3);
  float pixT = CrtsRcpF1(pix0 + pix1 + pix2 + pix3);

  #ifdef CRTS_WARP
  pixT *= max(MIN_VIN, vin);
  #endif

  scanA *= pixT;
  scanB *= pixT;

  vec3 color =
    (colA0 * pix0 + colA1 * pix1 + colA2 * pix2 + colA3 * pix3) * scanA +
      (colB0 * pix0 + colB1 * pix1 + colB2 * pix2 + colB3 * pix3) * scanB;

  color *= CrtsMask(ipos, mask);

  float peak = max(1.0 / (256.0 * 65536.0),
      CrtsMax3F1(color.r, color.g, color.b));
  vec3 ratio = color * CrtsRcpF1(peak);
  peak = peak * CrtsRcpF1(peak * tone.x + tone.y);
  return ratio * peak;
}

float ToSrgb1(float c) {
  return (c < 0.0031308 ? c * 12.92 : 1.055 * pow(c, 0.41666) - 0.055);
}
vec3 ToSrgb(vec3 c) {
  return vec3(
    ToSrgb1(c.r), ToSrgb1(c.g), ToSrgb1(c.b));
}

vec3 sampleBloom(vec2 uv, vec2 texelSize) {
  vec3 bloom = vec3(0.0);
  float spread = BLOOM_SPREAD;
  bloom += FromSrgb(texture(iChannel0, uv + vec2(-spread, -spread) * texelSize).rgb);
  bloom += FromSrgb(texture(iChannel0, uv + vec2( spread, -spread) * texelSize).rgb);
  bloom += FromSrgb(texture(iChannel0, uv + vec2(-spread,  spread) * texelSize).rgb);
  bloom += FromSrgb(texture(iChannel0, uv + vec2( spread,  spread) * texelSize).rgb);
  bloom += FromSrgb(texture(iChannel0, uv + vec2( 0.0,    -spread) * texelSize).rgb);
  bloom += FromSrgb(texture(iChannel0, uv + vec2( 0.0,     spread) * texelSize).rgb);
  bloom += FromSrgb(texture(iChannel0, uv + vec2(-spread,   0.0)   * texelSize).rgb);
  bloom += FromSrgb(texture(iChannel0, uv + vec2( spread,   0.0)   * texelSize).rgb);
  return bloom / 8.0;
}

#if defined(ENABLE_NOISE) || defined(ENABLE_JITTER)
float hash(vec2 p) {
    p = fract(p * vec2(443.897, 441.423));
    p += dot(p, p + 19.19);
    return fract(p.x * p.y);
}
#endif

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
  float aspect = iResolution.x / iResolution.y;
  vec2 texelSize = 1.0 / iResolution.xy;
  vec2 uv = fragCoord.xy * texelSize;

  vec2 warp = vec2(1.0 / (45.0 * aspect), 1.0 / 45.0);

  // --- Horizontal jitter ---
  vec2 jitteredCoord = fragCoord.xy;
  #ifdef ENABLE_JITTER
  {
      float scanline = floor(fragCoord.y);
      float jit = (hash(vec2(scanline, iTime * 60.0)) - 0.5) * 2.0;
      jitteredCoord.x += jit * JITTER_AMOUNT * iResolution.x;
  }
  #endif

  // Core CRT filter
  fragColor.rgb = CrtsFilter(
      jitteredCoord,
      vec2(1.0),
      iResolution.xy * SCALE * 0.5,
      1.0 / (iResolution.xy * SCALE),
      texelSize,
      2.0 * texelSize,
      iResolution.y,
      warp,
      INPUT_THIN,
      INPUT_BLUR,
      INPUT_MASK,
      CrtsTone(INPUT_THIN, INPUT_MASK)
    );

  // --- Phosphor bloom ---
  vec3 bloom = sampleBloom(uv, texelSize);
  fragColor.rgb += bloom * BLOOM_AMOUNT;

  // --- Halation ---
  #ifdef ENABLE_HALATION
  {
      vec3 wideBloom = sampleBloom(uv, texelSize * 4.0);
      fragColor.rgb += wideBloom * HALATION_AMOUNT;
  }
  #endif

  // --- Phosphor glow (monochrome tint) ---
  #if defined(PHOSPHOR_GREEN) || defined(PHOSPHOR_AMBER) || defined(PHOSPHOR_WHITE)
  {
      float luma = dot(bloom, vec3(0.2126, 0.7152, 0.0722));
      fragColor.rgb += GLOW_COLOR * luma * PHOSPHOR_GLOW;
  }
  #endif

  // --- Phosphor persistence ---
  fragColor.rgb *= PHOSPHOR_BOOST;

  // --- Chromatic aberration ---
  vec2 center = uv - 0.5;
  float dist = length(center);
  float shift = dist * CHROMA_SHIFT * texelSize.x * 3.0;
  vec2 dir = normalize(center + 0.0001);
  float rShift = FromSrgb1(texture(iChannel0, uv + dir * shift).r);
  float bShift = FromSrgb1(texture(iChannel0, uv - dir * shift).b);
  float chromaBlend = smoothstep(0.1, 0.6, dist) * 0.4;
  fragColor.r = mix(fragColor.r, rShift * PHOSPHOR_BOOST, chromaBlend);
  fragColor.b = mix(fragColor.b, bShift * PHOSPHOR_BOOST, chromaBlend);

  // --- Interlacing ---
  #ifdef ENABLE_INTERLACE
  {
      float field = mod(floor(iTime * 60.0), 2.0);
      float scanline = mod(floor(fragCoord.y), 2.0);
      float interlace = 1.0 - INTERLACE_AMOUNT * step(0.5, abs(scanline - field));
      fragColor.rgb *= interlace;
  }
  #endif

  // --- 60Hz flicker ---
  float flicker = 1.0 - FLICKER_AMOUNT * sin(iTime * 120.0 * 3.14159);
  fragColor.rgb *= flicker;

  // --- Static noise ---
  #ifdef ENABLE_NOISE
  {
      float noise = (hash(fragCoord.xy + iTime * 1000.0) - 0.5) * 2.0 * NOISE_AMOUNT;
      fragColor.rgb += noise;
  }
  #endif

  // --- Edge shadow ---
  vec2 edgeUV = uv * 2.0 - 1.0;
  float edgeDist = max(abs(edgeUV.x), abs(edgeUV.y));
  float edgeFade = smoothstep(EDGE_SHADOW, 1.0, edgeDist);
  fragColor.rgb *= 1.0 - edgeFade * 0.6;

  // --- Black out beyond curved tube ---
  vec2 curvedUV = edgeUV * vec2(
    1.0 + (edgeUV.y * edgeUV.y) * warp.x,
    1.0 + (edgeUV.x * edgeUV.x) * warp.y);
  if (abs(curvedUV.x) > 1.02 || abs(curvedUV.y) > 1.02) {
    fragColor.rgb = vec3(0.0);
  }

  fragColor = vec4(ToSrgb(fragColor.rgb), 1.0);
}
