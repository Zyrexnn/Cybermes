/**
 * npm/src/commands/update.js
 * Upgrade cybermes-mcp global package, fetch latest native binary, and prune stale cache.
 * 100% Zero external dependencies.
 */

const { execSync } = require('child_process');
const { ANSI, printBanner, printDivider, badge, padVisible } = require('../utils/ui');
const { VERSION, getOrDownloadBinary, pruneOldBinaries, testBinaryExecution } = require('../utils/binary');
const { fetchLatestNpmVersion, compareSemver } = require('../utils/updater');

function printStatusBox(current, latest, statusLabel, statusColor = ANSI.green) {
  const width = 74;
  const innerWidth = width - 4;
  console.log(`  ${ANSI.cyan}╭${'─'.repeat(innerWidth)}╮${ANSI.reset}`);

  const rows = [
    `  ${ANSI.gray}Installed Version :${ANSI.reset} ${ANSI.bold}v${current}${ANSI.reset}`,
    `  ${ANSI.gray}NPM Registry      :${ANSI.reset} ${ANSI.bold}v${latest}${ANSI.reset}`,
    `  ${ANSI.gray}Sync Status       :${ANSI.reset} ${statusColor}${statusLabel}${ANSI.reset}`,
  ];

  for (const r of rows) {
    console.log(`  ${ANSI.cyan}│${ANSI.reset}${padVisible(r, innerWidth)}${ANSI.cyan}│${ANSI.reset}`);
  }

  console.log(`  ${ANSI.cyan}╰${'─'.repeat(innerWidth)}╯${ANSI.reset}\n`);
}

async function runUpdate(options = {}) {
  printBanner('Universal MCP Updater & Engine Synchronizer');

  process.stdout.write(`  ${badge('CHECKING', 'info')} Querying npm registry for latest release... `);

  let latestVersion;
  try {
    latestVersion = await fetchLatestNpmVersion();
    process.stdout.write(`${ANSI.green}OK${ANSI.reset}\n\n`);
  } catch (err) {
    process.stdout.write(`${ANSI.red}FAILED${ANSI.reset}\n\n`);
    console.log(`  ${badge('ERROR', 'error')} Unable to fetch release info: ${ANSI.red}${err.message}${ANSI.reset}`);
    console.log(`  ${ANSI.gray}Tip: Check your internet connection or try again later.${ANSI.reset}\n`);
    return;
  }

  const hasUpdate = compareSemver(VERSION, latestVersion);

  if (!hasUpdate && !options.force) {
    printStatusBox(VERSION, latestVersion, 'UP TO DATE (Running Latest)', ANSI.green);
    console.log(`  ${ANSI.teal}✨ Everything is up to date!${ANSI.reset} No action required.`);
    console.log(`  ${ANSI.dim}Use --force to re-download native engine binaries and refresh cache.${ANSI.reset}\n`);
    return;
  }

  if (options.force && !hasUpdate) {
    printStatusBox(VERSION, latestVersion, 'FORCED RE-SYNC REQUESTED', ANSI.yellow);
  } else {
    printStatusBox(VERSION, latestVersion, 'UPDATE AVAILABLE', ANSI.yellow);
  }

  if (options.dryRun) {
    console.log(`  ${badge('DRY RUN', 'warn')} ${ANSI.gray}Simulation mode active — No files or packages will be modified.${ANSI.reset}\n`);
    return;
  }

  printDivider(74);

  // 1. Upgrade global npm package
  console.log(`  ${badge('STEP 1/3', 'info')} Updating NPM package (${ANSI.cyan}npm install -g cybermes-mcp@latest${ANSI.reset})...`);
  try {
    execSync('npm install -g cybermes-mcp@latest', { stdio: 'inherit' });
    console.log(`  ${badge('SUCCESS', 'success')} Global NPM package upgraded successfully.\n`);
  } catch (_) {
    console.log(`  ${badge('NOTICE', 'warn')} ${ANSI.gray}Global npm install skipped (not installed globally or missing permissions).${ANSI.reset}\n`);
  }

  // 2. Pre-cache native engine binary
  console.log(`  ${badge('STEP 2/3', 'info')} Synchronizing native binary engine...`);
  try {
    const binPath = await getOrDownloadBinary('auto');
    const check = testBinaryExecution(binPath);
    if (check.ok) {
      console.log(`  ${badge('READY', 'success')} Engine verified: ${ANSI.dim}${binPath}${ANSI.reset}\n`);
    } else {
      console.log(`  ${badge('WARN', 'warn')} Binary present but execution check returned: ${check.error}\n`);
    }
  } catch (err) {
    console.log(`  ${badge('ERROR', 'error')} Failed to retrieve native binary: ${ANSI.red}${err.message}${ANSI.reset}\n`);
  }

  // 3. Prune outdated cache files
  console.log(`  ${badge('STEP 3/3', 'info')} Pruning obsolete cached binaries...`);
  const pruned = pruneOldBinaries(VERSION);
  if (pruned > 0) {
    console.log(`  ${badge('CLEANED', 'success')} Removed ${pruned} older binary build(s).\n`);
  } else {
    console.log(`  ${badge('CLEAN', 'dim')} Binary cache is already optimal.\n`);
  }

  printDivider(74);
  console.log(`\n  ${ANSI.green}${ANSI.bold}✔ Cybermes MCP update completed successfully!${ANSI.reset}`);
  console.log(`  ${ANSI.gray}Run ${ANSI.cyan}cybermes-mcp doctor${ANSI.gray} to run end-to-end diagnostic checks.${ANSI.reset}\n`);
}

module.exports = {
  runUpdate,
  printStatusBox
};
