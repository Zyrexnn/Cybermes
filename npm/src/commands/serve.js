/**
 * npm/src/commands/serve.js
 * Spawns and supervises native Cybermes MCP server Go binary over stdio.
 * 100% Zero external dependencies.
 */

const { spawn } = require('child_process');
const { getOrDownloadBinary } = require('../utils/binary');
const { loadConfig } = require('../utils/config-store');

async function runServe(extraArgs = []) {
  try {
    const config = loadConfig();
    const preferMode = config.binaryMode || 'auto';
    const binPath = await getOrDownloadBinary(preferMode);

    const args = [...extraArgs];
    if (config.workspace && !args.includes('-workspace')) {
      args.push('-workspace', config.workspace);
    }

    const child = spawn(binPath, args, {
      stdio: 'inherit',
      env: process.env,
    });

    child.on('error', (err) => {
      process.stderr.write(`[cybermes-mcp] Failed to spawn child process: ${err.message}\n`);
      process.exit(1);
    });

    child.on('exit', (code, signal) => {
      if (signal) {
        process.kill(process.pid, signal);
      } else {
        process.exit(code || 0);
      }
    });

    process.on('SIGINT', () => child.kill('SIGINT'));
    process.on('SIGTERM', () => child.kill('SIGTERM'));

  } catch (err) {
    process.stderr.write(`[cybermes-mcp] Error: ${err.message}\n`);
    process.exit(1);
  }
}

module.exports = {
  runServe
};
