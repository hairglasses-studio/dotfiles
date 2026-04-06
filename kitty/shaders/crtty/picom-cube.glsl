#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Cube — terminal content projected onto a rotating 3D cube via pinhole camera
// Category: Post-FX | Cost: HIGH | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

#define PI 3.14159265

struct Camera {
    float focal_offset;
    vec3 rotations, translations, deformations;
    vec3 base_x, base_y, base_z, center_point, focal_point;
};

Camera setup_camera(Camera cam) {
    cam.center_point += cam.translations;
    cam.base_x = vec3(cam.deformations.x, 0.0, 0.0);
    cam.base_y = vec3(0.0, cam.deformations.y, 0.0);
    cam.base_z = vec3(0.0, 0.0, cam.deformations.z);

    float cx = cos(cam.rotations.x), sx = sin(cam.rotations.x);
    float cy = cos(cam.rotations.y), sy = sin(cam.rotations.y);
    float cz = cos(cam.rotations.z), sz = sin(cam.rotations.z);

    vec3 tmp;
    tmp = cam.base_x; tmp.y = cam.base_x.y*cx-cam.base_x.z*sx; tmp.z = cam.base_x.y*sx+cam.base_x.z*cx; cam.base_x = tmp;
    tmp.x = cam.base_x.x*cy+cam.base_x.z*sy; tmp.z = -cam.base_x.x*sy+cam.base_x.z*cy; cam.base_x = tmp;
    tmp.x = cam.base_x.x*cz-cam.base_x.y*sz; tmp.y = cam.base_x.x*sz+cam.base_x.y*cz; cam.base_x = tmp;
    tmp = cam.base_y; tmp.y = cam.base_y.y*cx-cam.base_y.z*sx; tmp.z = cam.base_y.y*sx+cam.base_y.z*cx; cam.base_y = tmp;
    tmp.x = cam.base_y.x*cy+cam.base_y.z*sy; tmp.z = -cam.base_y.x*sy+cam.base_y.z*cy; cam.base_y = tmp;
    tmp.x = cam.base_y.x*cz-cam.base_y.y*sz; tmp.y = cam.base_y.x*sz+cam.base_y.y*cz; cam.base_y = tmp;
    tmp = cam.base_z; tmp.y = cam.base_z.y*cx-cam.base_z.z*sx; tmp.z = cam.base_z.y*sx+cam.base_z.z*cx; cam.base_z = tmp;
    tmp.x = cam.base_z.x*cy+cam.base_z.z*sy; tmp.z = -cam.base_z.x*sy+cam.base_z.z*cy; cam.base_z = tmp;
    tmp.x = cam.base_z.x*cz-cam.base_z.y*sz; tmp.y = cam.base_z.x*sz+cam.base_z.y*cz; cam.base_z = tmp;

    cam.focal_point = cam.center_point + cam.base_z * cam.focal_offset;
    return cam;
}

vec4 alpha_composite(vec4 c1, vec4 c2) {
    float ar = c1.w + c2.w - c1.w * c2.w;
    if (ar < 0.001) return vec4(0.0);
    float asr = c2.w / ar;
    vec4 o;
    o.xyz = c1.xyz * (1.0 - asr) + c2.xyz * asr;
    o.w = ar;
    return o;
}

void main() {
    vec2 res = u_resolution;
    vec2 center = res * 0.5;
    float wss = min(res.x, res.y);

    float time_cyclic = mod(u_time * 0.1, 2.0);

    Camera cam = Camera(
        -wss,
        vec3(PI/6.0*sin(2.0*time_cyclic*PI), -time_cyclic*PI - PI/2.0, 0.0),
        vec3(cos(time_cyclic*PI)*wss, wss/2.0*sin(2.0*time_cyclic*PI), sin(time_cyclic*PI)*wss),
        vec3(1.0), vec3(0.0), vec3(0.0), vec3(0.0), vec3(0.0), vec3(0.0)
    );
    cam = setup_camera(cam);

    vec2 coords = gl_FragCoord.xy - center;
    vec3 pixPos = cam.center_point + coords.x * cam.base_x + coords.y * cam.base_y;
    vec3 fv = pixPos - cam.focal_point;

    float half = wss * 0.5;
    float planeA[6] = float[6](0.0, 0.0, 1.0, 1.0, 0.0, 0.0);
    float planeB[6] = float[6](0.0, 0.0, 0.0, 0.0, 1.0, 1.0);
    float planeC[6] = float[6](1.0, 1.0, 0.0, 0.0, 0.0, 0.0);
    float planeD[6] = float[6](-half, half, -half, half, -half, half);

    float tVals[6];
    int faceIds[6];
    for (int i = 0; i < 6; i++) {
        float denom = planeA[i]*fv.x + planeB[i]*fv.y + planeC[i]*fv.z;
        tVals[i] = (abs(denom) < 0.0001) ? -1.0 :
            (planeD[i] - planeA[i]*cam.focal_point.x - planeB[i]*cam.focal_point.y - planeC[i]*cam.focal_point.z) / denom;
        faceIds[i] = i;
    }

    for (int i = 0; i < 5; i++) {
        for (int j = 0; j < 5; j++) {
            if (tVals[j] > tVals[j+1]) {
                float tt = tVals[j]; tVals[j] = tVals[j+1]; tVals[j+1] = tt;
                int fi = faceIds[j]; faceIds[j] = faceIds[j+1]; faceIds[j+1] = fi;
            }
        }
    }

    vec4 result = vec4(0.0);
    for (int i = 0; i < 6; i++) {
        if (tVals[i] < 1.0) continue;
        vec3 hit = fv * tVals[i] + cam.focal_point;
        vec2 uvCoord;
        int face = faceIds[i];
        if (face < 2) uvCoord = hit.xy;
        else if (face < 4) uvCoord = hit.zy;
        else uvCoord = hit.zx;

        if (res.x > res.y) uvCoord.x /= res.y / res.x;
        else if (res.x < res.y) uvCoord.y /= res.x / res.y;
        uvCoord += center;

        if (uvCoord.x < 1.0 || uvCoord.y < 1.0 || uvCoord.x >= res.x-1.0 || uvCoord.y >= res.y-1.0) continue;

        vec4 px = texture(u_input, uvCoord / res);
        if (px.a > 0.0) result = alpha_composite(px, result);
    }
    o_color = result;
}
