#!/usr/bin/env node

/**
 * Cybermes MCP Server — Zero-Go Modular NPX Launcher & Universal CLI Suite
 * https://github.com/Zyrexnn/Cybermes
 */

const { resolveClientTarget } = require('../src/adapters/clients');
const { findLocalDevBinary } = require('../src/utils/binary');
const { runInstaller } = require('../src/commands/install');
const { runUninstaller } = require('../src/commands/uninstall');
const { runDoctor } = require('../src/commands/doctor');
const { runStatus } = require('../src/commands/status');
const { runTools } = require('../src/commands/tools');
const { runSkills } = require('../src/commands/skills');
const { runConfig } = require('../src/commands/config');
const { runHelp } = require('../src/commands/help');
const { runUpdate } = require('../src/commands/update');
const { runServe } = require('../src/commands/serve');
const { checkUpdate } = require('../src/utils/updater');
const { padVisible } = require('../src/utils/ui');

function parseTargetClients(args) {
  const targeted = new Set();
  const knownGeneralFlags = new Set([
    'install', 'uninstall', 'remove', 'status', 'doctor', 'help', '-help', '--help', '-h',
    'tools', 'skills', 'config', 'update', 'upgrade', '--dry-run', '--force', '--local', '--global', '-g', '--all'
  ]);

  for (const arg of args) {
    if (arg.startsWith('--clients=')) {
      const parts = arg.replace('--clients=', '').split(',');
      for (const p of parts) {
        const id = resolveClientTarget(p);
        if (id) targeted.add(id);
      }
      continue;
    }

    if (knownGeneralFlags.has(arg.toLowerCase())) continue;

    const matchedId = resolveClientTarget(arg);
    if (matchedId) {
      targeted.add(matchedId);
    }
  }

  return targeted.size > 0 ? Array.from(targeted) : null;
}

async function main() {
  const args = process.argv.slice(2);
  const firstArg = (args[0] || '').toLowerCase();
  const targetClients = parseTargetClients(args);
  const useGlobal = (args.includes('--global') || args.includes('-g')) ? true : (args.includes('--npx') ? false : undefined);
  const useNpx = args.includes('--npx');
  const dryRun = args.includes('--dry-run');
  const useLocal = args.includes('--local');
  const localBin = useLocal ? findLocalDevBinary() : null;

  // 1. Help Commands: -help, --help, -h, help, help <cmd>
  if (firstArg === 'help' || firstArg === '-help' || firstArg === '--help' || firstArg === '-h') {
    runHelp(args.slice(1));
    return;
  }

  // 2. Install Command: install, -i, --install
  if (firstArg === 'install' || firstArg === '--install' || firstArg === '-i') {
    const force = args.includes('--force');
    const all = args.includes('--all');
    await runInstaller({
      dryRun,
      force,
      all,
      useLocal,
      useGlobal,
      useNpx,
      localBinPath: localBin,
      clients: targetClients ? targetClients.join(',') : null,
    });
    return;
  }

  // 3. Doctor Command: doctor
  if (firstArg === 'doctor') {
    await runDoctor();
    return;
  }

  // 4. Status Command: status
  if (firstArg === 'status') {
    await runStatus();
    return;
  }

  // 5. Tools Catalog: tools, --tools
  if (firstArg === 'tools' || firstArg === '--tools') {
    await runTools();
    return;
  }

  // 6. Skills SOP Search: skills
  if (firstArg === 'skills' || firstArg === '--skills') {
    await runSkills(args.slice(1));
    return;
  }

  // 7. Config Manager: config
  if (firstArg === 'config') {
    await runConfig(args.slice(1));
    return;
  }

  // 8. Update Command: update, upgrade
  if (firstArg === 'update' || firstArg === 'upgrade') {
    const force = args.includes('--force');
    await runUpdate({ dryRun, force });
    return;
  }

  // 9. Uninstall Command: uninstall, remove
  if (firstArg === 'uninstall' || firstArg === '--uninstall' || firstArg === 'remove') {
    const all = args.includes('--all');
    await runUninstaller({
      dryRun,
      all,
      clients: targetClients ? targetClients.join(',') : null,
    });
    return;
  }

  // 10. Quick target flag directly (e.g. npx cybermes-mcp --kilo)
  if (targetClients && (firstArg.startsWith('--') || firstArg.startsWith('-'))) {
    const force = args.includes('--force');
    await runInstaller({
      dryRun,
      force,
      useLocal,
      useGlobal,
      localBinPath: localBin,
      clients: targetClients.join(','),
    });
    return;
  }

  // Background update check on interactive CLI commands only (never in stdio serve mode)
  const interactiveCmds = new Set(['doctor', 'status', 'tools', 'skills', 'config', 'help', '-help', '--help', '-h']);
  if (interactiveCmds.has(firstArg) && process.stderr.isTTY) {
    checkUpdate().then((info) => {
      if (info && info.updateAvailable) {
        const boxWidth = 56;
        const line1 = `  \x1b[1mUpdate available:\x1b[0m v${info.currentVersion} → \x1b[32mv${info.latestVersion}\x1b[0m`;
        const line2 = `  Run: \x1b[36mcybermes-mcp update\x1b[0m or \x1b[36mnpm i -g cybermes-mcp\x1b[0m`;
        process.stderr.write(`\n  \x1b[33m╭${'─'.repeat(boxWidth)}╮\x1b[0m\n`);
        process.stderr.write(`  \x1b[33m│\x1b[0m${padVisible(line1, boxWidth)}\x1b[33m│\x1b[0m\n`);
        process.stderr.write(`  \x1b[33m│\x1b[0m${padVisible(line2, boxWidth)}\x1b[33m│\x1b[0m\n`);
        process.stderr.write(`  \x1b[33m╰${'─'.repeat(boxWidth)}╯\x1b[0m\n\n`);
      }
    }).catch(() => {});
  }

  // 11. Default: Execute native MCP Server over stdio
  await runServe(args);
}

main();
