/**
 * npm/src/commands/install.js
 * Universal AI client installer with Interactive Checklist Wizard & CLI flag support.
 * Supports selective provider injection in both Global and NPX execution modes.
 * 100% Zero external dependencies.
 */

const fs = require('fs');
const { ANSI, printBanner, printDivider, badge } = require('../utils/ui');
const { VERSION } = require('../utils/binary');
const { getClientDefinitions } = require('../adapters/clients');
const { injectClientConfig } = require('../adapters/injector');
const { promptRadio, promptCheckbox } = require('../utils/prompt');

async function runInstaller(options = {}) {
  printBanner('Universal AI Client Auto-Injector & Setup Wizard');

  if (options.dryRun) {
    console.log(`  ${ANSI.yellow}[ DRY RUN ]${ANSI.reset} ${ANSI.gray}Simulation mode active — No files will be modified on disk.${ANSI.reset}\n`);
  }

  let useGlobal = options.useGlobal;

  // Step 1: Interactive Execution Mode Prompt (if not explicitly passed via flag and in interactive TTY)
  if (useGlobal === undefined && !options.useNpx && process.stdin.isTTY && !options.dryRun && !options.clients && !options.all) {
    const selectedMode = await promptRadio('Choose how AI Clients should launch Cybermes MCP', [
      { id: 'global', name: 'Global Executable (cybermes-mcp)', desc: '— [Faster, Zero-Latency, Offline Ready]' },
      { id: 'npx', name: 'NPX On-Demand (npx -y cybermes-mcp)', desc: '— [Always Latest Release, No npm -g needed]' }
    ], 0);

    useGlobal = (selectedMode === 'global');
  } else if (useGlobal === undefined) {
    useGlobal = false;
  }

  const clients = getClientDefinitions(options.useLocal, options.localBinPath, options.workspaceRoot, useGlobal);
  let targetClients = options.clients ? options.clients.split(',').map(s => s.trim().toLowerCase()) : null;

  // Step 2: Interactive Client Selection (if no specific provider flags were passed and not --all)
  if (!targetClients && !options.all && process.stdin.isTTY && !options.dryRun) {
    const promptItems = clients.map(c => {
      const existingPath = c.paths.find(p => fs.existsSync(p));
      return {
        id: c.id,
        name: c.name,
        checked: Boolean(existingPath),
        installed: Boolean(existingPath),
        path: existingPath || c.paths[0]
      };
    });

    const modeSubtitle = useGlobal 
      ? 'Mode: Global Command (cybermes-mcp)' 
      : 'Mode: NPX Command (npx -y cybermes-mcp)';

    const selectedIds = await promptCheckbox('Select AI Clients to configure for Cybermes MCP', promptItems, modeSubtitle);
    if (!selectedIds || selectedIds.length === 0) {
      console.log(`  ${ANSI.yellow}[INFO]${ANSI.reset} No clients selected. Exiting installer without making changes.\n`);
      return;
    }
    targetClients = selectedIds;
  }

  let injectedCount = 0;
  let readyCount = 0;
  let evaluatedCount = 0;

  const modeLabel = useGlobal ? 'Global Executable (cybermes-mcp)' : 'NPX On-Demand (npx -y cybermes-mcp)';
  console.log(`  ${ANSI.bold}INJECTION RESULTS:${ANSI.reset} ${ANSI.dim}(Mode: ${modeLabel})${ANSI.reset}`);
  printDivider(74);

  for (const client of clients) {
    if (targetClients && !targetClients.includes(client.id) && !targetClients.includes(client.name.toLowerCase())) {
      continue;
    }

    let targetPath = client.paths.find(p => fs.existsSync(p));
    if (!targetPath) {
      if (options.force || options.createAll || options.all || (targetClients && targetClients.includes(client.id))) {
        targetPath = client.paths[0];
      } else {
        continue;
      }
    }

    evaluatedCount++;
    const res = injectClientConfig(client, targetPath, options.dryRun);

    if (res.status === 'injected' || res.status === 'dry-run') {
      injectedCount++;
      console.log(`  ${badge('INJECTED', 'injected')} ${ANSI.white}${client.name.padEnd(24)}${ANSI.reset} ${ANSI.dim}-> ${res.path}${ANSI.reset}`);
    } else if (res.status === 'unchanged') {
      readyCount++;
      console.log(`  ${badge('READY', 'ready')} ${ANSI.white}${client.name.padEnd(24)}${ANSI.reset} ${ANSI.dim}-> ${res.path} (Up-to-date)${ANSI.reset}`);
    } else if (res.status === 'error') {
      console.log(`  ${badge('ERROR', 'error')} ${ANSI.white}${client.name.padEnd(24)}${ANSI.reset} ${ANSI.red}${res.details}${ANSI.reset}`);
    }
  }

  printDivider(74);

  if (injectedCount > 0 || readyCount > 0) {
    console.log(`\n  ${ANSI.green}✔ Installation Complete!${ANSI.reset} Configured: ${ANSI.bold}${injectedCount + readyCount}${ANSI.reset} client(s) [${ANSI.cyan}${useGlobal ? 'Global' : 'NPX'}${ANSI.reset} mode]`);
    console.log(`
  ${ANSI.bold}📌 NEXT STEPS:${ANSI.reset}
    1. Restart your selected AI IDEs/Clients (Cursor, Gemini, Claude Desktop, etc.).
    2. In your AI chat window, ask:
       ${ANSI.cyan}"Search Cybermes knowledge base for JWT vulnerabilities"${ANSI.reset}
    3. Run ${ANSI.cyan}cybermes-mcp doctor${ANSI.reset} anytime to verify system & handshake health.\n`);
  } else {
    console.log(`  ${ANSI.yellow}[INFO]${ANSI.reset} No target client configuration detected. Use --force to generate automatically.\n`);
  }
}

module.exports = {
  runInstaller
};
