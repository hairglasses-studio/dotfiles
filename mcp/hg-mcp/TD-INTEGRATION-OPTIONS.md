# TouchDesigner Permanent Integration Options

## Overview

You want to integrate AFTRS-MCP with TouchDesigner permanently across all projects. Here are your options, from simplest to most robust.

---

## Option 1: Component-Based Integration (Recommended)

### How It Works
Create a reusable `.tox` component that includes all MCP integration code. Drop it into every project.

### Setup

**Create: `AFTRS-MCP-Bridge.tox`**

```
AFTRS-MCP-Bridge/
├── webserverDAT (port 9980)
│   └── Handles HTTP requests from MCP
├── executeDAT
│   └── Python code execution endpoint
├── statusDAT
│   └── Gathers project metrics
├── parameterDAT
│   └── Get/set parameter operations
└── callbacksDAT
    └── Handles incoming requests
```

**Python inside the component:**

```python
# AFTRS_MCP_Callbacks.py
import json

def onHTTPRequest(webserverDAT, request, response):
    """Handle HTTP requests from AFTRS MCP"""

    path = request['path']
    method = request['method']

    # Status endpoint
    if path == '/status':
        status = {
            'connected': True,
            'project_name': project.name,
            'fps': root.time.rate,
            'realtime_fps': 1.0 / root.time.frame if root.time.frame > 0 else 0,
            'cook_time_ms': root.cookTime * 1000,
            'error_count': len([op for op in root.findChildren(depth=999) if op.errors]),
            'warning_count': len([op for op in root.findChildren(depth=999) if op.warnings]),
            'version': app.version
        }
        response['statusCode'] = 200
        response['body'] = json.dumps(status)

    # Operators list
    elif path == '/operators':
        network_path = request.get('pars', {}).get('path', '/')
        op_node = op(network_path)
        if op_node:
            operators = []
            for child in op_node.children:
                operators.append({
                    'name': child.name,
                    'path': child.path,
                    'type': child.type,
                    'family': child.family,
                    'has_errors': len(child.errors) > 0,
                    'cook_time_ms': child.cookTime * 1000
                })
            response['statusCode'] = 200
            response['body'] = json.dumps(operators)
        else:
            response['statusCode'] = 404
            response['body'] = json.dumps({'error': 'Operator not found'})

    # Parameters
    elif path == '/parameters':
        op_path = request.get('pars', {}).get('path', '')
        target_op = op(op_path)
        if target_op:
            params = {}
            for par in target_op.pars():
                params[par.name] = par.eval()
            response['statusCode'] = 200
            response['body'] = json.dumps(params)
        else:
            response['statusCode'] = 404
            response['body'] = json.dumps({'error': 'Operator not found'})

    # Execute Python
    elif path == '/execute' and method == 'POST':
        try:
            data = json.loads(request['body'])
            code = data.get('code', '')
            # Execute in safe context
            exec_globals = {'op': op, 'root': root, 'project': project}
            exec(code, exec_globals)
            response['statusCode'] = 200
            response['body'] = json.dumps({'success': True, 'output': 'Executed'})
        except Exception as e:
            response['statusCode'] = 500
            response['body'] = json.dumps({'success': False, 'error': str(e)})

    # Network health
    elif path == '/network_health':
        all_ops = root.findChildren(depth=999)
        error_ops = [op for op in all_ops if op.errors]
        warning_ops = [op for op in all_ops if op.warnings]
        slow_ops = sorted([op for op in all_ops if op.cookTime > 0.01],
                         key=lambda x: x.cookTime, reverse=True)[:10]

        health = {
            'score': max(0, 100 - (len(error_ops) * 10) - (len(warning_ops) * 2)),
            'status': 'healthy' if len(error_ops) == 0 else 'degraded',
            'total_operators': len(all_ops),
            'error_count': len(error_ops),
            'warning_count': len(warning_ops),
            'slow_operators': [{'path': op.path, 'cook_time_ms': op.cookTime * 1000}
                              for op in slow_ops[:5]]
        }
        response['statusCode'] = 200
        response['body'] = json.dumps(health)

    else:
        response['statusCode'] = 404
        response['body'] = json.dumps({'error': 'Endpoint not found'})
```

### Pros
- ✅ **Easy to deploy**: Drag and drop into projects
- ✅ **Consistent**: Same integration everywhere
- ✅ **Updateable**: Update .tox, reload in projects
- ✅ **Portable**: Works across all TD versions
- ✅ **No dependencies**: Pure TouchDesigner

### Cons
- ⚠️ Must manually add to each project
- ⚠️ Need to update component in all projects when changed

### Best For
- Your current workflow
- Multiple independent projects
- Quick setup

---

## Option 2: Project Template with Auto-Load

### How It Works
Create a master project template that includes the MCP bridge and auto-loads on TD startup.

