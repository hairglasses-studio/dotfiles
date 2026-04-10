const { spawnSync } = require('node:child_process');
const crypto = require('node:crypto');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');

function ensureBuiltBinary(repoRoot, name, pkgPath) {
  const cacheRoot = path.join(os.tmpdir(), 'hg-promptfoo-bins');
  const repoHash = crypto.createHash('sha1').update(repoRoot).digest('hex').slice(0, 12);
  const binDir = path.join(cacheRoot, repoHash);
  const binPath = path.join(binDir, `${name}${process.platform === 'win32' ? '.exe' : ''}`);
  if (fs.existsSync(binPath)) {
    return binPath;
  }

  fs.mkdirSync(binDir, { recursive: true });
  const build = spawnSync('go', ['build', '-o', binPath, pkgPath], {
    cwd: repoRoot,
    env: { ...process.env, GOWORK: 'off' },
    encoding: 'utf8',
    maxBuffer: 10 * 1024 * 1024,
  });
  if (build.status !== 0) {
    throw new Error((build.stderr || build.stdout || `go build failed with exit ${build.status}`).trim());
  }
  return binPath;
}

class HgMcpOllamaEvalProvider {
  constructor(options) {
    this.providerId = options.id || 'hg-mcp-ollama-eval';
    this.config = options.config || {};
    this.repoRoot = path.resolve(__dirname, '..');
    this.binaryPath = ensureBuiltBinary(this.repoRoot, 'hg-mcp-ollama-eval', './cmd/hg-mcp-ollama-eval');
  }

  id() {
    return this.providerId;
  }

  async callApi(_prompt, context) {
    const vars = context?.vars || {};
    const payload = {
      tool_name: vars.tool_name,
      args: vars.args || {},
    };

    const result = spawnSync(this.binaryPath, [], {
      cwd: this.repoRoot,
      env: {
        ...process.env,
        OLLAMA_BASE_URL: process.env.OLLAMA_BASE_URL || 'http://127.0.0.1:11434',
        OLLAMA_API_KEY: process.env.OLLAMA_API_KEY || 'ollama',
      },
      input: JSON.stringify(payload),
      encoding: 'utf8',
      maxBuffer: 10 * 1024 * 1024,
    });

    if (result.status !== 0) {
      throw new Error((result.stderr || result.stdout || `command failed with exit ${result.status}`).trim());
    }

    const parsed = JSON.parse(result.stdout || '{}');
    return {
      output: JSON.stringify(parsed.result || {}, null, 2),
    };
  }
}

module.exports = HgMcpOllamaEvalProvider;