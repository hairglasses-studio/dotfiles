#!/usr/bin/env python3
"""
TouchDesigner Python Bridge Server

Bridges HTTP requests to TouchDesigner's Python Server (port 8090)
Allows AFTRS-MCP to communicate with TouchDesigner via REST API
"""

from flask import Flask, request, jsonify
import socket
import json
import threading
import time
import sys
from typing import Optional, Dict, Any

app = Flask(__name__)

class TDConnection:
    """Manages connection to TouchDesigner Python Server"""

    def __init__(self, host='localhost', port=8090):
        self.host = host
        self.port = port
        self.sock: Optional[socket.socket] = None
        self.connected = False
        self.lock = threading.Lock()

    def connect(self) -> bool:
        """Connect to TouchDesigner Python Server"""
        try:
            if self.sock:
                try:
                    self.sock.close()
                except:
                    pass

            self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.sock.settimeout(5.0)
            self.sock.connect((self.host, self.port))
            self.connected = True
            print(f"✓ Connected to TouchDesigner Python Server at {self.host}:{self.port}")
            return True
        except Exception as e:
            self.connected = False
            print(f"✗ Failed to connect to TouchDesigner: {e}")
            return False

    def execute(self, code: str) -> Dict[str, Any]:
        """Execute Python code in TouchDesigner"""
        with self.lock:
            if not self.connected:
                if not self.connect():
                    return {"error": "Not connected to TouchDesigner", "success": False}

            try:
                # Wrap code to capture output
                wrapped_code = f"""
import sys
from io import StringIO
import json

__output = StringIO()
__old_stdout = sys.stdout
sys.stdout = __output

try:
{chr(10).join('    ' + line for line in code.split(chr(10)))}
    __result = {{"success": True, "output": __output.getvalue()}}
except Exception as __e:
    __result = {{"success": False, "error": str(__e), "output": __output.getvalue()}}
finally:
    sys.stdout = __old_stdout

print(json.dumps(__result))
"""

                # Send code
                self.sock.sendall((wrapped_code + '\n').encode('utf-8'))

                # Receive response
                response = b''
                self.sock.settimeout(2.0)
                while True:
                    try:
                        chunk = self.sock.recv(4096)
                        if not chunk:
                            break
                        response += chunk
                        if b'\n' in chunk:
                            break
                    except socket.timeout:
                        break

                if response:
                    result = json.loads(response.decode('utf-8').strip())
                    return result
                else:
                    return {"success": False, "error": "No response from TouchDesigner"}

            except Exception as e:
                self.connected = False
                return {"success": False, "error": str(e)}

    def query(self, code: str) -> Any:
        """Execute code and return the result"""
        result = self.execute(code)
        if result.get("success"):
            output = result.get("output", "").strip()
            if output:
                try:
                    return json.loads(output)
                except:
                    return output
            return None
        else:
            raise Exception(result.get("error", "Unknown error"))

# Global TD connection
td_conn = TDConnection()

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({"status": "ok", "connected": td_conn.connected})

@app.route('/status', methods=['GET'])
def get_status():
    """Get TouchDesigner project status"""
    try:
        code = """
import json
status = {
    'connected': True,
    'project_name': project.name if project else 'Unknown',
    'fps': float(root.time.rate) if root and root.time else 0.0,
    'realtime_fps': 1.0 / root.time.frame if root and root.time and root.time.frame > 0 else 0.0,
    'cook_time_ms': float(root.cookTime) * 1000 if root and hasattr(root, 'cookTime') else 0.0,
    'error_count': len([op for op in root.findChildren(depth=999) if op.errors]) if root else 0,
    'warning_count': len([op for op in root.findChildren(depth=999) if op.warnings]) if root else 0,
    'version': str(app.version) if app else 'Unknown'
}
print(json.dumps(status))
"""
        result = td_conn.query(code)
        return jsonify(result)
    except Exception as e:
        return jsonify({"error": str(e), "connected": False}), 500

