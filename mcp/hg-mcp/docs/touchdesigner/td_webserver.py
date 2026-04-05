"""
TouchDesigner WebServer DAT Callback Handler
=============================================

This script runs inside TouchDesigner as a WebServer DAT callback handler.
It provides an HTTP API for external tools (like hg-mcp) to control TD.

SETUP:
1. Create a WebServer DAT in your project
2. Set the port to 9980 (or your preferred port)
3. In the "Callbacks DAT" parameter, reference a Text DAT containing this script
4. Enable "Active" on the WebServer DAT

The Go client (internal/clients/touchdesigner.go) expects these endpoints.
"""

import json
import traceback

# Response helper
def json_response(response, data, status=200):
    """Send a JSON response."""
    response['statusCode'] = status
    response['statusReason'] = 'OK' if status == 200 else 'Error'
    response['data'] = json.dumps(data)
    return response


def error_response(response, message, status=500):
    """Send an error response."""
    return json_response(response, {'error': message, 'success': False}, status)


# ============================================================================
# ENDPOINT HANDLERS
# ============================================================================

def handle_health(request, response):
    """GET /health - Basic health check."""
    return json_response(response, {
        'status': 'ok',
        'connected': True,
        'version': app.version,
        'build': app.build
    })


def handle_status(request, response):
    """GET /status - Detailed project status."""
    perf = op('/perform1') if op('/perform1') else None

    status = {
        'connected': True,
        'project_name': project.name,
        'project_folder': project.folder,
        'fps': me.time.rate,
        'realtime_fps': perf.par.fps.eval() if perf else me.time.rate,
        'cook_time_ms': me.time.cookTime * 1000 if hasattr(me.time, 'cookTime') else 0,
        'frame': me.time.frame,
        'playing': me.time.play,
        'timeline_seconds': me.time.seconds,
        'version': app.version,
        'build': app.build,
        'os': app.osName,
        'gpu': app.gpuName if hasattr(app, 'gpuName') else 'unknown'
    }
    return json_response(response, status)


def handle_operators(request, response):
    """GET /operators?path=/project1 - List operators in a network."""
    params = parse_query_params(request)
    path = params.get('path', '/project1')

    container = op(path)
    if not container:
        return error_response(response, f'Container not found: {path}', 404)

    operators = []
    for child in container.children:
        op_info = {
            'name': child.name,
            'path': child.path,
            'type': child.type,
            'family': child.family,
            'active': child.cooking if hasattr(child, 'cooking') else True,
            'viewer': child.viewer if hasattr(child, 'viewer') else False,
            'warnings': len(child.warnings()) if hasattr(child, 'warnings') else 0,
            'errors': len(child.errors()) if hasattr(child, 'errors') else 0
        }
        operators.append(op_info)

    return json_response(response, {
        'path': path,
        'count': len(operators),
        'operators': operators
    })


def handle_get_parameters(request, response):
    """GET /parameters?op=/project1/geo1 - Get operator parameters."""
    params = parse_query_params(request)
    op_path = params.get('op', '')

    if not op_path:
        return error_response(response, 'Missing required parameter: op', 400)

    target = op(op_path)
    if not target:
        return error_response(response, f'Operator not found: {op_path}', 404)

    parameters = []
    for p in target.pars():
        if p.isCustom or p.page.name not in ['About', 'Info', 'Common']:
            par_info = {
                'name': p.name,
                'label': p.label,
                'value': p.eval(),
                'default': p.default,
                'page': p.page.name,
                'mode': str(p.mode),
                'type': str(p.style),
                'readonly': p.readOnly,
                'enabled': p.enable
            }
            parameters.append(par_info)

    return json_response(response, {
        'operator': op_path,
        'parameter_count': len(parameters),
        'parameters': parameters
    })


