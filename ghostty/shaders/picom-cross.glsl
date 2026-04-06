precision highp float;
// Cross — terminal content projected onto a rotating 3D cross shape (two perpendicular planes)
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
    float ar = c1.w + c2.w - c1.w*c2.w;
    if (ar < 0.001) return vec4(0.0);
    float asr = c2.w / ar;
    return vec4(c1.xyz*(1.0-asr) + c2.xyz*asr, ar);
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 res = iResolution.xy;
    vec2 center = res * 0.5;
    float diag = length(res);
    float time_cyclic = mod(iTime * 0.1, 2.0);

    Camera cam = Camera(
        -diag,
        vec3(0.0, -time_cyclic*PI - PI/2.0, 0.0),
        vec3(cos(time_cyclic*PI)*diag, 0.0, sin(time_cyclic*PI)*diag),
        vec3(1.0), vec3(0.0), vec3(0.0), vec3(0.0), vec3(0.0), vec3(0.0)
    );
    cam = setup_camera(cam);

    vec2 coords = fragCoord - center;
    vec3 pixPos = cam.center_point + coords.x*cam.base_x + coords.y*cam.base_y;
    vec3 fv = pixPos - cam.focal_point;

    // Two perpendicular planes: z=0 and x=0
    // Plane 0: z=0 → t = -fp.z / fv.z
    // Plane 1: x=0 → t = -fp.x / fv.x
    float t0 = (abs(fv.z) < 0.0001) ? -1.0 : -cam.focal_point.z / fv.z;
    float t1 = (abs(fv.x) < 0.0001) ? -1.0 : -cam.focal_point.x / fv.x;
    int face0 = 0, face1 = 1;

    // Sort by t
    if (t0 > t1) {
        float tt = t0; t0 = t1; t1 = tt;
        face0 = 1; face1 = 0;
    }

    vec4 result = vec4(0.0);
    float tArr[2] = float[2](t0, t1);
    int faces[2] = int[2](face0, face1);

    for (int i = 0; i < 2; i++) {
        if (tArr[i] < 1.0) continue;
        vec3 hit = fv * tArr[i] + cam.focal_point;
        vec2 uvCoord;
        if (faces[i] == 0) uvCoord = hit.xy + center;
        else uvCoord = hit.zy + center;

        if (uvCoord.x < 0.0 || uvCoord.y < 0.0 || uvCoord.x >= res.x || uvCoord.y >= res.y) continue;
        vec4 px = texture(iChannel0, uvCoord / res);
        result = alpha_composite(px, result);
    }
    fragColor = result;
}
