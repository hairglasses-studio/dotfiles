// Plane — terminal content on a rotating flat plane in 3D space
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

void windowShader(inout vec4 color) {
    vec2 res = x_WindowSize;
    vec2 center = res * 0.5;
    float diag = length(res);
    float time_cyclic = mod(x_Time * 0.1, 2.0);

    Camera cam = Camera(
        -diag,
        vec3(0.0, -time_cyclic*PI - PI/2.0, 0.0),
        vec3(cos(time_cyclic*PI)*diag, 0.0, sin(time_cyclic*PI)*diag),
        vec3(1.0), vec3(0.0), vec3(0.0), vec3(0.0), vec3(0.0), vec3(0.0)
    );
    cam = setup_camera(cam);

    vec2 coords = x_PixelPos - center;
    vec3 pixPos = cam.center_point + coords.x*cam.base_x + coords.y*cam.base_y;
    vec3 fv = pixPos - cam.focal_point;

    if (abs(fv.z) < 0.0001) { color = vec4(0.0); return; }
    float t = -cam.focal_point.z / fv.z;
    if (t < 1.0) { color = vec4(0.0); return; }

    vec3 hit = fv * t + cam.focal_point;
    vec2 uvCoord = hit.xy + center;

    if (uvCoord.x < 0.0 || uvCoord.y < 0.0 || uvCoord.x >= res.x || uvCoord.y >= res.y) {
        color = vec4(0.0);
        return;
    }
    color = x_Texture(uvCoord / res);
}