### Setup

**1. Create Master Template**
```
~/Documents/TouchDesigner/Templates/
└── AFTRS-MCP-Template.toe
    ├── AFTRS-MCP-Bridge.tox (from Option 1)
    ├── Auto-start script
    └── Your standard project structure
```

**2. Auto-Start Script**

Add to project's `local/` module or create an Execute DAT that runs on project load:

```python
# Auto_Start_MCP.py
def onProjectLoad():
    """Automatically start MCP WebServer when project loads"""

    # Find or create MCP bridge
    mcp_bridge = op.TDF.op('AFTRS_MCP_Bridge')
    if not mcp_bridge:
        # Create it if missing
        mcp_bridge = root.create(containerCOMP, 'AFTRS_MCP_Bridge')
        # Load the .tox
        mcp_bridge.par.externaltox = '~/Documents/TouchDesigner/Components/AFTRS-MCP-Bridge.tox'
        mcp_bridge.par.reinitnet.pulse()

    # Ensure WebServer is active
    webserver = mcp_bridge.op('webserver')
    if webserver:
        webserver.par.active = True
        print(f"AFTRS MCP Bridge active on port {webserver.par.port}")
```

### Pros
- ✅ **Automatic**: Works as soon as project opens
- ✅ **Template-based**: All new projects have it
- ✅ **Centralized component**: Update once, reload everywhere

### Cons
- ⚠️ Existing projects need manual update
- ⚠️ Need to use template for new projects

### Best For
- Starting fresh with standardized projects
- Team workflows with shared templates

---

## Option 3: External Python Application (TouchEngine-like)

### How It Works
Run a separate Python application that connects to TouchDesigner and exposes an API.

### Architecture

```
┌─────────────────────┐
│   AFTRS-MCP Server  │
│   (Go)              │
└──────────┬──────────┘
           │ HTTP
           ▼
┌─────────────────────┐
│  TD Python Bridge   │
│  (Python + Flask)   │
└──────────┬──────────┘
           │ Python sockets
           ▼
┌─────────────────────┐
│   TouchDesigner     │
│   (any project)     │
└─────────────────────┘
```

**Python Bridge (`td_bridge.py`):**

```python
from flask import Flask, jsonify, request
import socket
import json

app = Flask(__name__)

class TDConnection:
    def __init__(self, host='localhost', port=8090):
        """Connect to TD Python Server"""
        self.host = host
        self.port = port
        self.sock = None

    def connect(self):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect((self.host, self.port))

    def execute(self, code):
        """Execute Python code in TouchDesigner"""
        self.sock.sendall(code.encode() + b'\n')
        response = self.sock.recv(4096)
        return response.decode()

td = TDConnection()

@app.route('/status')
def get_status():
    code = """
import json
status = {
    'project_name': project.name,
    'fps': root.time.rate,
    'cook_time': root.cookTime * 1000
}
print(json.dumps(status))
"""
    result = td.execute(code)
    return jsonify(json.loads(result))

@app.route('/operators/<path:op_path>')
def get_operators(op_path):
    code = f"""
import json
target = op('/{op_path}')
ops = [{{'name': c.name, 'path': c.path, 'type': c.type}} for c in target.children]
print(json.dumps(ops))
"""
    result = td.execute(code)
    return jsonify(json.loads(result))

if __name__ == '__main__':
    td.connect()
    app.run(port=9980)
```

### Pros
- ✅ **No project modification**: Works with any TD project
- ✅ **Uses Python Server**: Just enable in preferences
- ✅ **Centralized**: One bridge for all projects
- ✅ **Easy updates**: Update bridge, affects all projects

### Cons
- ⚠️ Separate process to manage
- ⚠️ Python Server must be enabled
- ⚠️ Additional complexity
- ⚠️ Socket communication overhead

### Best For
- Running multiple TD projects simultaneously
- Don't want to modify projects
- Need centralized control

---

## Option 4: TouchDesigner Plugin/Extension

### How It Works
Create a proper TouchDesigner extension that auto-loads with TD.

### Setup

**1. Create Extension Directory**
```
~/Documents/Derivative/Plugins/
└── AFTRS_MCP/
    ├── AFTRS_MCP.py
    ├── __init__.py
    └── config.json
```

**2. Extension Code (`AFTRS_MCP.py`)**

```python
class AFTRSMCPExtension:
    """
    Auto-loaded TD extension for AFTRS MCP integration
    """
    def __init__(self, ownerComp):
        self.ownerComp = ownerComp
        self.webserver = None
        self.initialize()

    def initialize(self):
        """Setup MCP integration"""
        # Create WebServer if not exists
        if not op.TDF.op('AFTRS_MCP_WebServer'):
            self.webserver = root.create(webserverDAT, 'AFTRS_MCP_WebServer')
            self.webserver.par.port = 9980
            self.webserver.par.active = True
            self.setupCallbacks()

    def setupCallbacks(self):
        """Setup HTTP callbacks"""
        # Link callback scripts
        pass

    def onProjectLoad(self):
        """Called when any project loads"""
        self.initialize()
        print("AFTRS MCP Extension loaded")

    def onProjectSave(self):
        """Called when project saves"""
        pass
```