def handle_set_parameters(request, response):
    """POST /parameters - Set operator parameters.

    Body: {"op": "/project1/geo1", "parameters": {"tx": 1.0, "ty": 2.0}}
    """
    try:
        body = json.loads(request.get('data', '{}'))
    except json.JSONDecodeError:
        return error_response(response, 'Invalid JSON body', 400)

    op_path = body.get('op', '')
    params_to_set = body.get('parameters', {})

    if not op_path:
        return error_response(response, 'Missing required field: op', 400)

    target = op(op_path)
    if not target:
        return error_response(response, f'Operator not found: {op_path}', 404)

    results = []
    for name, value in params_to_set.items():
        try:
            par = getattr(target.par, name, None)
            if par is None:
                results.append({'name': name, 'success': False, 'error': 'Parameter not found'})
            elif par.readOnly:
                results.append({'name': name, 'success': False, 'error': 'Parameter is read-only'})
            else:
                old_value = par.eval()
                par.val = value
                results.append({
                    'name': name,
                    'success': True,
                    'old_value': old_value,
                    'new_value': par.eval()
                })
        except Exception as e:
            results.append({'name': name, 'success': False, 'error': str(e)})

    return json_response(response, {
        'operator': op_path,
        'results': results,
        'success': all(r['success'] for r in results)
    })


def handle_execute(request, response):
    """POST /execute - Execute arbitrary Python code.

    Body: {"code": "op('/project1/geo1').par.tx = 5"}
    """
    try:
        body = json.loads(request.get('data', '{}'))
    except json.JSONDecodeError:
        return error_response(response, 'Invalid JSON body', 400)

    code = body.get('code', '')
    if not code:
        return error_response(response, 'Missing required field: code', 400)

    # Execute in a namespace that includes common TD globals
    namespace = {
        'op': op,
        'ops': ops,
        'me': me,
        'project': project,
        'app': app,
        'tdu': tdu,
        'ui': ui,
        'parent': parent,
        'mod': mod,
        'result': None
    }

    try:
        # Try exec first (for statements)
        exec(code, namespace)
        result = namespace.get('result')
        return json_response(response, {
            'success': True,
            'result': str(result) if result is not None else None
        })
    except SyntaxError:
        # Try eval (for expressions)
        try:
            result = eval(code, namespace)
            return json_response(response, {
                'success': True,
                'result': str(result) if result is not None else None
            })
        except Exception as e:
            return error_response(response, f'Execution error: {str(e)}', 500)
    except Exception as e:
        return error_response(response, f'Execution error: {str(e)}', 500)


def handle_performance(request, response):
    """GET /performance - Get performance metrics."""
    perf = op('/perform1')

    metrics = {
        'fps': me.time.rate,
        'realtime_fps': perf.par.fps.eval() if perf else me.time.rate,
        'cook_time_ms': me.time.cookTime * 1000 if hasattr(me.time, 'cookTime') else 0,
        'frame': me.time.frame,
        'playing': me.time.play,
        'gpu_memory_used_mb': perf.par.gpumemused.eval() if perf and hasattr(perf.par, 'gpumemused') else 0,
        'cpu_cook_time_ms': perf.par.cpucook.eval() * 1000 if perf and hasattr(perf.par, 'cpucook') else 0,
        'gpu_cook_time_ms': perf.par.gpucook.eval() * 1000 if perf and hasattr(perf.par, 'gpucook') else 0,
    }

    # Get cook realtime ratio if available
    if perf and hasattr(perf.par, 'cookrealtime'):
        metrics['cook_realtime_ratio'] = perf.par.cookrealtime.eval()

    return json_response(response, metrics)


def handle_errors(request, response):
    """GET /errors - Get all errors in the project."""
    errors = []

    def collect_errors(container):
        for child in container.children:
            child_errors = child.errors() if hasattr(child, 'errors') else []
            child_warnings = child.warnings() if hasattr(child, 'warnings') else []

            if child_errors or child_warnings:
                errors.append({
                    'operator': child.path,
                    'type': child.type,
                    'errors': list(child_errors),
                    'warnings': list(child_warnings)
                })

            # Recurse into containers
            if hasattr(child, 'children'):
                collect_errors(child)

    collect_errors(root)

    return json_response(response, {
        'total_operators_with_issues': len(errors),
        'errors': errors
    })


def handle_project_info(request, response):
    """GET /project_info - Get project metadata."""
    return json_response(response, {
        'name': project.name,
        'folder': project.folder,
        'save_version': project.saveVersion if hasattr(project, 'saveVersion') else 'unknown',
        'save_build': project.saveBuild if hasattr(project, 'saveBuild') else 'unknown',
        'save_time': str(project.saveTime) if hasattr(project, 'saveTime') else 'unknown',
        'app_version': app.version,
        'app_build': app.build,
        'os': app.osName,
        'product': app.product
    })


