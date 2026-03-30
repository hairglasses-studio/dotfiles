#!/usr/bin/env python3
"""Parse ralphglasses fleet_status JSON and output lines for SketchyBar."""
import json, sys
from collections import Counter

raw = json.loads(sys.stdin.read())
d = json.loads(raw['content'][0]['text'])
loops = d.get('loops', [])
summary = d.get('summary', {})
repos = d.get('repos', [])

st = Counter(l['status'] for l in loops)
running = st.get('running', 0)
completed = st.get('completed', 0) + st.get('converged', 0)
failed = st.get('failed', 0)
pending = st.get('pending', 0)

# Line 1: Fleet label
parts = []
if running: parts.append(f'\u21bb{running}')
if completed: parts.append(f'\u2713{completed}')
if failed: parts.append(f'\u2717{failed}')
if pending: parts.append(f'{pending}pend')
print(' '.join(parts) if parts else 'idle')

# Line 2: Fleet color
if running > 0: print('0xff5af78e')
elif failed > 0 and completed == 0: print('0xffff5c57')
elif pending > 0: print('0xfff3f99d')
else: print('0xff686868')

# Line 3: Loops detail
total_runs = len(loops)
converge_pct = (completed / total_runs * 100) if total_runs else 0
print(f'{total_runs} runs {converge_pct:.0f}% converge')

# Line 4: Cost
spend = summary.get('total_spend_usd', 0)
print(f'${spend:.2f}')

# Line 5: Models (top planners abbreviated)
planner = Counter(l.get('planner_model', '?') for l in loops)
ab = {'o1-pro':'o1p','o4-mini':'o4m','gpt-4o':'4o','claude-sonnet-4-6':'son',
      'claude-opus-4-6':'opus','sonnet':'son','codex-mini-latest':'cdx','gpt-5.4-xhigh':'5.4x'}
top_p = planner.most_common(3)
print(' '.join(f'{ab.get(m,m[:5])}:{n}' for m,n in top_p))

# Line 6: Repos
unique_repos = len(set(l['repo'] for l in loops))
scanned = len(repos)
print(f'{scanned} scan {unique_repos} tgt')

# Line 7: Iterations
total_iters = sum(l['iterations'] for l in loops)
avg = (total_iters / total_runs) if total_runs else 0
print(f'{total_iters} iters ~{avg:.1f}/run')
