precision highp float;
// Sphere — terminal content projected onto a rotating 3D sphere via raycasting
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
    float wss = min(res.x, res.y);
    float time_cyclic = mod(iTime * 0.1, 2.0);

    Camera cam = Camera(
        -wss,
        vec3(PI/6.0*sin(2.0*time_cyclic*PI), -time_cyclic*PI - PI/2.0, 0.0),
        vec3(cos(time_cyclic*PI)*wss, wss/2.0*sin(2.0*time_cyclic*PI), sin(time_cyclic*PI)*wss),
        vec3(1.0), vec3(0.0), vec3(0.0), vec3(0.0), vec3(0.0), vec3(0.0)
    );
    cam = setup_camera(cam);

    vec2 coords = fragCoord - center;
    vec3 pixPos = cam.center_point + coords.x*cam.base_x + coords.y*cam.base_y;
    vec3 fv = pixPos - cam.focal_point;

    // Sphere radius
    float r = wss / (PI * 0.5);

    // Ray-sphere intersection: |fp + fv*t|^2 = r^2
    float a = dot(fv, fv);
    float b = 2.0 * dot(cam.focal_point, fv);
    float c = dot(cam.focal_point, cam.focal_point) - r*r;
    float disc = b*b - 4.0*a*c;

    if (disc < 0.0) {
        fragColor = vec4(0.0);
        return;
    }

    float sqrtDisc = sqrt(disc);
    float t0 = (-b - sqrtDisc) / (2.0*a);
    float t1 = (-b + sqrtDisc) / (2.0*a);

    // Sort and blend both intersections
    if (t0 > t1) { float tmp2 = t0; t0 = t1; t1 = tmp2; }
    float tArr[2] = float[2](t0, t1);

    vec4 result = vec4(0.0);
    for (int i = 0; i < 2; i++) {
        if (tArr[i] < 1.0) continue;
        vec3 hit = fv * tArr[i] + cam.focal_point;
        vec2 uvCoord = hit.xy;

        if (res.x > res.y) uvCoord.x /= res.y / res.x;
        else if (res.x < res.y) uvCoord.y /= res.x / res.y;
        uvCoord += center;

        if (uvCoord.x < 0.0 || uvCoord.y < 0.0 || uvCoord.x >= res.x || uvCoord.y >= res.y) {
            // Dim border color
            vec4 border = texture(iChannel0, vec2(0.5 / res.x, 0.5));
            result = alpha_composite(border * 0.5, result);
            continue;
        }
        vec4 px = texture(iChannel0, uvCoord / res);
        if (px.a > 0.0) result = alpha_composite(px, result);
    }
    fragColor = result;
}
