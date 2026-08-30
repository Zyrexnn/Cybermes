/**
 * npm/src/commands/config.js
 * CLI command to view and edit ~/.cybermes/config.json
 * 100% Zero external dependencies.
 */

const { ANSI, printHeader, printDivider, badge } = require('../utils/ui');
const { loadConfig, saveConfig, setConfigValue, getConfigValue, resetConfig, getConfigPath } = require('../utils/config-store');

async function runConfig(args = []) {
  const action = (args[0] || 'list').toLowerCase();
  const configPath = getConfigPath();

  if (action === 'list' || action === 'show') {
    printHeader('CYBERMES MCP — CONFIGURATION', `Store: ${configPath}`);
    const cfg = loadConfig();

    if (cfg.error) {
      console.log(`  ${badge('ERROR', 'error')} ${cfg.error}\n`);
      return;
    }

    console.log(`  ${ANSI.bold}${'KEY'.padEnd(20)} VALUE${ANSI.reset}`);
    printDivider(68);

    for (const [k, v] of Object.entries(cfg)) {
      if (k === 'autoApprove' && Array.isArray(v)) {
        console.log(`  ${ANSI.cyan}${k.padEnd(20)}${ANSI.reset} ${ANSI.gray}[${v.length} tools enabled]${ANSI.reset}`);
      } else {
        const valStr = typeof v === 'object' ? JSON.stringify(v) : String(v);
        console.log(`  ${ANSI.cyan}${k.padEnd(20)}${ANSI.reset} ${ANSI.white}${valStr || '(not set)'}${ANSI.reset}`);
      }
    }

    printDivider(68);
    console.log(`  ${ANSI.gray}To change a setting: ${ANSI.cyan}cybermes-mcp config set <key> <val>${ANSI.reset}\n`);
    return;
  }

  if (action === 'get') {
    const key = args[1];
    if (!key) {
      console.log(`  ${badge('USAGE', 'warn')} ${ANSI.cyan}cybermes-mcp config get <key>${ANSI.reset}\n`);
      return;
    }
    const val = getConfigValue(key);
    console.log(typeof val === 'object' ? JSON.stringify(val, null, 2) : String(val));
    return;
  }

  if (action === 'set') {
    const key = args[1];
    const val = args.slice(2).join(' ');
    if (!key || val === undefined || val === '') {
      console.log(`  ${badge('USAGE', 'warn')} ${ANSI.cyan}cybermes-mcp config set <key> <value>${ANSI.reset}\n`);
      return;
    }
    try {
      setConfigValue(key, val);
      console.log(`  ${badge('UPDATED', 'success')} Configuration updated: ${ANSI.cyan}${key}${ANSI.reset} = ${ANSI.white}${val}${ANSI.reset}\n`);
    } catch (err) {
      console.log(`  ${badge('ERROR', 'error')} ${err.message}\n`);
    }
    return;
  }

  if (action === 'reset') {
    resetConfig();
    console.log(`  ${badge('RESET', 'success')} Configuration restored to default values.\n`);
    return;
  }

  console.log(`  ${badge('UNKNOWN ACTION', 'warn')} ${action}. Use ${ANSI.cyan}cybermes-mcp config list${ANSI.reset} or ${ANSI.cyan}cybermes-mcp help config${ANSI.reset}\n`);
}

module.exports = {
  runConfig
};