@app.route('/operators', methods=['GET'])
def get_operators():
    """List operators in a network"""
    path = request.args.get('path', '/')
    try:
        code = f"""
import json
target = op('{path}')
if target:
    operators = []
    for child in target.children:
        operators.append({{
            'name': child.name,
            'path': child.path,
            'type': str(child.type),
            'family': str(child.family) if hasattr(child, 'family') else 'unknown',
            'has_errors': len(child.errors) > 0 if hasattr(child, 'errors') else False,
            'cook_time_ms': float(child.cookTime) * 1000 if hasattr(child, 'cookTime') else 0.0
        }})
    print(json.dumps(operators))
else:
    print(json.dumps({{'error': 'Operator not found'}}))
"""
        result = td_conn.query(code)
        return jsonify(result)
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/parameters', methods=['GET'])
def get_parameters():
    """Get operator parameters"""
    path = request.args.get('path', '')
    if not path:
        return jsonify({"error": "path parameter required"}), 400

    try:
        code = f"""
import json
target = op('{path}')
if target:
    params = {{}}
    for par in target.pars():
        try:
            params[par.name] = par.eval()
        except:
            params[par.name] = str(par)
    print(json.dumps(params))
else:
    print(json.dumps({{'error': 'Operator not found'}}))
"""
        result = td_conn.query(code)
        return jsonify(result)
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/parameters', methods=['POST'])
def set_parameter():
    """Set operator parameter"""
    data = request.get_json()
    path = data.get('path', '')
    param = data.get('param', '')
    value = data.get('value')

    if not path or not param:
        return jsonify({"error": "path and param required"}), 400

    try:
        code = f"""
import json
target = op('{path}')
if target and hasattr(target.par, '{param}'):
    target.par.{param} = {repr(value)}
    print(json.dumps({{'success': True}}))
else:
    print(json.dumps({{'error': 'Operator or parameter not found'}}))
"""
        result = td_conn.query(code)
        return jsonify(result)
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/execute', methods=['POST'])
def execute_python():
    """Execute Python code in TouchDesigner"""
    data = request.get_json()
    code = data.get('code', '')

    if not code:
        return jsonify({"error": "code required"}), 400

    try:
        result = td_conn.execute(code)
        return jsonify(result)
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/network_health', methods=['GET'])
def get_network_health():
    """Analyze network health"""
    path = request.args.get('path', '/')
    try:
        code = f"""
import json
target = op('{path}')
if target:
    all_ops = target.findChildren(depth=999) if hasattr(target, 'findChildren') else []
    error_ops = [op for op in all_ops if hasattr(op, 'errors') and op.errors]
    warning_ops = [op for op in all_ops if hasattr(op, 'warnings') and op.warnings]
    slow_ops = sorted([op for op in all_ops if hasattr(op, 'cookTime') and op.cookTime > 0.01],
                     key=lambda x: x.cookTime, reverse=True)[:10]

    health = {{
        'score': max(0, 100 - (len(error_ops) * 10) - (len(warning_ops) * 2)),
        'status': 'healthy' if len(error_ops) == 0 else 'degraded',
        'total_operators': len(all_ops),
        'error_count': len(error_ops),
        'warning_count': len(warning_ops),
        'slow_operators': [{{'path': op.path, 'cook_time_ms': float(op.cookTime) * 1000}}
                          for op in slow_ops[:5]]
    }}
    print(json.dumps(health))
else:
    print(json.dumps({{'error': 'Operator not found'}}))
"""
        result = td_conn.query(code)
        return jsonify(result)
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/errors', methods=['GET'])
def get_errors():
    """Get all errors and warnings"""
    try:
        code = """
import json
all_ops = root.findChildren(depth=999) if root else []
errors = []
warnings = []

for op in all_ops:
    if hasattr(op, 'errors') and op.errors:
        for err in op.errors:
            errors.append({'path': op.path, 'error': str(err)})
    if hasattr(op, 'warnings') and op.warnings:
        for warn in op.warnings:
            warnings.append({'path': op.path, 'warning': str(warn)})

print(json.dumps({'errors': errors, 'warnings': warnings}))
"""
        result = td_conn.query(code)
        return jsonify(result)
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/performance', methods=['GET'])
def get_performance():
    """Get performance metrics"""
    try:
        code = """
import json
perf = {
    'fps': float(root.time.rate) if root and root.time else 0.0,
    'cook_time_ms': float(root.cookTime) * 1000 if root and hasattr(root, 'cookTime') else 0.0,
    'frame': int(root.time.frame) if root and root.time else 0,
    'seconds': float(root.time.seconds) if root and root.time else 0.0
}
print(json.dumps(perf))
"""
        result = td_conn.query(code)
        return jsonify(result)
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/project_info', methods=['GET'])
def get_project_info():
    """Get project information"""
    try:
        code = """
import json
info = {
    'name': project.name if project else 'Unknown',
    'folder': project.folder if project and hasattr(project, 'folder') else '',
    'save_file': project.saveFile if project and hasattr(project, 'saveFile') else ''
}
print(json.dumps(info))
"""
        result = td_conn.query(code)
        return jsonify(result)
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    print("=" * 60)
    print("TouchDesigner Python Bridge Server")
    print("=" * 60)
    print(f"Connecting to TouchDesigner Python Server at localhost:8090...")

    # Try to connect on startup
    if td_conn.connect():
        print("✓ Ready to accept requests")
    else:
        print("⚠ TouchDesigner not connected - will retry on first request")
        print("  Make sure TouchDesigner is running with Python Server enabled:")
        print("  Edit → Preferences → Network → Python Server (port 8090)")

    print("\nStarting HTTP server on http://localhost:9980")
    print("Endpoints:")
    print("  GET  /health           - Health check")
    print("  GET  /status           - Project status")
    print("  GET  /operators?path=/ - List operators")
    print("  GET  /parameters?path=/null1 - Get parameters")
    print("  POST /parameters       - Set parameter")
    print("  POST /execute          - Execute Python code")
    print("  GET  /network_health?path=/ - Network health")
    print("  GET  /errors           - All errors/warnings")
    print("  GET  /performance      - Performance metrics")
    print("  GET  /project_info     - Project information")
    print("=" * 60)

    app.run(host='0.0.0.0', port=9980, debug=False)