**3. Register in TD Preferences**

TouchDesigner → Preferences → Extensions → Add Path:
```
~/Documents/Derivative/Plugins/AFTRS_MCP
```

### Pros
- ✅ **Truly automatic**: Loads with TD, works in all projects
- ✅ **No per-project setup**: Just works
- ✅ **Professional**: Real TD extension
- ✅ **Persistent**: Survives project reloads

### Cons
- ⚠️ More complex to develop
- ⚠️ TD extension API learning curve
- ⚠️ Requires extension mechanism understanding

### Best For
- Professional long-term solution
- Multiple users/machines
- Want it "just to work" everywhere

---

## Option 5: Hybrid Approach (Recommended for You)

### Combination Strategy

**For Your Workflow:**

1. **Create Component Library** (Option 1)
   - Build `AFTRS-MCP-Bridge.tox`
   - Store in `~/hairglasses-studio/visual-projects/touchdesigner/components/`

2. **Use Python Bridge** (Option 3) for existing projects
   - No need to modify current 15 projects
   - Works with Python Server (port 8090)
   - Can control any open TD project

3. **Add Component to New Projects** (Option 1)
   - Future projects get component baked in
   - More robust, direct integration

### Implementation

**Phase 1: Quick Win (Python Bridge)**
```bash
# Create bridge server
~/hairglasses-studio/hg-mcp/td-python-bridge/
├── bridge.py          # Flask server
├── requirements.txt   # Python deps
└── start.sh          # Startup script
```

**Phase 2: Component Library**
```bash
# Create reusable component
~/hairglasses-studio/visual-projects/touchdesigner/components/
└── AFTRS-MCP-Bridge.tox
```

**Phase 3: Update AFTRS-MCP Client**
- Point to Python bridge by default
- Add option to use direct HTTP (component mode)
- Support both integration methods

---

## Comparison Matrix

| Feature | Component | Template | Python Bridge | Extension | Hybrid |
|---------|-----------|----------|---------------|-----------|--------|
| **Setup Time** | 🟡 Medium | 🟡 Medium | 🟢 Fast | 🔴 Slow | 🟡 Medium |
| **Works with Existing Projects** | 🔴 No | 🔴 No | 🟢 Yes | 🟢 Yes | 🟢 Yes |
| **Auto-Start** | 🔴 No | 🟢 Yes | 🟢 Yes | 🟢 Yes | 🟢 Yes |
| **Easy Updates** | 🟡 Medium | 🟢 Yes | 🟢 Yes | 🟢 Yes | 🟢 Yes |
| **No Project Modification** | 🔴 No | 🔴 No | 🟢 Yes | 🟢 Yes | 🟡 Partial |
| **Performance** | 🟢 Best | 🟢 Best | 🟡 Good | 🟢 Best | 🟢 Best |
| **Maintenance** | 🟡 Medium | 🟡 Medium | 🟢 Easy | 🔴 Complex | 🟡 Medium |
| **Reliability** | 🟢 High | 🟢 High | 🟡 Medium | 🟢 High | 🟢 High |

---

## My Recommendation: Hybrid Approach

### Why This Works Best for You

1. **Immediate Results**: Python bridge works with all 15 existing projects
2. **Future-Proof**: Component for new projects is more robust
3. **Flexible**: Can use whichever method fits the project
4. **Gradual Migration**: Update projects to component over time

### Implementation Order

**Week 1: Python Bridge (Quick Win)**
- ✅ Create `td-python-bridge.py`
- ✅ Update AFTRS-MCP TouchDesigner client to use it
- ✅ Test with existing projects (no modification needed)
- ✅ Enable Python Server in TD preferences once

**Week 2: Component Library**
- ✅ Build `AFTRS-MCP-Bridge.tox`
- ✅ Test in one project
- ✅ Document usage
- ✅ Add to component library

**Week 3: Integration**
- ✅ Update AFTRS-MCP to support both methods
- ✅ Add config option: `td_integration_mode: "python_bridge" | "http_component"`
- ✅ Test all 25 tools with both methods

**Ongoing:**
- When opening old projects for updates: add component
- New projects: start from template with component
- Both methods work simultaneously

---

## What Would You Like to Do?

I can help you implement any of these options:

1. **Start with Python Bridge** - Get it working today with existing projects
2. **Build the Component** - Create reusable .tox for future projects
3. **Hybrid Approach** - Implement both (recommended)
4. **Something else** - Customize based on your needs

What sounds best for your workflow?