def handle_network_health(request, response):
    """GET /network_health?path=/project1 - Get health metrics for a network."""
    params = parse_query_params(request)
    path = params.get('path', '/project1')

    container = op(path)
    if not container:
        return error_response(response, f'Container not found: {path}', 404)

    total = 0
    cooking = 0
    errored = 0
    warned = 0

    for child in container.children:
        total += 1
        if hasattr(child, 'cooking') and child.cooking:
            cooking += 1
        if hasattr(child, 'errors') and child.errors():
            errored += 1
        if hasattr(child, 'warnings') and child.warnings():
            warned += 1

    health_score = 1.0
    if total > 0:
        error_penalty = errored / total * 0.5
        warning_penalty = warned / total * 0.2
        health_score = max(0, 1.0 - error_penalty - warning_penalty)

    return json_response(response, {
        'path': path,
        'total_operators': total,
        'cooking': cooking,
        'errored': errored,
        'warned': warned,
        'health_score': health_score
    })


def handle_pulse(request, response):
    """POST /pulse - Pulse a parameter.

    Body: {"op": "/project1/button1", "par": "click"}
    """
    try:
        body = json.loads(request.get('data', '{}'))
    except json.JSONDecodeError:
        return error_response(response, 'Invalid JSON body', 400)

    op_path = body.get('op', '')
    par_name = body.get('par', '')

    if not op_path or not par_name:
        return error_response(response, 'Missing required fields: op, par', 400)

    target = op(op_path)
    if not target:
        return error_response(response, f'Operator not found: {op_path}', 404)

    par = getattr(target.par, par_name, None)
    if par is None:
        return error_response(response, f'Parameter not found: {par_name}', 404)

    try:
        par.pulse()
        return json_response(response, {'success': True, 'operator': op_path, 'parameter': par_name})
    except Exception as e:
        return error_response(response, f'Pulse failed: {str(e)}', 500)


def handle_cook(request, response):
    """POST /cook - Force cook an operator.

    Body: {"op": "/project1/geo1"}
    """
    try:
        body = json.loads(request.get('data', '{}'))
    except json.JSONDecodeError:
        return error_response(response, 'Invalid JSON body', 400)

    op_path = body.get('op', '')
    if not op_path:
        return error_response(response, 'Missing required field: op', 400)

    target = op(op_path)
    if not target:
        return error_response(response, f'Operator not found: {op_path}', 404)

    try:
        target.cook(force=True)
        return json_response(response, {'success': True, 'operator': op_path})
    except Exception as e:
        return error_response(response, f'Cook failed: {str(e)}', 500)


# ============================================================================
# ROUTING
# ============================================================================

def parse_query_params(request):
    """Parse query parameters from request URI."""
    uri = request.get('uri', '')
    params = {}
    if '?' in uri:
        query_string = uri.split('?', 1)[1]
        for pair in query_string.split('&'):
            if '=' in pair:
                key, value = pair.split('=', 1)
                params[key] = value
    return params


# Route table: (method, path) -> handler
ROUTES = {
    ('GET', '/health'): handle_health,
    ('GET', '/status'): handle_status,
    ('GET', '/operators'): handle_operators,
    ('GET', '/parameters'): handle_get_parameters,
    ('POST', '/parameters'): handle_set_parameters,
    ('POST', '/execute'): handle_execute,
    ('GET', '/performance'): handle_performance,
    ('GET', '/errors'): handle_errors,
    ('GET', '/project_info'): handle_project_info,
    ('GET', '/network_health'): handle_network_health,
    ('POST', '/pulse'): handle_pulse,
    ('POST', '/cook'): handle_cook,
}


def onHTTPRequest(webServerDAT, request, response):
    """Main entry point for HTTP requests."""
    method = request.get('method', 'GET')
    uri = request.get('uri', '/')
    path = uri.split('?')[0]  # Remove query string

    # Set content type
    response['content-type'] = 'application/json'

    # Find handler
    handler = ROUTES.get((method, path))

    if handler:
        try:
            return handler(request, response)
        except Exception as e:
            tb = traceback.format_exc()
            return error_response(response, f'Internal error: {str(e)}\n{tb}', 500)
    else:
        # 404 for unknown routes
        return json_response(response, {
            'error': f'Unknown endpoint: {method} {path}',
            'available_endpoints': list(f'{m} {p}' for m, p in ROUTES.keys())
        }, 404)


# For debugging - prints to textport when script is compiled
print(f"TD WebServer API loaded. {len(ROUTES)} endpoints registered.")