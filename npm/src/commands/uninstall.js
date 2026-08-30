/**
 * npm/src/commands/uninstall.js
 * Cleanly remove Cybermes configuration from AI clients.
 * 100% Zero external dependencies.
 */

const fs = require('fs');
const { ANSI, printHeader, printDivider, badge } = require('../utils/ui');
const { VERSION } = require('../utils/binary');
const { getClientDefinitions } = require('../adapters/clients');
const { removeClientConfig } = require('../adapters/injector');
const { promptCheckbox } = require('../utils/prompt');

async function runUninstaller(options = {}) {
  printHeader('CYBERMES MCP — UNINSTALLER & CLEANUP', `Version: v${VERSION}`);

  if (options.dryRun) {
    console.log(`  ${ANSI.yellow}[ DRY RUN ]${ANSI.reset} ${ANSI.gray}Simulation mode active — No files will be modified.${ANSI.reset}\n`);
  }

  const clients = getClientDefinitions();
  let targetClients = options.clients ? options.clients.split(',').map(s => s.trim().toLowerCase()) : null;

  if (!targetClients && !options.all && process.stdin.isTTY && !options.dryRun) {
    const installedItems = clients
      .map(c => {
        const existingPath = c.paths.find(p => fs.existsSync(p));
        return {
          id: c.id,
          name: c.name,
          checked: Boolean(existingPath),
          path: existingPath
        };
      })
      .filter(i => i.path);

    if (installedItems.length > 0) {
      const selectedIds = await promptCheckbox('Select AI Clients to remove Cybermes configuration from', installedItems);
      if (!selectedIds || selectedIds.length === 0) {
        console.log(`  ${ANSI.yellow}[INFO]${ANSI.reset} No clients selected. Exiting without changes.\n`);
        return;
      }
      targetClients = selectedIds;
    }
  }

  let removedCount = 0;

  for (const client of clients) {
    if (targetClients && !targetClients.includes(client.id) && !targetClients.includes(client.name.toLowerCase())) {
      continue;
    }

    const existingPath = client.paths.find(p => fs.existsSync(p));
    if (!existingPath) continue;

    const res = removeClientConfig(client, existingPath, options.dryRun);
    if (res.status === 'removed' || res.status === 'dry-run') {
      removedCount++;
      console.log(`  ${badge('REMOVED', 'injected')} ${client.name.padEnd(24)} -> ${res.path}`);
    } else if (res.status === 'unchanged') {
      console.log(`  ${badge('NO CHANGE', 'dim')} ${client.name.padEnd(24)} -> Not present`);
    }
  }

  printDivider(74);
  console.log(`  ${ANSI.green}[SUCCESS]${ANSI.reset} Cybermes configuration cleaned from ${removedCount} client(s).\n`);
}

module.exports = {
  runUninstaller
};
