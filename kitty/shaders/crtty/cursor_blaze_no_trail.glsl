#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

float getSdfRectangle(in vec2 p, in vec2 xy, in vec2 b)
{
    vec2 d = abs(p - xy) - b;
    return length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);
}

vec2 norm(vec2 value, float isPosition) {
    return (value * 2.0 - (u_resolution * isPosition)) / u_resolution.y;
}

float antialising(float distance) {
    return 1. - smoothstep(0., norm(vec2(2., 2.), 0.).x, distance);
}

vec2 getRectangleCenter(vec4 rectangle) {
    return vec2(rectangle.x + (rectangle.z / 2.), rectangle.y - (rectangle.w / 2.));
}
float ease(float x) {
    return pow(1.0 - x, 3.0);
}

const vec4 TRAIL_COLOR = vec4(1.0, 0.725, 0.161, 1.0);
const vec4 TRAIL_COLOR_ACCENT = vec4(1.0, 0., 0., 1.0);
const float DURATION = 0.3; //IN SECONDS

void main()
{
    o_color = texture(u_input, gl_FragCoord.xy.xy / u_resolution);
    // Normalization for gl_FragCoord.xy to a space of -1 to 1;
    vec2 vu = norm(gl_FragCoord.xy, 1.);
    vec2 offsetFactor = vec2(-.5, 0.5);

    // Normalization for cursor position and size;
    // cursor xy has the postion in a space of -1 to 1;
    // zw has the width and height
    vec4 currentCursor = vec4(norm(vec2(0.0).xy, 1.), norm(vec2(0.0).zw, 0.));
    vec4 previousCursor = vec4(norm(vec2(0.0).xy, 1.), norm(vec2(0.0).zw, 0.));

    vec2 centerCC = getRectangleCenter(currentCursor);
    vec2 centerCP = getRectangleCenter(previousCursor);

    float sdfCurrentCursor = getSdfRectangle(vu, currentCursor.xy - (currentCursor.zw * offsetFactor), currentCursor.zw * 0.5);

    float progress = clamp((u_time - 0.0) / DURATION, 0.0, 1.0);
    float easedProgress = ease(progress);
    float lineLength = distance(centerCC, centerCP);

    // Compute fade factor based on distance along the trail

    //cursorblaze
    vec4 trail = mix(TRAIL_COLOR_ACCENT, o_color, 1. - smoothstep(0., sdfCurrentCursor + .002, 0.004));
    trail = mix(TRAIL_COLOR, trail, 1. - smoothstep(0., sdfCurrentCursor + .002, 0.004));
    o_color = mix(trail, o_color, 1. - smoothstep(0., sdfCurrentCursor, easedProgress * lineLength));
}
