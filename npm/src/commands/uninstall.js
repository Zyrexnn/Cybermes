/**
 * npm/src/commands/uninstall.js
 * Cleanly remove Cybermes configuration from AI clients.
 * 100% Zero external dependencies.
 */

const fs = require('fs');
const { ANSI, printHeader, printDivider, badge } = require('../utils/ui');
const { VERSION } = require('../utils/binary');
const { getClientDefinitions } = require('../adapters/clients');
const { removeClientConfig, checkClientStatus } = require('../adapters/injector');
const { promptCheckbox } = require('../utils/prompt');

async function runUninstaller(options = {}) {
  printHeader('CYBERMES MCP — UNINSTALLER & CLEANUP', `Version: v${VERSION}`);

  if (options.dryRun) {
    console.log(`  ${ANSI.yellow}[ DRY RUN ]${ANSI.reset} ${ANSI.gray}Simulation mode active — No files will be modified.${ANSI.reset}\n`);
  }

  const clients = getClientDefinitions();
  let targetClients = options.clients ? options.clients.split(',').map(s => s.trim().toLowerCase()) : null;

  const configuredItems = clients
    .map(c => {
      const existingPath = c.paths.find(p => fs.existsSync(p));
      if (!existingPath) return null;
      const status = checkClientStatus(c, existingPath);
      if (!status.configured) return null;
      return {
        id: c.id,
        name: c.name,
        checked: true,
        installed: true,
        path: existingPath
      };
    })
    .filter(Boolean);

  if (!targetClients && !options.all && process.stdin.isTTY && !options.dryRun) {
    if (configuredItems.length === 0) {
      console.log(`  ${ANSI.green}✔ System is already clean!${ANSI.reset} ${ANSI.gray}No AI clients currently have Cybermes MCP configured.${ANSI.reset}\n`);
      return;
    }

    const selectedIds = await promptCheckbox('Select AI Clients to remove Cybermes configuration from', configuredItems);
    if (!selectedIds || selectedIds.length === 0) {
      console.log(`  ${ANSI.yellow}[INFO]${ANSI.reset} No clients selected. Exiting uninstaller without changes.\n`);
      return;
    }
    targetClients = selectedIds;
  }

  let removedCount = 0;
  let evaluatedCount = 0;

  for (const client of clients) {
    if (targetClients && !targetClients.includes(client.id) && !targetClients.includes(client.name.toLowerCase())) {
      continue;
    }

    const existingPath = client.paths.find(p => fs.existsSync(p));
    if (!existingPath) continue;

    evaluatedCount++;
    const res = removeClientConfig(client, existingPath, options.dryRun);
    if (res.status === 'removed' || res.status === 'dry-run') {
      removedCount++;
      console.log(`  ${badge('REMOVED', 'injected')} ${ANSI.white}${client.name.padEnd(24)}${ANSI.reset} ${ANSI.dim}-> ${res.path}${ANSI.reset}`);
    } else if (res.status === 'unchanged') {
      console.log(`  ${badge('NOT PRESENT', 'dim')} ${ANSI.white}${client.name.padEnd(24)}${ANSI.reset} ${ANSI.dim}-> ${res.path} (Not present)${ANSI.reset}`);
    } else if (res.status === 'error') {
      console.log(`  ${badge('ERROR', 'error')} ${ANSI.white}${client.name.padEnd(24)}${ANSI.reset} ${ANSI.red}${res.details}${ANSI.reset}`);
    }
  }

  printDivider(74);

  if (removedCount > 0) {
    console.log(`\n  ${ANSI.green}✔ Cleanup Complete!${ANSI.reset} Successfully removed Cybermes configuration from ${ANSI.bold}${removedCount}${ANSI.reset} client(s).\n`);
  } else if (evaluatedCount > 0) {
    console.log(`\n  ${ANSI.yellow}[INFO]${ANSI.reset} No active Cybermes MCP configuration was found on evaluated client(s).\n`);
  } else {
    console.log(`\n  ${ANSI.green}✔ System is already clean!${ANSI.reset} No target client configuration found.\n`);
  }
}

module.exports = {
  runUninstaller
};
